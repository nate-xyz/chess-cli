package pkg

import (
	"testing"

	tc "github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/notnil/chess"
)

func TestBoardGeometryFitsAndGrows(t *testing.T) {
	prev := 0
	for _, sz := range [][2]int{{30, 10}, {60, 20}, {80, 26}, {120, 40}, {200, 60}, {300, 90}} {
		w, h := sz[0], sz[1]
		sqW, sqH := boardGeometry(w, h, false)
		if sqW == 0 {
			t.Fatalf("%dx%d: no board", w, h)
		}
		if 8*sqW+labelW > w || 8*sqH+labelH > h {
			t.Errorf("%dx%d: board %dx%d squares overflows", w, h, sqW, sqH)
		}
		if sqH < prev {
			t.Errorf("%dx%d: squares shrank (%d < %d)", w, h, sqH, prev)
		}
		prev = sqH
	}
	if w, _ := boardGeometry(10, 5, false); w != 0 {
		t.Errorf("tiny area should report no board")
	}
}

func TestSquareMappingRoundTrip(t *testing.T) {
	for _, white := range []bool{true, false} {
		for sq := chess.A1; sq <= chess.H8; sq++ {
			r, c := displayOf(sq, white)
			if got := squareAt(r, c, white); got != sq {
				t.Fatalf("white=%v %v -> (%d,%d) -> %v", white, sq, r, c, got)
			}
		}
	}
	if squareAt(7, 0, true) != chess.A1 || squareAt(7, 0, false) != chess.H8 {
		t.Errorf("bottom-left square wrong for orientation")
	}
}

func TestHandleSquareSelectAndMove(t *testing.T) {
	b := NewBoardView(nil)
	g := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	var played string
	do := func(uci string) { played = uci }

	b.HandleSquare(g, chess.E7, true, do) // opponent's piece: ignored
	if b.Selected() != chess.NoSquare {
		t.Fatalf("selected a black piece on white's turn")
	}
	b.HandleSquare(g, chess.E2, true, do)
	if b.Selected() != chess.E2 || !b.targets[chess.E4] || b.targets[chess.E5] {
		t.Fatalf("bad selection/targets: %v %v", b.Selected(), b.targets)
	}
	b.HandleSquare(g, chess.E4, true, do)
	if played != "e2e4" || b.Selected() != chess.NoSquare {
		t.Fatalf("played %q, selected %v", played, b.Selected())
	}
	b.HandleSquare(g, chess.G1, true, do)
	b.HandleSquare(g, chess.G1, true, do) // click again: deselect
	if b.Selected() != chess.NoSquare {
		t.Fatalf("second click should deselect")
	}
}

func TestPromotionPicker(t *testing.T) {
	fen, _ := chess.FEN("8/4P3/8/8/8/8/k7/4K3 w - - 0 1")
	for _, tc := range []struct {
		name  string
		pick  func(b *BoardView, g *chess.Game, do func(string))
		want  string
		white bool
	}{
		{"queen by click", func(b *BoardView, g *chess.Game, do func(string)) { b.HandleSquare(g, chess.E8, true, do) }, "e7e8q", true},
		{"knight by click", func(b *BoardView, g *chess.Game, do func(string)) { b.HandleSquare(g, chess.E7, true, do) }, "e7e8n", true},
		{"rook by click", func(b *BoardView, g *chess.Game, do func(string)) { b.HandleSquare(g, chess.E6, true, do) }, "e7e8r", true},
		{"bishop by click", func(b *BoardView, g *chess.Game, do func(string)) { b.HandleSquare(g, chess.E5, true, do) }, "e7e8b", true},
		{"knight by key", func(b *BoardView, g *chess.Game, do func(string)) {
			sq, ok := b.promoSquareFor("N")
			if !ok {
				t.Fatal("no picker square for n")
			}
			b.HandleSquare(g, sq, true, do)
		}, "e7e8n", true},
		// board flipped (black at bottom): e8 is on the bottom row, picker goes up
		{"flipped rook by click", func(b *BoardView, g *chess.Game, do func(string)) { b.HandleSquare(g, chess.E6, true, do) }, "e7e8r", false},
		{"cancel by clicking away", func(b *BoardView, g *chess.Game, do func(string)) { b.HandleSquare(g, chess.A1, true, do) }, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := chess.NewGame(fen)
			b := NewBoardView(nil)
			b.SetPosition(g.Position(), tc.white)
			played := ""
			do := func(u string) { played = u }
			b.HandleSquare(g, chess.E7, true, do)
			if b.HandleSquare(g, chess.E8, true, do) || played != "" {
				t.Fatalf("promotion should wait for a choice, played %q", played)
			}
			if !b.PromotionOpen() {
				t.Fatal("picker not open")
			}
			tc.pick(b, g, do)
			if played != tc.want {
				t.Fatalf("played %q, want %q", played, tc.want)
			}
			if b.PromotionOpen() {
				t.Fatal("picker still open")
			}
		})
	}
}

// newTestScreen is a real terminfo screen attached to tcell's mock terminal,
// finalised when the test ends.
func newTestScreen(t *testing.T, w, h int) tc.Screen {
	t.Helper()
	s := newMockScreen(t, w, h)
	t.Cleanup(s.Fini)
	return s
}

func newMockScreen(t *testing.T, w, h int) tc.Screen {
	t.Helper()
	mt := vt.NewMockTerm(vt.MockOptSize{X: vt.Col(w), Y: vt.Row(h)}, vt.MockOptColors(1<<24))
	s, err := tc.NewTerminfoScreenFromTty(mt)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestParseMoveInput(t *testing.T) {
	promo, _ := chess.FEN("r3k2r/1P6/8/8/8/8/8/R3K2R w KQkq - 0 1")
	start := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for _, tc := range []struct {
		game *chess.Game
		in   string
		want string
	}{
		{start, "e2e4", "e2e4"},
		{start, "E2E4", "e2e4"},
		{start, "e2-e4", "e2e4"},
		{start, "Nf3", "g1f3"},
		{start, "e4", "e2e4"},
		{start, "Ng1f3", "g1f3"},
		{chess.NewGame(promo), "b7b8=N", "b7b8n"},
		{chess.NewGame(promo), "b7b8n", "b7b8n"},
		{chess.NewGame(promo), "b7b8N", "b7b8n"},
		{chess.NewGame(promo), "b8=R", "b7b8r"},
		{chess.NewGame(promo), "bxa8=Q", "b7a8q"},
		{chess.NewGame(promo), "O-O", "e1g1"},
		{chess.NewGame(promo), "0-0-0", "e1c1"},
		{chess.NewGame(promo), "b7b8", ""}, // promotion needs a piece
		{start, "e2e5", ""},
		{start, "Nf6", ""},
		{start, "", ""},
	} {
		got, err := parseMoveInput(tc.game, tc.in)
		if tc.want == "" {
			if err == nil {
				t.Errorf("%q: expected an error, got %q", tc.in, got)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("%q: got %q, %v; want %q", tc.in, got, err, tc.want)
		}
	}
}

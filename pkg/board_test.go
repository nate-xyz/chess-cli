package pkg

import (
	"fmt"
	"os"
	"strings"
	"testing"

	tc "github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/notnil/chess"
)

func TestBoardGeometryFitsAndGrows(t *testing.T) {
	prev := 0
	for _, sz := range [][2]int{{30, 10}, {60, 20}, {80, 26}, {120, 40}, {200, 60}, {300, 90}} {
		w, h := sz[0], sz[1]
		sqW, sqH := boardGeometry(w, h)
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
	if w, _ := boardGeometry(10, 5); w != 0 {
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

// TestRenderDump writes ASCII-ish renders of the board at several sizes to
// $BOARD_DUMP (if set), for eyeballing the art without running the TUI.
func TestRenderDump(t *testing.T) {
	out := os.Getenv("BOARD_DUMP")
	if out == "" {
		t.Skip("set BOARD_DUMP=path to write renders")
	}
	var sb strings.Builder
	g := chess.NewGame()
	for _, m := range []string{"e4", "e5", "Nf3", "Nc6", "Bc4"} {
		_ = g.MoveStr(m)
	}
	for _, sz := range [][2]int{{40, 12}, {80, 30}, {150, 45}, {190, 75}} {
		s := newTestScreen(t, sz[0], sz[1])
		b := NewBoardView(nil)
		b.SetRect(0, 0, sz[0], sz[1])
		b.SetPosition(g.Position(), true)
		b.SetLastMove(chess.F1, chess.C4)
		b.HandleSquare(g, chess.G8, true, func(string) {})
		b.Draw(s)
		w, h := s.Size()
		sqW, sqH := boardGeometry(w, h)
		fmt.Fprintf(&sb, "=== size %dx%d, square %dx%d\n", w, h, sqW, sqH)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				sb.WriteRune(shadeAt(s, x, y))
			}
			sb.WriteByte('\n')
		}
	}
	if err := os.WriteFile(out, []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// newTestScreen is a real terminfo screen attached to tcell's mock terminal.
func newTestScreen(t *testing.T, w, h int) tc.Screen {
	t.Helper()
	mt := vt.NewMockTerm(vt.MockOptSize{X: vt.Col(w), Y: vt.Row(h)}, vt.MockOptColors(1<<24))
	s, err := tc.NewTerminfoScreenFromTty(mt)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Fini)
	return s
}

// shadeAt renders one screen cell as a plain-text character so renders can be
// eyeballed: piece pixels as W/B, detail lines as o, target dots as *, squares
// as . (light) and : (dark), tinted squares as ~.
func shadeAt(s tc.Screen, x, y int) rune {
	str, style, _ := s.Get(x, y)
	r := ' '
	if str != "" {
		r = []rune(str)[0]
	}
	return shade(r, style.GetForeground(), style.GetBackground())
}

func shade(r rune, fg, bg tc.Color) rune {
	cls := func(c tc.Color) rune {
		switch c {
		case colWhiteBody:
			return 'W'
		case colBlackBody:
			return 'B'
		case colWhiteLine, colBlackLine:
			return 'o'
		case colDot:
			return '*'
		case colLight:
			return '.'
		case colDark:
			return ':'
		}
		return '~' // tinted square
	}
	switch r {
	case '▀':
		top, bot := cls(fg), cls(bg)
		if top == 'W' || top == 'B' || top == 'o' || top == '*' {
			return top
		}
		return bot
	case ' ':
		return cls(bg)
	}
	return r
}

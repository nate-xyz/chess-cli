package pkg

import (
	"testing"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/notnil/chess"
)

func TestBoardGeometryFitsAndGrows(t *testing.T) {
	prev := 0
	for _, sz := range [][2]int{{30, 10}, {60, 20}, {80, 26}, {120, 40}, {200, 60}, {300, 90}} {
		w, h := sz[0], sz[1]
		sqW, sqH := boardGeometry(w, h, styleGlyph, 2)
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
	if w, _ := boardGeometry(10, 5, styleGlyph, 2); w != 0 {
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

// The keyboard cursor stays hidden until a key is used (the first key only
// shows it), a click hides it, and a study board never shows it.
func TestBoardCursorShowsOnKeys(t *testing.T) {
	key := func(b *BoardView, k tc.Key) {
		b.InputHandler()(tc.NewEventKey(k, "", tc.ModNone), func(cv.Primitive) {})
	}
	var picked []chess.Square
	b := NewBoardView(func(sq chess.Square) { picked = append(picked, sq) })
	if b.keyboard {
		t.Fatal("cursor shown before any key")
	}
	key(b, tc.KeyUp)
	if !b.keyboard || b.cursor != chess.E2 {
		t.Fatalf("first key: shown %v on %v, want shown on e2", b.keyboard, b.cursor)
	}
	key(b, tc.KeyUp)
	if b.cursor != chess.E3 {
		t.Errorf("second key moved cursor to %v, want e3", b.cursor)
	}
	key(b, tc.KeyEnter)
	if len(picked) != 1 || picked[0] != chess.E3 {
		t.Errorf("Enter picked %v, want [e3]", picked)
	}

	study := NewBoardView(func(chess.Square) {})
	study.browse = true
	key(study, tc.KeyUp)
	key(study, tc.KeyEnter)
	if study.keyboard {
		t.Error("study board shows a cursor")
	}
}

// The game layout adapts: sidebar beside a good-sized board, the board on
// top in a narrow tall terminal, the board alone in a small one, and in a
// wide one the sidebar stops growing and the two are centred.
func TestGameLayoutAdapts(t *testing.T) {
	sq := func(w, h int) int {
		_, s := boardGeometry(w, h, styleGlyph, 2)
		return s
	}
	for _, c := range []struct {
		w, h int
		want layoutMode
	}{
		{120, 36, layoutSide},
		{250, 60, layoutSide},
		{100, 30, layoutSide},
		{80, 24, layoutBoard}, // beside a sidebar the squares would be one row high
		{70, 50, layoutStacked},
		{56, 22, layoutBoard},
	} {
		if got := layoutFor(c.w, c.h, sq); got != c.want {
			t.Errorf("%dx%d: layout %d, want %d", c.w, c.h, got, c.want)
		}
	}

	startTestApp(t, 160, 50)
	onUI(func() {
		Root.Switch("study")
		Root.study.show(parseStudyPGN(testStudyPGN), "")
	})
	// draw lays the screen out at w x h and runs check on the result, all in
	// one go on the UI goroutine: onUI redraws the app at its own size
	// afterwards, which would move everything again before a later look.
	draw := func(board *BoardView, w, h int, check func(l *boardLayout)) {
		onUI(func() {
			_, p := Root.nav.GetFrontPanel()
			scr := newMockScreen(t, w, h)
			p.SetRect(0, 0, w, h)
			p.Draw(scr)
			check(layoutOf[board])
		})
	}
	defer onUI(func() { Root.nav.SetRect(0, 0, 160, 50) })

	draw(Root.study.Board, 250, 60, func(l *boardLayout) {
		_, _, sw, _ := l.side.GetRect()
		lx, _, lw, _ := l.left.GetRect()
		_, _, rw, _ := l.right.GetRect()
		if sw != maxSidebar || lw == 0 || lx != 0 || lw-rw > 1 || rw-lw > 1 {
			t.Errorf("wide: sidebar %d (want %d), margins %d and %d (want equal, non-zero)", sw, maxSidebar, lw, rw)
		}
	})
	draw(Root.study.Board, 70, 50, func(l *boardLayout) {
		_, by, _, bh := l.board.GetRect()
		_, sy, _, _ := l.side.GetRect()
		if by != 0 || sy != by+bh {
			t.Errorf("narrow: board at row %d (%d high), sidebar at row %d; want board on top, sidebar below", by, bh, sy)
		}
	})
	draw(Root.study.Board, 56, 22, func(l *boardLayout) {
		if _, _, bw, _ := l.board.GetRect(); bw != 56 {
			t.Errorf("small: board %d wide, want the whole width", bw)
		}
	})
	// the menu focused: the board stays, the menu shows under it
	onUI(func() { Root.App.SetFocus(Root.study.List) })
	draw(Root.study.Board, 56, 22, func(l *boardLayout) {
		_, by, bw, bh := l.board.GetRect()
		_, my, mw, mh := Root.study.List.GetRect()
		if bw != 56 || bh < 14 || my != by+bh || mw != 56 || mh != Root.study.List.GetItemCount() {
			t.Errorf("small with the menu focused: board %dx%d at row %d, menu %dx%d at row %d; want the board, the menu under it",
				bw, bh, by, mw, mh, my)
		}
	})
	// typing a move in a small local game: the board stays in view
	onUI(func() {
		startNewLocalGame()
		Root.lgame.Move.show()
	})
	draw(Root.lgame.Board, 56, 22, func(l *boardLayout) {
		_, by, bw, bh := l.board.GetRect()
		_, iy, _, ih := Root.lgame.Move.input.GetRect()
		if bw != 56 || bh < 14 || iy != by+bh || ih != 1 {
			t.Errorf("typing a move: board %dx%d, move box at row %d (%d high); want the board with the box under it", bw, bh, iy, ih)
		}
	})
	onUI(func() {
		Root.lgame.Move.hide()
		gotoWelcome()
		Root.Switch("study")
	})
	onUI(func() { Root.App.SetFocus(Root.study.Board) })
}

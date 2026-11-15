package pkg

import (
	"testing"

	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

func boardAfter(t *testing.T, moves ...string) *BoardView {
	t.Helper()
	g := chess.NewGame()
	for _, m := range moves {
		if err := g.MoveStr(m); err != nil {
			t.Fatal(err)
		}
	}
	b := NewBoardView(nil)
	b.SetPosition(g.Position(), true)
	b.SetHistory(g.Positions(), g.Moves())
	return b
}

func TestCheckSquare(t *testing.T) {
	b := boardAfter(t, "e4", "f5", "Qh5+")
	b.mu.Lock()
	if sq := b.checkSquare(); sq != chess.E8 {
		t.Errorf("check: glow on %v, want e8", sq)
	}
	b.step(1) // back to before the check
	if sq := b.checkSquare(); sq != chess.NoSquare {
		t.Errorf("stepped back: glow on %v, want none", sq)
	}
	b.mu.Unlock()

	mate := boardAfter(t, "f3", "e5", "g4", "Qh4#")
	mate.mu.Lock()
	if sq := mate.checkSquare(); sq != chess.E1 {
		t.Errorf("checkmate: glow on %v, want e1", sq)
	}
	mate.mu.Unlock()

	quiet := boardAfter(t, "e4", "e5")
	quiet.mu.Lock()
	if sq := quiet.checkSquare(); sq != chess.NoSquare {
		t.Errorf("no check, but glow on %v", sq)
	}
	quiet.mu.Unlock()
}

func TestCheckGlowImage(t *testing.T) {
	img := checkGlowImage(64)
	centre, corner := img.RGBAAt(32, 32), img.RGBAAt(0, 0)
	if centre.R < 200 || centre.A < 200 || centre.G != 0 {
		t.Errorf("centre %v: want bright opaque red", centre)
	}
	if corner.A != 0 {
		t.Errorf("corner %v: want transparent", corner)
	}
	if mid := img.RGBAAt(32, 4); mid.A == 0 || mid.A >= centre.A {
		t.Errorf("edge %v: want fading", mid)
	}
}

// In text and basic rendering the checked king's square is red.
func TestCheckGlowDrawn(t *testing.T) {
	for _, mode := range []string{PiecesText, PiecesBasic} {
		previewPieces(mode)
		scr := newMockScreen(t, 120, 40)
		b := boardAfter(t, "e4", "f5", "Qh5+")
		b.SetRect(0, 0, 120, 40)
		b.Draw(scr)
		b.mu.Lock()
		row, col := displayOf(chess.E8, b.white)
		x, y := b.ox+col*b.sqW, b.oy+row*b.sqH
		b.mu.Unlock()
		_, st, _ := scr.Get(x, y)
		want := blend(colLight, colCheck, 0.7) // e8 is a light square
		if mode == PiecesBasic {
			want = tc.ColorMaroon
		}
		if st.GetBackground() != want {
			t.Errorf("%s: e8 background %v, want %v", mode, st.GetBackground(), want)
		}
		scr.Fini()
	}
	previewPieces("")
}

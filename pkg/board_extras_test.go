package pkg

import (
	"testing"

	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

func TestHistoryStepping(t *testing.T) {
	game, _ := gameFromSAN("", "e4 e5 Nf3")
	b := NewBoardView(nil)
	b.SetPosition(game.Position(), true)
	b.SetHistory(game.Positions(), game.Moves())

	b.mu.Lock()
	b.step(1) // back one: after 1. e4 e5
	pos, from, to := b.shown()
	b.mu.Unlock()
	if pos.Board().Piece(chess.F3) != chess.NoPiece || from != chess.E7 || to != chess.E5 {
		t.Errorf("one back: %v %v-%v", pos, from, to)
	}
	b.mu.Lock()
	b.step(100) // clamps at the start
	pos, from, _ = b.shown()
	b.mu.Unlock()
	if pos.String() != chess.StartingPosition().String() || from != chess.NoSquare {
		t.Errorf("start: %v", pos)
	}
	if !b.ViewingHistory() {
		t.Error("should be viewing history")
	}
	// a new move brings the board back to the present
	game.MoveStr("Nc6")
	b.SetHistory(game.Positions(), game.Moves())
	if b.ViewingHistory() {
		t.Error("a new move should return to the present")
	}
}

func TestShapes(t *testing.T) {
	b := NewBoardView(nil)
	arrow := boardShape{from: chess.E2, to: chess.E4, color: 0}
	b.toggleShape(arrow)
	b.toggleShape(boardShape{from: chess.D4, to: chess.D4, color: 1})
	if len(b.shapes) != 2 {
		t.Fatalf("shapes: %v", b.shapes)
	}
	// same arrow in another colour replaces it; same colour removes it
	b.toggleShape(boardShape{from: chess.E2, to: chess.E4, color: 2})
	if len(b.shapes) != 2 || b.shapes[1].color != 2 {
		t.Errorf("recolour: %v", b.shapes)
	}
	b.toggleShape(boardShape{from: chess.E2, to: chess.E4, color: 2})
	if len(b.shapes) != 1 {
		t.Errorf("remove: %v", b.shapes)
	}
	if shapeColorFor(tc.ModShift|tc.ModAlt) != 3 || shapeColorFor(tc.ModShift) != 1 || shapeColorFor(tc.ModAlt) != 2 || shapeColorFor(0) != 0 {
		t.Error("brush colours")
	}
	// the arrow image has paint along the arrow and none in the far corner
	img := shapesImage([]boardShape{arrow}, true, 20, 20)
	if img.RGBAAt(4*20+10, 5*20+10).A == 0 { // e3, mid-shaft
		t.Error("arrow shaft not drawn")
	}
	if img.RGBAAt(5, 5).A != 0 {
		t.Error("paint where there's no shape")
	}
	// a new position clears the shapes
	b.SetPosition(chess.NewGame().Position(), false)
	if len(b.shapes) != 0 {
		t.Error("shapes kept after the position changed")
	}
}

func TestResultBadges(t *testing.T) {
	b := NewBoardView(nil)
	b.SetResult(chess.Black, "mate")
	if b.badgeFor(chess.Black) != badgeWin || b.badgeFor(chess.White) != badgeMate {
		t.Error("mate badges")
	}
	b.SetResult(chess.White, "resign")
	if b.badgeFor(chess.Black) != badgeLoss {
		t.Error("resignation badge")
	}
	b.SetResult(chess.NoColor, "stalemate")
	if b.badgeFor(chess.White) != badgeDraw || b.badgeFor(chess.Black) != badgeDraw {
		t.Error("draw badges")
	}
	for k := badgeWin; k <= badgeDraw; k++ {
		img := badgeImage(k, 24)
		if img.RGBAAt(12, 12).A == 0 {
			t.Errorf("badge %d empty in the middle", k)
		}
	}
	if resultText(chess.White, "timeout") != "[::b]1–0[::-]\nWhite wins by time" {
		t.Error(resultText(chess.White, "timeout"))
	}
}

func TestDrawBoardStyles(t *testing.T) {
	// every style draws without panicking on a small and a large screen
	for _, size := range [][2]int{{40, 12}, {200, 60}} {
		s := newMockScreen(t, size[0], size[1])
		b := NewBoardView(nil)
		b.SetRect(0, 0, size[0], size[1])
		game, _ := gameFromSAN("", "f3 e5 g4 Qh4")
		b.SetPosition(game.Position(), true)
		b.SetResult(chess.Black, "mate")
		b.SetClockBars(clockBar{label: "1:00", frac: 0.5, right: "3+2", active: true}, clockBar{label: "2:00", frac: 1})
		b.toggleShape(boardShape{from: chess.E2, to: chess.E4})
		b.Draw(s)
		b.mu.Lock()
		b.step(1)
		b.mu.Unlock()
		b.Draw(s)
	}
}

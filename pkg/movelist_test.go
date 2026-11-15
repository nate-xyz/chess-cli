package pkg

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

func TestMoveListRegions(t *testing.T) {
	game := chess.NewGame()
	for _, m := range []string{"e4", "e5", "Nf3"} {
		if err := game.MoveStr(m); err != nil {
			t.Fatal(err)
		}
	}
	text := moveListText(game)
	for _, want := range []string{`["m1"]e4[""]`, `["m2"]e5[""]`, `["m3"]Nf3[""]`} {
		if !strings.Contains(text, want) {
			t.Errorf("move list missing %s:\n%s", want, text)
		}
	}
}

// Clicking a move shows the position after it; stepping highlights the move.
func TestMoveListClick(t *testing.T) {
	startTestApp(t, 160, 50)
	g := Root.lgame
	onUI(func() {
		startNewLocalGame()
		for _, m := range []string{"e2e4", "e7e5", "g1f3"} {
			if err := g.DoMove(m); err != nil {
				t.Fatal(err)
			}
		}
		g.History.Highlight("m1") // what a click on "e4" does
		if !g.Board.ViewingHistory() {
			t.Fatal("clicking a move didn't go back")
		}
		g.Board.mu.Lock()
		pos, _, to := g.Board.shown()
		g.Board.mu.Unlock()
		if pos.String() != Root.gameState.Game.Positions()[1].String() || to != chess.E4 {
			t.Errorf("shown %v (last move to %v), want the position after 1. e4", pos, to)
		}
		g.Board.ShowPly(2) // e.g. ] on the board
	})
	waitFor(t, "move list follows the board", func() bool {
		hl := g.History.GetHighlights() // waitFor runs this on the UI goroutine
		return len(hl) == 1 && hl[0] == "m2"
	})
	onUI(func() {
		g.History.Highlight("m3") // the last move: back to the present
		if g.Board.ViewingHistory() {
			t.Error("clicking the last move should show the current position")
		}
	})
}

// The game-over screen has the game layout: players, result, moves, final
// clocks, with the menu focused.
func TestPostLocalLayout(t *testing.T) {
	startTestApp(t, 160, 50)
	g := Root.lgame
	onUI(func() {
		startNewLocalGame()
		for _, m := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} { // fool's mate
			if err := g.DoMove(m); err != nil {
				t.Fatal(err)
			}
		}
	})
	waitFor(t, "game over screen", func() bool {
		name, _ := Root.nav.GetFrontPanel()
		return name == "postlocal"
	})
	onUI(func() {
		p := Root.pgame
		if r := p.Result.GetText(true); !strings.Contains(r, "Black wins by checkmate") {
			t.Errorf("result = %q", r)
		}
		if h := p.History.GetText(true); !strings.Contains(h, "Qh4") {
			t.Errorf("moves = %q", h)
		}
		if p.Top.GetText(true) == "" || p.Bottom.GetText(true) == "" {
			t.Error("players not shown")
		}
		if Root.App.GetFocus() != p.List {
			t.Error("menu not focused")
		}
		p.History.Highlight("m2")
		if !p.Board.ViewingHistory() {
			t.Error("can't step back on the game-over board")
		}
	})
}

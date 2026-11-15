package pkg

import (
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func TestLocalClock(t *testing.T) {
	startTestApp(t, 160, 50)
	g := Root.lgame
	onUI(func() {
		api.UserInfo.LocalClock = 3 // 3+2
		startNewLocalGame()
		if !g.clock.on || g.clock.white != 3*time.Minute {
			t.Fatalf("clock not set up: %+v", g.clock)
		}
		// White's clock doesn't run before White's first move
		if !g.clock.since.IsZero() {
			t.Error("clock running before the first move")
		}
		if err := g.DoMove("e2e4"); err != nil {
			t.Fatal(err)
		}
		if g.clock.white != 3*time.Minute+2*time.Second {
			t.Errorf("white after first move = %v, want 3:02 (no time used, +2 increment)", g.clock.white)
		}
		// black's clock is running now; make it run out
		g.clock.black = time.Millisecond
	})
	time.Sleep(20 * time.Millisecond)
	onUI(func() { g.tickClock() })
	waitFor(t, "loss on time", func() bool {
		name, _ := Root.nav.GetFrontPanel()
		return name == "postlocal" && strings.Contains(Root.pgame.Result.GetText(true), "White wins by time")
	})
	onUI(func() { api.UserInfo.LocalClock = 0 })
}

// FIDE classical: each side gets 30 minutes more after their 40th move, once.
func TestLocalClockFIDEBonus(t *testing.T) {
	startTestApp(t, 160, 50)
	g := Root.lgame
	onUI(func() {
		defer func() { api.UserInfo.LocalClock = 0 }()
		api.UserInfo.LocalClock = len(localClocks) - 1
		if name := localClockName(api.UserInfo.LocalClock); name != "90/40+30+30s" {
			t.Fatalf("last clock is %q, want the FIDE control", name)
		}
		startNewLocalGame()
		game := Root.gameState.Game
		rng := rand.New(rand.NewPCG(1, 2))
		for ply := 1; ply <= 82; ply++ {
			// a random game that keeps going (no mate, no forced draw)
			var played bool
			for try := 0; try < 50 && !played; try++ {
				valid := game.ValidMoves()
				g2 := game.Clone()
				m := valid[rng.IntN(len(valid))]
				if g2.Move(m) == nil && g2.Outcome() == chess.NoOutcome {
					played = game.Move(m) == nil
				}
			}
			if !played {
				t.Fatalf("couldn't find a move at ply %d", ply)
			}
			mover := chess.White
			if ply%2 == 0 {
				mover = chess.Black
			}
			wBefore, bBefore := g.clock.white, g.clock.black
			g.clock.since = time.Time{} // no thinking time, so only increments and bonuses count
			g.clockMoved(mover)
			gained := g.clock.white - wBefore
			if mover == chess.Black {
				gained = g.clock.black - bBefore
			}
			want := 30 * time.Second
			if ply == 79 || ply == 80 { // White's and Black's 40th moves
				want += 30 * time.Minute
			}
			if gained != want {
				t.Errorf("ply %d (%v): gained %v, want %v", ply, mover, gained, want)
			}
		}
	})
}

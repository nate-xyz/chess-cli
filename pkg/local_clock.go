package pkg

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// Clocks for local games (two players, or against Stockfish). White's clock
// starts after White's first move; each move adds the increment; running out
// loses on time.

// localClocks are the time controls on offer: minutes + seconds per move,
// and for the FIDE classical control, extra minutes after move 40.
var localClocks = []struct{ minutes, increment, bonusAt40 int }{
	{0, 0, 0}, // none
	{1, 0, 0}, {3, 0, 0}, {3, 2, 0}, {5, 0, 0}, {5, 3, 0}, {10, 0, 0}, {10, 5, 0}, {15, 10, 0},
	{30, 0, 0}, {30, 20, 0}, {45, 15, 0}, {60, 30, 0}, {90, 30, 0}, {120, 30, 0},
	{90, 30, 30}, // FIDE classical: 90 min for 40 moves, then +30 min; 30 s per move throughout
}

// localClockIndex is the chosen time control (0 = no clock).
func localClockIndex() int {
	i := api.UserInfo.LocalClock
	if i < 0 || i >= len(localClocks) {
		return 0
	}
	return i
}

func localClockName(i int) string {
	c := localClocks[i]
	switch {
	case c.minutes == 0:
		return "none"
	case c.bonusAt40 > 0:
		return fmt.Sprintf("%d/40+%d+%ds", c.minutes, c.bonusAt40, c.increment)
	}
	return fmt.Sprintf("%d+%d", c.minutes, c.increment)
}

// nextLocalClock cycles the time control.
func nextLocalClock() {
	next := (localClockIndex() + 1) % len(localClocks)
	_ = api.UpdateConfig(func(c *api.UserConfig) { c.LocalClock = next })
}

// clockActive mirrors GameScreen.clock.on for the background ticker.
var clockActive atomic.Bool

// localClock is a running pair of clocks.
type localClock struct {
	on        bool
	white     time.Duration
	black     time.Duration
	initial   time.Duration
	increment time.Duration
	since     time.Time // when the side to move's clock last started (zero = not running)
	tc        string
	bonusAt40 time.Duration // added to each side's clock after their 40th move
	lowWhite  bool          // the low-time sound has played for this side
	lowBlack  bool
}

// startClock sets up the clocks for a new game with the chosen control.
func (g *GameScreen) startClock() {
	c := localClocks[localClockIndex()]
	g.clock = localClock{}
	clockActive.Store(c.minutes > 0)
	if c.minutes == 0 {
		return
	}
	d := time.Duration(c.minutes) * time.Minute
	g.clock = localClock{on: true, white: d, black: d, initial: d,
		increment: time.Duration(c.increment) * time.Second, tc: localClockName(localClockIndex()),
		bonusAt40: time.Duration(c.bonusAt40) * time.Minute}
}

// clockMoved updates the clocks after a move by mover.
func (g *GameScreen) clockMoved(mover chess.Color) {
	c := &g.clock
	if !c.on {
		return
	}
	now := time.Now()
	if !c.since.IsZero() {
		used := now.Sub(c.since)
		if mover == chess.White {
			c.white -= used
		} else {
			c.black -= used
		}
	}
	bonus := time.Duration(0)
	// the mover's 40th move has just been played: second time period starts
	if c.bonusAt40 > 0 && len(Root.gameState.Game.Moves()) == 80-int(chess.Black-mover) {
		bonus = c.bonusAt40
	}
	if mover == chess.White {
		c.white += c.increment + bonus
	} else {
		c.black += c.increment + bonus
	}
	c.since = now // the other side's clock starts (after White's first move)
}

// remaining is each side's time now, counting the running clock down.
func (g *GameScreen) remaining() (w, b time.Duration) {
	c := g.clock
	w, b = c.white, c.black
	if !c.since.IsZero() {
		if Root.gameState.Game.Position().Turn() == chess.White {
			w -= time.Since(c.since)
		} else {
			b -= time.Since(c.since)
		}
	}
	return max(0, w), max(0, b)
}

// tickClock shows the clocks and checks for a flag fall. UI goroutine.
func (g *GameScreen) tickClock() {
	if !g.clock.on || Root.gameState == nil {
		g.Board.HideClockBars()
		g.UserTimer.SetText("")
		g.OppTimer.SetText("")
		return
	}
	w, b := g.remaining()
	running := !g.clock.since.IsZero()
	whiteToMove := Root.gameState.Game.Position().Turn() == chess.White
	if running && (w == 0 || b == 0) {
		winner := chess.Black
		if w > 0 {
			winner = chess.White
		}
		gotoPostLocalResult(winner, "timeout") // stops the clock
		return
	}
	if running && g.clock.white > 0 {
		low, left := &g.clock.lowWhite, w
		if !whiteToMove {
			low, left = &g.clock.lowBlack, b
		}
		if !*low && lowTime(left, g.clock.initial) && !g.engineTurn() {
			*low = true
			playSound(soundLowTime)
		}
	}
	bar := func(d time.Duration, active bool) clockBar {
		return clockBar{label: clockLabel(int(d.Round(time.Second) / time.Second)), right: g.clock.tc,
			frac: min(1, float64(d)/float64(g.clock.initial)), active: active && running}
	}
	wc, bc := bar(w, whiteToMove), bar(b, !whiteToMove)
	top, bottom := bc, wc
	if !g.whiteAtBottom() {
		top, bottom = wc, bc
	}
	if clockBarsOn() {
		g.Board.SetClockBars(top, bottom)
		g.UserTimer.SetText("")
		g.OppTimer.SetText("")
		return
	}
	g.Board.HideClockBars()
	text := func(c clockBar) string {
		if c.active {
			return "[black:#c8c8c8] " + c.label + " [-:-] [#a0a0a0]" + c.right + "[-]"
		}
		return " " + c.label + " [#a0a0a0]" + c.right + "[-]"
	}
	g.OppTimer.SetText(text(top))
	g.UserTimer.SetText(text(bottom))
}

// clockTicker refreshes the local game's clocks while it's on screen.
func (g *GameScreen) clockTicker() {
	for range time.Tick(200 * time.Millisecond) {
		if !showing("localgame") || !clockActive.Load() {
			continue
		}
		Root.App.QueueUpdateDraw(func() {
			if name, _ := Root.nav.GetFrontPanel(); name == "localgame" && g.clock.on {
				g.tickClock()
			}
		})
	}
}

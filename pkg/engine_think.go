package pkg

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/notnil/chess"
)

// How long Stockfish takes over a move in a local game. It finds its move
// in well under a second at any level, so the reply is held back to a time a
// person might take: a share of the clock, less for the opening, a forced
// move or an obvious recapture, more for a position with lots going on,
// with some randomness so it doesn't feel mechanical.

const (
	thinkNoClock  = 3 * time.Second // the typical move without a clock
	thinkMaxNoClk = 8 * time.Second
	thinkMin      = 400 * time.Millisecond
	thinkMax      = 30 * time.Second
)

// engineThinkTime is thinkTime; tests replace it so they don't wait.
var engineThinkTime = thinkTime

// thinkTime is how long to take over the move in pos. left and inc are the
// engine's time and increment; clock is false in a game without one.
func thinkTime(game *chess.Game, left, inc time.Duration, clock bool, r *rand.Rand) time.Duration {
	pos := game.Position()
	legal := pos.ValidMoves()
	if len(legal) <= 1 { // nothing to think about
		return jitter(500*time.Millisecond, r, 0.3)
	}

	// the usual time for a move: a share of what's left, or a fixed
	// amount without a clock
	moveNo := len(game.Moves())/2 + 1
	budget := thinkNoClock
	if clock {
		movesLeft := max(40, 70-moveNo) // people plan for a long game
		budget = left/time.Duration(movesLeft) + inc*3/4
	}

	// how much there is to think about
	f := 1.0
	switch {
	case moveNo <= 4: // known opening moves come quickly
		f *= 0.25
	case moveNo <= 10:
		f *= 0.6
	}
	captures, checks := 0, 0
	for _, m := range legal {
		if m.HasTag(chess.Capture) {
			captures++
		}
		if m.HasTag(chess.Check) {
			checks++
		}
	}
	f *= 0.5 + math.Min(float64(len(legal)), 45)/45*0.5 + math.Min(float64(captures+checks), 6)*0.06
	if recapture(game, legal) {
		f *= 0.5
	}

	t := jitter(time.Duration(float64(budget)*f), r, 0.35)
	limit := thinkMaxNoClk
	if clock {
		limit = min(thinkMax, left/10) // never spend much of what's left
	}
	return max(min(thinkMin, limit), min(t, limit))
}

// recapture reports whether the last move took a piece that can be taken
// back: usually an obvious reply.
func recapture(game *chess.Game, legal []*chess.Move) bool {
	moves := game.Moves()
	if len(moves) == 0 || !moves[len(moves)-1].HasTag(chess.Capture) {
		return false
	}
	to := moves[len(moves)-1].S2()
	for _, m := range legal {
		if m.S2() == to {
			return true
		}
	}
	return false
}

// jitter scales d by a random factor around 1 (log-normal with spread s,
// kept within ½–2×).
func jitter(d time.Duration, r *rand.Rand, s float64) time.Duration {
	k := math.Exp(r.NormFloat64() * s)
	return time.Duration(float64(d) * math.Max(0.5, math.Min(2, k)))
}

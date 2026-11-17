package pkg

import (
	"math/rand/v2"
	"testing"
	"time"
)

// Stockfish's thinking time looks like a person's: quick in the opening and
// for forced moves and recaptures, longer in a busy middlegame, a fair share
// of the clock, never much of what's left.
func TestThinkTime(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	// the median of many draws, so the randomness doesn't decide the test
	median := func(sans string, left, inc time.Duration, clock bool) time.Duration {
		t.Helper()
		g, err := gameFromSAN("", sans)
		if err != nil {
			t.Fatal(err)
		}
		var ts []time.Duration
		for range 201 {
			d := thinkTime(g, left, inc, clock, r)
			if d < 0 || (clock && d > max(thinkMin, left/10)) || (!clock && d > thinkMaxNoClk) {
				t.Fatalf("%q: %v with %v left", sans, d, left)
			}
			ts = append(ts, d)
		}
		for i := range ts { // (insertion sort; it's small)
			for j := i; j > 0 && ts[j] < ts[j-1]; j-- {
				ts[j], ts[j-1] = ts[j-1], ts[j]
			}
		}
		return ts[len(ts)/2]
	}
	const ten = 10 * time.Minute
	middlegame := "e4 e5 Nf3 Nc6 Bc4 Bc5 c3 Nf6 d4 exd4 cxd4 Bb4+ Bd2 Bxd2+ Nbxd2 d5 exd5 Nxd5 Qb3 Nce7 O-O O-O Rfe1 c6"
	opening := median("e4", ten, 0, true)
	busy := median(middlegame, ten, 0, true)
	if !(opening < busy/2) {
		t.Errorf("opening %v, middlegame %v: the opening should be much quicker", opening, busy)
	}
	if busy < 5*time.Second || busy > 25*time.Second {
		t.Errorf("middlegame move at 10+0 takes %v, want 5-25 s", busy)
	}
	// one legal move: after 1. e4 f5 2. Qh5+ Black must play g6
	if forced := median("e4 f5 Qh5+", ten, 0, true); forced > time.Second {
		t.Errorf("forced move takes %v", forced)
	}
	// taking back: after ...Bxd2+ White recaptures quickly
	retake := median("e4 e5 Nf3 Nc6 Bc4 Bc5 c3 Nf6 d4 exd4 cxd4 Bb4+ Bd2 Bxd2+", ten, 0, true)
	if !(retake < busy) {
		t.Errorf("recapture %v, middlegame %v", retake, busy)
	}
	// low on time: quick
	if low := median(middlegame, 20*time.Second, 0, true); low > 2*time.Second {
		t.Errorf("with 20 s left it takes %v", low)
	}
	// an increment buys time; no clock is a few seconds
	if inc := median(middlegame, time.Minute, 5*time.Second, true); inc < 2*time.Second {
		t.Errorf("1+5 with a minute left: %v", inc)
	}
	if none := median(middlegame, 0, 0, false); none < time.Second || none > thinkMaxNoClk {
		t.Errorf("no clock: %v", none)
	}
}

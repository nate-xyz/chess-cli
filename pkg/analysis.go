package pkg

import (
	"context"
	"math"
	"time"

	"github.com/notnil/chess"
)

// Computer analysis of a finished game with the local Stockfish, worked out
// the way Lichess does it, so the numbers mean the same:
//   - win% from the engine's centipawns (capped at ±1000; mate counts as
//     ±1000): 50 + 50·(2/(1+e^(−0.00368208·cp)) − 1)
//   - a move's accuracy from the win% it threw away:
//     103.1668·e^(−0.04354·drop) − 3.1669, +1, kept within 0–100
//   - a side's game accuracy: the mean of its move accuracies weighted by
//     how volatile the position was (win% spread over a sliding window),
//     averaged with their harmonic mean
//   - inaccuracy / mistake / blunder: the move cost 0.1 / 0.2 / 0.3 in
//     winning chances (win% on a −1…1 scale)
//   - average centipawn loss, with the same ±1000 cap
//   - phases: the middlegame starts once there are 10 or fewer queens,
//     rooks, bishops and knights on the board, or either side's back rank
//     has fewer than 4 of its pieces left; the endgame at 6 or fewer.
//     (Lichess also starts the middlegame when the pieces get mixed up
//     across the board; that rule isn't here, so the middlegame can start a
//     little later than on Lichess.)

type judgment int

const (
	judgeNone judgment = iota
	judgeInaccuracy
	judgeMistake
	judgeBlunder
)

func (j judgment) glyph() string {
	return [...]string{"", "?!", "?", "??"}[j]
}

func (j judgment) name() string {
	return [...]string{"", "Inaccuracy", "Mistake", "Blunder"}[j]
}

func (j judgment) plural() string {
	return [...]string{"", "Inaccuracies", "Mistakes", "Blunders"}[j]
}

// colour is the tag the move list and summary use for a judgment (Lichess's
// colours).
func (j judgment) colour() string {
	return [...]string{"", "[#56b4e9]", "[#e69f00]", "[#df5353]"}[j]
}

type phase int

const (
	phaseOpening phase = iota
	phaseMiddlegame
	phaseEndgame
)

var phaseNames = [...]string{"Opening", "Middlegame", "Endgame"}

// sideStats are one side's numbers, for the whole game or a phase.
type sideStats struct {
	moves    int
	accuracy float64 // %; -1 if no moves
	acpl     int
	counts   [4]int // by judgment
}

// analysis is a game's computer analysis, filled in as the engine goes.
type analysis struct {
	positions []*chess.Position // the game: positions[i] is before move i+1
	moves     []*chess.Move
	scores    []Score  // per position, from White's side
	best      []string // the engine's move in each position (UCI)
	done      int      // positions analysed so far

	// worked out once complete (summarize)
	win     []float64  // White's win% per position
	acc     []float64  // the mover's accuracy per move
	judged  []judgment // per move
	mid     int        // the position the middlegame starts at (len(positions) if never)
	end     int        // the position the endgame starts at
	total   [2]sideStats
	byPhase [3][2]sideStats // [phase][white, black]
}

func newAnalysis(positions []*chess.Position, moves []*chess.Move) *analysis {
	return &analysis{
		positions: positions,
		moves:     moves,
		scores:    make([]Score, len(positions)),
		best:      make([]string, len(positions)),
	}
}

func (a *analysis) complete() bool { return a.done == len(a.positions) }

// cpOf is a score in centipawns from White's side, capped at ±1000 (mate
// counts as the cap).
func cpOf(s Score) int {
	switch {
	case s.Mate > 0:
		return 1000
	case s.Mate < 0:
		return -1000
	}
	return max(-1000, min(1000, s.CP))
}

// winningChances is −1 (Black wins) … 1 (White wins).
func winningChances(s Score) float64 {
	return 2/(1+math.Exp(-0.00368208*float64(cpOf(s)))) - 1
}

func winPercent(s Score) float64 { return 50 + 50*winningChances(s) }

// moveAccuracy is how accurate a move was, from the mover's win% before and
// after it.
func moveAccuracy(before, after float64) float64 {
	if after >= before {
		return 100
	}
	raw := 103.1668100711649*math.Exp(-0.04354415386753951*(before-after)) - 3.166924740191411
	return max(0, min(100, raw+1))
}

func judge(drop float64) judgment {
	switch {
	case drop >= 0.3:
		return judgeBlunder
	case drop >= 0.2:
		return judgeMistake
	case drop >= 0.1:
		return judgeInaccuracy
	}
	return judgeNone
}

// gamePhase finds where the middlegame and endgame start.
func gamePhases(positions []*chess.Position) (mid, end int) {
	mid, end = len(positions), len(positions)
	for i, p := range positions {
		pieces, back := 0, [2]int{}
		for sq, pc := range p.Board().SquareMap() {
			if t := pc.Type(); t != chess.Pawn && t != chess.King {
				pieces++
			}
			if pc.Color() == chess.White && sq.Rank() == chess.Rank1 {
				back[0]++
			}
			if pc.Color() == chess.Black && sq.Rank() == chess.Rank8 {
				back[1]++
			}
		}
		if mid == len(positions) && (pieces <= 10 || back[0] < 4 || back[1] < 4) {
			mid = i
		}
		if mid < len(positions) && pieces <= 6 {
			end = i
			break
		}
	}
	return mid, end
}

// phaseOf is the phase a move (made from position i) belongs to.
func (a *analysis) phaseOf(i int) phase {
	switch {
	case i >= a.end:
		return phaseEndgame
	case i >= a.mid:
		return phaseMiddlegame
	}
	return phaseOpening
}

// side is 0 for White, 1 for Black: who made the move from position i.
func (a *analysis) side(i int) int {
	if a.positions[i].Turn() == chess.White {
		return 0
	}
	return 1
}

// summarize works out everything from the scores.
func (a *analysis) summarize() {
	n := len(a.moves)
	a.win = make([]float64, len(a.scores))
	for i, s := range a.scores {
		a.win[i] = winPercent(s)
	}
	a.mid, a.end = gamePhases(a.positions)
	a.acc = make([]float64, n)
	a.judged = make([]judgment, n)

	// volatility weights, one per move (Lichess's windows)
	window := max(2, min(8, len(a.win)/10))
	weights := make([]float64, n)
	for i := 0; i < n; i++ {
		lo := max(0, i+2-window)
		hi := min(len(a.win), lo+window)
		weights[i] = max(0.5, min(12, stdDev(a.win[lo:hi])))
	}

	type acc struct{ v, w float64 }
	var accs [2][]acc
	var phaseAccs [3][2][]float64
	var loss [2]int
	var phaseLoss [3][2]int
	for i := 0; i < n; i++ {
		s := a.side(i)
		before, after := a.win[i], a.win[i+1]
		wcBefore, wcAfter := winningChances(a.scores[i]), winningChances(a.scores[i+1])
		cpBefore, cpAfter := cpOf(a.scores[i]), cpOf(a.scores[i+1])
		if s == 1 { // from Black's side
			before, after = 100-before, 100-after
			wcBefore, wcAfter = -wcBefore, -wcAfter
			cpBefore, cpAfter = -cpBefore, -cpAfter
		}
		a.acc[i] = moveAccuracy(before, after)
		a.judged[i] = judge(wcBefore - wcAfter)
		l := max(0, cpBefore-cpAfter)
		ph := a.phaseOf(i)

		accs[s] = append(accs[s], acc{a.acc[i], weights[i]})
		phaseAccs[ph][s] = append(phaseAccs[ph][s], a.acc[i])
		loss[s] += l
		phaseLoss[ph][s] += l
		a.total[s].moves++
		a.total[s].counts[a.judged[i]]++
		a.byPhase[ph][s].moves++
		a.byPhase[ph][s].counts[a.judged[i]]++
	}
	for s := 0; s < 2; s++ {
		t := &a.total[s]
		t.accuracy = -1
		if t.moves > 0 {
			var sum, wsum, inv float64
			for _, x := range accs[s] {
				sum += x.v * x.w
				wsum += x.w
				inv += 1 / max(x.v, 0.0001)
			}
			weighted := sum / wsum
			harmonic := float64(len(accs[s])) / inv
			t.accuracy = (weighted + harmonic) / 2
			t.acpl = loss[s] / t.moves
		}
		for ph := 0; ph < 3; ph++ {
			p := &a.byPhase[ph][s]
			p.accuracy = -1
			if p.moves > 0 {
				p.accuracy = mean(phaseAccs[ph][s])
				p.acpl = phaseLoss[ph][s] / p.moves
			}
		}
	}
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func stdDev(xs []float64) float64 {
	m := mean(xs)
	var s float64
	for _, x := range xs {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(xs)))
}

// analysisTime is how long the engine looks at each position.
const (
	analysisTime  = 300 * time.Millisecond
	analysisDepth = 20
)

// run analyses the game with the engine at path, handing each position's
// score and best move to result (on the engine's goroutine; record them with
// set on the UI goroutine). Positions where the game is over aren't searched.
func (a *analysis) run(ctx context.Context, path string, result func(i int, sc Score, best string)) error {
	e, err := StartEngine(path, 8)
	if err != nil {
		return err
	}
	defer e.Close()
	start := a.positions[0].String()
	uci := make([]string, len(a.moves))
	for i, m := range a.moves {
		uci[i] = chess.UCINotation{}.Encode(a.positions[i], m)
	}
	for i, p := range a.positions {
		var sc Score
		best := ""
		switch p.Status() {
		case chess.Checkmate:
			sc = Score{Mate: 1}
			if p.Turn() == chess.White {
				sc = Score{Mate: -1}
			}
		case chess.Stalemate, chess.InsufficientMaterial:
		default:
			sc, best, err = e.Evaluate(ctx, start, uci[:i], analysisTime, analysisDepth)
			if err != nil {
				return err
			}
		}
		result(i, sc, best)
	}
	return nil
}

// set records position i's result; once all are in, the numbers are worked
// out.
func (a *analysis) set(i int, sc Score, best string) {
	a.scores[i], a.best[i] = sc, best
	a.done = max(a.done, i+1)
	if a.complete() {
		a.summarize()
	}
}

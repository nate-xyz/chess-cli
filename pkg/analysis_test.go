package pkg

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/notnil/chess"
)

// gameOf plays SAN moves from the start.
func gameOf(t *testing.T, sans ...string) *chess.Game {
	t.Helper()
	g := chess.NewGame()
	for _, s := range sans {
		if err := g.MoveStr(s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	return g
}

func TestAnalysisFormulas(t *testing.T) {
	near := func(got, want float64) bool { return math.Abs(got-want) < 0.01 }
	if !near(winPercent(Score{}), 50) {
		t.Errorf("win%% at 0 = %v, want 50", winPercent(Score{}))
	}
	if winPercent(Score{CP: 5000}) != winPercent(Score{Mate: 3}) || winPercent(Score{Mate: -2}) != winPercent(Score{CP: -1000}) {
		t.Error("mate and big scores should share the ±1000 cap")
	}
	// Lichess: a 10-point win% drop is 103.1668·e^(−0.4354) − 3.1669 + 1
	if got := moveAccuracy(60, 50); !near(got, 103.1668100711649*math.Exp(-0.4354415386753951)-3.166924740191411+1) {
		t.Errorf("accuracy for a 10%% drop = %v", got)
	}
	if moveAccuracy(50, 55) != 100 || moveAccuracy(100, 0) != 0 {
		t.Error("accuracy should be 100 for no loss and 0 for throwing a won game away")
	}
	for drop, want := range map[float64]judgment{0.05: judgeNone, 0.1: judgeInaccuracy, 0.25: judgeMistake, 0.31: judgeBlunder} {
		if got := judge(drop); got != want {
			t.Errorf("drop %v judged %v, want %v", drop, got, want)
		}
	}
}

func TestGamePhases(t *testing.T) {
	g := gameOf(t, "e4", "e5")
	if mid, end := gamePhases(g.Positions()); mid != 3 || end != 3 {
		t.Errorf("opening only: mid %d end %d, want neither (3)", mid, end)
	}
	// a king and pawn ending is an endgame from the start
	kp, _ := chess.FEN("4k3/8/8/8/8/8/4P3/4K3 w - - 0 1")
	pos := chess.NewGame(kp).Positions()
	if mid, end := gamePhases(pos); mid != 0 || end != 0 {
		t.Errorf("K+P: mid %d end %d, want 0 and 0", mid, end)
	}
}

// White plays well, then blunders: the blunder is found, costs accuracy and
// centipawns, and only White's numbers suffer.
func TestAnalysisSummary(t *testing.T) {
	g := gameOf(t, "e4", "e5", "Nf3", "Nc6", "Bc4", "Nf6", "Ng5", "d5")
	a := newAnalysis(g.Positions(), g.Moves())
	// position 7 is after Ng5 (White's move from 6); make that the blunder
	cps := []int{20, 30, 25, 30, 30, 35, 30, -400, -380}
	for i, cp := range cps {
		a.set(i, Score{CP: cp}, "")
	}
	if !a.complete() {
		t.Fatal("not complete")
	}
	if a.judged[6] != judgeBlunder {
		t.Errorf("Ng5 judged %v, want blunder", a.judged[6])
	}
	w, b := a.total[0], a.total[1]
	if w.counts[judgeBlunder] != 1 || b.counts[judgeBlunder]+b.counts[judgeMistake]+b.counts[judgeInaccuracy] != 0 {
		t.Errorf("counts white %v black %v", w.counts, b.counts)
	}
	if !(w.accuracy < b.accuracy) || w.acpl < 100 || b.acpl > 10 {
		t.Errorf("white %.1f%% acpl %d, black %.1f%% acpl %d", w.accuracy, w.acpl, b.accuracy, b.acpl)
	}
	if a.byPhase[phaseOpening][0].moves != 4 || a.byPhase[phaseOpening][0].acpl != w.acpl {
		t.Errorf("opening stats %+v", a.byPhase[phaseOpening][0])
	}
}

// The engine run with the real Stockfish (skipped when it isn't
// installed): a result for every position, the final mate scored without
// searching.
func TestAnalysisRunStockfish(t *testing.T) {
	path := realStockfish(t)
	g := gameOf(t, "f3", "e5", "g4", "Qh4#")
	a := newAnalysis(g.Positions(), g.Moves())
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	got := 0
	err := a.run(ctx, path, func(i int, sc Score, best string) {
		a.set(i, sc, best)
		got++
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != 5 || !a.complete() {
		t.Fatalf("%d results, complete %v", got, a.complete())
	}
	if a.scores[4] != (Score{Mate: -1}) || a.best[4] != "" {
		t.Errorf("final position %+v best %q, want White mated", a.scores[4], a.best[4])
	}
	if a.scores[3].Mate >= 0 { // after g4, Black mates in one
		t.Errorf("after g4: %+v, want Black mating", a.scores[3])
	}
	if a.judged[2] != judgeBlunder {
		t.Errorf("g4 judged %v, want blunder", a.judged[2])
	}
}

// The analysis on the game-over screen, from scores (no engine): graph,
// table, marks in the move list and the note for the move shown.
func TestAnalysisPanel(t *testing.T) {
	startTestApp(t, 160, 50)
	sans := []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Nf6", "Ng5", "d5"}
	cps := []int{20, 30, 25, 30, 30, 35, 30, -400, -380}
	var p *analysisPanel
	onUI(func() {
		Root.NewGame()
		g := Root.gameState.Game
		for _, s := range sans {
			m, err := chess.AlgebraicNotation{}.Decode(g.Position(), s)
			if err == nil {
				err = g.Move(m)
			}
			if err != nil {
				t.Error(s, err) // not Fatal: this is the UI goroutine
				return
			}
		}
		gotoPostLocal()
		pg := Root.pgame.postGame
		p = pg.analysis
		a := newAnalysis(pg.positions, pg.moves)
		p.a, p.graph.a = a, a
		p.side.ResizeItem(p.graph, 9, 0)
		for i, cp := range cps {
			a.set(i, Score{CP: cp}, "")
		}
		a.best[6] = "c4f7" // Bxf7+ instead of Ng5
		p.finished()
		pg.Board.ShowPly(7)
	})
	waitFor(t, "the note for Ng5", func() bool { return strings.Contains(p.note.GetText(true), "4. Ng5??") })
	onUI(func() {
		note := p.note.GetText(true)
		if !strings.Contains(note, "Blunder") || !strings.Contains(note, "best was Bxf7+") {
			t.Errorf("note %q", note)
		}
		table := p.table.GetText(true)
		for _, want := range []string{"Accuracy", "Avg cp loss", "Blunders", "Opening"} {
			if !strings.Contains(table, want) {
				t.Errorf("table has no %q:\n%s", want, table)
			}
		}
		if !strings.Contains(p.history.GetText(true), "Ng5??") {
			t.Errorf("move list not marked:\n%s", p.history.GetText(true))
		}
		_, _, _, gh := p.graph.GetRect()
		if gh != 9 {
			t.Errorf("graph %d rows, want 9", gh)
		}
	})
	onUI(func() { gotoWelcome() })
}

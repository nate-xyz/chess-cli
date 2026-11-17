package pkg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

// The analysis on a game-over screen: an evaluation graph (White's winning
// chances over the game, the phases marked, the move shown highlighted; click
// it to go to a move), a table of each side's accuracy, centipawn loss and
// mistakes, overall and by phase, and ?! ? ?? marks in the move list.

// evalGraph draws White's win% per position as a filled area: light is
// White's share, dark Black's, so the middle row is equality.
type evalGraph struct {
	*cv.Box
	a     *analysis
	board *BoardView
}

var (
	colGraphWhite   = tc.NewRGBColor(0xe0, 0xdc, 0xcf)
	colGraphBlack   = tc.NewRGBColor(0x3a, 0x36, 0x32)
	colGraphPending = tc.NewRGBColor(0x24, 0x22, 0x20)
	colGraphPhase   = tc.NewRGBColor(0xb0, 0x8f, 0x3c)
	colGraphShown   = tc.NewRGBColor(0x5a, 0x7f, 0xbf)
)

func newEvalGraph(board *BoardView) *evalGraph {
	g := &evalGraph{Box: cv.NewBox(), board: board}
	return g
}

// positionAt is the position drawn in column c of a w-wide graph.
func (g *evalGraph) positionAt(c, w int) int {
	n := len(g.a.positions)
	return min(n-1, c*n/max(1, w))
}

func (g *evalGraph) Draw(screen tc.Screen) {
	g.Box.Draw(screen)
	if g.a == nil {
		return
	}
	x, y, w, h := g.GetInnerRect()
	if w <= 0 || h <= 0 {
		return
	}
	limited := uiLimited()
	white, black, pending, phaseCol, shownCol := colGraphWhite, colGraphBlack, colGraphPending, colGraphPhase, colGraphShown
	if limited {
		white, black, pending, phaseCol, shownCol = tc.ColorSilver, tc.ColorGray, tc.ColorBlack, tc.ColorOlive, tc.ColorNavy
	}
	g.board.mu.Lock()
	shown := g.board.shownPly
	g.board.mu.Unlock()
	a := g.a
	blocks := []rune(" ▁▂▃▄▅▆▇█")
	for c := 0; c < w; c++ {
		i := g.positionAt(c, w)
		first := c == 0 || g.positionAt(c-1, w) != i
		bg := black
		mark := false
		switch {
		case i >= a.done:
			bg = pending
		case a.complete() && first && i > 0 && (i == a.mid || i == a.end):
			mark = true
		}
		level := 0
		if i < a.done {
			level = int(winPercent(a.scores[i]) / 100 * float64(h*8))
		}
		for r := 0; r < h; r++ { // r = 0 is the bottom row
			eighths := max(0, min(8, level-r*8))
			fg, cellBg := white, bg
			ch := blocks[eighths]
			if mark {
				cellBg = phaseCol
				if eighths == 8 {
					fg = phaseCol
				}
			}
			if i == shown && i < a.done {
				cellBg = blend(cellBg, shownCol, 0.6)
				if eighths == 8 {
					fg = blend(fg, shownCol, 0.4)
				}
			}
			screen.SetContent(x+c, y+h-1-r, ch, nil, tc.StyleDefault.Foreground(fg).Background(cellBg))
		}
	}
}

// MouseHandler: a click (or drag) goes to that move.
func (g *evalGraph) MouseHandler() func(action cv.MouseAction, event *tc.EventMouse, setFocus func(p cv.Primitive)) (bool, cv.Primitive) {
	return g.WrapMouseHandler(func(action cv.MouseAction, event *tc.EventMouse, setFocus func(p cv.Primitive)) (bool, cv.Primitive) {
		mx, my := event.Position()
		if !g.InRect(mx, my) || g.a == nil {
			return false, nil
		}
		switch action {
		case cv.MouseLeftDown, cv.MouseLeftClick:
			x, _, w, _ := g.GetInnerRect()
			if mx >= x && mx < x+w {
				g.board.ShowPly(g.positionAt(mx-x, w))
			}
		case cv.MouseScrollUp:
			g.board.mu.Lock()
			g.board.step(1)
			g.board.mu.Unlock()
		case cv.MouseScrollDown:
			g.board.mu.Lock()
			g.board.step(-1)
			g.board.mu.Unlock()
		}
		return true, nil
	})
}

// analysisTable is the summary: each side's numbers, overall and by phase.
func analysisTable(a *analysis) string {
	var b strings.Builder
	pct := func(s sideStats) string {
		if s.accuracy < 0 {
			return "  –"
		}
		return fmt.Sprintf("%3.0f%%", s.accuracy)
	}
	row := func(label, w, bl string) {
		fmt.Fprintf(&b, "%-12s %-8s %s\n", label, w, bl)
	}
	b.WriteString("[::b]             White    Black[::-]\n")
	row("Accuracy", pct(a.total[0]), pct(a.total[1]))
	row("Avg cp loss", fmt.Sprint(a.total[0].acpl), fmt.Sprint(a.total[1].acpl))
	for _, j := range []judgment{judgeInaccuracy, judgeMistake, judgeBlunder} {
		label := j.colour() + fmt.Sprintf("%-12s", j.plural()) + "[-]"
		fmt.Fprintf(&b, "%s %-8d %d\n", label, a.total[0].counts[j], a.total[1].counts[j])
	}
	b.WriteString("[::b]By phase     acc·acpl acc·acpl[::-]\n")
	for ph := phaseOpening; ph <= phaseEndgame; ph++ {
		w, bl := a.byPhase[ph][0], a.byPhase[ph][1]
		if w.moves+bl.moves == 0 {
			continue
		}
		cell := func(s sideStats) string {
			if s.moves == 0 {
				return "–"
			}
			return fmt.Sprintf("%.0f%%·%d", s.accuracy, s.acpl)
		}
		row(phaseNames[ph], cell(w), cell(bl))
	}
	return b.String()
}

// moveNote describes the move that led to position ply: its judgment and
// what the engine preferred.
func moveNote(a *analysis, ply int) string {
	if !a.complete() || ply < 1 || ply > len(a.moves) {
		return ""
	}
	i := ply - 1
	p := a.positions[i]
	san := chess.AlgebraicNotation{}.Encode(p, a.moves[i])
	num := fmt.Sprintf("%d. ", moveNumber(p))
	if p.Turn() == chess.Black {
		num = fmt.Sprintf("%d… ", moveNumber(p))
	}
	j := a.judged[i]
	text := fmt.Sprintf("[::b]%s%s[::-]", num, san)
	if j != judgeNone {
		text += j.colour() + j.glyph() + " " + j.name() + "[-]"
	}
	if best := a.best[i]; best != "" && j != judgeNone {
		for _, m := range p.ValidMoves() { // these know about check, for the +
			if (chess.UCINotation{}).Encode(p, m) == best {
				text += " · best was " + chess.AlgebraicNotation{}.Encode(p, m)
			}
		}
	}
	return text + fmt.Sprintf("\n[#a0a0a0]accuracy %.0f%% · eval %s[-]", a.acc[i], scoreText(a.scores[ply]))
}

// scoreText is a score as Lichess writes it: +1.3, -0.4, #3, #-2.
func scoreText(s Score) string {
	if s.Mate != 0 {
		return fmt.Sprintf("#%d", s.Mate)
	}
	return fmt.Sprintf("%+.1f", float64(s.CP)/100)
}

// moveMarks are the analysis's ?! ? ?? for the move list.
func moveMarks(a *analysis) []moveMark {
	marks := make([]moveMark, len(a.moves))
	for i, j := range a.judged {
		if j != judgeNone {
			marks[i] = moveMark{text: j.colour() + j.glyph() + "[-]", width: len(j.glyph())}
		}
	}
	return marks
}

// analysisPanel is the analysis part of a game-over sidebar.
type analysisPanel struct {
	side    *cv.Flex
	history *cv.TextView
	board   *BoardView
	graph   *evalGraph
	setT    func(string) // the graph box's title
	table   *cv.TextView
	note    *cv.TextView

	a      *analysis
	cancel context.CancelFunc
	gen    int // which run's results are wanted
}

func newAnalysisPanel(side *cv.Flex, history *cv.TextView, board *BoardView) *analysisPanel {
	p := &analysisPanel{side: side, history: history, board: board, graph: newEvalGraph(board), table: textBox(cv.AlignLeft), note: textBox(cv.AlignLeft)}
	p.setT = boxed(p.graph.Box, "")
	return p
}

// reset hides the analysis (a new game ended).
func (p *analysisPanel) reset() {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.gen++
	p.a = nil
	p.graph.a = nil
	for _, it := range []cv.Primitive{p.graph, p.table, p.note} {
		p.side.ResizeItem(it, 0, 0)
	}
	p.board.SetGlyph(chess.NoSquare, 0)
}

// start analyses the game with the local Stockfish; status reports
// problems (no Stockfish, the engine failing).
func (p *analysisPanel) start(positions []*chess.Position, moves []*chess.Move, status func(string)) {
	if p.a != nil || len(moves) == 0 {
		return // running or done already, or nothing to analyse
	}
	path := findEngine()
	if path == "" {
		status("[yellow]Computer analysis needs Stockfish.[-] Get it from Home → Local → Play local Stockfish.")
		return
	}
	a := newAnalysis(positions, moves)
	p.a, p.graph.a = a, a
	p.side.ResizeItem(p.graph, 9, 0)
	p.setT(fmt.Sprintf("Analysing… 0/%d", len(positions)))
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.gen++
	gen := p.gen
	go func() {
		err := a.run(ctx, path, func(i int, sc Score, best string) {
			Root.App.QueueUpdateDraw(func() {
				if gen != p.gen {
					return
				}
				a.set(i, sc, best)
				if a.complete() {
					p.finished()
				} else {
					p.setT(fmt.Sprintf("Analysing… %d/%d", a.done, len(positions)))
				}
			})
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			Root.App.QueueUpdateDraw(func() {
				if gen == p.gen {
					p.setT("Analysis stopped")
					status(fmt.Sprintf("[red]Analysis failed: %v[-]", cv.Escape(err.Error())))
				}
			})
		}
	}()
}

// finished shows the numbers once every position is analysed.
func (p *analysisPanel) finished() {
	a := p.a
	p.setT("Stockfish analysis") // short: the sidebar can be 30 wide
	p.table.SetText(analysisTable(a))
	p.side.ResizeItem(p.table, strings.Count(p.table.GetText(false), "\n"), 0)
	p.side.ResizeItem(p.note, 2, 0)
	p.history.SetText(moveListFrom(a.positions, a.moves, moveMarks(a)))
	p.board.mu.Lock()
	ply := p.board.shownPly
	p.board.mu.Unlock()
	p.shown(ply)
}

// shown updates the note for the move the board is showing, and marks the
// square it went to with its ?! ? or ?? (as studies show their symbols).
func (p *analysisPanel) shown(ply int) {
	if p.a == nil {
		return
	}
	p.note.SetText(moveNote(p.a, ply))
	if p.a.complete() && ply >= 1 && ply <= len(p.a.moves) {
		if j := p.a.judged[ply-1]; j != judgeNone {
			p.board.SetGlyph(p.a.moves[ply-1].S2(), j.nag())
			return
		}
	}
	p.board.SetGlyph(chess.NoSquare, 0)
}

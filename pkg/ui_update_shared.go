package pkg

import (
	"fmt"
	"strings"

	cv "codeberg.org/tslocum/cview"
	"github.com/notnil/chess"
)

func DrawMoveHistory(win *cv.TextView) {
	win.SetText(moveListText(Root.gameState.Game))
	win.ScrollToEnd()
}

// moveListText is a game's moves in standard notation, numbered in pairs.
func moveListText(game *chess.Game) string {
	return moveListFrom(game.Positions(), game.Moves(), nil)
}

// moveListFrom is moveListText for positions and moves, with an optional
// mark after each move (a coloured ?! from the analysis): marks[i] is shown
// after move i+1, as text and its visible width.
func moveListFrom(positions []*chess.Position, moves []*chess.Move, marks []moveMark) string {
	var text strings.Builder
	text.WriteString("[::b]Moves[::-]\n")
	for i, m := range moves {
		san := chess.AlgebraicNotation{}.Encode(positions[i], m)
		width := len([]rune(san))
		// each move is a region, m<ply>, so it can be clicked (linkMoveList)
		tagged := fmt.Sprintf(`["m%d"]%s[""]`, i+1, san)
		if i < len(marks) && marks[i].width > 0 {
			tagged += marks[i].text
			width += marks[i].width
		}
		if positions[i].Turn() == chess.White {
			fmt.Fprintf(&text, "%3d. %s%s", moveNumber(positions[i]), tagged, strings.Repeat(" ", max(1, 8-width)))
		} else {
			if i == 0 { // game started with black to move
				fmt.Fprintf(&text, "%3d. %-8s", moveNumber(positions[0]), "…")
			}
			text.WriteString(tagged + "\n")
		}
	}
	return text.String()
}

// moveMark is something shown after a move in the move list.
type moveMark struct {
	text  string
	width int
}

// moveNumber is a position's full move number (the FEN's last field).
func moveNumber(pos *chess.Position) int {
	n := 1
	if f := strings.Fields(pos.String()); len(f) == 6 {
		fmt.Sscan(f[5], &n)
	}
	return n
}

// DrawBoard shows the current game position on a board, with white at the
// bottom if white is true, and highlights the last move played.
func DrawBoard(board *BoardView, white bool) {
	game := Root.gameState.Game
	board.SetPosition(game.Position(), white)
	board.SetHistory(game.Positions(), game.Moves())
	from, to := chess.NoSquare, chess.NoSquare
	if moves := game.Moves(); len(moves) > 0 {
		last := moves[len(moves)-1]
		from, to = last.S1(), last.S2()
	}
	board.SetLastMove(from, to)
	Root.App.QueueUpdateDraw(func() {}) // full draw, so big pieces are flushed too
}

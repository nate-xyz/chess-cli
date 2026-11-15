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
	moves, positions := game.Moves(), game.Positions()
	var text strings.Builder
	text.WriteString("[::b]Moves[::-]\n")
	for i, m := range moves {
		san := chess.AlgebraicNotation{}.Encode(positions[i], m)
		// each move is a region, m<ply>, so it can be clicked (linkMoveList)
		tagged := fmt.Sprintf(`["m%d"]%s[""]`, i+1, san)
		if positions[i].Turn() == chess.White {
			fmt.Fprintf(&text, "%3d. %s%s", moveNumber(positions[i]), tagged, strings.Repeat(" ", max(1, 8-len([]rune(san)))))
		} else {
			if i == 0 { // game started with black to move
				fmt.Fprintf(&text, "%3d. %-8s", moveNumber(positions[0]), "…")
			}
			text.WriteString(tagged + "\n")
		}
	}
	return text.String()
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

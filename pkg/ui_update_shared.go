package pkg

import (
	"fmt"

	cv "code.rocketnine.space/tslocum/cview"
	"github.com/notnil/chess"
)

func DrawMoveHistory(win *cv.TextView) {
	var text string
	moveArr := Root.gameState.MoveHistoryArray
	if moveArr == nil {
		return
	}
	pieceArr, err := GetPieceArr(moveArr)
	if err != nil {
		text += fmt.Sprintf("%v", err)
		return
	}
	if pieceArr == nil {
		return
	}
	for i := len(moveArr) - 1; i >= 0; i-- {
		move := moveArr[i]
		text += fmt.Sprintf("\t%v: %v %v", i, pieceArr[i], move)
		if i%2 == 0 {
			text += " ⬜\n"
		} else {
			text += " ⬛\n"
		}
	}
	win.SetText(text)
}

// DrawBoard shows the current game position on a board, with white at the
// bottom if white is true, and highlights the last move played.
func DrawBoard(board *BoardView, white bool) {
	game := Root.gameState.Game
	board.SetPosition(game.Position(), white)
	from, to := chess.NoSquare, chess.NoSquare
	if moves := game.Moves(); len(moves) > 0 {
		last := moves[len(moves)-1]
		from, to = last.S1(), last.S2()
	}
	board.SetLastMove(from, to)
	Root.App.QueueUpdateDraw(func() {}, board)
}

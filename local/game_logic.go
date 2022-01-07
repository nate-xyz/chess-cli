package local

import (
	"strings"

	. "github.com/nate-xyz/chess-cli/shared"
	ncurses "github.com/nate-xyz/goncurses_"
	"github.com/notnil/chess"
)

func game_logic(board_window *ncurses.Window) bool {
	//EnteredPromptStr = EnteredPromptStr.strip(' ').strip('\0').strip('^@')
	EnteredPromptStr = strings.Trim(EnteredPromptStr, " ^@")
	board_window.MovePrint(1, 1, EnteredPromptStr)
	legal_moves := CurrentGame.ValidMoves()
	LegalMoveStrArray = []string{}

	for _, move := range legal_moves {
		LegalMoveStrArray = append(LegalMoveStrArray, move.String())
	}

	if HasEnteredMove {
		HasEnteredMove = false

		if err := CurrentGame.MoveStr(EnteredPromptStr); err != nil {
			StatusMessage = "last input is invalid"
			EnteredPromptStr = ""
		} else {
			StatusMessage = "move is legal!"
			LastMoveString = EnteredPromptStr //set the last move string to be displayed in the info window
			MoveHistoryArray = append([]string{EnteredPromptStr}, MoveHistoryArray...)
			MoveAmount++ //increment the global move amount for the history window
			ncurses.Flash()
			ncurses.Beep()

			if CurrentGame.Outcome() != chess.NoOutcome { //check if the game is over
				StatusMessage = CurrentGame.Method().String()
				FinalBoardFEN = CurrentGame.Position().Board().Draw()
				return true
			}
		}

	}

	legal_moves = CurrentGame.ValidMoves()
	return false
}

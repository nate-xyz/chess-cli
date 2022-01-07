package main

import (
	"strings"

	ncurses "github.com/nate-xyz/goncurses_"
	"github.com/notnil/chess"
)

func game_logic(board_window *ncurses.Window) bool {
	//inputted_str = inputted_str.strip(' ').strip('\0').strip('^@')
	inputted_str = strings.Trim(inputted_str, " ^@")
	board_window.MovePrint(1, 1, inputted_str)
	legal_moves := game.ValidMoves()
	legal_move_str_array = []string{}

	for _, move := range legal_moves {
		legal_move_str_array = append(legal_move_str_array, move.String())
	}

	if entered_move {
		entered_move = false

		if err := game.MoveStr(inputted_str); err != nil {
			status_str = "last input is invalid"
			inputted_str = ""
		} else {
			status_str = "move is legal!"
			last_move_str = inputted_str //set the last move string to be displayed in the info window
			history_arr = append([]string{inputted_str}, history_arr...)
			move_amount++ //increment the global move amount for the history window
			ncurses.Flash()
			ncurses.Beep()

			if game.Outcome() != chess.NoOutcome { //check if the game is over
				status_str = game.Method().String()
				final_position = game.Position().Board().Draw()
				return true
			}
		}

	}

	legal_moves = game.ValidMoves()
	return false
}

func lichess_game_logic(board_window *ncurses.Window) bool {
	//inputted_str = inputted_str.strip(' ').strip('\0').strip('^@')
	inputted_str = strings.Trim(inputted_str, " ^@")
	board_window.MovePrint(1, 1, inputted_str)
	legal_moves := game.ValidMoves()
	legal_move_str_array = []string{}

	for _, move := range legal_moves {
		legal_move_str_array = append(legal_move_str_array, move.String())
	}

	if entered_move {
		entered_move = false

		if err := game.MoveStr(inputted_str); err != nil {
			status_str = "last input is invalid"
			inputted_str = ""
		} else {
			status_str = "move is legal!"
			last_move_str = inputted_str //set the last move string to be displayed in the info window
			history_arr = append([]string{inputted_str}, history_arr...)
			move_amount++ //increment the global move amount for the history window
			ncurses.Flash()
			ncurses.Beep()

			if game.Outcome() != chess.NoOutcome { //check if the game is over
				status_str = game.Method().String()
				final_position = game.Position().Board().Draw()
				return true
			}
		}

	}

	legal_moves = game.ValidMoves()
	return false
}

package online

import (
	"fmt"
	"strings"

	. "github.com/nate-xyz/chess-cli/shared"
	ncurses "github.com/nate-xyz/goncurses_"
)

func DisplayLichessInfoWindow(info_window *ncurses.Window) {
	height, width := info_window.MaxYX()
	type message struct {
		messageString string
		colorPairs    []int16
	}
	var arr []message

	var toMove string

	arr = append(arr, message{messageString: BoardFullGame.ID})
	arr = append(arr, message{messageString: LastMoveString})

	// if StatusMessage != "" {
	// 	arr = append(arr, message{messageString: StatusMessage})
	// }

	arr = append(arr, message{messageString: StatusMessage})

	n_moves := len(strings.Split(BoardGameState.Moves, " "))
	if n_moves%2 != 0 {
		toMove = fmt.Sprintf("white to move (%s)", BoardFullGame.White.Name)
	} else {
		toMove = fmt.Sprintf("black to move (%s)", BoardFullGame.Black.Name)

	}
	if CurrentStreamEventGame.Game.IsMyTurn {
		arr = append(arr, message{messageString: fmt.Sprintf("your turn!")})
	} else {
		arr = append(arr, message{messageString: fmt.Sprintf("%s's turn", CurrentStreamEventGame.Game.Opponent.Username)})
	}

	arr = append(arr, message{messageString: fmt.Sprintf("playing against %s, rating %v.",
		CurrentStreamEventGame.Game.Opponent.Username,
		CurrentStreamEventGame.Game.Opponent.Rating)})

	arr = append(arr, message{messageString: toMove})

	// arr = append(arr, message{
	// 	messageString: fmt.Sprintf("playing white: %s",
	// 		BoardFullGame.White.Name)})

	// arr = append(arr, message{
	// 	messageString: fmt.Sprintf("playing black: %s",
	// 		BoardFullGame.Black.Name)})

	arr = append(arr, message{
		messageString: fmt.Sprintf("game speed: %s",
			BoardFullGame.Speed)})

	arr = append(arr, message{
		messageString: fmt.Sprintf("game variant: %s",
			BoardFullGame.Variant.Name)})

	// info_window.MovePrint(3, 1, fmt.Sprintf("last move: %s", LastMoveString))

	var text_colour int16
	if StatusMessage == "move is legal!" {
		text_colour = 8
	} else {
		text_colour = 9
	}
	arr = append(arr, message{messageString: StatusMessage, colorPairs: []int16{text_colour}})

	// info_window.MovePrint(5, 1, fmt.Sprintf("input: %s", EnteredPromptStr))

	//wrap_y := 0
	san_move_str := fmt.Sprintf("legal moves: %s", strings.Join(LegalMoveStrArray[:], ", "))
	arr = append(arr, message{messageString: san_move_str, colorPairs: []int16{8}})

	y := 1
	for i := 0; i < len(arr) && y < height; i++ {
		for c := 0; c < len(arr[i].colorPairs); c++ {
			info_window.AttrOn(ncurses.ColorPair(arr[i].colorPairs[c]))
		}

		info_window.MovePrint(y, 1, arr[i].messageString)
		y++
		for len(arr[i].messageString)-width > 0 && y < height {
			arr[i].messageString = arr[i].messageString[width-2:]
			info_window.MovePrint(y, 1, arr[i].messageString)
			y++
		}
		for c := 0; c < len(arr[i].colorPairs); c++ {
			info_window.AttrOff(ncurses.ColorPair(arr[i].colorPairs[c]))
		}

	}
	// for y := 7; y < height-1; y++ {
	// 	//wrap_y = y
	// 	if len(san_move_str) > width-2 {
	// 		info_window.MovePrint(y, 1, san_move_str[:width-2])
	// 		san_move_str = san_move_str[width-2:]
	// 	} else {
	// 		info_window.MovePrint(y, 1, san_move_str)
	// 		break
	// 	}

	// }
	// legal_move_str := fmt.Sprintf("legal moves (uci): %s", legal_move_str)

	// for y := wrap_y + 2; y < height-1; y++ {
	// 	if len(legal_move_str) > width-2 {
	// 		info_window.MovePrint(y, 1, legal_move_str[:width-2])
	// 		legal_move_str = legal_move_str[width-2:]
	// 	} else {
	// 		info_window.MovePrint(y, 1, legal_move_str)
	// 		break
	// 	}
	// }
	// info_window.MovePrint(7, 1, "{}: {}".format("legal moves (uci)", legal_move_str))

	StatusMessage = ""
}

func DisplayLichessHistoryWindow(history_window *ncurses.Window) {
	height, width := history_window.MaxYX()
	var historyString string

	if len(BoardFullGame.State.Moves) > len(BoardGameState.Moves) {
		historyString = BoardFullGame.State.Moves
	} else {
		historyString = BoardGameState.Moves
	}

	y := 1
	for len(historyString) > 0 {
		if y >= height {
			y = 1
		}
		if len(historyString) < width-1 {
			history_window.MovePrint(y, 1, historyString)
			break
		} else {
			history_window.MovePrint(y, 1, historyString[:width-2])
			historyString = historyString[width-2:]
			y++
		}

	}
}

func DisplayLichessPostHistoryWindow(history_window *ncurses.Window, moves string, finalEvent string) {
	height, width := history_window.MaxYX()
	var historyString string

	if len(BoardFullGame.State.Moves) > len(BoardGameState.Moves) {
		historyString = BoardFullGame.State.Moves
	} else {
		historyString = BoardGameState.Moves
	}
	result_str, result_str_name := GameOutcome(moves)

	if historyString != "" {
		historyString = fmt.Sprintf("Move History: %s", historyString)
	}
	if finalEvent == "" {
		historyString = fmt.Sprintf("%s %s %s", result_str, result_str_name, historyString)
	} else if result_str == "" {
		historyString = fmt.Sprintf("%s %s", finalEvent, historyString)
	} else {
		historyString = fmt.Sprintf("%s %s %s %s", finalEvent, result_str, result_str_name, historyString)
	}
	y := 1
	for len(historyString) > 0 {
		if y >= height {
			y = 2
		}
		if len(historyString) < width-1 {
			history_window.MovePrint(y, 1, historyString)
			break
		} else {
			history_window.MovePrint(y, 1, historyString[:width-2])
			historyString = historyString[width-2:]
			y++
		}

	}
}

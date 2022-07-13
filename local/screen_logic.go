package local

import (
	. "github.com/nate-xyz/chess-cli/shared"
	ncurses "github.com/nate-xyz/goncurses_"
	"github.com/notnil/chess"
)

func WelcomeScreen(screen *ncurses.Window) ncurses.Key {
	var key ncurses.Key
	height, width := screen.MaxYX()
	//start windows
	//options := []string{"<<press '1' to play locally>>", "<<press '2' to play online>>", "<<press '3' to play stockfish>>", "<<quit>>"}

	options := []string{"play locally", "play online", "play stockfish", "quit", "test challenge a friend", "test challenge the ai"}
	op_info := WinInfo{H: (height / 2) - 4, W: width / 2, Y: (height / 2) + 2, X: width / 4}
	options_window, _ := ncurses.NewWindow(op_info.H, op_info.W, op_info.Y, op_info.X)
	windows_array := [1]*ncurses.Window{options_window}
	windows_info_arr := [1]WinInfo{op_info}
	var option_index int = 0
	var selected bool
	DrawWelcomeScreen(screen, key, windows_array, windows_info_arr, options)
	for {
		select {
		case <-Sigs:
			tRow, tCol, _ := OsTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			DrawWelcomeScreen(screen, key, windows_array, windows_info_arr, options)
		default: //normal character loop here
			key = screen.GetChar()
			option_index, selected = OptionsInput(options_window, key, options, option_index)
			if selected {
				switch option_index {
				case 0:
					key = OneKey
				case 1:
					key = TwoKey
				case 2:
					// key = ThreeKey
				case 3:
					key = CtrlO_Key
				case 4:
					key = TwoKey
				case 5:
					key = TwoKey
				}
			}
			switch key {
			case CtrlO_Key:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 3)
				}
				return key
			case OneKey:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 0)
				}
				return key
			case TwoKey:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 1)
				}
				return key
			case ThreeKey:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 2)
				}
				return key
			}
		}
	}
}

func LocalGameScreen(stdscr *ncurses.Window) int {
	var key ncurses.Key
	// var mouse_event *ncurses.MouseEvent
	// _ = mouse_event

	height, width := stdscr.MaxYX()

	//start windows
	bw_info := WinInfo{H: (height / 4) * 3, W: width / 2, Y: 0, X: 0}
	iw_info := WinInfo{H: (height / 4) - 1, W: width / 2, Y: (height / 4) * 3, X: 0}
	pw_info := WinInfo{H: height / 2, W: width / 2, Y: 0, X: width / 2}
	hw_info := WinInfo{H: (height / 2) - 1, W: width / 2, Y: height / 2, X: width / 2}

	board_window, _ := ncurses.NewWindow(bw_info.H, bw_info.W, bw_info.Y, bw_info.X)
	prompt_window, _ := ncurses.NewWindow(pw_info.H, pw_info.W, pw_info.Y, pw_info.X)
	info_window, _ := ncurses.NewWindow(iw_info.H, iw_info.W, iw_info.Y, iw_info.X)
	history_window, _ := ncurses.NewWindow(hw_info.H, hw_info.W, hw_info.Y, hw_info.X)

	windows_array := [4]*ncurses.Window{board_window, info_window, prompt_window, history_window}
	windows_info_arr := [4]WinInfo{bw_info, iw_info, pw_info, hw_info}

	DrawLocalGameScreen(stdscr, key, windows_array, windows_info_arr)
	for {
		select {
		case <-Sigs:
			tRow, tCol, _ := OsTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			DrawLocalGameScreen(stdscr, key, windows_array, windows_info_arr)
		default: //normal character loop here
			//external function calls
			UpdateInput(prompt_window, key)
			if game_logic(board_window) {
				return 2
			}
			if CurrentGame.Position().Turn() == chess.Black {
				DrawBoardWindow(board_window, CurrentGame.Position().Board().Flip(chess.UpDown).Flip(chess.LeftRight).String(), true)
			} else {
				DrawBoardWindow(board_window, CurrentGame.Position().String(), false)
			}

			DisplayInfoWindow(info_window)
			DisplayHistoryWindow(history_window)
			//board_window_mouse_input(board_window, key, width, height)

			//TODO: move refresh call to within window function
			for _, win := range windows_array {
				win.NoutRefresh()
			}
			ncurses.Update() // var board_window *ncurses.Window
			// var prompt_window *ncurses.Window
			// var info_window *ncurses.Window
			// var history_window *ncurses.Window
			key = stdscr.GetChar()
			switch key {
			case CtrlO_Key, ZeroKey:
				UserInputString = ""
				EnteredPromptStr = ""
				HasEnteredMove = false
				if key == CtrlO_Key {
					return 3
				} else {
					return 0
				}
			}
		}
	}
}

func PostScreen(screen *ncurses.Window) int {
	var key ncurses.Key
	height, width := screen.MaxYX()

	//start windows
	bp_info := WinInfo{H: ((height) - (height / 3)), W: (width), Y: 0, X: 0}
	pp_info := WinInfo{H: ((height) / 4) - 1, W: width, Y: ((height / 4) * 3), X: 0}

	board_post_window, _ := ncurses.NewWindow(bp_info.H, bp_info.W, bp_info.Y, bp_info.X)
	history_window, _ := ncurses.NewWindow(pp_info.H, pp_info.W, pp_info.Y, pp_info.X)

	windows_array := [2]*ncurses.Window{board_post_window, history_window}
	windows_info_arr := [2]WinInfo{bp_info, pp_info}

	DrawPostScreen(screen, key, windows_array, windows_info_arr)
	for {
		select {
		case <-Sigs:
			tRow, tCol, _ := OsTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			DrawPostScreen(screen, key, windows_array, windows_info_arr)
		default: //normal character loop here
			DrawBoardWindow(board_post_window, CurrentGame.Position().String(), false)
			//UpdateInput(prompt_post_window, key)
			for _, win := range windows_array {
				win.NoutRefresh()
			}
			ncurses.Update()
			key = screen.GetChar()
			switch key {
			case CtrlO_Key, OneKey, ZeroKey:
				UserInputString = ""
				EnteredPromptStr = ""
				HasEnteredMove = false
				if key == CtrlO_Key {
					return 3
				} else if key == OneKey {
					return 1
				} else {
					return 0
				}
			}
		}
	}
}

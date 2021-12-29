package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/nate-xyz/goncurses"
)

func local_game_screen(stdscr *goncurses.Window) {
	var key goncurses.Key = 0
	var mouse_event *goncurses.MouseEvent
	_ = mouse_event
	// cursor_x := 0
	// cursor_y := 0
	//stdscr = goncurses.initscr()
	height, width := stdscr.MaxYX()

	//Clear and refresh the screen for a blank canvas
	stdscr.Clear()
	stdscr.Refresh()

	//start windows
	board_window, _ := goncurses.NewWindow(int(math.Floor(float64((height/4)*3))), int(math.Floor(float64(width/2))), 0, 0)
	prompt_window, _ := goncurses.NewWindow(int(math.Floor(float64((height/4)-1))), int(math.Floor(float64(width/2))), int(math.Floor(float64((height/4)*3))), 0)
	info_window, _ := goncurses.NewWindow(int(math.Floor(float64((height / 2)))), int(math.Floor(float64(width/2))), 0, int(math.Floor(float64(width/2))))
	history_window, _ := goncurses.NewWindow(int(math.Floor(float64((height/2)-1))), int(math.Floor(float64(width/2))), int(math.Floor(float64(height/2))), int(math.Floor(float64(width/2))))

	windows_array := [4]*goncurses.Window{board_window, info_window, prompt_window, history_window}

	// Loop where key is the last character pressed
	for key != 15 { // while not quitting (ctrl+o)
		if quit_game {
			break
		}

		// Initialization
		stdscr.Clear()
		board_window.Clear()
		info_window.Clear()
		history_window.Clear()

		//Resize everything if necessary
		if goncurses.IsTermResized(height, width) {
			height, width := stdscr.MaxYX() //get new height and width

			//Resize the terminal and refresh
			goncurses.ResizeTerm(height, width)
			stdscr.Clear()
			stdscr.Refresh()

			var height_float float64 = float64(height)
			var width_float float64 = float64(width)

			//Resize windows based on new dimensions
			board_window.Resize(int(math.Floor(float64((height_float/4)*3))), int(math.Floor(float64(width_float/2))))
			prompt_window.Resize(int(math.Floor(float64((height_float/4)-1))), int(math.Floor(float64(width_float/2))))
			info_window.Resize(int(math.Floor(float64((height_float / 2)))), int(math.Floor(float64(width_float/2))))
			history_window.Resize(int(math.Floor(float64((height_float/2)-1))), int(math.Floor(float64(width_float/2))))

			//move windows to appropriate locations
			board_window.MoveWindow(0, 0)
			prompt_window.MoveWindow(int(math.Floor(float64((height_float/4)*3))), 0)
			info_window.MoveWindow(0, int(math.Floor(float64(width_float/2))))
			history_window.MoveWindow(int(math.Floor(float64(height_float/2))), int(math.Floor(float64(width_float/2))))

			//Clear and refresh all windows
			for _, win := range windows_array {
				// win.Clear()
				win.Refresh()
			}
		}
		//get window dimensions
		height, width := stdscr.MaxYX()
		// board_window_height, board_window_width := board_window.MaxYX()
		// info_window_height, info_window_width := info_window.MaxYX()
		// prompt_window_height, prompt_window_width := prompt_window.MaxYX()
		// history_window_height, history_window_width := history_window.MaxYX()

		//get mouse location
		// cursor_x := math.Min(width-1, math.Max(0, cursor_x))
		// cursor_y := math.Min(height-1, math.Max(0, cursor_y))

		// Declaration of strings
		var board_title string = "board"
		var info_title string = "info"
		var prompt_title string = "prompt"
		var history_title string = "move_history"

		if len(board_title) >= width {
			board_title = "board"[:width-1]
		}
		if len(info_title) >= width {
			info_title = "info"[:width-1]
		}
		if len(prompt_title) >= width {
			prompt_title = "prompt"[:width-1]
		}
		if len(history_title) >= width {
			history_title = "move_history"[:width-1]
		}

		keystr := fmt.Sprintf("Last key pressed: %d\n", key)
		//statusbarstr = "Press 'Ctrl+o' to exit | CHESS-CLI | Pos: {}, {}".format(cursor_x, cursor_y)
		statusbarstr := "Press 'Ctrl+o' to exit | CHESS-CLI"

		statusbarfull := fmt.Sprintf("%s | %s\n", statusbarstr, keystr)

		if len(statusbarfull) >= width {
			statusbarfull = statusbarfull[:width-1]
		}
		//statusbarfull = ""

		if key == 0 {
			keystr := "No key press detected..."
			if len(keystr) >= width {
				keystr = "No key press detected..."[:width-1]
			}
		}

		// Render status bar
		stdscr.AttrOn(goncurses.ColorPair(3))
		stdscr.MovePrint(height-1, 0, statusbarfull)
		padding := fmt.Sprintf("+%s", strings.Repeat(" ", (width-len(statusbarfull)-1)))
		stdscr.MovePrint(height-1, len(statusbarfull), padding)
		stdscr.AttrOff(goncurses.ColorPair(3))

		for _, win := range windows_array {
			win.Box('|', '-')
		}

		//// EXTERNAL FUNCTION CALL !!! //////
		//external function calls
		if !mouse_event_bool {
			update_input(prompt_window, key)
		}

		game_logic(board_window)
		if post_screen_toggle {
			post_screen_toggle = false
			post_screen(stdscr)
			if quit_game {
				break
			}
			welcome_screen(stdscr)
			continue
		}
		display_info(info_window)
		display_history(history_window)
		// board_window_mouse_input(board_window, key, width, height)

		// Turning on attributes for title
		for _, win := range windows_array {
			win.AttrOn(goncurses.ColorPair(2))
			win.AttrOn(goncurses.A_BOLD)
		}

		// Rendering title
		board_window.MovePrint(0, 1, board_title)
		info_window.MovePrint(0, 1, info_title)
		prompt_window.MovePrint(0, 1, prompt_title)
		history_window.MovePrint(0, 1, history_title)

		// Turning off attributes for title
		for _, win := range windows_array {
			win.AttrOff(goncurses.ColorPair(2))
			win.AttrOff(goncurses.A_BOLD)
		}

		// Refresh the screen
		stdscr.Refresh()
		for _, win := range windows_array {
			win.Refresh()
		}
		// Wait for next input
		key = stdscr.GetChar()
		if key == goncurses.KEY_MOUSE {
			mouse_event = goncurses.GetMouse()
			mouse_event_bool = true
			continue
		}
		mouse_event_bool = false

	}
}

// //                      888
// //                       888
// //                       888
// //888  888  888  .d88b.  888  .d8888b .d88b.  88888b.d88b.   .d88b.         .d8888b   .d8888b 888d888 .d88b.   .d88b.  88888b.
// //888  888  888 d8P  Y8b 888 d88P"   d88""88b 888 "888 "88b d8P  Y8b        88K      d88P"    888P"  d8P  Y8b d8P  Y8b 888 "88b
// //888  888  888 88888888 888 888     888  888 888  888  888 88888888        "Y8888b. 888      888    88888888 88888888 888  888
// //Y88b 888 d88P Y8b.     888 Y88b.   Y88..88P 888  888  888 Y8b.                 X88 Y88b.    888    Y8b.     Y8b.     888  888
// // "Y8888888P"   "Y8888  888  "Y8888P "Y88P"  888  888  888  "Y8888 88888888 88888P'  "Y8888P 888     "Y8888   "Y8888  888  888

func welcome_screen(screen *goncurses.Window) {
	height, width := screen.MaxYX()
	var key goncurses.Key

	prompt_welcome_window, _ := goncurses.NewWindow(((height)/4)-1, width, ((height / 4) * 3), 0)

	for key != 12 { // while not quitting
		if key == 15 {
			quit_game = true
			break
		}
		screen.Clear()

		// Declaration of strings
		title := "chess-tui"
		subtitle := "play locally with a friend, against stockfish, or online with lichess!"
		keystr := fmt.Sprintf("Last key pressed: %v", key)
		statusbarstr := "Press 'Ctrl-l' to skip to local | Press 'Ctrl-o' to quit"

		// Centering calculations
		start_x_title := int((width / 2) - (len(title) / 2) - len(title)%2)
		start_x_subtitle := int((width / 2) - (len(subtitle) / 2) - len(subtitle)%2)
		start_x_keystr := int((width / 2) - (len(keystr) / 2) - len(keystr)%2)
		start_y := int((height / 2) - 2)

		// Rendering some text
		whstr := fmt.Sprintf("Width: %v, Height: %v", width, height)
		screen.MovePrint(0, 0, whstr, goncurses.ColorPair(1))

		// Render status bar
		screen.AttrOn(goncurses.ColorPair(3))
		screen.MovePrint(height-1, 0, statusbarstr)
		padding := fmt.Sprintf("+%s", strings.Repeat(" ", (width-len(statusbarstr)-1)))
		screen.MovePrint(height-1, len(statusbarstr), padding)
		screen.AttrOff(goncurses.ColorPair(3))

		// Turning on attributes for title
		screen.AttrOn(goncurses.ColorPair(2))
		screen.AttrOn(goncurses.A_BOLD)

		// Rendering title
		screen.MovePrint(start_y, start_x_title, title)

		// Turning off attributes for title
		screen.AttrOff(goncurses.ColorPair(2))
		screen.AttrOff(goncurses.A_BOLD)

		// Print rest of text
		screen.MovePrint(start_y+1, start_x_subtitle, subtitle)
		screen.MovePrint(start_y+3, (width/2)-2, "----")
		screen.MovePrint(start_y+5, start_x_keystr, keystr)

		update_input(prompt_welcome_window, key)

		prompt_welcome_window.Box('|', '-')
		screen.Refresh()
		prompt_welcome_window.Refresh()
		key = screen.GetChar()
	}
	//reset global strings that may have been set in the prompt window
	user_input_string = ""
	inputted_str = ""
	entered_move = false
}

// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// //                                 .
// //                                .o8
// // oo.ooooo.   .ooooo.   .oooo.o .o888oo              .oooo.o  .ooooo.  oooo d8b  .ooooo.   .ooooo.  ooo. .oo.
// //  888' `88b d88' `88b d88(  "8   888               d88(  "8 d88' `"Y8 `888""8P d88' `88b d88' `88b `888P"Y88b
// //  888   888 888   888 `"Y88b.    888               `"Y88b.  888        888     888ooo888 888ooo888  888   888
// //  888   888 888   888 o.  )88b   888 .             o.  )88b 888   .o8  888     888    .o 888    .o  888   888
// //  888bod8P' `Y8bod8P' 8""888P'   "888" ooooooooooo 8""888P' `Y8bod8P' d888b    `Y8bod8P' `Y8bod8P' o888o o888o
// //  888
// // o888o
// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func post_screen(screen1 *goncurses.Window) {
	screen1.Clear()
	screen1.Refresh()

	height, width := screen1.MaxYX()
	var key goncurses.Key

	prompt_post_window, _ := goncurses.NewWindow(((height)/4)-1, width, ((height / 4) * 3), 0)
	board_post_window, _ := goncurses.NewWindow(((height) - (height / 3)), (width), 0, 0)

	for i, j := 0, len(history_arr)-1; i < j; i, j = i+1, j-1 {
		history_arr[i], history_arr[j] = history_arr[j], history_arr[i]
	}

	for key != 12 { // while not quitting ctrl-l
		if key == 15 { //ctrl-o
			quit_game = true
			break
		}
		screen1.Clear()

		// Declaration of strings
		title := "Game has ended."
		final_position_str := "Final position: "
		final_history_str := fmt.Sprintf("Last key pressed: %v", key)
		outcome_str := fmt.Sprint("outcome: ", game.Outcome().String(), game.Method().String())
		statusbarstr := "Press 'Ctrl-l' to play again | Press 'Ctrl-o' to quit"

		// Centering calculations
		start_x_title := int((width / 2) - (len(title) / 2) - len(title)%2)
		start_x_final_position_str := int((width / 2) - (len(final_position_str) / 2) - len(final_position_str)%2)
		start_x_final_history_str := int((width / 2) - (len(final_history_str) / 2) - len(final_history_str)%2)
		start_y := int((height / 2) - 2)

		// Render status bar
		screen1.AttrOn(goncurses.ColorPair(3))
		screen1.MovePrint(height-1, 0, statusbarstr)
		padding := fmt.Sprintf("+%s", strings.Repeat(" ", (width-len(statusbarstr)-1)))
		screen1.MovePrint(height-1, len(statusbarstr), padding)
		screen1.AttrOff(goncurses.ColorPair(3))

		// Turning on attributes for title
		board_post_window.AttrOn(goncurses.ColorPair(2))
		board_post_window.AttrOn(goncurses.A_BOLD)

		// Rendering title
		board_post_window.MovePrint(start_y, start_x_title, title)

		// Turning off attributes for title
		screen1.AttrOff(goncurses.ColorPair(2))
		screen1.AttrOff(goncurses.A_BOLD)

		// Print rest of text
		board_post_window.MovePrint(start_y+1, start_x_final_position_str, final_position_str)
		//history = " -> ".join([str(elem) for elem in [ele for ele in reversed(history_arr)][1:]])[:width-2]
		//revese history array

		history := fmt.Sprintf(strings.Join(history_arr[:], " -> "))
		board_post_window.MovePrint(start_y+3, ((width / 2) - (len(history) / 2)), history)
		board_post_window.MovePrint(start_y+5, start_x_final_history_str, outcome_str)
		board_post_window.MovePrint(start_y+6, start_x_final_history_str, final_history_str)

		draw_board(board_post_window)

		update_input(prompt_post_window, key)

		prompt_post_window.Box('|', '-')
		board_post_window.Box('|', '-')
		screen1.Refresh()
		prompt_post_window.Refresh()
		board_post_window.Refresh()
		key = screen1.GetChar()
	}
	//reset global strings that may have been set in the prompt window
	user_input_string = ""
	inputted_str = ""
	entered_move = false
}

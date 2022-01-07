package main

import (
	"fmt"
	"strings"
	"time"

	ncurses "github.com/nate-xyz/goncurses_"
)

func draw_welcome_screen(screen *ncurses.Window, key ncurses.Key, windows_array [1]*ncurses.Window, windows_info_arr [1]windowSizePos, op []string) {
	screen.Clear()
	height, width := screen.MaxYX()

	//update window dimensions
	max_len := getMaxLenStr(op) + 6

	//options window
	windows_info_arr[0] = windowSizePos{len(op) + 2, max_len, (height / 2) + 2, (width / 2) - (max_len / 2) - max_len%2}

	//Clear, refresh, update all windows
	for i, win := range windows_array {
		win.Clear()
		info := windows_info_arr[i]
		win.Resize(info.h, info.w)     //Resize windows based on new dimensions
		win.MoveWindow(info.y, info.x) //move windows to appropriate locations
		win.NoutRefresh()
	}

	// Declaration of strings
	title := "chess-cli"
	subtitle := "play locally with a friend, against stockfish, or online with lichess!"
	additional_info := []string{"play locally", "play online", "play stockfish"}
	//keystr := fmt.Sprintf("Last key pressed: %v", key)
	statusbarstr := "WELCOME TO CHESS-CLI ! | Press 'Ctrl-o' to quit"

	// Centering calculations
	start_x_title := int((width / 2) - (len(title) / 2) - len(title)%2)
	if start_x_title < 1 {
		start_x_title = 1
	}
	start_x_subtitle := int((width / 2) - (len(subtitle) / 2) - len(subtitle)%2)
	if start_x_subtitle < 1 {
		start_x_subtitle = 1
	}
	start_y := int((height / 2) - 2)
	if start_y < 1 {
		start_y = 1
	}

	// Rendering some text
	screen.MovePrint(0, 0, fmt.Sprintf("Width: %d, Height: %d\n", width, height))

	// Render status bar
	screen.AttrOn(ncurses.ColorPair(3))
	screen.MovePrint(height-1, 0, statusbarstr)
	var padding string
	if (width - len(statusbarstr) - 1) > 0 {
		padding = fmt.Sprintf("%s", strings.Repeat(" ", (width-len(statusbarstr)-1)))
	}
	screen.MovePrint(height-1, len(statusbarstr), padding)
	screen.AttrOff(ncurses.ColorPair(3))

	// Turning on attributes for title
	screen.AttrOn(ncurses.ColorPair(15))
	screen.AttrOn(ncurses.A_BOLD)
	screen.AttrOn(ncurses.A_UNDERLINE)

	// Rendering title
	screen.MovePrint(start_y, start_x_title, title)

	// Turning off attributes for title
	screen.AttrOff(ncurses.A_UNDERLINE)
	screen.AttrOff(ncurses.ColorPair(15))
	screen.AttrOff(ncurses.A_BOLD)

	// Print rest of text
	screen.AttrOn(ncurses.A_DIM)
	screen.AttrOn(ncurses.ColorPair(17))
	screen.MovePrint(start_y+1, start_x_subtitle, subtitle)
	screen.AttrOff(ncurses.ColorPair(17))
	screen.AttrOff(ncurses.A_DIM)

	screen.MovePrint(start_y+3, (width/2)-2, "----")
	for i, str := range additional_info {
		screen.MovePrint(start_y+4+i, (width/2)-(len(str)/2), str)
	}
	screen.MovePrint(start_y+7, (width/2)-2, "----")

	quote := GetRandomQuote()
	x_quote := width/2 - (len(quote) / 2)
	if x_quote < 0 {
		x_quote = 0
	}
	screen.AttrOn(ncurses.ColorPair(1))
	screen.AttrOn(ncurses.A_DIM)
	screen.MovePrint(start_y+11, x_quote, quote)
	screen.AttrOff(ncurses.A_DIM)
	screen.AttrOff(ncurses.ColorPair(1))
	//prompt_welcome_window.Box('|', '-')
	screen.NoutRefresh()
	for _, win := range windows_array {
		win.Box('|', '-')
		win.NoutRefresh()
	}
	ncurses.Update()
}

func draw_local_game_screen(stdscr *ncurses.Window, key ncurses.Key, windows_array [4]*ncurses.Window, windows_info_arr [4]windowSizePos) {
	//Clear and refresh the screen for a blank canvas
	stdscr.Clear()
	height, width := stdscr.MaxYX()

	//update window dimensions
	// windows_info_arr[0] = windowSizePos{(height / 4) * 3, width / 2, 0, 0}
	// windows_info_arr[1] = windowSizePos{height / 2, width / 2, 0, width / 2}
	// windows_info_arr[2] = windowSizePos{(height / 4) - 1, width / 2, (height / 4) * 3, 0}
	// windows_info_arr[3] = windowSizePos{(height / 2) - 1, width / 2, height / 2, width / 2}

	//h, w, y, x
	windows_info_arr[0] = windowSizePos{height / 2, width, 0, 0}                                           //bw
	windows_info_arr[1] = windowSizePos{(height / 2) - 1, width / 2, (height / 2), 0}                      //iw
	windows_info_arr[2] = windowSizePos{(height / 4) - 1, width / 2, height / 2, width / 2}                //pw
	windows_info_arr[3] = windowSizePos{(height / 4), width / 2, int(float64(height)*0.75) - 1, width / 2} //hw

	//Clear, refresh, update all windows
	for i, win := range windows_array {
		win.Clear()
		info := windows_info_arr[i]
		win.Resize(info.h, info.w)     //Resize windows based on new dimensions
		win.MoveWindow(info.y, info.x) //move windows to appropriate locations
		win.NoutRefresh()
	}

	//get mouse location
	// cursor_x := math.Min(width-1, math.Max(0, cursor_x))
	// cursor_y := math.Min(height-1, math.Max(0, cursor_y))

	// Declaration of strings
	board_title := "board"
	info_title := "info"
	prompt_title := "prompt"
	history_title := "move_history"
	title_array := []string{board_title, info_title, prompt_title, history_title}
	//keystr := fmt.Sprintf("Last key pressed: %v", key)
	statusbarstr := "CHESS-CLI | Press '0' to return to main | Press 'Ctrl+o' to exit"
	// if key == zero_key {
	// 	keystr = "No key press detected..."
	// }
	statusbarfull := fmt.Sprintf("%s", statusbarstr)

	// Turning on attributes for title
	for i, win := range windows_array {
		win.Box('|', '-')
		// Rendering title
		win.AttrOn(ncurses.ColorPair(2))
		win.AttrOn(ncurses.A_BOLD)
		win.MovePrint(0, 1, title_array[i])
		win.AttrOff(ncurses.ColorPair(2))
		win.AttrOff(ncurses.A_BOLD)
	}

	// Render status bar
	stdscr.AttrOn(ncurses.ColorPair(3))
	stdscr.MovePrint(height-1, 0, statusbarfull)
	var padding string
	if (width - len(statusbarstr) - 1) > 0 {
		padding = fmt.Sprintf("%s", strings.Repeat(" ", (width-len(statusbarstr)-1)))
	}
	stdscr.MovePrint(height-1, len(statusbarfull), padding)
	stdscr.AttrOff(ncurses.ColorPair(3))

	// Refresh the screen
	stdscr.NoutRefresh()
	for _, win := range windows_array {
		win.NoutRefresh()
	}
	ncurses.Update()

	// // Wait for next input
	// key = stdscr.GetChar()
	// if key == ncurses.KEY_MOUSE {
	// 	mouse_event = ncurses.GetMouse()
	// 	mouse_event_bool = true
	// 	continue
	// }
}

func draw_post_screen(screen1 *ncurses.Window, key ncurses.Key, windows_array [2]*ncurses.Window, windows_info_arr [2]windowSizePos) {
	height, width := screen1.MaxYX()
	screen1.Clear()

	//update window dimensions
	windows_info_arr[0] = windowSizePos{height / 2, width, 0, 0}
	windows_info_arr[1] = windowSizePos{height / 2, width, height / 2, 0}

	//Clear, refresh, update all windows
	for i, win := range windows_array {
		win.Clear()
		info := windows_info_arr[i]
		win.Resize(info.h, info.w)     //Resize windows based on new dimensions
		win.MoveWindow(info.y, info.x) //move windows to appropriate locations
		win.NoutRefresh()
	}

	//revese history array
	for i, j := 0, len(history_arr)-1; i < j; i, j = i+1, j-1 {
		history_arr[i], history_arr[j] = history_arr[j], history_arr[i]
	}

	// Declaration of strings
	title := "Game has ended."
	board_title := "board"
	history_title := "outcome"
	title_array := []string{board_title, history_title}
	final_position_str := "Final position: "
	//final_history_str := fmt.Sprintf("Last key pressed: %v", key)
	outcome_str := fmt.Sprintf("outcome: %s, %s\n", game.Outcome().String(), game.Method().String())
	statusbarstr := "CHESS-CLI | Press '0' to return to main | Press '1' to play again | Press 'Ctrl-o' to quit"

	// Centering calculations
	width = windows_info_arr[1].w
	height = windows_info_arr[1].h
	start_x_title := int((width / 2) - (len(title) / 2) - len(title)%2)
	if start_x_title < 1 {
		start_x_title = 1
	}
	start_x_final_position_str := int((width / 2) - (len(final_position_str) / 2) - len(final_position_str)%2)
	if start_x_final_position_str < 1 {
		start_x_final_position_str = 1
	}

	// start_x_final_history_str := int((width / 2) - (len(final_history_str) / 2) - len(final_history_str)%2)
	// if start_x_final_history_str < 1 {
	// 	start_x_final_history_str = 1
	// }
	start_y := int((height / 2) - 2)
	if start_y < 1 {
		start_y = 1
	}
	height, width = screen1.MaxYX()

	// Render status bar
	screen1.AttrOn(ncurses.ColorPair(3))
	screen1.MovePrint(height-1, 0, statusbarstr)
	var padding string
	if (width - len(statusbarstr) - 1) > 0 {
		padding = fmt.Sprintf("%s", strings.Repeat(" ", (width-len(statusbarstr)-1)))
	}
	screen1.MovePrint(height-1, len(statusbarstr), padding)
	screen1.AttrOff(ncurses.ColorPair(3))

	// Turning on attributes for title
	for i, win := range windows_array {
		win.Box('|', '-')
		// Rendering title
		win.AttrOn(ncurses.ColorPair(2))
		win.AttrOn(ncurses.A_BOLD)
		win.MovePrint(0, 1, title_array[i])
		if i == 0 {
			win.MovePrint(start_y, start_x_title, title)
		}
		win.AttrOff(ncurses.ColorPair(2))
		win.AttrOff(ncurses.A_BOLD)

	}

	// Print rest of text
	windows_array[1].MovePrint(start_y+1, start_x_final_position_str, final_position_str)
	history := fmt.Sprintf(strings.Join(history_arr[:], " -> "))
	windows_array[1].MovePrint(start_y+3, ((width / 2) - (len(history) / 2)), history)
	windows_array[1].MovePrint(start_y+5, ((width / 2) - (len(outcome_str) / 2)), outcome_str)
	//windows_array[1].MovePrint(start_y+6, start_x_final_history_str, final_history_str)

	// Refresh the screen
	screen1.NoutRefresh()
	for _, win := range windows_array {
		win.NoutRefresh()
	}
	ncurses.Update()

}

func draw_options_input(window *ncurses.Window, options []string, selected_index int) {
	_, width := window.MaxYX()

	piece := "♟︎ "

	//draw standout for currently selected option
	for i, str := range options {
		if i == selected_index {
			window.AttrOn(ncurses.ColorPair(3))
			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
			window.AttrOff(ncurses.ColorPair(3))
			window.AttrOn(ncurses.A_DIM)
			window.AttrOn(ncurses.A_BLINK)
			window.MovePrint(i+1, 1, piece)
			window.MovePrint(i+1, width-3, piece)
			window.AttrOff(ncurses.A_BLINK)
			window.AttrOff(ncurses.A_DIM)
		} else {
			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
			window.MovePrint(i+1, 1, " ")
			window.MovePrint(i+1, width-3, " ")
		}
	}
	window.Refresh()
}

func loading_screen(screen *ncurses.Window, message string) {
	height, width := screen.MaxYX()
	screen.MovePrint(height/2, width/2-len(message)/2, message)
	// dt := time.Now().Unix() % 10
	// screen.MovePrint((height/2)+1, width/2, fmt.Sprintf("%v", loader[dt]))
	dt := time.Now().Unix() % 8
	screen.MovePrint((height/2)+1, width/2, fmt.Sprintf("%v", knight_loader[dt]))
	screen.Refresh()
}

// func draw_slider_input(window *ncurses.Window, titles []string, intervals [][]interface{}, selected_index int) {

// 	_, width := window.MaxYX()
// 	nav_options := []string{"submit", "back"}

// 	//draw standout for currently selected option
// 	for i, str := range titles {
// 		if i == selected_index {
// 			window.AttrOn(ncurses.ColorPair(3))
// 			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
// 			window.AttrOff(ncurses.ColorPair(3))
// 		} else {
// 			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
// 			window.MovePrint(i+1, 1, " ")
// 			window.MovePrint(i+1, width-3, " ")
// 		}
// 		arr := intervals[i]
// 		tic_line := fmt.Sprintf("%s", strings.Repeat("-", (width-4)))
// 		tic_loc := int((float64(len(tic_line)) / float64(len(arr))) * float64(tic_index[0]))
// 		window.MovePrint((height/3)+4, 2, tic_line) //print the tic_line
// 		window.MovePrint((height/3)+4, tic_loc, "|")
// 	}

// 	for i, str := range nav_options {
// 		if i == selected_index {
// 			window.AttrOn(ncurses.ColorPair(3))
// 			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
// 			window.AttrOff(ncurses.ColorPair(3))
// 		} else {
// 			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
// 			window.MovePrint(i+1, 1, " ")
// 			window.MovePrint(i+1, width-3, " ")
// 		}
// 	}
// 	window.Refresh()

// }

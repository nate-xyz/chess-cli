package main

import (
	"fmt"
	"strings"

	ncurses "github.com/nate-xyz/goncurses"
)

func draw_welcome_screen(screen *ncurses.Window, key ncurses.Key, windows_array [1]*ncurses.Window, windows_info_arr [1]windowSizePos, op []string) {
	screen.Clear()
	height, width := screen.MaxYX()

	//update window dimensions
	max_len := 0
	for _, str := range op {
		if max_len < len(str) {
			max_len = len(str)
		}

	}
	windows_info_arr[0] = windowSizePos{len(op) + 2, max_len + 2, (height / 2) + 2, (width / 2) - ((max_len + 2) / 2)}

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
	screen.MovePrint(0, 0, fmt.Sprintf("Width: %d, Height: %d\n", width, height), ncurses.ColorPair(1))

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
	screen.AttrOn(ncurses.ColorPair(2))
	screen.AttrOn(ncurses.A_BOLD)
	screen.AttrOn(ncurses.A_UNDERLINE)

	// Rendering title
	screen.MovePrint(start_y, start_x_title, title)

	// Turning off attributes for title
	screen.AttrOff(ncurses.A_UNDERLINE)
	screen.AttrOff(ncurses.ColorPair(2))
	screen.AttrOff(ncurses.A_BOLD)

	// Print rest of text
	screen.AttrOn(ncurses.A_DIM)
	screen.MovePrint(start_y+1, start_x_subtitle, subtitle)
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
	screen.AttrOn(ncurses.A_DIM)
	screen.AttrOff(ncurses.ColorPair(1))
	//prompt_welcome_window.Box('|', '-')
	screen.NoutRefresh()
	for _, win := range windows_array {
		win.Box('|', '-')
		win.NoutRefresh()
	}
	ncurses.Update()
}

func draw_lichess_welcome(screen *ncurses.Window, key ncurses.Key, windows_array [1]*ncurses.Window, windows_info_arr [1]windowSizePos, op []string) {
	screen.Clear()
	height, width := screen.MaxYX()

	//update window dimensions
	max_len := 0
	for _, str := range op {
		if max_len < len(str) {
			max_len = len(str)
		}

	}
	windows_info_arr[0] = windowSizePos{len(op) + 2, max_len + 2, (height / 2) + 2, (width / 2) - ((max_len + 2) / 2)}

	//Clear, refresh, update all windows
	for i, win := range windows_array {
		win.Clear()
		info := windows_info_arr[i]
		win.Resize(info.h, info.w)     //Resize windows based on new dimensions
		win.MoveWindow(info.y, info.x) //move windows to appropriate locations
		win.NoutRefresh()
	}

	// Declaration of strings
	title := "chess-cli: lichess client"
	var subtitle string
	var additional_info []string
	if UserInfo.ApiToken == "" {
		subtitle = fmt.Sprintf("not logged into lichess.")
		additional_info = []string{"please login through your browser.", "press (ctrl-l) to login through lichess.org"}

	} else {
		subtitle = fmt.Sprintf("logged in as: %s", Username)
		additional_info = []string{"<<Press 0 to return to welcome screen>>", "<<Press 2 to view / create challenges>>", "<<Press 3 to view / join ongoing games>>", "etc"}
	}
	keystr := fmt.Sprintf("Last key pressed: %v", key)

	var statusbarstr string
	// if UserInfo.ApiToken == "" {
	// 	statusbarstr = fmt.Sprintf("LICHESS CLIENT | Press 'Ctrl-l' to login | Press 'Ctrl-o' to quit")
	// } else {
	// 	statusbarstr = fmt.Sprintf("LICHESS CLIENT | Press 'Ctrl-o' to quit")
	// }
	statusbarstr = fmt.Sprintf("LICHESS CLIENT | Press '0' to return to main | Press 'Ctrl-o' to quit")

	// Centering calculations
	start_x_title := int((width / 2) - (len(title) / 2) - len(title)%2)
	start_x_subtitle := int((width / 2) - (len(subtitle) / 2) - len(subtitle)%2)
	start_x_keystr := int((width / 2) - (len(keystr) / 2) - len(keystr)%2)
	start_y := int((height / 2) - 2)

	// Rendering some text
	whstr := fmt.Sprintf("Width: %d, Height: %d\n", width, height)
	screen.MovePrint(0, 0, whstr, ncurses.ColorPair(1))

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
	screen.AttrOn(ncurses.ColorPair(2))
	screen.AttrOn(ncurses.A_BOLD)

	// Rendering title
	screen.MovePrint(start_y, start_x_title, title)

	// Turning off attributes for title
	screen.AttrOff(ncurses.ColorPair(2))
	screen.AttrOff(ncurses.A_BOLD)

	// Print rest of text
	screen.MovePrint(start_y+1, start_x_subtitle, subtitle)
	screen.MovePrint(start_y+3, (width/2)-2, "----")
	for i, str := range additional_info {
		screen.MovePrint(start_y+4+i, (width/2)-(len(str)/2), str)
	}
	screen.MovePrint(start_y+7, (width/2)-2, "----")
	screen.MovePrint(start_y+9, start_x_keystr, keystr)
	screen.NoutRefresh()
	for _, win := range windows_array {
		win.Box('|', '-')
		win.NoutRefresh()
	}
	ncurses.Update()
}

func draw_lichess_challenges(screen *ncurses.Window, key ncurses.Key, windows_array [3]*ncurses.Window, windows_info_arr [3]windowSizePos, op []string) {
	screen.Clear()
	height, width := screen.MaxYX()

	//update window dimensions
	max_len := 0
	for _, str := range op {
		if max_len < len(str) {
			max_len = len(str)
		}
	}
	windows_info_arr[0] = windowSizePos{len(op) + 2, max_len + 2, 2, (width / 2) - ((max_len + 2) / 2)}
	windows_info_arr[1] = windowSizePos{(height / 4) * 3, width / 2, (height / 4), 0}
	windows_info_arr[2] = windowSizePos{(height / 4) * 3, width / 2, (height / 4), width / 2}

	//Clear, refresh, update all windows
	for i, win := range windows_array {
		win.Clear()
		info := windows_info_arr[i]
		win.Resize(info.h, info.w)     //Resize windows based on new dimensions
		win.MoveWindow(info.y, info.x) //move windows to appropriate locations
		win.NoutRefresh()
	}
	// Declaration of strings
	title := "lichess challenges"
	options_title := "options"
	incoming_title := "incoming challenges"
	outgoing_title := "outgoing challenges"

	title_array := []string{options_title, incoming_title, outgoing_title}

	var subtitle string
	//var additional_info []string
	var statusbarstr string = fmt.Sprintf("LICHESS CHALLENGES | Press '0' to return to main | Press '1' to return to lichess main | Press 'Ctrl-o' to quit")

	// Centering calculations
	start_x_title := int((width / 2) - (len(title) / 2) - len(title)%2)
	start_x_subtitle := int((width / 2) - (len(subtitle) / 2) - len(subtitle)%2)
	_ = start_x_subtitle
	//start_y := int((height / 2) - 2)
	start_y := 1

	// Rendering some text
	whstr := fmt.Sprintf("Width: %d, Height: %d\n", width, height)
	screen.MovePrint(0, 0, whstr, ncurses.ColorPair(1))

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
	for i, win := range windows_array {
		win.Box('|', '-')
		// Rendering title
		win.AttrOn(ncurses.ColorPair(2))
		win.AttrOn(ncurses.A_BOLD)
		win.MovePrint(0, 1, title_array[i])
		win.AttrOff(ncurses.ColorPair(2))
		win.AttrOff(ncurses.A_BOLD)
	}

	draw_challenge_windows(windows_array[1], windows_array[2])

	// Turning on attributes for main title
	screen.AttrOn(ncurses.ColorPair(2))
	screen.AttrOn(ncurses.A_BOLD)
	screen.MovePrint(start_y, start_x_title, title) // Rendering title
	screen.AttrOff(ncurses.ColorPair(2))            // Turning off attributes for title
	screen.AttrOff(ncurses.A_BOLD)

	// Print rest of text
	// screen.MovePrint(start_y+1, start_x_subtitle, subtitle)
	// for i, friend := range allFriends {
	// 	screen.MovePrint(start_y+3+i, (width/2)-2, "Friend: "+(friend))
	// }

	// for i, str := range additional_info {
	// 	screen.MovePrint(start_y+4+i, (width/2)-(len(str)/2), str)
	// }
	//screen.MovePrint(start_y+7, (width/2)-2, "----")
	screen.NoutRefresh()
	for _, win := range windows_array {
		win.NoutRefresh()
	}
	ncurses.Update()
}

func draw_challenge_windows(inc *ncurses.Window, out *ncurses.Window) {
	for i, challenge := range IncomingChallenges {
		inc.MovePrint(i+1, 1, fmt.Sprintf("%s -> %s", challenge.Challenger.Id, challenge.DestUser.Id))

	}
	for i, challenge := range OutgoingChallenges {
		out.MovePrint(i+1, 1, fmt.Sprintf("%s -> %s", challenge.Challenger.Id, challenge.DestUser.Id))
	}

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
	windows_info_arr[0] = windowSizePos{height / 2, width, 0, 0}                                         //bw
	windows_info_arr[1] = windowSizePos{(height / 2) - 1, width / 2, height / 2, 0}                      //hw
	windows_info_arr[2] = windowSizePos{(height / 4) - 1, width / 2, height / 2, width / 2}              //pw
	windows_info_arr[3] = windowSizePos{(height / 4), width / 2, int(float64(height) * 0.75), width / 2} //iw

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
	keystr := fmt.Sprintf("Last key pressed: %v", key)
	statusbarstr := "CHESS-CLI | Press '0' to return to main | Press 'Ctrl+o' to exit"
	if key == 0 {
		keystr = "No key press detected..."
	}
	statusbarfull := fmt.Sprintf("%s | %s", statusbarstr, keystr)

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
	statusbarstr := "Press '0' to return to main | Press '1' to play again | Press 'Ctrl-o' to quit"

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

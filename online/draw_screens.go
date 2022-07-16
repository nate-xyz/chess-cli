package online

import (
	"fmt"
	"strings"

	. "github.com/nate-xyz/chess-cli/shared"
	ncurses "github.com/nate-xyz/goncurses_"
)

func DrawLichessWelcome(screen *ncurses.Window, key ncurses.Key, windows_array [1]*ncurses.Window, windows_info_arr [1]WinInfo, op []string) {
	screen.Clear()
	height, width := screen.MaxYX()

	//update window dimensions
	max_len := GetMaxLenStr(op) + 6
	windows_info_arr[0] = WinInfo{len(op) + 2, max_len, (height / 2) + 2, (width / 2) - (max_len / 2) - max_len%2}

	//Clear, refresh, update all windows
	for i, win := range windows_array {
		win.Clear()
		info := windows_info_arr[i]
		win.Resize(info.H, info.W)     //Resize windows based on new dimensions
		win.MoveWindow(info.Y, info.X) //move windows to appropriate locations
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
		subtitle = fmt.Sprintf("logged in as: %s, %s", Username, UserEmail)
		additional_info = []string{}
	}
	//keystr := fmt.Sprintf("Last key pressed: %v", key)

	var statusbarstr string
	// if UserInfo.ApiToken == "" {
	// 	statusbarstr = fmt.Sprintf("LICHESS CLIENT | Press 'Ctrl-l' to login | Press 'Ctrl-o' to quit")
	// } else {
	// 	statusbarstr = fmt.Sprintf("LICHESS CLIENT | Press 'Ctrl-o' to quit")
	// }
	statusbarstr = fmt.Sprintf("CHESS-CLI | LICHESS CLIENT | Press '0' to return to main | Press 'Ctrl-o' to quit")

	// Centering calculations
	start_x_title := int((width / 2) - (len(title) / 2) - len(title)%2)
	start_x_subtitle := int((width / 2) - (len(subtitle) / 2) - len(subtitle)%2)
	//start_x_keystr := int((width / 2) - (len(keystr) / 2) - len(keystr)%2)
	start_y := int((height / 2) - 2)

	//background
	screen.AttrOn(ncurses.A_DIM)
	screen.MovePrint(0, 0, LichessBg)
	screen.AttrOff(ncurses.A_DIM)

	// Rendering some text
	screen.MovePrint(0, 0, fmt.Sprintf("Width: %d, Height: %d\n", width, height))

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
	//screen.MovePrint(start_y+9, start_x_keystr, keystr)

	// Render status bar
	screen.AttrOn(ncurses.ColorPair(3))
	screen.MovePrint(height-1, 0, statusbarstr)
	var padding string
	if (width - len(statusbarstr) - 1) > 0 {
		padding = fmt.Sprintf("%s", strings.Repeat(" ", (width-len(statusbarstr)-1)))
	}
	screen.MovePrint(height-1, len(statusbarstr), padding)
	screen.AttrOff(ncurses.ColorPair(3))

	screen.NoutRefresh()
	for _, win := range windows_array {
		win.Box('|', '-')
		win.NoutRefresh()
	}
	ncurses.Update()
}

func DrawLichessChallenges(screen *ncurses.Window, key ncurses.Key, windows_array [3]*ncurses.Window, windows_info_arr [3]WinInfo, op []string) {
	screen.Clear()
	height, width := screen.MaxYX()

	//Clear, refresh, update all windows
	for i, win := range windows_array {
		win.Clear()
		info := windows_info_arr[i]
		win.Resize(info.H, info.W)     //Resize windows based on new dimensions
		win.MoveWindow(info.Y, info.X) //move windows to appropriate locations
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
	var statusbarstr string = fmt.Sprintf("CHESS-CLI | LICHESS CHALLENGES | Press '0' to return to main | Press '1' to return to lichess main | Press 'Ctrl-o' to quit")

	// Centering calculations
	start_x_title := int((width / 2) - (len(title) / 2) - len(title)%2)
	start_x_subtitle := int((width / 2) - (len(subtitle) / 2) - len(subtitle)%2)
	_ = start_x_subtitle
	//start_y := int((height / 2) - 2)
	start_y := 1

	//background
	screen.AttrOn(ncurses.A_DIM)
	screen.MovePrint(0, 0, LichessBg)
	screen.AttrOff(ncurses.A_DIM)

	// Rendering some text
	whstr := fmt.Sprintf("Width: %d, Height: %d\n", width, height)
	screen.MovePrint(0, 0, whstr)

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

	DrawChallengeWindows(windows_array[1], windows_array[2])

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

func DrawChallengeWindows(inc *ncurses.Window, out *ncurses.Window) {
	for i, challenge := range IncomingChallenges {
		inc.MovePrint(i+1, 1, fmt.Sprintf("%s -> %s", challenge.Challenger.Id, challenge.DestUser.Id))

	}
	for i, challenge := range OutgoingChallenges {
		out.MovePrint(i+1, 1, fmt.Sprintf("%s -> %s", challenge.Challenger.Id, challenge.DestUser.Id))
	}

}

func DrawCreateGame(screen *ncurses.Window, op []string, sel []string, title string, win *ncurses.Window, info WinInfo) {
	screen.Clear()
	height, width := screen.MaxYX()
	y := height / 4

	// Declaration of strings

	//title_array := []string{"options", "variants", "time options", "time interval", "rated/casual", "choose color", "select friend to challenge"}
	//title_array := []string{"options", "variants", "time options", "rated/casual", "choose color", "select friend to challenge"}

	var statusbarstr string = fmt.Sprintf("CHESS CLI | CREATE A LICHESS GAME | Press 'Ctrl-o' to quit")

	//background
	// screen.AttrOn(ncurses.A_DIM)
	// screen.MovePrint(0, 0, LichessBg)
	// screen.AttrOff(ncurses.A_DIM)

	// size info
	screen.MovePrint(0, 0, fmt.Sprintf("Width: %d, Height: %d\n", width, height))

	// print status bar
	screen.AttrOn(ncurses.ColorPair(3))
	screen.MovePrint(height-1, 0, statusbarstr)
	var padding string
	if (width - len(statusbarstr) - 1) > 0 {
		padding = fmt.Sprintf("%s", strings.Repeat(" ", (width-len(statusbarstr)-1)))
	}
	screen.MovePrint(height-1, len(statusbarstr), padding)
	screen.AttrOff(ncurses.ColorPair(3))

	// main title
	main_title := "create lichess game"
	screen.AttrOn(ncurses.ColorPair(2))
	screen.AttrOn(ncurses.A_BOLD)
	screen.AttrOn(ncurses.A_UNDERLINE)
	screen.MovePrint(y, ((width / 2) - (len(main_title) / 2) - len(main_title)%2), main_title)
	screen.AttrOff(ncurses.A_UNDERLINE)
	screen.AttrOff(ncurses.A_BOLD)
	screen.AttrOff(ncurses.ColorPair(2))
	y++

	sep := "----"
	screen.MovePrint(y, ((width / 2) - (len(sep) / 2) - len(sep)%2), sep)
	y++

	if len(sel) > 0 {
		selections := fmt.Sprintf(strings.Join(sel[:], " -> "))
		screen.MovePrint(y, ((width / 2) - (len(selections) / 2) - len(selections)%2), selections)
		y++

		screen.MovePrint(y, ((width / 2) - (len(sep) / 2) - len(sep)%2), sep)
		y++
	}
	op_wid := GetMaxLenStr(append(op, title)) + 6

	//determin option screen size based on what window you're on. different for slider option vs normal menu option
	if title == "time interval" {
		info = WinInfo{int(float64(height) / 2.5), width - 2, y, 1}
	} else {
		info = WinInfo{len(op) + 2, op_wid, y, (width / 2) - (op_wid / 2) - op_wid%2}
	}

	y += info.H

	//Clear, refresh, update all windows
	win.Clear()
	win.Resize(info.H, info.W)     //Resize windows based on new dimensions
	win.MoveWindow(info.Y, info.X) //move windows to appropriate locations
	win.NoutRefresh()

	// print windows
	//win_title := title_array[len(sel)]
	win.Box('|', '-')
	// Rendering title
	win.AttrOn(ncurses.ColorPair(2))
	win.AttrOn(ncurses.A_BOLD)
	win.MovePrint(0, 1, title)
	win.AttrOff(ncurses.ColorPair(2))
	win.AttrOff(ncurses.A_BOLD)

	piece := " ♟︎ "
	screen.AttrOn(ncurses.ColorPair(1))
	screen.AttrOn(ncurses.A_DIM)
	screen.AttrOn(ncurses.A_BLINK)
	screen.MovePrint(y, ((width / 2) - 2), piece)
	screen.AttrOff(ncurses.A_BLINK)
	screen.AttrOff(ncurses.A_DIM)
	screen.AttrOff(ncurses.ColorPair(1))

	screen.NoutRefresh()
	win.NoutRefresh()
	ncurses.Update()
}

func DrawLichessGame(stdscr *ncurses.Window, key ncurses.Key, windows_array [4]*ncurses.Window, windows_info_arr [4]WinInfo) {
	//Clear and refresh the screen for a blank canvas
	stdscr.Clear()
	height, width := stdscr.MaxYX()

	//h, w, y, x
	windows_info_arr[0] = WinInfo{H: height / 2, W: width, Y: 0, X: 0}                                           //bw
	windows_info_arr[1] = WinInfo{H: (height / 2) - 1, W: width / 2, Y: (height / 2), X: 0}                      //iw
	windows_info_arr[2] = WinInfo{H: (height / 4) - 1, W: width / 2, Y: height / 2, X: width / 2}                //pw
	windows_info_arr[3] = WinInfo{H: (height / 4), W: width / 2, Y: int(float64(height)*0.75) - 1, X: width / 2} //hw

	//Clear, refresh, update all windows
	for i, win := range windows_array {
		win.Clear()
		info := windows_info_arr[i]
		win.Resize(info.H, info.W)     //Resize windows based on new dimensions
		win.MoveWindow(info.Y, info.X) //move windows to appropriate locations
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
	statusbarstr := "CHESS-CLI | LICHESS CLIENT | Press 'Ctrl+o' to exit"
	// if key == ZeroKey {
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
}

func DrawLichessPostGame(stdscr *ncurses.Window, windows_array [3]*ncurses.Window, windows_info_arr [3]WinInfo) {
	//Clear and refresh the screen for a blank canvas
	stdscr.Clear()
	height, width := stdscr.MaxYX()

	//h, w, y, x
	windows_info_arr[0] = WinInfo{H: (height / 3) * 2, W: (width / 3) * 2, Y: 0, X: 0}
	windows_info_arr[1] = WinInfo{H: (height / 3) * 2, W: (width / 3), Y: 0, X: (width / 3) * 2}
	windows_info_arr[2] = WinInfo{H: (height / 3), W: width, Y: (height / 3) * 2, X: 0}

	//Clear, refresh, update all windows
	for i, win := range windows_array {
		win.Clear()
		info := windows_info_arr[i]
		win.Resize(info.H, info.W)     //Resize windows based on new dimensions
		win.MoveWindow(info.Y, info.X) //move windows to appropriate locations
		win.NoutRefresh()
	}

	//get mouse location
	// cursor_x := math.Min(width-1, math.Max(0, cursor_x))
	// cursor_y := math.Min(height-1, math.Max(0, cursor_y))

	// Declaration of strings
	board_title := "board"
	info_title := "choices"
	history_title := "move_history"
	title_array := []string{board_title, info_title, history_title}
	//keystr := fmt.Sprintf("Last key pressed: %v", key)
	statusbarstr := "CHESS-CLI | LICHESS CLIENT | Press 'Ctrl+o' to exit"
	// if key == ZeroKey {
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
}

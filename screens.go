package main

// #include <sys/ioctl.h>
import "C"

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	ncurses "github.com/nate-xyz/goncurses"
)

func local_game_screen(stdscr *ncurses.Window) int {
	var key ncurses.Key
	// var mouse_event *ncurses.MouseEvent
	// _ = mouse_event

	height, width := stdscr.MaxYX()

	//start windows
	bw_info := windowSizePos{(height / 4) * 3, width / 2, 0, 0}
	iw_info := windowSizePos{(height / 4) - 1, width / 2, (height / 4) * 3, 0}
	pw_info := windowSizePos{height / 2, width / 2, 0, width / 2}
	hw_info := windowSizePos{(height / 2) - 1, width / 2, height / 2, width / 2}

	board_window, _ := ncurses.NewWindow(bw_info.h, bw_info.w, bw_info.y, bw_info.x)
	prompt_window, _ := ncurses.NewWindow(pw_info.h, pw_info.w, pw_info.y, pw_info.x)
	info_window, _ := ncurses.NewWindow(iw_info.h, iw_info.w, iw_info.y, iw_info.x)
	history_window, _ := ncurses.NewWindow(hw_info.h, hw_info.w, hw_info.y, hw_info.x)

	windows_array := [4]*ncurses.Window{board_window, info_window, prompt_window, history_window}
	windows_info_arr := [4]windowSizePos{bw_info, iw_info, pw_info, hw_info}

	draw_local_game_screen(stdscr, key, windows_array, windows_info_arr)
	for {
		select {
		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			draw_local_game_screen(stdscr, key, windows_array, windows_info_arr)
		default: //normal character loop here
			//external function calls
			update_input(prompt_window, key)
			if game_logic(board_window) {
				return 2
			}
			draw_board(board_window)
			display_info(info_window)
			display_history(history_window)
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
			case control_o_key, zero_key:
				user_input_string = ""
				inputted_str = ""
				entered_move = false
				if key == control_o_key {
					return 3
				} else {
					return 0
				}
			}
		}
	}
}

func post_screen(screen *ncurses.Window) int {
	var key ncurses.Key
	height, width := screen.MaxYX()

	//start windows
	bp_info := windowSizePos{((height) - (height / 3)), (width), 0, 0}
	pp_info := windowSizePos{((height) / 4) - 1, width, ((height / 4) * 3), 0}

	board_post_window, _ := ncurses.NewWindow(bp_info.h, bp_info.w, bp_info.y, bp_info.x)
	history_window, _ := ncurses.NewWindow(pp_info.h, pp_info.w, pp_info.y, pp_info.x)

	windows_array := [2]*ncurses.Window{board_post_window, history_window}
	windows_info_arr := [2]windowSizePos{bp_info, pp_info}

	draw_post_screen(screen, key, windows_array, windows_info_arr)
	for {
		select {
		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			draw_post_screen(screen, key, windows_array, windows_info_arr)
		default: //normal character loop here
			draw_board(board_post_window)
			//update_input(prompt_post_window, key)
			for _, win := range windows_array {
				win.NoutRefresh()
			}
			ncurses.Update()
			key = screen.GetChar()
			switch key {
			case control_o_key, one_key, zero_key:
				user_input_string = ""
				inputted_str = ""
				entered_move = false
				if key == control_o_key {
					return 3
				} else if key == one_key {
					return 1
				} else {
					return 0
				}
			}
		}
	}
}

func welcome_screen(screen *ncurses.Window) ncurses.Key {
	var key ncurses.Key
	height, width := screen.MaxYX()
	//start windows
	options := []string{"<<press '1' to play locally>>", "<<press '2' to play online>>", "<<press '3' to play stockfish>>", "<<quit>>"}
	op_info := windowSizePos{(height / 2) - 4, width / 2, (height / 2) + 2, width / 4}
	options_window, _ := ncurses.NewWindow(op_info.h, op_info.w, op_info.y, op_info.x)
	windows_array := [1]*ncurses.Window{options_window}
	windows_info_arr := [1]windowSizePos{op_info}
	var option_index int = 0
	var selected bool
	draw_welcome_screen(screen, key, windows_array, windows_info_arr, options)
	for {
		select {
		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			draw_welcome_screen(screen, key, windows_array, windows_info_arr, options)
		default: //normal character loop here
			key = screen.GetChar()
			option_index, selected = options_input(options_window, key, options, option_index)
			if selected {
				switch option_index {
				case 0:
					key = one_key
				case 1:
					key = two_key
				case 2:
					key = three_key
				case 3:
					key = control_o_key
				}
			}
			switch key {
			case control_o_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 3)
				}
				return key
			case one_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 0)
				}
				return key
			case two_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 1)
				}
				return key
			case three_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 2)
				}
				return key
			}
		}
	}
}

func lichess_welcome(screen *ncurses.Window) ncurses.Key {
	var key ncurses.Key
	do_oauth()
	if UserInfo.ApiToken != "" {
		err := GetEmail()
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		err = GetUsername()
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}
	height, width := screen.MaxYX()
	//start windows
	options := []string{"<<Press 0 to return to welcome screen>>", "<<Press 1 to view / create challenges>>", "<<Press 2 to view / join ongoing games>>", "etc", "quit"}
	op_info := windowSizePos{(height / 2) - 4, width / 2, (height / 2) + 2, width / 4}
	options_window, _ := ncurses.NewWindow(op_info.h, op_info.w, op_info.y, op_info.x)
	windows_array := [1]*ncurses.Window{options_window}
	windows_info_arr := [1]windowSizePos{op_info}
	var option_index int = 0
	var selected bool
	draw_lichess_welcome(screen, key, windows_array, windows_info_arr, options)
	for {
		select {
		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			draw_lichess_welcome(screen, key, windows_array, windows_info_arr, options)
		default: //normal character loop here
			key = screen.GetChar()
			option_index, selected = options_input(options_window, key, options, option_index)
			if selected {
				switch option_index {
				case 0:
					key = zero_key
				case 1:
					key = one_key
				case 2:
					key = two_key
				case 3:
				case 4:
					key = control_o_key
				}
			}
			switch key {
			case zero_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 0)
				}
				return key
			case one_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 1)
				}
				return two_key
			case two_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 1)
				}
				return three_key
			case three_key:
			case control_o_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 4)
				}
				return key
			}

		}
	}
}

func lichess_challenges(screen *ncurses.Window) ncurses.Key {
	var key ncurses.Key
	if UserInfo.ApiToken != "" {
		err := GetFriends()
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}
	draw_lichess_challenges(screen)
	for {
		select {
		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			draw_lichess_challenges(screen)
		default: //normal character loop here
			key = screen.GetChar()
			switch key {
			case control_o_key, zero_key, one_key, two_key, three_key:
				user_input_string = ""
				inputted_str = ""
				entered_move = false
				return key
			}
		}
	}
}

func osTermSize() (int, int, error) {
	w := &C.struct_winsize{}
	res, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(w)),
	)
	if int(res) == -1 {
		return 0, 0, err
	}
	return int(w.ws_row), int(w.ws_col), nil
}

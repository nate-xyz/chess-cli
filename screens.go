package main

// #include <sys/ioctl.h>
import "C"

import (
	"fmt"
	"os"
	"syscall"
	"time"
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
	//options := []string{"<<press '1' to play locally>>", "<<press '2' to play online>>", "<<press '3' to play stockfish>>", "<<quit>>"}

	options := []string{"play locally", "play online", "play stockfish", "quit"}
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
					// key = three_key
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

func lichess_welcome(screen *ncurses.Window) int {
	var key ncurses.Key
	height, width := screen.MaxYX()
	done := make(chan struct{})
	ticker := time.NewTicker(time.Second)
	go func() {
		do_oauth()
		if UserInfo.ApiToken != "" {
			if UserEmail == "" {
				err := GetEmail()
				if err != nil {
					fmt.Printf("%s\n", err)
					UserEmail = "could not retrieve email"
				}
			}
			if Username == "" {
				err := GetUsername()
				if err != nil {
					fmt.Printf("%s\n", err)
					UserEmail = "could not retrieve username"
				}
			}

		}
		close(done)
	}()
	loading_screen(screen, "Please login through lichess.org")
blocking_loop:
	for {
		select {
		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			loading_screen(screen, "Please login through lichess.org")
		case <-ticker.C:
			loading_screen(screen, "Please login through lichess.org")
		case <-done:
			break blocking_loop
		}

	}

	//start windows
	//options := []string{"<<Press 0 to return to welcome screen>>", "<<Press 1 to view / create challenges>>", "<<Press 2 to view / join ongoing games>>", "etc", "quit"}
	options := []string{"new game", "ongoing games", "back", "quit"}
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
				case 0: //view / create challenges
					key = one_key
				case 1: //view / join ongoing games
					//key = two_key
				case 2:
					key = zero_key //return to welcome screen
				case 3:
					key = control_o_key

				}
			}
			switch key {
			case zero_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 0)
				}
				return 0
			case one_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 1)
				}
				return 2
			case two_key:
			case three_key:
			case control_o_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 4)
				}
				return 5
			}

		}
	}
}

func loading_screen(screen *ncurses.Window, message string) {
	screen.Clear()
	height, width := screen.MaxYX()
	dt := time.Now().Unix() % 10
	screen.MovePrint(height/2, width/2-len(message)/2, message)
	screen.MovePrint((height/2)+1, width/2, fmt.Sprintf("%v", loader[dt]))
	screen.Refresh()
}

func lichess_challenges(screen *ncurses.Window) int {
	var key ncurses.Key
	var option_index int = 0
	var selected bool
	//var choosing_from_challenges bool
	height, width := screen.MaxYX()
	done := make(chan struct{})
	ticker := time.NewTicker(time.Second)
	go func() {
		if UserInfo.ApiToken != "" {
			GetChallenges()
		}
		close(done)
	}()
	load_msg := "Requesting your challenges ... "
	loading_screen(screen, load_msg)
blocking_loop:
	for {
		select {
		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			loading_screen(screen, load_msg)
		case <-ticker.C:
			loading_screen(screen, load_msg)
		case <-done:
			break blocking_loop
		}

	}

	//start windows
	options := []string{"create a new game", "select a challenge", "etc", "back", "quit"}
	max_len := getMaxLenStr(options) + 6
	op_info := windowSizePos{len(options) + 2, max_len + 2, 2, (width / 2) - ((max_len + 2) / 2)}
	in_info := windowSizePos{int(float64(height)*0.75) - 1, width / 2, len(options) + 4, 0}
	out_info := windowSizePos{int(float64(height)*0.75) - 1, width / 2, len(options) + 4, width / 2}
	options_window, _ := ncurses.NewWindow(op_info.h, op_info.w, op_info.y, op_info.x)
	in_challenge, _ := ncurses.NewWindow(in_info.h, in_info.w, in_info.y, in_info.x)
	out_challenge, _ := ncurses.NewWindow(out_info.h, out_info.w, out_info.y, out_info.x)
	windows_array := [3]*ncurses.Window{options_window, in_challenge, out_challenge}
	windows_info_arr := [3]windowSizePos{op_info, in_info, out_info}

	var mode int = 0
	if UserInfo.ApiToken != "" {
		err := GetFriends()
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}
	draw_lichess_challenges(screen, key, windows_array, windows_info_arr, options)
	for {
		select {
		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)

			height, width := screen.MaxYX()

			switch mode {
			case 0:
				//update window dimensions
				max_len := getMaxLenStr(options) + 6
				windows_info_arr[0] = windowSizePos{len(options) + 2, max_len + 2, 2, (width / 2) - ((max_len + 2) / 2)}
				windows_info_arr[1] = windowSizePos{int(float64(height)*0.75) - 1, width / 2, len(options) + 4, 0}
				windows_info_arr[2] = windowSizePos{int(float64(height)*0.75) - 1, width / 2, len(options) + 4, width / 2}
				draw_lichess_challenges(screen, key, windows_array, windows_info_arr, options)
			case 1:
			case 3:

			}
		default: //normal character loop here
			key = screen.GetChar()
			option_index, selected = options_input(options_window, key, options, option_index)
			if selected {
				switch option_index {
				case 0: //create a new challenge
					return 3
				case 1: //accept a challenge
				case 2: //view a challenge
				case 3: //back
					return 1
				case 4: //quit
					return 5

				}
			}
			switch key {
			case zero_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 0)
				}
				return 3
			case three_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 2)
				}
				return 2
			case control_o_key:
				if !selected {
					_, _ = options_input(options_window, key, options, 4)
				}
				return 5
			}
		}
	}
}

type CreateChallengeType struct {
	Type           int
	Username       string
	DestUser       string
	Variant        string
	VariantIndex   int
	TimeOption     int
	ClockLimit     int
	ClockIncrement int
	Days           int
	Rated          bool
	Color          string
	ColorIndex     int
}

func create_game(screen *ncurses.Window) int {
	var key ncurses.Key
	//_, width := screen.MaxYX()
	//start windows
	options := []string{"seek a random player", "challenge a friend", "play with the computer", "back", "quit"}
	variant_options := []string{"standard", "chess960", "crazyhouse", "antichess", "atomic", "horde", "kingOfTheHill", "racingKings", "threeCheck"}
	time_options := []string{"real time", "correspondence", "unlimited"}
	rated_options := []string{"casual", "rated"}
	color_options := []string{"random", "white", "black"}
	selection := []string{}

	//options_arrays := [][]string{options, variant_options, time_options, rated_options, color_options}

	//mode 0 : select game type (friend, bot)
	//mode 1 : select game variant
	//mode 2 : select time control
	//mode 3 : select time interval for real time / correspondence
	//mode 4 : select rated / casual
	//mode 5 : select color
	//mode 6 : choose user

	//windows_info_arr := []windowSizePos{{0, 0, 0, 0}}
	op_info := windowSizePos{0, 0, 0, 0}
	options_window, _ := ncurses.NewWindow(0, 0, 0, 0)

	//windows_array := []*ncurses.Window{options_window}
	var window_mode int = 0
	var option_index int = 0
	var selected bool
	var newChallenge CreateChallengeType
	draw_create_game_screen(screen, options, selection, options_window, op_info)
	for {
		select {
		case <-sigs: //resize on os resize signal
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			switch window_mode {
			case 0:
				draw_create_game_screen(screen, options, selection, options_window, op_info)
			case 1:
				draw_create_game_screen(screen, variant_options, selection, options_window, op_info)
			case 2:
				draw_create_game_screen(screen, time_options, selection, options_window, op_info)
			case 3:
				//draw_create_game_screen(screen, options, options_window, op_info)
			case 4:
				draw_create_game_screen(screen, rated_options, selection, options_window, op_info)
			case 5:
				draw_create_game_screen(screen, color_options, selection, options_window, op_info)
			case 6:
				draw_create_game_screen(screen, allFriends, selection, options_window, op_info)
			}

		default: //normal character loop here
			key = screen.GetChar()
			switch window_mode {
			case 0:
				option_index, selected = options_input(options_window, key, options, option_index)
				if selected {
					switch option_index {
					case -1: //go to challenges screen
						return 2
					case 0, 1, 2: //game mode type
						//selection[0] = options[option_index]
						selection = append(selection, options[option_index])
						newChallenge.Type = option_index
						window_mode = 1
						option_index = 0
						draw_create_game_screen(screen, variant_options, selection, options_window, op_info)
					// case 1: //bot
					// 	window_mode = 1
					case 3: //go to challenges screen
						return 2
					case 4: //quit
						return 5
					}

				}
			case 1: //mode 1 : select game variant
				option_index, selected = options_input(options_window, key, variant_options, option_index)
				if selected {
					switch option_index {
					case -1:
						window_mode = 0 //go back to main options
						selection = selection[:len(selection)-1]
						option_index = newChallenge.Type
						draw_create_game_screen(screen, options, selection, options_window, op_info)

					case 0, 1, 2, 3, 4, 5, 6, 7, 8: //all variants
						//selection[1] = variant_options[option_index]
						selection = append(selection, variant_options[option_index])
						newChallenge.Variant = variant_options[option_index] //save selected variant
						window_mode = 2
						option_index = 0 //go to time
						draw_create_game_screen(screen, time_options, selection, options_window, op_info)

					}

				}
			case 2: //mode 2 : select time control
				option_index, selected = options_input(options_window, key, time_options, option_index)
				if selected {
					switch option_index {
					case -1: //go back to variant
						window_mode = 1
						selection = selection[:len(selection)-1]
						option_index = newChallenge.VariantIndex
						draw_create_game_screen(screen, variant_options, selection, options_window, op_info)

					case 0, 1, 2: //realtime, corrspondence, unlimited
						//selection[2] = time_options[option_index]
						selection = append(selection, time_options[option_index])
						window_mode = 4
						newChallenge.TimeOption = option_index
						draw_create_game_screen(screen, rated_options, selection, options_window, op_info)
					}

				}
			case 3: //mode 3 : select time interval for real time / correspondence
			case 4: //mode 4 : select rated / casual

				option_index, selected = options_input(options_window, key, rated_options, option_index)
				if selected {
					switch option_index {
					case -1: //go back to time
						window_mode = 2
						selection = selection[:len(selection)-1]
						option_index = newChallenge.TimeOption
						draw_create_game_screen(screen, time_options, selection, options_window, op_info)

					case 0, 1: //rated or casual
						//selection[4] = rated_options[option_index]
						selection = append(selection, rated_options[option_index])
						window_mode = 5 // go to color
						option_index = 0
						if option_index == 1 {
							newChallenge.Rated = true

						} else {
							newChallenge.Rated = false
						}
						draw_create_game_screen(screen, color_options, selection, options_window, op_info)
					}

				}
			case 5: //mode 5 : select color

				option_index, selected = options_input(options_window, key, color_options, option_index)
				if selected {
					switch option_index {
					case -1: //go back to rated
						window_mode = 4
						selection = selection[:len(selection)-1]
						if newChallenge.Rated {
							option_index = 1
						} else {
							option_index = 0
						}
						draw_create_game_screen(screen, rated_options, selection, options_window, op_info)

					case 0, 1, 2: //choose color: random, white, black
						//selection[5] = color_options[option_index]
						selection = append(selection, color_options[option_index])
						newChallenge.Color = color_options[option_index]
						newChallenge.ColorIndex = option_index
						window_mode = 6 // go to user
						option_index = 0
						draw_create_game_screen(screen, append(allFriends, "get url"), selection, options_window, op_info)
					}

				}
			case 6: //mode 6 : choose user
				option_index, selected = options_input(options_window, key, append(allFriends, "get url"), option_index)
				if selected {
					switch option_index {
					case -1: //go back to color choice
						window_mode = 5
						selection = selection[:len(selection)-1]
						option_index = newChallenge.ColorIndex
						draw_create_game_screen(screen, color_options, selection, options_window, op_info)
					}
				}
			}
		}
	}
}

func getMaxLenStr(arr []string) int {
	max_len := 0
	for _, str := range arr {
		if max_len < len(str) {
			max_len = len(str)
		}
	}
	return max_len
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

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

	options := []string{"play locally", "play online", "play stockfish", "quit", "test"}
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
				case 4:
					//curChallenge = testChallenge
					lichessScreenHandler(screen, 1)
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

	var loading_msg string
	if UserInfo.ApiToken != "" {
		loading_msg = "Please login through lichess.org"
	} else {
		loading_msg = "Logging into lichess.org with your token"
	}

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
			err := GetChallenges()
			if err != nil {
				error_message <- fmt.Errorf("unable to retrieve challenges: %v", err)
			} else {
				noti_message <- fmt.Sprintf("retrieved challenges")
			}
			err = GetOngoingGames()
			if err != nil {
				error_message <- fmt.Errorf("unable to retrieve ongoing games: %v", err)
			} else {
				noti_message <- fmt.Sprintf("retrieved ongoing games")
			}
			ready <- struct{}{} //unblock event stream
		}
		close(done)
	}()

	screen.Erase()
	//ticker := time.NewTicker(time.Second)
	ticker := time.NewTicker(time.Millisecond * 500)
	loading_screen(screen, loading_msg)
blocking_loop:
	for {
		select {
		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			loading_screen(screen, loading_msg)
		case <-ticker.C:
			loading_screen(screen, loading_msg)
		case <-done:
			break blocking_loop
		}

	}

	//start windows
	//options := []string{"<<Press 0 to return to welcome screen>>", "<<Press 1 to view / create challenges>>", "<<Press 2 to view / join ongoing games>>", "etc", "quit"}
	options := []string{"new game", "ongoing games", "back", "quit", "test"}
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
				case 4:
					curChallenge = testChallenge
					lichessScreenHandler(screen, 4)

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

func lichess_challenges(screen *ncurses.Window) int {
	var key ncurses.Key
	var option_index int = 0
	var selected bool
	//var choosing_from_challenges bool
	screen.Clear()
	height, width := screen.MaxYX()
	//done := make(chan struct{})
	//ticker := time.NewTicker(time.Second)
	// 	go func() {
	// 		if UserInfo.ApiToken != "" {
	// 			err := GetChallenges()
	// 			if err != nil {
	// 				error_message <- fmt.Errorf("unable to retrieve challenges: %v", err)
	// 			} else {
	// 				noti_message <- fmt.Sprintf("retrieved challenges")
	// 			}
	// 			err = GetOngoingGames()
	// 			if err != nil {
	// 				error_message <- fmt.Errorf("unable to retrieve ongoing games: %v", err)
	// 			} else {
	// 				noti_message <- fmt.Sprintf("retrieved ongoing games")
	// 			}
	// 		}
	// 		close(done)
	// 	}()
	// 	load_msg := "Requesting your challenges ... "
	// 	loading_screen(screen, load_msg)

	// blocking_loop:
	// 	for {
	// 		select {
	// 		case <-sigs:
	// 			tRow, tCol, _ := osTermSize()
	// 			ncurses.ResizeTerm(tRow, tCol)
	// 			loading_screen(screen, load_msg)
	// 		case <-ticker.C:
	// 			loading_screen(screen, load_msg)
	// 		case <-done:
	// 			break blocking_loop
	// 		}

	// 	}

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
	submit_options := []string{"challenge ok? submit", "back", "back to challenge screen", "quit chess-cli"}
	title_array := []string{"options", "variants", "time options", "time interval", "rated/casual", "choose color", "select friend to challenge", "submit challenge"}
	op_arr := [][]string{options, variant_options, time_options, {}, rated_options, color_options, append(allFriends, "get url", "back"), submit_options}
	//empty_strings := []string{"", "", "", "", "", "", "", "", "", "", "", "", ""}

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
	var slider_index int = 0
	var selected bool
	tic_index := []int{0, 0}
	var newChallenge CreateChallengeType
	draw_create_game_screen(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)
	for {
		select {
		case <-sigs: //resize on os resize signal
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			draw_create_game_screen(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)

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
						//draw_create_game_screen(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)
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

					case 0, 1, 2, 3, 4, 5, 6, 7, 8: //all variants
						selection = append(selection, variant_options[option_index])
						newChallenge.Variant = variant_options[option_index] //save selected variant
						window_mode = 2
						option_index = 0 //go to time

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
					case 0, 1, 2: //realtime, corrspondence, unlimited
						selection = append(selection, time_options[option_index])
						newChallenge.TimeOption = option_index
						if option_index == 2 { //time interval window skip if unlimited
							window_mode = 4
						} else {
							window_mode = 3
						}

					}

				}
			case 3: //mode 3 : select time interval for real time / correspondence
				var min_time float64
				tic_index, slider_index, min_time, selected = slider_input(options_window, key, newChallenge.TimeOption, tic_index, slider_index)
				if selected {
					switch newChallenge.TimeOption {
					case 0: // real time, two sliders
						switch slider_index {
						case 0, 1, 2: //
							window_mode = 4
							option_index = 0
							newChallenge.MinTurn = min_time
							newChallenge.ClockLimit = fmt.Sprintf("%v", int(min_time*60))
							newChallenge.ClockIncrement = fmt.Sprintf("%v", tic_index[1])
							selection = append(selection, fmt.Sprintf("%v+%v", min_time, tic_index[1])) //show select
						case 3: // go back to time control window (real time, corr, unl)
							tic_index = []int{0, 0}
							slider_index = 0
							option_index = newChallenge.TimeOption
							window_mode = 2
							selection = selection[:len(selection)-1]
						}

					case 1: //correspondence, one slider for day
						switch slider_index {
						case 0, 1:
							window_mode = 4
							option_index = 0
							newChallenge.Days = fmt.Sprintf("%v", tic_index[0]+1)
							selection = append(selection, fmt.Sprintf("%v days", newChallenge.Days)) //show select
						case 2: // go back to time control window (real time, corr, unl)
							tic_index = []int{0, 0}
							slider_index = 0
							window_mode = 2
							selection = selection[:len(selection)-1]
							option_index = newChallenge.TimeOption
						}
					}
					//draw_create_game_screen(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)

				}
			case 4: //mode 4 : select rated / casual
				option_index, selected = options_input(options_window, key, rated_options, option_index)
				if selected {
					switch option_index {
					case -1: //go back to time interval window, time control if unlimited was selected

						selection = selection[:len(selection)-1]
						tic_index = []int{0, 0}
						slider_index = 0
						if newChallenge.TimeOption != 2 { //not unlimited
							option_index = 0
							window_mode = 3
						} else {
							option_index = newChallenge.TimeOption
							window_mode = 2
						}
					case 0, 1: //go to rated or casual window
						selection = append(selection, rated_options[option_index])
						window_mode = 5 // go to color
						option_index = 0
						if option_index == 1 {
							newChallenge.Rated = "true"
							newChallenge.RatedBool = true
						} else {
							newChallenge.Rated = "false"
							newChallenge.RatedBool = false
						}
					}

				}
			case 5: //mode 5 : select color

				option_index, selected = options_input(options_window, key, color_options, option_index)
				if selected {
					switch option_index {
					case -1: //go back to rated window
						window_mode = 4
						selection = selection[:len(selection)-1]
						if newChallenge.RatedBool {
							option_index = 1
						} else {
							option_index = 0
						}
					case 0, 1, 2: //choose color: random, white, black,
						//only friend mode go to user screen, seek random / or bot go to confirmation screen
						selection = append(selection, color_options[option_index])
						newChallenge.Color = color_options[option_index]
						newChallenge.ColorIndex = option_index
						if newChallenge.Type == 1 {
							window_mode = 6 // go to user
						} else {
							window_mode = 7
						}
						option_index = 0
					}

				}
			case 6: //mode 6 : choose user to challenge
				option_index, selected = options_input(options_window, key, append(allFriends, "get url", "back"), option_index)
				if selected {
					if option_index == -1 || option_index == len(allFriends) { //go back to color choice
						window_mode = 5
						selection = selection[:len(selection)-1]
						option_index = newChallenge.ColorIndex
					} else {
						if option_index == len(append(allFriends, "get url", "back"))-1 { //get url option for open ended challenge

							newChallenge.OpenEnded = true
						} else { //specfic user chosen
							newChallenge.DestUser = allFriends[option_index]

							newChallenge.OpenEnded = false
						}
						window_mode = 7
						option_index = 0
					}

				}
			case 7: //confirmation screen
				option_index, selected = options_input(options_window, key, submit_options, option_index)
				if selected {

					switch option_index {
					case -1, 1: //go back to user screen if friend mode, color screen otherwise
						selection = selection[:len(selection)-1]
						option_index = 0
						if newChallenge.Type == 1 { //not unlimited
							option_index = 0
							window_mode = 6
						} else {
							option_index = newChallenge.ColorIndex
							window_mode = 5
						}
					case 0: //submit the challenge
						//return to lichessScreenHandler to go to waiting screen
						curChallenge = newChallenge
						//noti_message <- fmt.Sprintf("going to wait screen")
						return 4
					case 2: //go back to the main challege screen
						return 2
					case 4: //quit chess cli
						return 5

					}
				}

			}

			//redraw screen for all window mode if something is selected
			if selected {
				draw_create_game_screen(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)

			}

		}
	}
}

var curChallenge CreateChallengeType
var waiting_alert chan StreamEventType

func lichess_game_wait(screen *ncurses.Window) int {
	//svar key ncurses.Key
	screen.Clear()
	//height, width := screen.MaxYX()

	//done := make(chan struct{})
	var localid string
	getgameid := make(chan string, 1)
	waiting_alert = make(chan StreamEventType, 1)
	ticker := time.NewTicker(time.Second)

	go func(gchan chan<- string) {
		if UserInfo.ApiToken != "" {
			noti_message <- fmt.Sprintf("type is %v", curChallenge.Type)
			switch curChallenge.Type {
			case 0: //random seek
				//TODO: api call CREATE A SEEK
				noti_message <- fmt.Sprintf("creating a random seek")
			case 1: //challenge friend

				if curChallenge.OpenEnded {
					noti_message <- fmt.Sprintf("creating an open ended challenge")
					//TODO: api call CREATE A OPEN END CHALLENGE

				} else {
					// api call CREATE A CHALLENGE
					noti_message <- fmt.Sprintf("creating a challenge")

					err, id := CreateChallenge(curChallenge)
					if err != nil {
						error_message <- err
						//os.Exit(1)
					} else {
						noti_message <- fmt.Sprintf("posted challenge, id: %v", id)
					}
					gchan <- id
				}
			case 2: //lichess ai
				noti_message <- fmt.Sprintf("challenging the lichess ai")
				//TODO: api call CHALLENGE THE AI
			}
		} else {
			error_message <- fmt.Errorf("no token")
		}
		//close(done)
	}(getgameid)
	noti_message <- fmt.Sprintf("sent goroutine ")
	// load_msg := "requesting game from lichess ... "
	// loading_screen(screen, load_msg)
	// blocking_loop:
	// 	for {
	// 		select {
	// 		case <-sigs:
	// 			tRow, tCol, _ := osTermSize()
	// 			ncurses.ResizeTerm(tRow, tCol)
	// 			loading_screen(screen, load_msg)
	// 		case <-ticker.C:
	// 			loading_screen(screen, load_msg)
	// 		case <-done:
	// 			break blocking_loop
	// 		}

	// 	}
	//load_msg = fmt.Sprintf("waiting for stream event from lichess...")
	load_msg := "requesting game from lichess ... "
	loading_screen(screen, load_msg)
	for {
		select {
		// case <-stream_done:
		// 	load_msg = fmt.Sprintf("stream closed")
		// 	loading_screen(screen, load_msg)
		case id := <-getgameid:
			noti_message <- fmt.Sprintf("got getgameid")
			noti_message <- fmt.Sprintf("event %v retrieved from channel", id)
			localid = id
			s, b := containedInEventStream(EventStreamArr, id)

			if b {
				switch s {
				case "challenge":
					noti_message <- fmt.Sprintf("event: %v:%v", id, s)
					load_msg = fmt.Sprintf("waiting for %v to accept the challenge %v/%v.", curChallenge.DestUser, hostUrl, id)
					noti_message <- load_msg
				case "gameFinish":
					noti_message <- fmt.Sprintf("event %v wrong type s", s)
					time.Sleep(time.Second * 3)
					return 2
				case "challengeCanceled", "challengeDeclined":
					noti_message <- fmt.Sprintf("challenge rejected: %v", s)
					error_message <- fmt.Errorf("challenge rejected: %v", s)
					time.Sleep(time.Second * 3)
					return 2
				case "gameStart":
					noti_message <- fmt.Sprintf("event id: %v", id)
					load_msg = fmt.Sprintf("load event: %v!!!", s)
					loading_screen(screen, load_msg)
					currentGameID = id
					//call stream board state
					gameStateChan = make(chan BoardState, 1)
					board_state_sig = make(chan bool, 1)
					go StreamBoardState(gameStateChan, board_state_sig, id, error_message)
					go BoardConsumer(gameStateChan, noti_message)
					time.Sleep(time.Second * 2)
					return 6
				}
			} else {
				load_msg = fmt.Sprintf("waiting for %v to accept the challenge %v/%v.", curChallenge.DestUser, hostUrl, id)

				noti_message <- fmt.Sprintf("event %v not in stream ... waiting", id)

				time.Sleep(time.Second * 3)
				//return 2
			}
		case e := <-stream_channel:
			if e.Id == localid {
				noti_message <- fmt.Sprintf("event id: %v", e.Id)
				error_message <- fmt.Errorf("event id: %v", e.Id)
				load_msg = fmt.Sprintf("load event: %v!!!", e.Event)
				loading_screen(screen, load_msg)
				currentGameID = e.Id
				//call stream board state
				gameStateChan = make(chan BoardState, 1)
				board_state_sig = make(chan bool, 1)
				go StreamBoardState(gameStateChan, board_state_sig, e.Id, error_message)
				go BoardConsumer(gameStateChan, noti_message)
				time.Sleep(time.Second * 2)
				return 6
			} else {
				noti_message <- fmt.Sprintf("event %v is not your game, it is %v", e.Id, e.Event)
			}

		case <-sigs:
			noti_message <- fmt.Sprintf("resize")
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			loading_screen(screen, load_msg)
		case <-ticker.C:
			noti_message <- fmt.Sprintf("ticker")
			//load_msg = fmt.Sprintf("waiting for stream event from lichess...")
			loading_screen(screen, load_msg)
			// case <-done: //normal character loop here
			// 	key = screen.GetChar()
			// 	switch key {
			// 	case control_o_key, one_key, zero_key:
			// 		user_input_string = ""
			// 		inputted_str = ""
			// 		entered_move = false
			// 		if key == control_o_key {
			// 			return 5
			// 		} else if key == one_key {
			// 			return 1
			// 		} else {
			// 			return 0
			// 		}
			// 	}
		}
		screen.Refresh()
	}
}
func containedInEventStream(a []StreamEventType, gameid string) (string, bool) {
	for _, e := range a {
		if e.Id == gameid {
			return e.Event, true
		}
	}
	return "", false
}

func containedInOngoingGames(a []OngoingGameInfo, gameid string) bool {
	for _, g := range a {
		if g.GameID == gameid {
			return true
		}
	}

	return false
}

func lichess_game(screen *ncurses.Window) int {
	var key ncurses.Key
	screen.Clear()
	height, width := screen.MaxYX()

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

	draw_local_game_screen(screen, key, windows_array, windows_info_arr)
	for {
		select {
		case <-board_state_sig:
			//get updated board position from board event
			screen.Clear()
			//Clear, refresh, update all windows
			for i, win := range windows_array {
				win.Clear()
				info := windows_info_arr[i]
				win.Resize(info.h, info.w)     //Resize windows based on new dimensions
				win.MoveWindow(info.y, info.x) //move windows to appropriate locations
				win.NoutRefresh()
			}
			board_window.MovePrint(height/2, width/2, "something happened!")
			time.Sleep(time.Second)
			screen.Refresh()

		case <-sigs:
			tRow, tCol, _ := osTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			draw_local_game_screen(screen, key, windows_array, windows_info_arr)
		default: //normal character loop here
			//external function calls
			update_input(prompt_window, key)
			if lichess_game_logic(board_window) {
				<-ready
				os.Exit(0) //game done no post lichess screen tho
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
			key = screen.GetChar()
			switch key {
			case control_o_key, zero_key:
				user_input_string = ""
				inputted_str = ""
				entered_move = false
				if key == control_o_key {
					return 5
				} else {
					return 1
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

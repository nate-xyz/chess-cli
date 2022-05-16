package online

import (
	"fmt"
	"os"
	"time"

	. "github.com/nate-xyz/chess-cli/shared"
	ncurses "github.com/nate-xyz/goncurses_"
)

func LichessChallenges(screen *ncurses.Window) int {
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
	// 				ErrorMessage <- fmt.Errorf("unable to retrieve challenges: %v", err)
	// 			} else {
	// 				NotiMessage <- fmt.Sprintf("retrieved challenges")
	// 			}
	// 			err = GetOngoingGames()
	// 			if err != nil {
	// 				ErrorMessage <- fmt.Errorf("unable to retrieve ongoing games: %v", err)
	// 			} else {
	// 				NotiMessage <- fmt.Sprintf("retrieved ongoing games")
	// 			}
	// 		}
	// 		close(done)
	// 	}()
	// 	load_msg := "Requesting your challenges ... "
	// 	LoadingScreen(screen, load_msg)

	// blocking_loop:
	// 	for {
	// 		select {
	// 		case <-Sigs:
	// 			tRow, tCol, _ := OsTermSize()
	// 			ncurses.ResizeTerm(tRow, tCol)
	// 			LoadingScreen(screen, load_msg)
	// 		case <-ticker.C:
	// 			LoadingScreen(screen, load_msg)
	// 		case <-done:
	// 			break blocking_loop
	// 		}

	// 	}

	//start windows
	options := []string{"create a new game", "select a challenge", "etc", "back", "quit"}
	max_len := GetMaxLenStr(options) + 6
	op_info := WinInfo{len(options) + 2, max_len + 2, 2, (width / 2) - ((max_len + 2) / 2)}
	in_info := WinInfo{int(float64(height)*0.75) - 1, width / 2, len(options) + 4, 0}
	out_info := WinInfo{int(float64(height)*0.75) - 1, width / 2, len(options) + 4, width / 2}
	options_window, _ := ncurses.NewWindow(op_info.H, op_info.W, op_info.Y, op_info.X)
	in_challenge, _ := ncurses.NewWindow(in_info.H, in_info.W, in_info.Y, in_info.X)
	out_challenge, _ := ncurses.NewWindow(out_info.H, out_info.W, out_info.Y, out_info.X)
	windows_array := [3]*ncurses.Window{options_window, in_challenge, out_challenge}
	windows_info_arr := [3]WinInfo{op_info, in_info, out_info}

	var mode int = 0
	if UserInfo.ApiToken != "" {
		err := GetFriends()
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}
	DrawLichessChallenges(screen, key, windows_array, windows_info_arr, options)
	for {
		select {
		case <-Sigs:
			tRow, tCol, _ := OsTermSize()
			ncurses.ResizeTerm(tRow, tCol)

			height, width := screen.MaxYX()

			switch mode {
			case 0:
				//update window dimensions
				max_len := GetMaxLenStr(options) + 6
				windows_info_arr[0] = WinInfo{len(options) + 2, max_len + 2, 2, (width / 2) - ((max_len + 2) / 2)}
				windows_info_arr[1] = WinInfo{int(float64(height)*0.75) - 1, width / 2, len(options) + 4, 0}
				windows_info_arr[2] = WinInfo{int(float64(height)*0.75) - 1, width / 2, len(options) + 4, width / 2}
				DrawLichessChallenges(screen, key, windows_array, windows_info_arr, options)
			case 1:
			case 3:

			}
		default: //normal character loop here
			key = screen.GetChar()
			option_index, selected = OptionsInput(options_window, key, options, option_index)
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
			case ZeroKey:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 0)
				}
				return 3
			case ThreeKey:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 2)
				}
				return 2
			case CtrlO_Key:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 4)
				}
				return 5
			}
		}
	}
}

func CreateLichessGame(screen *ncurses.Window) int {
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

	//windows_info_arr := []WinInfo{{0, 0, 0, 0}}
	op_info := WinInfo{0, 0, 0, 0}
	options_window, _ := ncurses.NewWindow(0, 0, 0, 0)

	//windows_array := []*ncurses.Window{options_window}
	var window_mode int = 0
	var option_index int = 0
	var slider_index int = 0
	var selected bool
	tic_index := []int{0, 0}
	var newChallenge CreateChallengeType
	DrawCreateGame(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)
	for {
		select {
		case <-Sigs: //resize on os resize signal
			tRow, tCol, _ := OsTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			DrawCreateGame(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)

		default: //normal character loop here
			key = screen.GetChar()
			switch window_mode {
			case 0:
				option_index, selected = OptionsInput(options_window, key, options, option_index)
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
						//DrawCreateGame(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)
						// case 1: //bot
					// 	window_mode = 1
					case 3: //go to challenges screen
						return 2
					case 4: //quit
						return 5
					}

				}
			case 1: //mode 1 : select game variant
				option_index, selected = OptionsInput(options_window, key, variant_options, option_index)
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
				option_index, selected = OptionsInput(options_window, key, time_options, option_index)
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
				tic_index, slider_index, min_time, selected = SliderInput(options_window, key, newChallenge.TimeOption, tic_index, slider_index)
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
					//DrawCreateGame(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)

				}
			case 4: //mode 4 : select rated / casual
				option_index, selected = OptionsInput(options_window, key, rated_options, option_index)
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

				option_index, selected = OptionsInput(options_window, key, color_options, option_index)
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
				option_index, selected = OptionsInput(options_window, key, append(allFriends, "get url", "back"), option_index)
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
				option_index, selected = OptionsInput(options_window, key, submit_options, option_index)
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
						//return to LichessScreenHandler to go to waiting screen
						CurrentChallenge = newChallenge
						//NotiMessage <- fmt.Sprintf("going to wait screen")
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
				DrawCreateGame(screen, op_arr[window_mode], selection, title_array[window_mode], options_window, op_info)

			}

		}
	}
}

// case e := <-StreamChannelForWaiter:
// 	if e.Id == localid {
// 		NotiMessage <- fmt.Sprintf("(stream channel)  event id: %v", e.Id)
// 		ErrorMessage <- fmt.Errorf("(stream channel)  event id: %v", e.Id)
// 		load_msg = fmt.Sprintf("(stream channel) load event: %v!!!", e.Event)
// 		screen.Clear()
// 		screen.Refresh()
// 		LoadingScreen(screen, load_msg)
// 		currentGameID = e.Id
// 		//call stream board state
// 		// gameStateChan = make(chan BoardState, 1)
// 		// board_state_sig = make(chan bool, 1)
// 		// go StreamBoardState(gameStateChan, board_state_sig, e.Id, ErrorMessage)
// 		// go BoardConsumer(gameStateChan, NotiMessage)

// 		//time.Sleep(time.Second * 5)
// 		return 6
// 	} else {
// 		NotiMessage <- fmt.Sprintf("event %v is not your game, it is %v (waiting for %v)", e.Id, e.Event, localid)
// 	}

func sendApiRequest(gchan chan<- string) {
	if UserInfo.ApiToken != "" {
		NotiMessage <- fmt.Sprintf("type is %v", CurrentChallenge.Type)
		switch CurrentChallenge.Type {
		case 0: //random seek
			//TODO: api call CREATE A SEEK
			NotiMessage <- fmt.Sprintf("creating a random seek")
		case 1: //challenge friend

			if CurrentChallenge.OpenEnded {
				NotiMessage <- fmt.Sprintf("creating an open ended challenge")
				//TODO: api call CREATE A OPEN END CHALLENGE

			} else {
				// api call CREATE A CHALLENGE
				NotiMessage <- fmt.Sprintf("creating a challenge")

				err, id := CreateChallenge(CurrentChallenge)
				if err != nil {
					ErrorMessage <- err
					NotiMessage <- fmt.Sprintf("%v", err)
					time.Sleep(time.Second * 5)
					os.Exit(1)
				} else {
					NotiMessage <- fmt.Sprintf("posted challenge, id: %v", id)
					gchan <- id
				}

			}
		case 2: //lichess ai
			NotiMessage <- fmt.Sprintf("challenging the lichess ai")
			//TODO: api call CHALLENGE THE AI
			err, id := CreateAiChallenge(CurrentChallenge)
			if err != nil {
				ErrorMessage <- err
				NotiMessage <- fmt.Sprintf("%v", err)
				time.Sleep(time.Second * 5)
				os.Exit(1)
			} else {
				NotiMessage <- fmt.Sprintf("posted challenge, id: %v", id)
				gchan <- id
			}
		}
	} else {
		ErrorMessage <- fmt.Errorf("no token")
	}
}

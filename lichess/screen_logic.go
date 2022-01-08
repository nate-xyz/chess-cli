package lichess

import (
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/nate-xyz/chess-cli/shared"
	ncurses "github.com/nate-xyz/goncurses_"
	"github.com/notnil/chess"
)

func LichessWelcome(screen *ncurses.Window) int {
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
					NotiMessage <- "could not retrieve email"
					time.Sleep(time.Second * 5)
					os.Exit(1)
				}
			}
			if Username == "" {
				err := GetUsername()
				if err != nil {
					fmt.Printf("%s\n", err)
					NotiMessage <- "could not retrieve username"
					time.Sleep(time.Second * 5)
					os.Exit(1)
				}
			}
			err := GetChallenges()
			if err != nil {
				ErrorMessage <- fmt.Errorf("unable to retrieve challenges: %v", err)
				time.Sleep(time.Second * 5)
				os.Exit(1)
			} else {
				NotiMessage <- fmt.Sprintf("retrieved challenges")
			}
			err = GetOngoingGames()
			if err != nil {
				ErrorMessage <- fmt.Errorf("unable to retrieve ongoing games: %v", err)
				time.Sleep(time.Second * 5)
				os.Exit(1)
			} else {
				NotiMessage <- fmt.Sprintf("retrieved ongoing games")
			}
			Ready <- struct{}{} //unblock event stream
		}
		close(done)
	}()

	screen.Erase()
	//ticker := time.NewTicker(time.Second)
	ticker := time.NewTicker(time.Millisecond * 500)
	LoadingScreen(screen, loading_msg)

blocking_loop:
	for {
		select {
		case <-Sigs:
			tRow, tCol, _ := OsTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			LoadingScreen(screen, loading_msg)
		case <-ticker.C:
			LoadingScreen(screen, loading_msg)
		case <-done:
			break blocking_loop
		}

	}

	//start windows
	//options := []string{"<<Press 0 to return to welcome screen>>", "<<Press 1 to view / create challenges>>", "<<Press 2 to view / join ongoing games>>", "etc", "quit"}
	options := []string{"new game", "ongoing games", "back", "quit", "test friend", "test ai"}
	op_info := WinInfo{H: (height / 2) - 4, W: width / 2, Y: (height / 2) + 2, X: width / 4}
	options_window, _ := ncurses.NewWindow(op_info.H, op_info.W, op_info.Y, op_info.X)
	windows_array := [1]*ncurses.Window{options_window}
	windows_info_arr := [1]WinInfo{op_info}
	var option_index int = 0
	var selected bool

	DrawLichessWelcome(screen, key, windows_array, windows_info_arr, options)
	for {
		select {
		case <-Sigs:
			tRow, tCol, _ := OsTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			DrawLichessWelcome(screen, key, windows_array, windows_info_arr, options)
		default: //normal character loop here
			key = screen.GetChar()
			option_index, selected = OptionsInput(options_window, key, options, option_index)
			if selected {
				switch option_index {
				case 0: //view / create challenges
					key = OneKey
				case 1: //view / join ongoing games
					//key = TwoKey
				case 2:
					key = ZeroKey //return to welcome screen
				case 3:
					key = CtrlO_Key
				case 4, 5: //testing out lichess requess, skip to wait
					if option_index == 4 {
						CurrentChallenge = testChallenge
					} else {
						CurrentAiChallenge = testAiChallenge
					}

					//LichessScreenHandler(screen, 4)
					func(stdscr *ncurses.Window, option int) {
						switch option {
						case 0:
							//return -1, ZeroKey //go back to welcome screen
						case 1:
							option = LichessWelcome(stdscr)
						case 2:
							option = LichessChallenges(stdscr) //go to challenge screen
						case 3:
							option = CreateLichessGame(stdscr)
						case 4:
							option = WaitForLichessGameResponse(stdscr)
						case 5:
							//return -1, CtrlO_Key //quit game
						case 6:
							screen.Clear()
							l, b := getEvents(EventStreamArr, currentGameID)
							if b {
								for i, e := range l {
									message := fmt.Sprintf("IN STREAM W/ ID: %v", e.Event)
									_, w := screen.MaxYX()
									screen.MovePrint(1+i, w/2, message)
									screen.Refresh()
									time.Sleep(time.Second)
									NotiMessage <- message
								}
							} else {
								os.Exit(2)
							}
							option = LichessGameScreen(stdscr, currentGameID)

						}
						os.Exit(1)
					}(screen, WaitForLichessGameResponse(screen))

				}
			}
			switch key {
			case ZeroKey:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 0)
				}
				return 0
			case OneKey:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 1)
				}
				return 2
			case TwoKey:
			case ThreeKey:
			case CtrlO_Key:
				if !selected {
					_, _ = OptionsInput(options_window, key, options, 4)
				}
				return 5
			}

		}
	}
}

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

func WaitForLichessGameResponse(screen *ncurses.Window) int {
	screen.Clear()

	var localid string
	getgameid := make(chan string, 1)
	ticker := time.NewTicker(time.Second)

	//send the request
	go sendApiRequest(getgameid)
	NotiMessage <- fmt.Sprintf("sent goroutine ")

	load_msg := "requesting game from lichess ... "
	LoadingScreen(screen, load_msg)

	for {
		select {
		case id := <-getgameid:
			NotiMessage <- fmt.Sprintf("event %v retrieved from channel", id)
			localid = id

		case e := <-StreamChannel: // receive event directly
			EventStreamArr = append([]StreamEventType{e}, EventStreamArr...)
			if e.Id == localid {
				NotiMessage <- fmt.Sprintf("(stream channel)  event id: %v", e.Id)
				ErrorMessage <- fmt.Errorf("(stream channel)  event id: %v", e.Id)
				load_msg = fmt.Sprintf("(stream channel) load event: %v!!!", e.Event)
				screen.Clear()
				screen.Refresh()
				LoadingScreen(screen, load_msg)
				currentGameID = e.Id
				//time.Sleep(time.Second * 2)
				screen.Clear()
				screen.Refresh()
				return 6
			} else {
				NotiMessage <- fmt.Sprintf("event %v is not your game, it is %v (waiting for %v)", e.Id, e.Event, localid)
			}
		case <-Sigs:
			NotiMessage <- fmt.Sprintf("resize")
			tRow, tCol, _ := OsTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			LoadingScreen(screen, load_msg)
		case <-ticker.C:
			if localid != "" {
				s, b := containedInEventStream(EventStreamArr, localid)
				if b {
					switch s {
					case "challenge":
						NotiMessage <- fmt.Sprintf("event: %v:%v", localid, s)
						load_msg = fmt.Sprintf("waiting for %v to accept the challenge %v/%v.", CurrentChallenge.DestUser, hostUrl, localid)
						NotiMessage <- load_msg
					case "gameFinish":
						NotiMessage <- fmt.Sprintf("event %v wrong type s", s)
						//time.Sleep(time.Second * 3)
						return 2
					case "challengeCanceled", "challengeDeclined":
						NotiMessage <- fmt.Sprintf("challenge rejected: %v", s)
						ErrorMessage <- fmt.Errorf("challenge rejected: %v", s)
						//time.Sleep(time.Second * 3)
						return 2
					case "gameStart":
						NotiMessage <- fmt.Sprintf("(direct from challenge) event id: %v", localid)
						load_msg = fmt.Sprintf("(direct from challenge) load event: %v!!!", s)
						LoadingScreen(screen, load_msg)
						currentGameID = localid
						//time.Sleep(time.Second * 3)
						screen.Clear()
						screen.Refresh()
						return 6
					}
				} else {
					load_msg = fmt.Sprintf("waiting for %v to accept the challenge %v/%v.", CurrentChallenge.DestUser, hostUrl, localid)
					NotiMessage <- fmt.Sprintf("event %v not in stream ... waiting", localid)
					//time.Sleep(time.Second * 3)
				}
			}
			LoadingScreen(screen, load_msg)
		default:
		}
		screen.Refresh()
	}
}

var streamed_move_sequence chan string

//screen for drawing, initialFEN for custom starting positions, and gameID for verifying streamed events
func LichessGameScreen(screen *ncurses.Window, gameID string) int {
	var key ncurses.Key
	screen.Clear()
	screen.Refresh()
	height, width := screen.MaxYX()
	var currentFEN string

	//call stream board state
	gameStateChan = make(chan BoardState, 1)
	//board_state_sig = make(chan bool, 1)
	streamed_move_sequence = make(chan string, 1)
	go BoardConsumer(gameStateChan, NotiMessage)
	go StreamBoardState(gameStateChan, gameID, ErrorMessage, streamed_move_sequence)

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

	//draw initial screen
	DrawLichessGame(screen, key, windows_array, windows_info_arr)
	for {
		select {
		case <-gameStateChan:
			return 1
		case s := <-streamed_move_sequence:
			ncurses.End()
			NotiMessage <- s
			currentFEN = move_translation(s)
			ncurses.Flash()
			ncurses.Beep()
		case <-Sigs:
			tRow, tCol, _ := OsTermSize()
			ncurses.ResizeTerm(tRow, tCol)
			DrawLichessGame(screen, key, windows_array, windows_info_arr)
		default: //normal character loop here
			//external function calls
			UpdateInput(prompt_window, key)
			if LichessGameLogic(board_window) {
				os.Exit(0) //game done no post lichess screen tho
			}
			DrawBoardWindow(board_window, currentFEN)
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
			key = screen.GetChar()
			switch key {
			case CtrlO_Key, ZeroKey:
				UserInputString = ""
				EnteredPromptStr = ""
				HasEnteredMove = false
				if key == CtrlO_Key {
					return 5
				} else {
					return 1
				}
			}
		}
	}
}

func BoardConsumer(event_chan <-chan BoardState, noti chan<- string) {
	for {
		select {
		case e := <-event_chan:
			//fmt.Printf("consumer: %v %v \n", e.Event, e.Id)
			BoardStreamArr = append([]BoardState{e}, BoardStreamArr...)
			noti <- fmt.Sprintf("event %v", e.Type)
		}
	}
}

func move_translation(sequence string) string {
	sequence_array := strings.Split(sequence, " ")
	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))

	for _, move := range sequence_array {
		if game.Outcome() == chess.NoOutcome {
			if err := game.MoveStr(move); err != nil {
				// handle error
				fmt.Printf("%v\n", err)
			}
			continue
		}
		break
	}
	return game.Position().String()
}

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
					gchan <- fmt.Sprintf("%v", err)
					//os.Exit(1)
				} else {
					NotiMessage <- fmt.Sprintf("posted challenge, id: %v", id)
					gchan <- id
				}

			}
		case 2: //lichess ai
			NotiMessage <- fmt.Sprintf("challenging the lichess ai")
			//TODO: api call CHALLENGE THE AI
			err, id := CreateAiChallenge(CurrentAiChallenge)
			if err != nil {
				ErrorMessage <- err
				gchan <- fmt.Sprintf("%v", err)
				//os.Exit(1)
			} else {
				NotiMessage <- fmt.Sprintf("posted challenge, id: %v", id)
				gchan <- id
			}
		}
	} else {
		ErrorMessage <- fmt.Errorf("no token")
	}
}

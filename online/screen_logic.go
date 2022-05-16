package online

import (
	"fmt"
	"os"
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
						CurrentChallenge = testAiChallenge
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

//screen for drawing, initialFEN for custom starting positions, and gameID for verifying streamed events
func LichessGameScreen(screen *ncurses.Window, gameID string) int {
	CurrentGame = chess.NewGame()
	var key ncurses.Key
	screen.Clear()
	screen.Refresh()
	height, width := screen.MaxYX()
	var currentFEN string

	// ticker := time.NewTicker(time.Second)

	// case <-ticker.C

	//call stream board state
	// gameStateChan = make(chan BoardState, 1)
	gameStateChan = make(chan BoardEvent, 1)
	//board_state_sig = make(chan bool, 1)
	// streamed_move_sequence = make(chan string, 1)
	// go BoardConsumer(gameStateChan, NotiMessage) //the board consumer should be THIS (LichessGameScreen) function
	go StreamBoardState(gameStateChan, gameID, ErrorMessage)

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

	//loop
	for {
		select {
		case b := <-gameStateChan:
			switch b {
			case GameFull:
				// if BoardFullGame.InitialFen == "startpos" {
				// 	currentFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
				// } else {
				// 	currentFEN = BoardFullGame.InitialFen
				// }
				currentFEN = MoveTranslation(BoardFullGame.State.Moves)
				if currentFEN != "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR" {
					fen, _ := chess.FEN(currentFEN)
					CurrentGame = chess.NewGame(fen)
				}
				DrawBoardWindow(board_window, currentFEN)
				DisplayLichessInfoWindow(info_window)
				DisplayLichessHistoryWindow(history_window)

			case GameState:

				currentFEN = MoveTranslation(BoardGameState.Moves)
				DrawBoardWindow(board_window, currentFEN)
				DisplayLichessInfoWindow(info_window)
				DisplayLichessHistoryWindow(history_window)

			case ChatLine:
			case ChatLineSpectator:
			case GameStateResign:

			}
			// ncurses.End()
			// NotiMessage <- s

			board_window.NoutRefresh()
			info_window.NoutRefresh()

			ncurses.Update()
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

			DisplayLichessInfoWindow(info_window)

			DisplayLichessHistoryWindow(history_window)

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

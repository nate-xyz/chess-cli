package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/notnil/chess"
)

func SendChallengeRequest(gchan chan<- string, errchan chan<- error) {
	if UserInfo.ApiToken != "" {
		var err error
		var id string
		switch CurrentChallenge.Type {
		case 0:
			err = CreateSeek(CurrentChallenge)
			if err == nil {
				gchan <- "Seek created. Searching for opponent..."
				return
			}
		case 1: //challenge friend
			err, id = CreateChallenge(CurrentChallenge)
		case 2: //lichess ai
			err, id = CreateAiChallenge(CurrentChallenge)
		}
		if err != nil {
			errchan <- err
		} else {
			gchan <- id
		}
	} else {
		log.Fatal(fmt.Errorf("tried to send api request with no api token"))
	}
}

func WaitForLichessGameResponse() {
	var localid string
	getgameid := make(chan string, 1)
	ticker := time.NewTicker(time.Millisecond * 250)
	var icon_index int = 0
	RequestError := make(chan error, 1)
	var RandomSeek bool = CurrentChallenge.Type == 0
	go SendChallengeRequest(getgameid, RequestError) //send the request
	load_msg := "Requesting new game from Lichess, please wait."
	UpdateLoaderMsg(load_msg)
	for {
		select {
		case e := <-RequestError:
			load_msg = fmt.Sprintf("ERROR: %v", e)
			UpdateLoaderMsg(load_msg)
			go func() {
				time.Sleep(time.Second * 5)
				root.app.QueueUpdate(gotoLichessAfterLogin)
			}()
		case id := <-getgameid:
			if !RandomSeek {
				localid = id
			} else {
				load_msg = id
				UpdateLoaderMsg(load_msg)
			}
		case e := <-StreamChannel: // receive event directly
			EventStreamArr = append([]StreamEventType{e}, EventStreamArr...)
			if !RandomSeek { //friend or AI challenge
				if e.GameID == localid {
					load_msg = "Game started, going to game screen. (FROM ARRAY)"
					UpdateLoaderMsg(load_msg)
					currentGameID = e.GameID
					root.app.QueueUpdate(startNewOnlineGame) //goto game
					return                                   //necessary for QueueUpdate
				}
			} else {
				if e.EventType == "gameStart" && e.Source != "friend" { //TODO: check to make sure match random seek request
					load_msg = "Found random opponent!"
					UpdateLoaderMsg(load_msg)
					go func() {
						currentGameID = e.GameID
						time.Sleep(time.Second)
						root.app.QueueUpdate(startNewOnlineGame)
					}()
				}
			}
		case <-ticker.C:
			if !RandomSeek {
				if localid != "" {
					s, b := containedInEventStream(EventStreamArr, localid)
					if b {
						switch s {
						case "challenge":
							load_msg = fmt.Sprintf("Waiting for %v to accept the challenge %v/%v.", CurrentChallenge.DestUser, hostUrl, localid)
							UpdateLoaderMsg(load_msg)
						case "gameFinish":
							return
						case "challengeCanceled", "challengeDeclined":
							return
						case "gameStart":
							load_msg = "Game started, going to game screen. (FROM STREAM)"
							UpdateLoaderMsg(load_msg)
							currentGameID = localid
							root.app.QueueUpdate(startNewOnlineGame) ///goto game
							return                                   //necessary for QueueUpdate
						}
					} else {
						load_msg = fmt.Sprintf("Waiting for %v to accept the challenge %v/%v.", CurrentChallenge.DestUser, hostUrl, localid)
						UpdateLoaderMsg(load_msg)
					}
				}
			} else {
				e, b := EventContainedInEventStream(EventStreamArr, "gameStart") //TODO: check to make sure match random seek request
				if b && (e.Source != "friend") {
					load_msg = "Found random opponent!"
					UpdateLoaderMsg(load_msg)
					go func() {
						currentGameID = e.GameID
						time.Sleep(time.Second)
						root.app.QueueUpdate(startNewOnlineGame)
					}()
				}
			}
			icon_index = UpdateLoaderIcon(icon_index)
		default:
			UpdateLoaderMsg(load_msg)
		}
	}
}

/*
LichessGame() is called after a gameID string has been retrieved from the event stream by WaitForLichessGameResponse()
This function then starts a board stream with the gameID and loops and modifies the gamestate and view based on events from the board stream.
*/

func LichessGame(gameID string) {
	killGame = make(chan string)
	gameStateChan = make(chan BoardEvent, 1)
	streamDoneErr := make(chan error)
	go StreamBoardState(gameStateChan, streamDoneErr, gameID)

	updateInc := make(chan BothInc)
	stopTicker := make(chan bool)
	ticker1 := time.NewTicker(time.Millisecond * 500)
	ticker2 := time.NewTicker(time.Millisecond)
	go TimerLoop(stopTicker, ticker1, ticker2, updateInc)

	for { //loop
		select {
		case s := <-killGame:
			root.app.QueueUpdate(func() {
				stopTicker <- true
				root.currentLocalGame.Status += "[green]Game has been aborted![white]\n"
				root.currentLocalGame.Status += fmt.Sprintf("Game ended due to %v.\n", s)
				gotoPostOnline()
			})
		case err := <-streamDoneErr:
			root.app.QueueUpdate(func() {
				stopTicker <- true
				root.currentLocalGame.Status += fmt.Sprintf("Game ended due to %v.\n", err)
				gotoPostOnline()
			})
			// root.app.QueueUpdate(gotoLoader)
			// // root.app.QueueUpdate(func() {
			// // 	ticker := time.NewTicker(time.Millisecond * 250)
			// // 	icon_index := 0
			// // 	start := time.Now()
			// // 	for time.Since(start).Seconds() > 5 {
			// // 		select {
			// // 		case <-ticker.C:
			// // 			icon_index = UpdateLoaderIcon(icon_index)
			// // 			UpdateLoaderMsg(fmt.Sprintf("%e", err))
			// // 		default:
			// // 			UpdateLoaderMsg(fmt.Sprintf("%e", err))
			// // 		}
			// // 	}
			// // UpdateLoaderMsg(fmt.Sprintf("%e", err))
			// // time.Sleep(5 * time.Second)

			// // })
			// root.app.QueueUpdate(gotoLichessAfterLogin)
			// return

		case b := <-gameStateChan:
			switch b {
			case GameFull: //full game state json
				currentFEN, err := MoveTranslationToFEN(BoardFullGame.State.Moves) //get the current FEN from the move history
				if err != nil {
					WriteLocal("BoardStateError MoveTranslation", fmt.Sprintf("%v ", err)+currentFEN)
					log.Fatal(err)
					os.Exit(1)
				}
				fen, err := chess.FEN(currentFEN)
				if err != nil {
					WriteLocal("BoardStateError chess FEN", fmt.Sprintf("%v ", err)+currentFEN)
					log.Fatal(err)
					os.Exit(1)
				}

				NewChessGame = chess.NewGame(fen)

				root.app.QueueUpdate(UpdateOnlineTimeView)
				root.app.QueueUpdate(UpdateChessGame)
				root.app.QueueUpdate(UpdateOnline)
				root.app.GetScreen().Beep()

				if BoardFullGame.State.Status != "started" { //the game has ended.
					root.app.QueueUpdate(func() {
						stopTicker <- true
						root.currentLocalGame.Status += fmt.Sprintf("Game ended due to %v.\n", BoardFullGame.State.Status)
						gotoPostOnline()
					})
					return
				}

			case GameState: // game state json
				MoveArr := strings.Split(BoardGameState.Moves, " ")
				MoveCount = len(MoveArr)
				_ = GetCapturePiecesArr(BoardGameState.Moves)
				currentFEN, err := MoveTranslationToFEN(BoardGameState.Moves)
				if err != nil {
					stopTicker <- true
					WriteLocal("BoardStateError MoveTranslation", fmt.Sprintf("%v", err))
					log.Fatal(err)
					os.Exit(1)
				}
				fen, err := chess.FEN(currentFEN)
				if err != nil {
					stopTicker <- true
					WriteLocal("BoardStateError chess FEN", fmt.Sprintf("%v", err))
					log.Fatal(err)
					os.Exit(1)
				}
				root.currentLocalGame.MoveHistoryArray = MoveArr
				NewChessGame = chess.NewGame(fen)

				updateInc <- BothInc{
					btime: int64(BoardGameState.Btime),
					wtime: int64(BoardGameState.Wtime),
				}

				root.app.QueueUpdate(UpdateChessGame)
				root.app.QueueUpdate(UpdateOnline)
				root.app.GetScreen().Beep()
				if BoardGameState.Status != "started" {
					root.app.QueueUpdate(func() {
						stopTicker <- true
						root.currentLocalGame.Status += fmt.Sprintf("Game ended due to %v.\n", BoardGameState.Status)
						gotoPostOnline()
					})
					return
				}

			case ChatLine:
			case ChatLineSpectator:
			case GameStateResign:
				root.app.QueueUpdate(func() {
					stopTicker <- true
					root.currentLocalGame.Status += "Game ended due to resignation.\n"
					gotoPostOnline()
				})
				return
			case EOF:
				root.app.QueueUpdate(func() {
					stopTicker <- true
					root.currentLocalGame.Status += "Game ended due lost connection.\n"
					gotoPostOnline()
				})
				root.app.QueueUpdate(gotoPostOnline)
				return

			}

		}
	}
}

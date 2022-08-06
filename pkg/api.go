package pkg

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func SendChallengeRequest(gchan chan<- string, errchan chan<- error) {
	if api.UserInfo.ApiToken != "" {
		var err error
		var id string
		switch CurrentChallenge.Type {
		case 0:
			err = api.CreateSeek(CurrentChallenge)
			if err == nil {
				gchan <- "Seek created. Searching for opponent..."
				return
			}
		case 1: //challenge friend
			err, id = api.CreateChallenge(CurrentChallenge)
		case 2: //lichess ai
			err, id = api.CreateAiChallenge(CurrentChallenge)
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
				Root.App.QueueUpdate(gotoLichessAfterLogin)
			}()
		case id := <-getgameid:
			if !RandomSeek {
				localid = id
			} else {
				load_msg = id
				UpdateLoaderMsg(load_msg)
			}
		case e := <-StreamChannel: // receive event directly
			api.EventStreamArr = append([]api.StreamEventType{e}, api.EventStreamArr...)
			if !RandomSeek { //friend or AI challenge
				if e.GameID == localid {
					load_msg = "Game started, going to game screen. (FROM ARRAY)"
					UpdateLoaderMsg(load_msg)
					currentGameID = e.GameID
					Root.App.QueueUpdate(startNewOnlineGame) //goto game
					return                                   //necessary for QueueUpdate
				}
			} else {
				if e.EventType == "gameStart" && e.Source != "friend" { //TODO: check to make sure match random seek request
					load_msg = "Found random opponent!"
					UpdateLoaderMsg(load_msg)
					go func() {
						currentGameID = e.GameID
						time.Sleep(time.Second)
						Root.App.QueueUpdate(startNewOnlineGame)
					}()
				}
			}
		case <-ticker.C:
			if !RandomSeek {
				if localid != "" {
					s, b := containedInEventStream(api.EventStreamArr, localid)
					if b {
						switch s {
						case "challenge":
							load_msg = fmt.Sprintf("Waiting for %v to accept the challenge %v/%v.", CurrentChallenge.DestUser, api.HostUrl, localid)
							UpdateLoaderMsg(load_msg)
						case "gameFinish":
							return
						case "challengeCanceled", "challengeDeclined":
							return
						case "gameStart":
							load_msg = "Game started, going to game screen. (FROM STREAM)"
							UpdateLoaderMsg(load_msg)
							currentGameID = localid
							Root.App.QueueUpdate(startNewOnlineGame) ///goto game
							return                                   //necessary for QueueUpdate
						}
					} else {
						load_msg = fmt.Sprintf("Waiting for %v to accept the challenge %v/%v.", CurrentChallenge.DestUser, api.HostUrl, localid)
						UpdateLoaderMsg(load_msg)
					}
				}
			} else {
				e, b := EventContainedInEventStream(api.EventStreamArr, "gameStart") //TODO: check to make sure match random seek request
				if b && (e.Source != "friend") {
					load_msg = "Found random opponent!"
					UpdateLoaderMsg(load_msg)
					go func() {
						currentGameID = e.GameID
						time.Sleep(time.Second)
						Root.App.QueueUpdate(startNewOnlineGame)
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
	gameStateChan = make(chan api.BoardEvent, 1)
	streamDoneErr := make(chan error)
	go api.StreamBoardState(gameStateChan, streamDoneErr, gameID)

	updateInc := make(chan BothInc)
	stopTicker := make(chan bool)
	ticker1 := time.NewTicker(time.Millisecond * 500)
	ticker2 := time.NewTicker(time.Millisecond)
	go TimerLoop(stopTicker, ticker1, ticker2, updateInc)

	for { //loop
		select {
		case s := <-killGame:
			Root.App.QueueUpdate(func() {
				stopTicker <- true
				if s != "GoHome" {
					Root.currentLocalGame.Status += fmt.Sprintf("[green]Game ended due to %v.[white]\n", s)
					gotoPostOnline()
				} else {
					gotoLichessAfterLogin()
				}

			})
		case err := <-streamDoneErr:
			Root.App.QueueUpdate(func() {
				stopTicker <- true

				Root.currentLocalGame.Status += fmt.Sprintf("Game ended due to %v.\n", err)
				gotoPostOnline()
			})
		case b := <-gameStateChan:
			switch b {
			case api.GameFull: //full game state json
				currentFEN, err := MoveTranslationToFEN(api.BoardFullGame.State.Moves) //get the current FEN from the move history
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

				Root.App.QueueUpdate(UpdateOnlineTimeView)
				Root.App.QueueUpdate(UpdateChessGame)
				Root.App.QueueUpdate(UpdateOnline)
				Root.App.GetScreen().Beep()

				if api.BoardFullGame.State.Status != "started" { //the game has ended.
					Root.App.QueueUpdate(func() {
						stopTicker <- true
						Root.currentLocalGame.Status += fmt.Sprintf("Game ended due to %v.\n", api.BoardFullGame.State.Status)
						gotoPostOnline()
					})
					return
				}

			case api.GameState: // game state json
				MoveArr := strings.Split(api.BoardGameState.Moves, " ")
				MoveCount = len(MoveArr)
				_ = GetCapturePiecesArr(api.BoardGameState.Moves)
				currentFEN, err := MoveTranslationToFEN(api.BoardGameState.Moves)
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
				Root.currentLocalGame.MoveHistoryArray = MoveArr
				NewChessGame = chess.NewGame(fen)

				updateInc <- BothInc{
					btime: int64(api.BoardGameState.Btime),
					wtime: int64(api.BoardGameState.Wtime),
				}

				Root.App.QueueUpdate(UpdateChessGame)
				Root.App.QueueUpdate(UpdateOnline)
				Root.App.GetScreen().Beep()

				if (api.BoardGameState.Bdraw && api.Username == api.BoardFullGame.White.ID) || (api.BoardGameState.Wdraw && api.Username == api.BoardFullGame.Black.ID) {
					api.BoardGameState.Bdraw = false
					api.BoardGameState.Wdraw = false
					Root.App.QueueUpdate(func() {
						modal := NewOptionWindow("Your opponent has offered a draw.", "Accept ✅ ", "Reject ❌ ", doAcceptDraw, doRejectDraw)
						Root.OnlineModal = modal
						Root.Online.AddItem(modal, 4, 2, 1, 1, 0, 0, false)
					})
				}

				if (api.BoardGameState.Btakeback && api.Username == api.BoardFullGame.White.ID) || (api.BoardGameState.Wtakeback && api.Username == api.BoardFullGame.Black.ID) {
					api.BoardGameState.Btakeback = false
					api.BoardGameState.Wtakeback = false
					Root.App.QueueUpdate(func() {
						modal := NewOptionWindow("Your opponent has proposed a takeback.", "Accept ✅ ", "Reject ❌ ", doAcceptTakeBack, doRejectTakeBack)
						Root.OnlineModal = modal
						Root.Online.AddItem(modal, 4, 2, 1, 1, 0, 0, false)
					})
				}

				if api.BoardGameState.Status != "started" {
					Root.App.QueueUpdate(func() {
						stopTicker <- true
						if api.BoardGameState.Winner != "" {
							Root.currentLocalGame.Status += fmt.Sprintf("Winner is [blue]%v![white]\n", api.BoardGameState.Winner)
						}
						Root.currentLocalGame.Status += fmt.Sprintf("Game ended due to [red]%v.[white]\n", api.BoardGameState.Status)
						gotoPostOnline()
					})
					return
				}

			case api.ChatLine:
			case api.ChatLineSpectator:
			case api.GameStateResign:
				Root.App.QueueUpdate(func() {
					stopTicker <- true

					Root.currentLocalGame.Status += "Game ended due to resignation.\n"
					gotoPostOnline()
				})
				return
			case api.EOF:
				Root.App.QueueUpdate(func() {
					stopTicker <- true

					Root.currentLocalGame.Status += "Game ended due lost connection.\n"
					gotoPostOnline()
				})
				Root.App.QueueUpdate(gotoPostOnline)
				return

			}

		}
	}
}

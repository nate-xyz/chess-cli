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

	load_msg := "requesting game from lichess ... "
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
			if !RandomSeek {
				if e.GameID == localid {
					load_msg = fmt.Sprintf("(Stream channel) load event: %v!!!", e.EventType)
					UpdateLoaderMsg(load_msg)
					currentGameID = e.GameID

					//goto game
					root.app.QueueUpdate(startNewOnlineGame)
					return

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
							// log.Fatal(fmt.Errorf("event %v wrong type s", s))
							return
						case "challengeCanceled", "challengeDeclined":

							// log.Fatal(fmt.Errorf("challenge rejected: %v", s))

							return
						case "gameStart":

							load_msg = fmt.Sprintf("(Direct from challenge) load event: %v!!!", s)
							UpdateLoaderMsg(load_msg)
							currentGameID = localid

							///goto game

							root.app.QueueUpdate(startNewOnlineGame)
							return

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

func LichessGame(gameID string) {
	gameStateChan = make(chan BoardEvent, 1)
	go StreamBoardState(gameStateChan, gameID)

	//loop
	for {
		select {
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
				root.app.QueueUpdate(UpdateChessGame)
				root.app.QueueUpdate(UpdateOnline)
				root.app.GetScreen().Beep()

				if BoardFullGame.State.Status != "started" {
					//the game has ended.
					root.app.QueueUpdate(func() {
						root.currentLocalGame.Status += fmt.Sprintf("Game ended due to %v.\n", BoardFullGame.State.Status)
						gotoPostOnline()
					})
					return
				}

			case GameState: // game state json
				MoveArr := strings.Split(BoardGameState.Moves, " ")

				currentFEN, err := MoveTranslationToFEN(BoardGameState.Moves)
				if err != nil {
					WriteLocal("BoardStateError MoveTranslation", fmt.Sprintf("%v", err))
					log.Fatal(err)
					os.Exit(1)
				}
				fen, err := chess.FEN(currentFEN)
				if err != nil {
					WriteLocal("BoardStateError chess FEN", fmt.Sprintf("%v", err))
					log.Fatal(err)
					os.Exit(1)
				}
				root.currentLocalGame.MoveHistoryArray = MoveArr
				NewChessGame = chess.NewGame(fen)
				root.app.QueueUpdate(UpdateChessGame)
				root.app.QueueUpdate(UpdateOnline)
				root.app.GetScreen().Beep()
				if BoardGameState.Status != "started" {
					root.app.QueueUpdate(func() {
						root.currentLocalGame.Status += fmt.Sprintf("Game ended due to %v.\n", BoardGameState.Status)
						gotoPostOnline()
					})
					return
				}

			case ChatLine:
			case ChatLineSpectator:
			case GameStateResign:
				root.app.QueueUpdate(func() {
					root.currentLocalGame.Status += "Game ended due to resignation.\n"
					gotoPostOnline()
				})
				return
			case EOF:
				root.app.QueueUpdate(func() {
					root.currentLocalGame.Status += "Game ended due lost connection.\n"
					gotoPostOnline()
				})
				root.app.QueueUpdate(gotoPostOnline)
				return

			}

		}
	}
}

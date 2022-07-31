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
		switch CurrentChallenge.Type {
		case 0: //random seek
			//TODO: api call CREATE A SEEK
		case 1: //challenge friend

			if CurrentChallenge.OpenEnded {
				//TODO: api call CREATE A OPEN END CHALLENGE

			} else {
				// api call CREATE A CHALLENGE
				err, id := CreateChallenge(CurrentChallenge)
				if err != nil {
					errchan <- err
				} else {
					gchan <- id
				}

			}
		case 2: //lichess ai

			//TODO: api call CHALLENGE THE AI
			err, id := CreateAiChallenge(CurrentChallenge)
			if err != nil {
				errchan <- err
			} else {

				gchan <- id
			}
		}
	} else {
		log.Fatal(fmt.Errorf("tried to send api request with no api token"))
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
				MoveArr := strings.Split(BoardFullGame.State.Moves, " ")
				if len(MoveArr)%2 != 0 { //white to move
					if BoardFullGame.White.Name == Username {
						canMove = true
					} else {
						canMove = false
					}
				} else {
					if BoardFullGame.Black.Name == Username {
						canMove = true
					} else {
						canMove = false
					}
				}

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
				if len(MoveArr)%2 != 0 { //white to move
					if BoardFullGame.White.Name == Username {
						canMove = true
					} else {
						canMove = false
					}
				} else {
					if BoardFullGame.Black.Name == Username {
						canMove = true
					} else {
						canMove = false
					}
				}

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

func WaitForLichessGameResponse() {
	var localid string
	getgameid := make(chan string, 1)
	ticker := time.NewTicker(time.Millisecond * 250)
	var icon_index int = 0
	RequestError := make(chan error, 1)
	//send the request
	go SendChallengeRequest(getgameid, RequestError)

	load_msg := "requesting game from lichess ... "
	UpdateLoaderMsg(load_msg)

	for {
		select {
		case e := <-RequestError:
			load_msg = fmt.Sprintf("ERROR: %v", e)
			UpdateLoaderMsg(load_msg)

			go func() {
				time.Sleep(time.Second * 5)
				os.Exit(1)
			}()

		case id := <-getgameid:
			localid = id
		case e := <-StreamChannel: // receive event directly
			EventStreamArr = append([]StreamEventType{e}, EventStreamArr...)
			if e.Id == localid {
				load_msg = fmt.Sprintf("(Stream channel) load event: %v!!!", e.Event)
				UpdateLoaderMsg(load_msg)
				currentGameID = e.Id

				//goto game
				root.app.QueueUpdate(startNewOnlineGame)
				return

			}
		case <-ticker.C:
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
			icon_index = UpdateLoaderIcon(icon_index)
		default:
			UpdateLoaderMsg(load_msg)
		}
	}
}

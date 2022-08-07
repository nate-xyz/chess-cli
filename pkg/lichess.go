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

/*
LichessGame() is called after a gameID string has been retrieved from the event stream by WaitForLichessGameResponse()
This function then starts a board stream with the gameID and loops and modifies the gamestate and view based on events from the board stream.
*/
func (online *OnlineGame) LichessGame(gameID string) {
	killGame = make(chan string)
	gameStateChan = make(chan api.BoardEvent, 1)
	streamDoneErr := make(chan error)
	go api.StreamBoardState(gameStateChan, streamDoneErr, gameID)

	updateInc := make(chan BothInc)
	stopTicker := make(chan bool)
	go online.TimerLoop(stopTicker, updateInc)

	for { //loop
		select {
		case s := <-killGame:
			Root.App.QueueUpdate(func() {
				stopTicker <- true
				if s != "GoHome" {
					Root.gameState.Status += fmt.Sprintf("[green]Game ended due to %v.[white]\n", s)
					online.gotoPostOnline()
				} else {
					gotoLichessAfterLogin()
				}
			})
			return
		case err := <-streamDoneErr:
			Root.App.QueueUpdate(func() {
				stopTicker <- true
				Root.gameState.Status += fmt.Sprintf("Game ended due to %v.\n", err)
				online.gotoPostOnline()
			})
			return
		case b := <-gameStateChan:
			switch b.Type {
			case api.GameFull: //full game state json
				online.Full = b.Full
				currentFEN, err := MoveTranslationToFEN(b.Full.State.Moves) //get the current FEN from the move history
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
				if b.Full.State.Moves == "" {
					Root.gameState.MoveHistoryArray = nil
				} else {
					Root.gameState.MoveHistoryArray = strings.Split(b.Full.State.Moves, " ")
				}
				NewChessGame = chess.NewGame(fen)

				Root.App.QueueUpdate(online.InitTimeView)
				Root.App.QueueUpdate(UpdateChessGame)
				Root.App.QueueUpdate(online.UpdateAll)
				Root.App.GetScreen().Beep()

				if b.Full.State.Status != "started" { //the game has ended.
					go Root.App.QueueUpdate(func() {
						stopTicker <- true
						Root.gameState.Status += fmt.Sprintf("Game ended due to %v.\n", b.Full.State.Status)
						online.gotoPostOnline()
					})
					return
				}

			case api.GameState: // game state json
				online.State = b.State
				MoveArr := strings.Split(b.State.Moves, " ")
				Root.gameState.MoveCount = len(MoveArr)
				_ = GetCapturePiecesArr(b.State.Moves)
				currentFEN, err := MoveTranslationToFEN(b.State.Moves)
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
				if b.State.Moves == "" {
					Root.gameState.MoveHistoryArray = nil
				} else {
					Root.gameState.MoveHistoryArray = MoveArr
				}
				NewChessGame = chess.NewGame(fen)

				updateInc <- BothInc{
					btime: int64(b.State.Btime),
					wtime: int64(b.State.Wtime),
				}

				Root.App.QueueUpdate(UpdateChessGame)
				Root.App.QueueUpdate(online.UpdateAll)
				Root.App.GetScreen().Beep()

				if (b.State.Bdraw && Root.User.Name == online.Full.White.ID) || (b.State.Wdraw && Root.User.Name == online.Full.Black.ID) {
					// b.State.Bdraw = false //this is to reset globals
					// b.State.Wdraw = false
					Root.App.QueueUpdate(func() {
						modal := NewOptionWindow("Your opponent has offered a draw.", "Accept ✅ ", "Reject ❌ ", online.doAcceptDraw, online.doRejectDraw)
						online.PopUp = modal
						online.Grid.AddItem(modal, 4, 2, 1, 1, 0, 0, false)
					})
				}

				if (b.State.Btakeback && Root.User.Name == online.Full.White.ID) || (b.State.Wtakeback && Root.User.Name == online.Full.Black.ID) {
					// api.BoardGameState.Btakeback = false
					// api.BoardGameState.Wtakeback = false
					Root.App.QueueUpdate(func() {
						modal := NewOptionWindow("Your opponent has proposed a takeback.", "Accept ✅ ", "Reject ❌ ", online.doAcceptTakeBack, online.doRejectTakeBack)
						online.PopUp = modal
						online.Grid.AddItem(modal, 4, 2, 1, 1, 0, 0, false)
					})
				}

				if b.State.Status != "started" {
					go Root.App.QueueUpdate(func() {
						stopTicker <- true
						if b.State.Winner != "" {
							Root.gameState.Status += fmt.Sprintf("Winner is [blue]%v![white]\n", b.State.Winner)
						}
						Root.gameState.Status += fmt.Sprintf("Game ended due to [red]%v.[white]\n", b.State.Status)
						online.gotoPostOnline()
					})
					return
				}

			case api.ChatLine:
			case api.ChatLineSpectator:
			case api.GameStateResign:
				online.Resign = b.Resign
				go Root.App.QueueUpdate(func() {
					stopTicker <- true
					Root.gameState.Status += "Game ended due to resignation.\n"
					online.gotoPostOnline()
				})
				return
			case api.EOF:
				Root.App.QueueUpdate(func() {
					stopTicker <- true
					Root.gameState.Status += "Game ended due lost connection.\n"
					online.gotoPostOnline()
				})
				Root.App.QueueUpdate(online.gotoPostOnline)
				return
			}

		}
	}
}

func (online *OnlineGame) TimerLoop(d <-chan bool, bi <-chan BothInc) {
	ticker1 := time.NewTicker(time.Millisecond * 500)
	ticker2 := time.NewTicker(time.Millisecond * 10)

	var Btime int64
	var Wtime int64
	var start time.Time
	for {
		select {
		case b := <-bi:
			Wtime = b.wtime
			Btime = b.btime
			start = time.Now()
		case <-d:
			return
		case <-ticker1.C: //every half second
			if Root.gameState.MoveCount >= 2 {
				var currB int64 = Btime
				var currW int64 = Wtime
				if Root.gameState.MoveCount%2 == 0 {
					currW -= time.Since(start).Milliseconds()
				} else {
					currB -= time.Since(start).Milliseconds()
				}
				online.LiveUpdateTime(currB, currW)
				Root.App.QueueUpdateDraw(func() {}, online.UserTimer, online.OppTimer)
			}
		case <-ticker2.C: //every ms
			if Root.gameState.MoveCount >= 2 {
				var currB int64 = Btime
				var currW int64 = Wtime
				if Root.gameState.MoveCount%2 == 0 {
					currW -= time.Since(start).Milliseconds()
				} else {
					currB -= time.Since(start).Milliseconds()
				}

				if currB < 10000 || currW < 10000 { //start drawing millis when less than ten seconds
					online.LiveUpdateTime(currB, currW)
					Root.App.QueueUpdateDraw(func() {}, online.UserTimer, online.OppTimer)
				}
			}

		}
	}
}

//consumes stream events for a user after stream for a user started after login
func StreamConsumer(EventChannel <-chan api.StreamEventType) {
	for {
		e := <-EventChannel
		if api.Online {
			n, _ := Root.nav.GetFrontPanel()
			if n == "lichesswelcome" {
				// Root.wonline.UpdateTitle(fmt.Sprintf("%s: %s", e.EventType, e.GameID))
				// time.Sleep(time.Second)
				Root.wonline.UpdateTitle("")
			}
			// if n == "ponline" && e.GameID == currentGameID {
			// 	Root.App.QueueUpdate(func() {
			// 		modal := NewOptionWindow(
			// 			fmt.Sprintf("%s: %s", e.EventType, e.GameID),
			// 			"Ok ✅ ", "Go Home 🏠",
			// 			func() {
			// 				Root.ponline.Grid.RemoveItem(Root.ponline.PopUp)
			// 			},
			// 			gotoLichessAfterLogin)
			// 		Root.ponline.PopUp = modal
			// 		Root.ponline.Grid.AddItem(modal, 2, 2, 1, 1, 0, 0, false)
			// 	})
			// }
		}
		EventStreamArr = append([]api.StreamEventType{e}, EventStreamArr...)
	}
}

package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
)

func UpdateLichessTitle() {
	var titlestr string = LichessTitle
	if Online {
		titlestr += " 🟢"
	} else {
		titlestr += " ⚪"

	}
	if UserInfo.ApiToken == "" {
		titlestr += "[red]\nNot logged into lichess.[blue]\nPlease login through your browser.[white]\n"
	} else {
		titlestr += fmt.Sprintf("\n[green]Logged in[white] as: [blue]%s, %s[white]", Username, UserEmail)
	}

	root.LichessTitle.SetText(titlestr)
	root.app.QueueUpdateDraw(func() {}, root.LichessTitle)
}

func UpdateOnline() {
	UpdateLegalMoves()
	UpdateGameHistory(root.OnlineHistory)
	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)
	//UpdateGameStatus(root.OnlineStatus)
	UpdateOnlineStatus(root.OnlineStatus)
	UpdateUserInfo()
	root.app.QueueUpdateDraw(func() {})
}

func UpdateOnlineStatus(s *cv.TextView) {
	var status string
	var ratestr string

	if BoardFullGame.Rated {
		ratestr = "Rated"
	} else {
		ratestr = "Casual"
	}

	if BoardFullGame.Speed == "correspondence" {
		status += fmt.Sprintf("\n\n%v • %v\n",
			ratestr,
			strings.Title(BoardFullGame.Speed))
	} else {
		status += fmt.Sprintf("\n\n%v+%v • %v • %v\n",
			timeFormat(int64(BoardFullGame.Clock.Initial)),
			BoardFullGame.Clock.Increment/1000,
			ratestr,
			strings.Title(BoardFullGame.Speed))
	}

	if root.currentLocalGame.Game.Position().Turn() == chess.White {
		status += "White's turn. \n\n"
	} else {
		status += "Black's turn. \n\n"
	}

	status += root.currentLocalGame.Status
	root.currentLocalGame.Status = ""

	s.SetText(status)
}

func UpdateUserInfo() {
	var OppName string
	var UserString string = "\n[blue]%v[white]"
	var OppString string = "[red]%v[white]"

	if BoardFullGame.White.Name == Username {
		OppName = BoardFullGame.Black.Name
		UserString = UserString + " (white)\n"
		OppString = OppString + " (black)\n"
	} else {
		OppName = BoardFullGame.White.Name
		OppString = OppString + " (white)\n"
		UserString = UserString + " (black)\n"
	}

	if CurrentChallenge.Type == 2 {
		OppName = "🤖"
	}

	UserString = fmt.Sprintf(UserString, Username)
	OppString = fmt.Sprintf(OppString, OppName)

	root.OnlineInfoUser.SetText(UserString)
	root.OnlineInfoOppo.SetText(OppString)
}

func UpdateLoaderIcon(i int) int {
	if i > 7 {
		i = 0
	}
	loadingstr := "\n\t ... [red]" + KnightIconMap[i] + "[white] ... \t\n"

	root.LoaderIcon.SetText(loadingstr + loadingstr + loadingstr + loadingstr + loadingstr + loadingstr)

	i++
	if i > 7 {
		i = 0
	}

	root.app.QueueUpdateDraw(func() {})

	return i
}

func UpdateLoaderMsg(msg string) {

	root.LoaderMsg.SetText(msg)
	root.app.QueueUpdateDraw(func() {})

}

func OnlineGameDoMove(move string) error {

	go func() {
		err := root.currentLocalGame.Game.MoveStr(move)
		if err == nil {
			UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)
			root.app.QueueUpdateDraw(func() {}, root.OnlineBoard)
		}
	}()

	//do the move
	err := MakeMove(currentGameID, move)
	if err != nil {
		return err
	}
	root.app.GetScreen().Beep()
	// err = root.currentLocalGame.Game.MoveStr(root.currentLocalGame.NextMove)

	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)

	root.currentLocalGame.NextMove = "" //clear the next move
	UpdateGameStatus(root.OnlineStatus)

	//check if game is done

	// if root.currentLocalGame.Game.Outcome() != chess.NoOutcome {
	// 	gotoPostOnline()
	// }
	//MOVED TO LICHESS GAME, (wait for api stream)
	return nil
}

func UpdateChessGame() {
	root.currentLocalGame.Game = NewChessGame
}

func UpdateOnlineTimeView() {
	b := int64(BoardFullGame.State.Btime)
	w := int64(BoardFullGame.State.Wtime)
	LiveUpdateOnlineTimeView(b, w)
}

func TimerLoop(d <-chan bool, v *time.Ticker, t *time.Ticker, bi <-chan BothInc) {
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
		case <-v.C: //every half second
			var currB int64 = Btime
			var currW int64 = Wtime
			if MoveCount >= 2 {
				if MoveCount%2 == 0 {
					currW -= time.Since(start).Milliseconds()
				} else {
					currB -= time.Since(start).Milliseconds()
				}
				LiveUpdateOnlineTimeView(currB, currW)
				root.app.QueueUpdateDraw(func() {}, root.OnlineTimeUser, root.OnlineTimeOppo)
			}
		case <-t.C: //every ms
			var currB int64 = Btime
			var currW int64 = Wtime
			if MoveCount >= 2 {
				if currB < 10000 || currW < 1000 { //start drawing millis when less than ten seconds
					if MoveCount%2 == 0 {
						currW -= time.Since(start).Milliseconds()
					} else {
						currB -= time.Since(start).Milliseconds()
					}
					LiveUpdateOnlineTimeView(currB, currW)
					root.app.QueueUpdateDraw(func() {}, root.OnlineTimeUser, root.OnlineTimeOppo)
				}
			}

		}
	}
}

func LiveUpdateOnlineTimeView(b int64, w int64) { //MoveCount
	if BoardFullGame.State.Btime == math.MaxInt32 {
		return
	}

	var White bool = BoardFullGame.White.Name == Username
	var UserStr string
	var OppoStr string

	if BoardFullGame.Speed == "correspondence" {
		if White {
			UserStr += (timeFormat(w))
			OppoStr += (timeFormat(b))
		} else {
			UserStr += (timeFormat(b))
			OppoStr += (timeFormat(w))
		}
	} else {
		binc := int64(BoardGameState.Binc)
		winc := int64(BoardGameState.Winc)
		if White {
			UserStr += (timeFormat(w) + fmt.Sprintf("+%d", winc/1000))
			OppoStr += (timeFormat(b) + fmt.Sprintf("+%d", binc/1000))
		} else {
			UserStr += (timeFormat(b) + fmt.Sprintf("+%d", binc/1000))
			OppoStr += (timeFormat(w) + fmt.Sprintf("+%d", winc/1000))
		}
	}

	if MoveCount > 1 {
		if MoveCount%2 == 0 {
			UserStr += " ⏲️\t"
			root.OnlineTimeUser.SetBackgroundColor(tc.ColorSeaGreen)
			root.OnlineTimeOppo.SetBackgroundColor(tc.ColorBlack.TrueColor())
		} else {
			OppoStr += " ⏲️\t"
			root.OnlineTimeOppo.SetBackgroundColor(tc.ColorSeaGreen)
			root.OnlineTimeUser.SetBackgroundColor(tc.ColorBlack.TrueColor())
		}
	}

	root.OnlineTimeUser.SetText(UserStr)
	root.OnlineTimeOppo.SetText(OppoStr)
}

func OnlineTableHandler(row, col int) {
	selectedCell := translateSelectedCell(row, col, BoardFullGame.White.Name == Username)
	if LastSelectedCell.Alg == selectedCell { //toggle selected status of this cell

		root.OnlineBoard.Select(100, 100)
		LastSelectedCell = PiecePosition{-1, -1, "", true, ""}
	} else { //try to do move

		todoMove := LastSelectedCell.Alg + selectedCell
		if contains(root.currentLocalGame.LegalMoves, todoMove) {
			err := OnlineGameDoMove(todoMove)
			if err != nil {
				root.currentLocalGame.Status += fmt.Sprintf("%v", err)
				UpdateGameStatus(root.OnlineStatus)
			}
		}
		//check if select is empty for updateBoard
		symbol := root.OnlineBoard.GetCell(row, col).GetText()
		LastSelectedCell = PiecePosition{row, col, selectedCell, (symbol == EmptyChar), symbol}
	}
	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)
}

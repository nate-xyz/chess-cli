package main

import (
	"fmt"
	"math"
	"strings"
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
	UpdateGameHistory(root.OnlineHistory)
	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)
	UpdateGameStatus(root.OnlineStatus)
	root.app.QueueUpdateDraw(func() {})
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

func LiveUpdateOnlineTimeView(b int64, w int64) {
	var timestr string

	binc := int64(BoardGameState.Binc)
	winc := int64(BoardGameState.Winc)

	var Opp string

	if BoardFullGame.White.Name == Username {
		Opp = BoardFullGame.Black.Name
	} else {
		Opp = BoardFullGame.White.Name
	}

	if CurrentChallenge.Type == 2 {
		Opp = "🤖"

	}

	if BoardFullGame.White.Name == Username {
		if BoardFullGame.State.Btime == math.MaxInt32 {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n\n[blue]%v[white]\n(white)",
				Opp,
				Username)
		} else if BoardFullGame.Speed == "correspondence" {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n%v\n\n\n%v\n[blue]%v[white]\n(white)",
				Opp,
				timeFormat(b),
				timeFormat(w),
				Username)
		} else {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n%v+%v\n\n\n%v+%v\n[blue]%v[white]\n(white)",
				Opp,
				timeFormat(b),
				timeFormat(binc),
				timeFormat(w),
				timeFormat(winc),
				Username)
		}
	} else {
		if BoardFullGame.State.Btime == math.MaxInt32 {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n\n[blue]%v[white]\n(white)",
				Opp,
				Username)
		} else if BoardFullGame.Speed == "correspondence" {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n%v\n\n\n%v\n[blue]%v[white]\n(white)",
				Opp,
				timeFormat(w),
				timeFormat(b),
				Username)
		} else {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n%v+%v\n\n\n%v+%v\n[blue]%v[white]\n(white)",
				Opp,
				timeFormat(w),
				timeFormat(winc),
				timeFormat(b),
				timeFormat(binc),
				Username)
		}
	}
	var ratestr string
	if BoardFullGame.Rated {
		ratestr = "Rated"
	} else {
		ratestr = "Casual"
	}
	if BoardFullGame.Speed == "correspondence" {
		timestr += fmt.Sprintf("\n\n%v • %v\n",
			ratestr,
			strings.Title(BoardFullGame.Speed))
	} else {
		timestr += fmt.Sprintf("\n\n%v+%v • %v • %v\n",
			timeFormat(int64(BoardFullGame.Clock.Initial)),
			timeFormat(int64(BoardFullGame.Clock.Increment)),
			ratestr,
			strings.Title(BoardFullGame.Speed))
	}
	root.OnlineTime.SetText(timestr)
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

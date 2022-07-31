package main

import "fmt"

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
	root.app.QueueUpdateDraw(func() {
	})
}

func UpdateOnline() {
	UpdateGameHistory(root.OnlineHistory)
	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)
	UpdateGameStatus(root.OnlineStatus)
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

func OnlineGameDoMove() {
	//TODO: handle game end
	//do the move
	err := MakeMove(currentGameID, root.currentLocalGame.NextMove)
	if err != nil {
		return
	}

	// err = root.currentLocalGame.Game.MoveStr(root.currentLocalGame.NextMove)

	//clear the next move

	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)

	root.currentLocalGame.NextMove = ""
	UpdateGameStatus(root.OnlineStatus)
	root.app.GetScreen().Beep()

	//check if game is done

	// if root.currentLocalGame.Game.Outcome() != chess.NoOutcome {
	// 	gotoPostOnline()
	// }
	//MOVED TO LICHESS GAME, (wait for api stream)
}

func UpdateChessGame() {
	root.currentLocalGame.Game = NewChessGame
}

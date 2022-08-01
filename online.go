package main

import (
	"fmt"
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

func UpdateOnlineTimeView() {
	b := int64(BoardFullGame.State.Btime)
	w := int64(BoardFullGame.State.Wtime)
	LiveUpdateOnlineTimeView(b, w)
}

func LiveUpdateOnlineTimeView(b int64, w int64) {
	var timestr string

	binc := int64(BoardGameState.Binc)
	winc := int64(BoardGameState.Winc)

	if BoardFullGame.White.Name == Username {
		if BoardFullGame.Speed == "unlimited" {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n\n[blue]%v[white]\n(white)",
				BoardFullGame.Black.Name,
				BoardFullGame.White.Name)
		} else if BoardFullGame.Speed == "correspondence" {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n%v\n\n\n%v\n[blue]%v[white]\n(white)",
				BoardFullGame.Black.Name,
				timeFormat(b),
				timeFormat(w),
				BoardFullGame.White.Name)
		} else {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n%v+%v\n\n\n%v+%v\n[blue]%v[white]\n(white)",
				BoardFullGame.Black.Name,
				timeFormat(b),
				timeFormat(binc),
				timeFormat(w),
				timeFormat(winc),
				BoardFullGame.White.Name)
		}

	} else {
		if BoardFullGame.Speed == "unlimited" {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n\n[blue]%v[white]\n(white)",
				BoardFullGame.White.Name,
				BoardFullGame.Black.Name)
		} else if BoardFullGame.Speed == "correspondence" {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n%v\n\n\n%v\n[blue]%v[white]\n(white)",
				BoardFullGame.White.Name,
				timeFormat(w),
				timeFormat(b),
				BoardFullGame.Black.Name)
		} else {
			timestr += fmt.Sprintf("\n[red]%v[white]\n(black)\n%v+%v\n\n\n%v+%v\n[blue]%v[white]\n(white)",
				BoardFullGame.White.Name,
				timeFormat(w),
				timeFormat(winc),
				timeFormat(b),
				timeFormat(binc),
				BoardFullGame.Black.Name)
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

func timeFormat(time int64) string {
	if time == 0 {
		return "0"
	}
	ms := time % 1000
	time /= 1000
	sec := time % 60
	time /= 60
	min := time % 60
	hours := time / 60
	if hours == 0 && min == 0 && sec <= 10 {
		return fmt.Sprintf("%02d:%02d:%03d", min, sec, ms)
	} else if hours == 0 {
		return fmt.Sprintf("%02d:%02d", min, sec)
	}
	days := hours / 24
	hours = hours % 24
	if days == 0 {
		return fmt.Sprintf("%d Hours", hours)
	} else if hours == 0 {
		return fmt.Sprintf("%d Days", days)
	} else {
		return fmt.Sprintf("%d Days %d Hours", days, hours)
	}
}

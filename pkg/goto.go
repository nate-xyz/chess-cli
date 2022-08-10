package pkg

import (
	"fmt"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func gotoWelcome() {
	Root.Switch("welcome")

}

func startNewLocalGame() {
	Root.NewGame()
	Root.Switch("localgame")

	DrawBoard(Root.lgame.Board, Root.gameState.Game.Position().Turn() == chess.White)

	Root.lgame.UpdateStatus()
}

func gotoPostLocal() {
	Root.gameState.Status += Root.gameState.Game.Outcome().String()
	Root.pgame.UpdateResult()
	DrawMoveHistory(Root.ponline.History)
	DrawBoard(Root.ponline.Board, true)
	Root.Switch("postlocal")
}

func gotoLichess() {
	err := Root.Login()
	if err != nil {
		Root.App.QueueUpdate(func() {
			Root.wonline.UpdateTitle(fmt.Sprintf("%v", err))
		})
	} else {
		Root.RefreshAll()
	}
	Root.Switch("lichesswelcome")
}

func gotoLichessAfterLogin() {
	Root.Switch("lichesswelcome")
}
func gotoLichessFromGame() {
	killGame <- "GoHome"
}

func startNewOnlineGame() {
	Root.NewGame()
	Root.Switch("onlinegame")
	go Root.ongame.LichessGame(currentGameID)
}

func (online *OnlineGame) gotoPostOnline() {
	Root.ponline.UpdateResult()
	DrawMoveHistory(Root.ponline.History)
	DrawBoard(Root.ponline.Board, true)
	Root.Switch("postonline")
	Root.RefreshAll()
}

func gotoChallengeConstruction() {
	Root.Switch("challenge")
}

func gotoLoaderFromChallenge() {
	CurrentChallenge = newChallenge
	newChallenge = api.CreateChallengeType{}
	Root.Switch("loader")           //goto loader
	go WaitForLichessGameResponse() //thread to update loading screen and wait for board event
}

func gotoOngoing() {
	err := Root.User.GetOngoing()
	if err != nil {
		Root.wonline.UpdateTitle(fmt.Sprintf("Ongoing Games: %v", err))
		if Root.User.OngoingGames == nil {
			return
		}
	}
	Root.ongoing.UpdateList()
	Root.Switch("ongoing")
}

func gotoChallenges() {
	err := Root.User.GetChallenges()
	if err != nil {
		Root.wonline.UpdateTitle(fmt.Sprintf("Challenges: %v", err))
		if Root.User.IncomingChallenges == nil && Root.User.OutgoingChallenges == nil {
			return
		}
	}
	Root.challenges.UpdateList()
	Root.Switch("listchallenge")
}

func gotoSaved() {
	checkForSavedGames()
	Root.sgame.UpdateList()
	Root.Switch("saved")
}

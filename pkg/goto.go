package pkg

import (
	"fmt"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func doNothing() {}

func gotoWelcome() {
	Root.nav.SetCurrentPanel("welcome")

}

func startNewLocalGame() {
	game := new(LocalGame)
	game.Init()
	Root.currentLocalGame = game
	Root.nav.SetCurrentPanel("localgame")

	UpdateBoard(Root.Board, Root.currentLocalGame.Game.Position().Turn() == chess.White)

	UpdateGameStatus(Root.Status)
}

func gotoPostLocal() {
	Root.currentLocalGame.Status += Root.currentLocalGame.Game.Outcome().String()
	UpdateResult(Root.PostStatus)
	UpdateGameHistory(Root.PostHistory)
	UpdateBoard(Root.PostBoard, true)
	Root.nav.SetCurrentPanel("postlocal")
}

func gotoLichess() {
	err := LichessLogin()
	if err != nil {
		UpdateLichessTitle(fmt.Sprintf("%v", err))
	} else {
		UpdateLichessTitle("")
	}
	Root.nav.SetCurrentPanel("lichesswelcome")
}

func LichessLogin() error {
	err := api.PerformOAuth()
	if err != nil {
		return err
	}
	err = api.GetLichessUserInfo()
	if err != nil {
		return err
	}
	return nil
}

func gotoLichessAfterLogin() {
	Root.nav.SetCurrentPanel("lichesswelcome")
}
func gotoLichessFromGame() {
	killGame <- "GoHome"
}

func startNewOnlineGame() {
	Root.currentLocalGame = new(LocalGame)
	Root.currentLocalGame.Init()
	Root.nav.SetCurrentPanel("onlinegame")
	go LichessGame(currentGameID)
}

func gotoPostOnline() {
	UpdateResult(Root.OnlinePostStatus)
	UpdateGameHistory(Root.OnlinePostHistory)
	UpdateBoard(Root.OnlinePostBoard, true)
	Root.nav.SetCurrentPanel("postonline")
	Root.App.QueueUpdateDraw(func() {})
}

func gotoChallengeConstruction() {
	Root.nav.SetCurrentPanel("challenge")
}

func TestFriend() {
	CurrentChallenge = testChallenge   //bypass creating a challenge
	Root.nav.SetCurrentPanel("loader") //goto loader
	go WaitForLichessGameResponse()    //thread to update loading screen and wait for board event
}

func TestAI() {
	CurrentChallenge = testAiChallenge //bypass creating a challenge
	Root.nav.SetCurrentPanel("loader") //goto loader
	go WaitForLichessGameResponse()    //thread to update loading screen and wait for board event
}

func gotoLoaderFromChallenge() {
	CurrentChallenge = newChallenge
	newChallenge = api.CreateChallengeType{}
	Root.nav.SetCurrentPanel("loader") //goto loader
	go WaitForLichessGameResponse()    //thread to update loading screen and wait for board event
}

func gotoOngoing() {
	err := api.GetOngoingGames()
	if err != nil {
		UpdateLichessTitle(fmt.Sprintf("Ongoing Games: %v", err))
		if api.OngoingGames == nil {
			return
		}
	}
	UpdateOngoingList()
	Root.nav.SetCurrentPanel("ongoing")
}

func gotoChallenges() {
	err := api.GetChallenges()
	if err != nil {
		UpdateLichessTitle(fmt.Sprintf("Challenges: %v", err))
		if api.IncomingChallenges == nil && api.OutgoingChallenges == nil {
			return
		}
	}
	UpdateChallengeList()
	Root.nav.SetCurrentPanel("listchallenge")
}

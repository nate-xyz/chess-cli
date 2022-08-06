package pkg

import (
	"fmt"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func doNothing() {}

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
	err := LichessLogin()
	if err != nil {
		Root.wonline.UpdateTitle(fmt.Sprintf("%v", err))
	} else {
		Root.wonline.UpdateTitle("")
	}
	Root.Switch("lichesswelcome")
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

func gotoPostOnline() {
	Root.pgame.UpdateResult()
	DrawMoveHistory(Root.ponline.History)
	DrawBoard(Root.ponline.Board, true)
	Root.Switch("postonline")
	Root.App.QueueUpdateDraw(func() {})
}

func gotoChallengeConstruction() {
	Root.Switch("challenge")
}

func TestFriend() {
	CurrentChallenge = testChallenge //bypass creating a challenge
	Root.Switch("loader")            //goto loader
	go WaitForLichessGameResponse()  //thread to update loading screen and wait for board event
}

func TestAI() {
	CurrentChallenge = testAiChallenge //bypass creating a challenge
	Root.Switch("loader")              //goto loader
	go WaitForLichessGameResponse()    //thread to update loading screen and wait for board event
}

func gotoLoaderFromChallenge() {
	CurrentChallenge = newChallenge
	newChallenge = api.CreateChallengeType{}
	Root.Switch("loader")           //goto loader
	go WaitForLichessGameResponse() //thread to update loading screen and wait for board event
}

func gotoOngoing() {
	err := api.GetOngoingGames()
	if err != nil {
		Root.wonline.UpdateTitle(fmt.Sprintf("Ongoing Games: %v", err))
		if api.OngoingGames == nil {
			return
		}
	}
	Root.ongoing.UpdateList()
	Root.Switch("ongoing")
}

func gotoChallenges() {
	err := api.GetChallenges()
	if err != nil {
		Root.wonline.UpdateTitle(fmt.Sprintf("Challenges: %v", err))
		if api.IncomingChallenges == nil && api.OutgoingChallenges == nil {
			return
		}
	}
	Root.challenges.UpdateList()
	Root.Switch("listchallenge")
}

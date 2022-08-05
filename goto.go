package main

import (
	"fmt"

	"github.com/notnil/chess"
)

func doNothing() {

}

func gotoWelcome() {
	root.nav.SetCurrentPanel("welcome")

}

func startNewLocalGame() {
	game := new(LocalGame)
	game.Init()
	root.currentLocalGame = game
	root.nav.SetCurrentPanel("localgame")

	UpdateBoard(root.Board, root.currentLocalGame.Game.Position().Turn() == chess.White)

	UpdateGameStatus(root.Status)
}

func gotoPostLocal() {
	root.currentLocalGame.Status += root.currentLocalGame.Game.Outcome().String()
	UpdateResult(root.PostStatus)
	UpdateGameHistory(root.PostHistory)
	UpdateBoard(root.PostBoard, true)
	root.nav.SetCurrentPanel("postlocal")
}

func gotoLichess() {
	err := LichessLogin()
	if err != nil {
		UpdateLichessTitle(fmt.Sprintf("%v", err))
	} else {
		UpdateLichessTitle("")
	}
	root.nav.SetCurrentPanel("lichesswelcome")
}

func LichessLogin() error {
	err := PerformOAuth()
	if err != nil {
		return err
	}
	err = GetLichessUserInfo()
	if err != nil {
		return err
	}
	return nil
}

func gotoLichessAfterLogin() {

	root.nav.SetCurrentPanel("lichesswelcome")
}

// func gotoLoader() {
// 	root.nav.SetCurrentPanel("loader")
// 	time.Sleep(2 * time.Second)
// }

func startNewOnlineGame() {
	root.currentLocalGame = new(LocalGame)
	root.currentLocalGame.Init()
	root.nav.SetCurrentPanel("onlinegame")

	go LichessGame(currentGameID)
	//UpdateOnline()

	// UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)
	// UpdateGameStatus(root.Status)

}

func gotoPostOnline() {
	// root.currentLocalGame.Status += root.currentLocalGame.Game.Outcome().String()
	var outcome string = root.currentLocalGame.Game.Outcome().String()
	if outcome == "1-0" {
		root.currentLocalGame.Status += fmt.Sprintf("White (%v) wins!\n", BoardFullGame.White.Name)
	} else if outcome == "0-1" {
		root.currentLocalGame.Status += fmt.Sprintf("Black (%v) wins!\n", BoardFullGame.Black.Name)
	}
	UpdateGameHistory(root.OnlinePostHistory)
	UpdateResult(root.OnlinePostStatus)
	UpdateBoard(root.OnlinePostBoard, true)
	root.nav.SetCurrentPanel("postonline")
	root.app.QueueUpdateDraw(func() {})
}

func gotoChallengeConstruction() {
	root.nav.SetCurrentPanel("challenge")
}

func TestFriend() {
	CurrentChallenge = testChallenge   //bypass creating a challenge
	root.nav.SetCurrentPanel("loader") //goto loader
	go WaitForLichessGameResponse()    //thread to update loading screen and wait for board event
}

func TestAI() {
	CurrentChallenge = testAiChallenge //bypass creating a challenge
	root.nav.SetCurrentPanel("loader") //goto loader
	go WaitForLichessGameResponse()    //thread to update loading screen and wait for board event
}

func gotoLoaderFromChallenge() {
	CurrentChallenge = newChallenge
	newChallenge = CreateChallengeType{}
	root.nav.SetCurrentPanel("loader") //goto loader
	go WaitForLichessGameResponse()    //thread to update loading screen and wait for board event
}

func gotoOngoing() {
	err := GetOngoingGames()
	if err != nil {
		UpdateLichessTitle(fmt.Sprintf("Ongoing Games: %v", err))
		if OngoingGames == nil {
			return
		}
	}
	UpdateOngoingList()
	root.nav.SetCurrentPanel("ongoing")
}

func gotoChallenges() {
	err := GetChallenges()
	if err != nil {
		UpdateLichessTitle(fmt.Sprintf("Challenges: %v", err))
		if IncomingChallenges == nil && OutgoingChallenges == nil {
			return
		}
	}
	UpdateChallengeList()
	root.nav.SetCurrentPanel("listchallenge")
}

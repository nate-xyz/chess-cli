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
	PerformOAuth()
	GetLichessUserInfo()
	UpdateLichessTitle()

	root.nav.SetCurrentPanel("lichesswelcome")
}

func gotoLichessAfterLogin() {

	root.nav.SetCurrentPanel("lichesswelcome")
}

func startNewOnlineGame() {
	root.currentLocalGame = new(LocalGame)
	root.currentLocalGame.Init()
	root.nav.SetCurrentPanel("onlinegame")

	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)

	UpdateGameStatus(root.Status)
	go LichessGame(currentGameID)
}

func gotoPostOnline() {
	// root.currentLocalGame.Status += root.currentLocalGame.Game.Outcome().String()
	var WhiteWon bool = root.currentLocalGame.Game.Outcome().String()[0] == '1'
	if WhiteWon {
		root.currentLocalGame.Status += fmt.Sprintf("White (%v) wins!\n", BoardFullGame.White.Name)
	} else {
		root.currentLocalGame.Status += fmt.Sprintf("Black (%v) wins!\n", BoardFullGame.Black.Name)
	}
	UpdateGameHistory(root.OnlinePostHistory)
	UpdateResult(root.OnlinePostStatus)
	UpdateBoard(root.OnlinePostBoard, true)
	root.nav.SetCurrentPanel("postonline")
}

func gotoChallengeConstruction() {
	root.nav.SetCurrentPanel("challenge")
}

func TestFriend() {
	CurrentChallenge = testChallenge //bypass creating a challenge

	root.nav.SetCurrentPanel("loader") //goto loader
	go WaitForLichessGameResponse()    //thread to update loading screen and wait for board event
}

func gotoLoaderFromChallenge() {
	CurrentChallenge = newChallenge
	newChallenge = CreateChallengeType{}
	root.nav.SetCurrentPanel("loader") //goto loader
	go WaitForLichessGameResponse()    //thread to update loading screen and wait for board event
}

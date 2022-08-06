package pkg

import (
	"fmt"

	"github.com/nate-xyz/chess-cli/api"
)

func UpdateChessGame() {
	Root.gameState.Game = NewChessGame
}

func (on *OnlineGame) doAbort() {
	err := api.AbortGame(currentGameID)
	if err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]%v[white]\n", err)
		Root.ongame.UpdateStatus()
		return
	}
	killGame <- "abort"
}

func (on *OnlineGame) doResign() {
	err := api.ResignGame(currentGameID)
	if err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]%v[white]\n", err)
		Root.ongame.UpdateStatus()
		return
	}
	killGame <- "resign"
}

func (on *OnlineGame) doOfferDraw() {
	err := api.HandleDraw(currentGameID, true)
	if err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]%v[white]\n", err)
		Root.ongame.UpdateStatus()
		return
	}

}

func (on *OnlineGame) doAcceptDraw() {
	err := api.HandleDraw(currentGameID, true)
	if err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]%v[white]\n", err)
		on.UpdateStatus()
		on.Grid.RemoveItem(on.PopUp)
		return
	}
	on.Grid.RemoveItem(on.PopUp)
}

func (on *OnlineGame) doRejectDraw() {
	err := api.HandleDraw(currentGameID, false)
	if err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]%v[white]\n", err)
		on.UpdateStatus()
		on.Grid.RemoveItem(on.PopUp)
		return
	}
	on.Grid.RemoveItem(on.PopUp)
}

func (on *OnlineGame) doProposeTakeBack() {
	err := api.HandleTakeback(currentGameID, true)
	if err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]%v[white]\n", err)
		on.UpdateStatus()
		return
	}
}

func (on *OnlineGame) doAcceptTakeBack() {
	err := api.HandleTakeback(currentGameID, true)
	if err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]%v[white]\n", err)
		on.UpdateStatus()
		on.Grid.RemoveItem(on.PopUp)
		return
	}
	on.Grid.RemoveItem(on.PopUp)
}

func (on *OnlineGame) doRejectTakeBack() {
	err := api.HandleTakeback(currentGameID, false)
	if err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]%v[white]\n", err)
		on.UpdateStatus()
		on.Grid.RemoveItem(on.PopUp)
		return
	}
	on.Grid.RemoveItem(on.PopUp)
}

package pkg

import (
	"context"
	"errors"
	"fmt"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func gotoWelcome() {
	if loginCancel != nil {
		loginCancel() // leaving the Lichess screen abandons a login in progress
	}
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
	DrawMoveHistory(Root.pgame.History)
	DrawBoard(Root.pgame.Board, true)
	Root.Switch("postlocal")
}

// Login state; UI goroutine only.
var (
	loginURL    string             // shown while waiting for the browser login
	loginCancel context.CancelFunc // non-nil while a login is in progress
)

func loggedIn() bool { return Root.User != nil && Root.User.Name != "" }

// gotoLichess opens the Lichess screen, logging in first if needed. The login
// runs in the background so the UI stays responsive while the browser is open.
func gotoLichess() {
	Root.Switch("lichesswelcome")
	if loggedIn() || loginCancel != nil {
		Root.wonline.UpdateTitle("")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	loginCancel = cancel
	Root.wonline.UpdateTitle("\n[yellow]Logging in…[-]")
	go func() {
		user := new(Login)
		err := user.Init(ctx, func(url string) {
			Root.App.QueueUpdateDraw(func() {
				loginURL = url
				Root.wonline.UpdateTitle("")
			})
		})
		Root.App.QueueUpdateDraw(func() {
			loginCancel, loginURL = nil, ""
			cancel()
			if err != nil {
				if ctx.Err() == nil || errors.Is(err, api.ErrLoginCancelled) {
					Root.wonline.UpdateTitle(fmt.Sprintf("\n[red]Login failed:[-] %v\nGo back and choose Lichess again to retry.", err))
				}
				return
			}
			Root.User = user
			Root.wonline.UpdateTitle("")
		})
	}()
}

// requireLogin wraps a Lichess menu action so it only runs once logged in.
func requireLogin(f func()) func() {
	return func() {
		if !loggedIn() {
			Root.wonline.UpdateTitle("\n[yellow]Finish logging in first.[-]")
			return
		}
		f()
	}
}

func gotoLichessAfterLogin() {
	Root.Switch("lichesswelcome")
	Root.wonline.UpdateTitle("")
}

func gotoLichessFromGame() {
	Root.ongame.endGame("")
}

func startNewOnlineGame() {
	Root.NewGame()
	og := Root.ongame
	if og.cancel != nil {
		og.cancel()
	}
	og.ctx, og.cancel = context.WithCancel(context.Background())
	og.ended, og.canClaim = false, false
	og.Full, og.State, og.Gone = api.StreamBoardGameFull{}, api.StreamBoardGameState{}, api.StreamOpponentGone{}
	og.hidePopup()
	og.Chat.Clear()
	og.Board.ClearSelection()
	og.UserTimer.SetText("")
	og.OppTimer.SetText("")
	og.UpdateList()
	Root.Switch("onlinegame")
	go og.LichessGame(currentGameID)
}

func (online *OnlineGame) gotoPostOnline() {
	Root.ponline.UpdateResult()
	DrawMoveHistory(Root.ponline.History)
	DrawBoard(Root.ponline.Board, online.IsWhite())
	Root.Switch("postonline")
	Root.RefreshAll()
}

func gotoChallengeConstruction() {
	Root.Switch("challenge")
}

func gotoLoaderFromChallenge() {
	c := newChallenge
	newChallenge = api.CreateChallengeType{}
	startRequest(c)
}

// fetchThen runs a Lichess fetch in the background, then show on the UI
// goroutine, reporting errors in the Lichess screen title.
func fetchThen(what string, fetch func() error, show func()) {
	Root.wonline.UpdateTitle(fmt.Sprintf("\n[gray]Loading %s…[-]", what))
	go func() {
		err := fetch()
		Root.App.QueueUpdateDraw(func() {
			if err != nil {
				Root.wonline.UpdateTitle(fmt.Sprintf("\n[red]%s: %v[-]", what, err))
				return
			}
			Root.wonline.UpdateTitle("")
			show()
		})
	}()
}

func gotoOngoing() {
	fetchThen("ongoing games", Root.User.GetOngoing, func() {
		Root.ongoing.UpdateList()
		Root.Switch("ongoing")
	})
}

func gotoChallenges() {
	fetchThen("challenges", Root.User.GetChallenges, func() {
		Root.challenges.UpdateList()
		Root.Switch("listchallenge")
	})
}

func gotoSaved() {
	checkForSavedGames()
	Root.sgame.UpdateList()
	Root.Switch("saved")
}

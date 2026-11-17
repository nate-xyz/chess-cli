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
	Root.lgame.stopEngine()
	previewPieces("") // first-run setup and Options preview styles; home shows the real setting
	resetWelcome()
	refreshNews()
	Root.Switch("welcome")
}

func startNewLocalGame() {
	Root.lgame.stopEngine()
	Root.NewGame()
	Root.lgame.startClock()
	Root.lgame.tickClock()
	Root.Switch("localgame")
	playSound(soundStart)

	DrawBoard(Root.lgame.Board, Root.gameState.Game.Position().Turn() == chess.White)

	Root.lgame.UpdateStatus()
}

func gotoPostLocal() {
	game := Root.gameState.Game
	gotoPostLocalResult(winnerOf(game.Outcome()), gameMethod(game))
}

// gotoPostLocalResult ends a local game with a result (e.g. on time).
func gotoPostLocalResult(winner chess.Color, method string) {
	Root.lgame.clock.on = false
	clockActive.Store(false)
	Root.gameState.Status = resultText(winner, method) + "\n\n" + Root.gameState.Status
	g := Root.lgame
	white := g.whiteAtBottom()
	me := chess.NoColor // two players: any win is a victory
	if g.engine != nil {
		me = g.engineColor.Other()
	}
	playSound(resultSound(winner, me))
	g.stopEngine()
	Root.pgame.show("postlocal", g.Board, g.OppInfo, g.UserInfo, white, winner, method)
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
	Root.wonline.UpdateTitle("\n[yellow]Logging in…[-]")
	startLogin(func() { Root.wonline.UpdateTitle("") }, func(err error) {
		if err != nil {
			Root.wonline.UpdateTitle(fmt.Sprintf("\n[red]Login failed:[-] %v\nGo back and choose Lichess again to retry.", err))
			return
		}
		Root.wonline.UpdateTitle("")
	})
}

// startLogin logs in to Lichess in the background (reusing a saved token if
// valid). onURL runs when the browser login link is ready (it's in loginURL);
// done runs when it finishes, unless it was cancelled by leaving the screen.
// Both run on the UI goroutine.
func startLogin(onURL func(), done func(err error)) {
	if loginCancel != nil {
		return // already in progress
	}
	ctx, cancel := context.WithCancel(context.Background())
	loginCancel = cancel
	go func() {
		user := new(Login)
		err := user.Init(ctx, func(url string) {
			Root.App.QueueUpdateDraw(func() {
				loginURL = url
				onURL()
			})
		})
		Root.App.QueueUpdateDraw(func() {
			loginCancel, loginURL = nil, ""
			cancel()
			if err == nil {
				Root.User = user
			} else if ctx.Err() != nil && !errors.Is(err, api.ErrLoginCancelled) {
				return // cancelled by leaving the screen
			}
			done(err)
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
	og.ended, og.canClaim, og.lowWarned = false, false, false
	og.Full, og.State, og.Gone = api.StreamBoardGameFull{}, api.StreamBoardGameState{}, api.StreamOpponentGone{}
	og.hidePopup()
	og.Chat.Clear()
	og.Board.ClearSelection()
	og.UserTimer.SetText("")
	og.OppTimer.SetText("")
	og.UpdateList()
	Root.Switch("onlinegame")
	playSound(soundStart)
	go og.LichessGame(currentGameID)
}

func (online *OnlineGame) gotoPostOnline() {
	winner, method := colorFromName(online.State.Winner), online.State.Status
	if method != "aborted" && method != "noStart" {
		Root.gameState.Status = resultText(winner, method) + "\n\n" + Root.gameState.Status
	}
	if method == "aborted" || method == "noStart" {
		method = "" // no result marks on the board
	} else {
		me := chess.Black
		if online.IsWhite() {
			me = chess.White
		}
		playSound(resultSound(winner, me))
	}
	Root.ponline.show("postonline", online.Board, online.OppInfo, online.UserInfo, online.IsWhite(), winner, method)
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

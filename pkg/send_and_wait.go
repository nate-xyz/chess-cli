package pkg

import (
	cv "codeberg.org/tslocum/cview"
	"context"
	"fmt"
	"time"

	"github.com/nate-xyz/chess-cli/api"
)

// loaderCancel withdraws the request currently shown on the loading screen.
// Only touched on the UI goroutine.
var loaderCancel context.CancelFunc

// startRequest sends a new game request (seek, challenge or AI game), shows the
// loading screen, and moves to the game once Lichess starts it.
func startRequest(c api.CreateChallengeType) {
	CurrentChallenge = c
	ctx, cancel := context.WithCancel(context.Background())
	loaderCancel = cancel
	Root.loader.Message.SetText("Requesting a new game from Lichess…")
	Root.Switch("loader")
	events, unsubscribe := subscribeEvents() // subscribe before sending, so no gameStart is missed
	go waitForGame(ctx, c, events, unsubscribe)
}

// cancelLoader is bound to Esc on the loading screen.
func cancelLoader() {
	if loaderCancel != nil {
		loaderCancel()
	}
}

func waitForGame(ctx context.Context, c api.CreateChallengeType, events <-chan api.StreamEventType, unsubscribe func()) {
	defer unsubscribe()
	show := func(msg string) {
		Root.App.QueueUpdateDraw(func() { Root.loader.Message.SetText(msg) })
	}
	start := func(gameID string) {
		Root.App.QueueUpdateDraw(func() {
			loaderCancel = nil
			currentGameID = gameID
			startNewOnlineGame()
		})
	}
	// leave shows a final message, then returns to the Lichess menu (or
	// straight away if the user presses Esc).
	leave := func(msg string) {
		show(msg + "\n\n[gray](returning to the menu, Esc to go now)[-]")
		select {
		case <-time.After(4 * time.Second):
		case <-ctx.Done():
		}
		Root.App.QueueUpdateDraw(func() {
			loaderCancel = nil
			gotoLichessAfterLogin()
		})
	}

	var (
		pending *api.Pending
		gameID  string
		err     error
	)
	switch c.Type {
	case 0: // random opponent
		pending, err = api.CreateSeek(c)
		if err == nil && c.TimeOption == api.TimeCorrespondence {
			leave("Correspondence seek created. When someone accepts, the game will appear under Ongoing Games.")
			return
		}
		show("Seek created. Searching for an opponent…\n\n[gray]Esc to cancel[-]")
	case 1: // a specific user
		pending, err = api.CreateChallenge(c)
		if err == nil {
			gameID = pending.ID
			show(fmt.Sprintf("Waiting for [blue]%s[-] to accept the challenge.\n%s/%s\n\n[gray]Esc to cancel[-]",
				cv.Escape(c.DestUser), api.HostUrl, gameID))
		}
	case 2: // Lichess AI: the game starts immediately
		gameID, err = api.CreateAiChallenge(c)
		if err == nil {
			start(gameID)
			return
		}
	}
	if err != nil {
		leave(fmt.Sprintf("[red]Could not create the game:[-] %v", err))
		return
	}
	defer pending.Cancel()

	done := pending.Done()
	spinner := time.NewTicker(250 * time.Millisecond)
	defer spinner.Stop()
	icon := 0
	for {
		select {
		case <-ctx.Done():
			pending.Cancel()
			Root.App.QueueUpdateDraw(func() {
				loaderCancel = nil
				gotoLichessAfterLogin()
			})
			return

		case e := <-events:
			switch {
			case e.EventType == "gameStart" && gameID != "" && e.GameID == gameID:
				start(e.GameID)
				return
			case e.EventType == "gameStart" && c.Type == 0 && e.Source != "friend" && e.Source != "ai":
				start(e.GameID)
				return
			case e.EventType == "challengeDeclined" && e.GameID == gameID:
				leave(fmt.Sprintf("[red]%s declined the challenge.[-]", cv.Escape(c.DestUser)))
				return
			case e.EventType == "challengeCanceled" && e.GameID == gameID:
				leave("The challenge was cancelled.")
				return
			}

		case result, ok := <-done:
			if !ok {
				done = nil
				continue
			}
			switch result {
			case "accepted":
				show("Accepted! Starting the game…")
				done = nil // gameStart arrives on the event stream
			case "declined":
				leave(fmt.Sprintf("[red]%s declined the challenge.[-]", cv.Escape(c.DestUser)))
				return
			case "canceled":
				leave("The challenge was cancelled.")
				return
			default:
				if c.Type == 0 {
					// a real-time seek's stream closes once paired; wait for gameStart
					done = nil
					continue
				}
				leave("[red]Lost the connection to Lichess; the request was withdrawn.[-]")
				return
			}

		case <-spinner.C:
			i := icon
			Root.App.QueueUpdateDraw(func() { Root.loader.DrawIcon(i) })
			icon = (icon + 1) % len(KnightIconMap)
		}
	}
}

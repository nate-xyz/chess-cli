package pkg

import (
	"context"
	"fmt"
	"strings"

	cv "codeberg.org/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
)

// IsWhite reports whether the logged-in user plays white in this game.
// Lichess ids are lower-case usernames, so compare case-insensitively.
func (og *OnlineGame) IsWhite() bool {
	me := Root.User.Name
	return strings.EqualFold(og.Full.White.ID, me) || strings.EqualFold(og.Full.White.Name, me)
}

// opponent returns the other side of the game.
func (og *OnlineGame) opponent() api.GamePlayer {
	if og.IsWhite() {
		return og.Full.Black
	}
	return og.Full.White
}

// background runs a Lichess call off the UI goroutine and reports a failure in
// the status panel. onOK, if set, runs on the UI goroutine after success.
func (og *OnlineGame) background(what string, call func() error, onOK func()) {
	go func() {
		err := call()
		Root.App.QueueUpdateDraw(func() {
			if err != nil {
				Root.gameState.Status += fmt.Sprintf("[red]%s failed: %v[-]\n", what, err)
				og.UpdateStatus()
				return
			}
			if onOK != nil {
				onOK()
			}
		})
	}()
}

// The game ends through the stream (Lichess sends the final status), so these
// only send the request.
func (og *OnlineGame) doAbort() {
	og.background("Abort", func() error { return api.AbortGame(currentGameID) }, nil)
}
func (og *OnlineGame) doResign() {
	og.background("Resign", func() error { return api.ResignGame(currentGameID) }, nil)
}

func (og *OnlineGame) doClaimVictory() {
	og.background("Claim victory", func() error { return api.ClaimVictory(currentGameID) }, nil)
}

func (og *OnlineGame) doClaimDraw() {
	og.background("Claim draw", func() error { return api.ClaimDraw(currentGameID) }, nil)
}

func (og *OnlineGame) doOfferDraw() {
	og.background("Draw offer", func() error { return api.HandleDraw(currentGameID, true) }, func() {
		Root.gameState.Status += "Draw offered.\n"
		og.UpdateStatus()
	})
}

func (og *OnlineGame) doProposeTakeBack() {
	og.background("Takeback", func() error { return api.HandleTakeback(currentGameID, true) }, func() {
		Root.gameState.Status += "Takeback proposed.\n"
		og.UpdateStatus()
	})
}

func (og *OnlineGame) answerOffer(what string, call func() error) {
	og.hidePopup()
	og.popupKind = what // don't reopen for the same offer while the answer is in flight
	og.background(what, call, nil)
}

func (og *OnlineGame) doAcceptDraw() {
	og.answerOffer("draw", func() error { return api.HandleDraw(currentGameID, true) })
}
func (og *OnlineGame) doRejectDraw() {
	og.answerOffer("draw", func() error { return api.HandleDraw(currentGameID, false) })
}
func (og *OnlineGame) doAcceptTakeBack() {
	og.answerOffer("takeback", func() error { return api.HandleTakeback(currentGameID, true) })
}
func (og *OnlineGame) doRejectTakeBack() {
	og.answerOffer("takeback", func() error { return api.HandleTakeback(currentGameID, false) })
}

// addChat appends a chat line. UI goroutine only.
func (og *OnlineGame) addChat(user, text, room string) {
	name := "[blue]" + cv.Escape(user) + "[-]"
	if strings.EqualFold(user, Root.User.Name) {
		name = "[green]" + cv.Escape(user) + "[-]"
	} else if user == "lichess" {
		name = "[gray]lichess[-]"
	}
	if room == "spectator" {
		name += " [gray](spectator)[-]"
	}
	fmt.Fprintf(og.Chat, "%s: %s\n", name, cv.Escape(text))
	og.Chat.ScrollToEnd()
}

func (og *OnlineGame) sendChat(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	gameID := currentGameID
	go func() {
		if err := api.WriteChat(gameID, "player", text); err != nil {
			Root.App.QueueUpdateDraw(func() {
				fmt.Fprintf(og.Chat, "[red]message not sent: %v[-]\n", err)
			})
		}
	}()
}

// loadChatHistory shows messages sent before we connected (e.g. when resuming
// a game). New messages arrive on the game stream.
func (og *OnlineGame) loadChatHistory(ctx context.Context, gameID string) {
	msgs, err := api.GetChat(gameID)
	if err != nil || ctx.Err() != nil {
		return
	}
	Root.App.QueueUpdateDraw(func() {
		if ctx.Err() != nil {
			return
		}
		og.Chat.Clear()
		for _, m := range msgs {
			og.addChat(m.User, m.Text, "player")
		}
	})
}

// rematch challenges the same opponent with the same settings and colours
// swapped (or starts a new game at the same AI level).
func (og *OnlineGame) rematch() {
	f := og.Full
	if f.ID == "" {
		return
	}
	c := api.CreateChallengeType{
		Variant:   f.Variant.Key,
		Rated:     fmt.Sprint(f.Rated),
		RatedBool: f.Rated,
		Color:     "white",
	}
	if og.IsWhite() {
		c.Color = "black"
	}
	if f.Variant.Key == "fromPosition" {
		c.Fen = f.InitialFen
	}
	switch {
	case f.Speed == "correspondence":
		c.TimeOption = api.TimeCorrespondence
		c.Days = fmt.Sprint(max(1, f.DaysPerTurn))
	case f.Clock.Initial > 0 || f.Clock.Increment > 0:
		c.TimeOption = api.TimeRealTime
		c.ClockLimit = fmt.Sprint(f.Clock.Initial / 1000)
		c.ClockIncrement = fmt.Sprint(f.Clock.Increment / 1000)
		c.MinTurn = float64(f.Clock.Initial) / 60000
	default:
		c.TimeOption = api.TimeUnlimited
	}
	opp := og.opponent()
	if opp.AiLevel > 0 {
		c.Type = 2
		c.Level = fmt.Sprint(opp.AiLevel)
	} else {
		c.Type = 1
		c.DestUser = opp.Name
	}
	startRequest(c)
}

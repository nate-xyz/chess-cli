package pkg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	cv "codeberg.org/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// streamRetries is how many times in a row the game stream may fail to
// reconnect before the game screen gives up.
const streamRetries = 5

/*
LichessGame runs one online game: it streams the game from Lichess
(reconnecting if the connection drops), applies each update on the UI
goroutine, and ticks the clocks. It returns when the game ends or the user
leaves (both cancel online.ctx via endGame).
*/
func (online *OnlineGame) LichessGame(gameID string) {
	ctx := online.ctx
	events := make(chan api.BoardEvent, 16)
	streamErr := make(chan error, 1)
	go runGameStream(ctx, gameID, events, streamErr)
	go online.loadChatHistory(ctx, gameID)

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-streamErr:
			Root.App.QueueUpdateDraw(func() { online.endGame(fmt.Sprintf("the connection to Lichess was lost (%v)", err)) })
			return
		case <-tick.C:
			Root.App.QueueUpdateDraw(online.tick)
		case ev := <-events:
			Root.App.QueueUpdateDraw(func() { online.handleEvent(ev) })
			if ev.Type == api.EOF {
				return
			}
		}
	}
}

// runGameStream keeps the game stream open, reconnecting after drops. Lichess
// resends the full game state on every connect, so reconnecting is seamless.
func runGameStream(ctx context.Context, gameID string, events chan<- api.BoardEvent, fatal chan<- error) {
	failures := 0
	for {
		errs := make(chan error, 1)
		connected := time.Now()
		api.StreamBoardState(ctx, events, errs, gameID)
		if ctx.Err() != nil {
			return
		}
		var err error
		select {
		case err = <-errs:
		default:
			return // the stream ended normally (EOF was delivered)
		}
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.Status < 500 && apiErr.Status != 429 {
			fatal <- err // e.g. game not found: retrying won't help
			return
		}
		if time.Since(connected) > 30*time.Second {
			failures = 0 // it was a working connection, not a failing reconnect
		}
		failures++
		if failures > streamRetries {
			fatal <- err
			return
		}
		Root.App.QueueUpdateDraw(func() {
			Root.gameState.Status += "[yellow]Connection dropped, reconnecting…[-]\n"
			Root.ongame.UpdateStatus()
		})
		select {
		case <-time.After(time.Duration(failures) * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

// handleEvent applies one game-stream event. UI goroutine only.
func (online *OnlineGame) handleEvent(ev api.BoardEvent) {
	if online.ended {
		return
	}
	switch ev.Type {
	case api.GameFull:
		online.Full = ev.Full
		online.applyState(ev.Full.State)
	case api.GameState:
		online.applyState(ev.State)
	case api.ChatLine:
		online.addChat(ev.Chat.Username, ev.Chat.Text, ev.Chat.Room)
	case api.OpponentGone:
		online.Gone = ev.Gone
		online.goneSince = time.Now()
		if ev.Gone.Gone {
			Root.gameState.Status += "[yellow]Your opponent left the game.[-]\n"
		} else {
			Root.gameState.Status += "[green]Your opponent is back.[-]\n"
		}
		online.UpdateList()
		online.UpdateStatus()
	case api.EOF:
		if online.State.Status == "" || online.State.Status == "started" {
			online.endGame("the game stream closed")
		}
	}
}

// applyState updates the board, history, clocks and offers from a game state.
func (online *OnlineGame) applyState(st api.StreamBoardGameState) {
	prevMoves := online.State.Moves
	online.State = st
	game, err := gameFromMoves(online.Full.InitialFen, st.Moves)
	if err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]Could not follow the game: %v[-]\n", err)
		online.UpdateStatus()
		return
	}
	Root.gameState.Game = game
	if st.Moves == "" {
		Root.gameState.MoveHistoryArray = nil
	} else {
		Root.gameState.MoveHistoryArray = strings.Split(st.Moves, " ")
	}
	Root.gameState.MoveCount = len(Root.gameState.MoveHistoryArray)
	_ = GetCapturePiecesArr(st.Moves)
	online.clockW, online.clockB = int64(st.Wtime), int64(st.Btime)
	online.clockAt = time.Now()

	online.UpdateAll()
	if st.Moves != prevMoves && prevMoves != "" {
		Root.App.GetScreen().Beep()
	}

	online.updateOffers(st)

	if st.Status != "" && st.Status != "started" {
		msg := fmt.Sprintf("the game ended: [red]%s[-]", st.Status)
		if st.Winner != "" {
			msg = fmt.Sprintf("%s wins (%s)", caser.String(st.Winner), st.Status)
		}
		online.endGame(msg)
	}
}

// updateOffers shows or hides the popup for the opponent's draw/takeback offer.
func (online *OnlineGame) updateOffers(st api.StreamBoardGameState) {
	white := online.IsWhite()
	drawOffered := (white && st.Bdraw) || (!white && st.Wdraw)
	takebackOffered := (white && st.Btakeback) || (!white && st.Wtakeback)
	switch {
	case drawOffered && online.popupKind != "draw":
		online.showPopup("draw", NewOptionWindow("Your opponent offers a draw.", "Accept ✅ ", "Decline ❌ ", online.doAcceptDraw, online.doRejectDraw))
	case takebackOffered && online.popupKind != "takeback":
		online.showPopup("takeback", NewOptionWindow("Your opponent proposes a takeback.", "Accept ✅ ", "Decline ❌ ", online.doAcceptTakeBack, online.doRejectTakeBack))
	case !drawOffered && !takebackOffered:
		online.hidePopup()
	}
}

func (online *OnlineGame) showPopup(kind string, p *cv.Flex) {
	online.hidePopup()
	online.PopUp = p
	online.popupKind = kind
	online.Side.AddItemAtIndex(2, p, 7, 0, false) // under the opponent's name and clock
}

func (online *OnlineGame) hidePopup() {
	if online.PopUp != nil {
		online.Side.RemoveItem(online.PopUp)
		online.PopUp = nil
	}
	online.popupKind = ""
}

// tick redraws the running clock and the opponent-gone countdown. UI goroutine.
func (online *OnlineGame) tick() {
	if online.ended || online.Full.ID == "" {
		return
	}
	w, b := online.clockW, online.clockB
	if Root.gameState.MoveCount >= 2 {
		elapsed := time.Since(online.clockAt).Milliseconds()
		if Root.gameState.Game.Position().Turn() == chess.White {
			w = max(0, w-elapsed)
		} else {
			b = max(0, b-elapsed)
		}
	}
	online.LiveUpdateTime(b, w)
	if online.Gone.Gone {
		before := online.canClaim
		online.canClaim = time.Since(online.goneSince) >= time.Duration(online.Gone.ClaimWinInSeconds)*time.Second
		if online.canClaim != before {
			online.UpdateList()
			online.UpdateStatus()
		}
	}
}

// endGame stops the game screen and shows the result. UI goroutine only.
func (online *OnlineGame) endGame(reason string) {
	if online.ended {
		return
	}
	online.ended = true
	online.cancel()
	online.hidePopup()
	if reason == "" {
		gotoLichessAfterLogin()
		return
	}
	Root.gameState.Status += "Game over: " + reason + "\n"
	online.gotoPostOnline()
}

// gameFromMoves rebuilds a game from its starting position and UCI moves.
// Keeping the full move list (rather than just a FEN) gives correct
// repetition detection and the last-move highlight.
func gameFromMoves(initialFEN, moves string) (*chess.Game, error) {
	opts := []func(*chess.Game){chess.UseNotation(chess.UCINotation{})}
	if initialFEN != "" && initialFEN != "startpos" {
		fen, err := chess.FEN(initialFEN)
		if err != nil {
			return nil, err
		}
		opts = append(opts, fen)
	}
	game := chess.NewGame(opts...)
	for _, m := range strings.Fields(moves) {
		if err := game.MoveStr(m); err != nil {
			return nil, fmt.Errorf("move %q: %w", m, err)
		}
	}
	return game, nil
}

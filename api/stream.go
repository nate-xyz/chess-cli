package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var (
	online        atomic.Bool
	streamRunning atomic.Bool
)

// IsOnline reports whether the account event stream is currently connected.
func IsOnline() bool { return online.Load() }

// StartEventStream makes sure the account's event stream (game starts and
// finishes, challenges) is running, sending events to events. Call it after
// logging in; it does nothing if the stream is already running. The stream
// reconnects with backoff if the connection drops, and stops if the token is
// rejected (a later login starts it again).
func StartEventStream(events chan<- StreamEventType) {
	if streamRunning.CompareAndSwap(false, true) {
		ctx, cancel := context.WithCancel(context.Background())
		streamMu.Lock()
		streamStop = cancel
		streamMu.Unlock()
		go streamEvents(ctx, events)
	}
}

// StopEventStream stops the account event stream (e.g. when logging out).
func StopEventStream() {
	streamMu.Lock()
	defer streamMu.Unlock()
	if streamStop != nil {
		streamStop()
		streamStop = nil
	}
}

var (
	streamMu   sync.Mutex
	streamStop context.CancelFunc
)

func streamEvents(ctx context.Context, events chan<- StreamEventType) {
	defer streamRunning.Store(false)
	backoff := time.Second
	for ctx.Err() == nil {
		start := time.Now()
		err := streamEventsOnce(ctx, events)
		online.Store(false)
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
			return // token revoked or expired; reconnecting won't help
		}
		if time.Since(start) > time.Minute {
			backoff = time.Second // it was a healthy connection; retry quickly
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, time.Minute)
	}
}

func streamEventsOnce(ctx context.Context, events chan<- StreamEventType) error {
	body, err := openStream(ctx, "GET", "/api/stream/event", nil)
	if err != nil {
		return err
	}
	defer body.Close()
	online.Store(true)

	d := json.NewDecoder(body)
	for {
		var ev struct {
			Type string `json:"type"`
			Game struct {
				ID     string `json:"id"`
				Source string `json:"source"`
			} `json:"game"`
			Challenge struct {
				ID string `json:"id"`
			} `json:"challenge"`
		}
		if err := decode(d, &ev); err != nil {
			if err == io.EOF {
				return io.ErrUnexpectedEOF
			}
			return err
		}
		out := StreamEventType{EventType: ev.Type}
		switch ev.Type {
		case "gameStart", "gameFinish":
			out.GameID, out.Source = ev.Game.ID, ev.Game.Source
		case "challenge", "challengeCanceled", "challengeDeclined":
			out.GameID = ev.Challenge.ID
		default:
			continue // unknown event types are ignored, not fatal
		}
		select {
		case events <- out:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// StreamBoardState streams one game's events (full state, moves, chat, the
// opponent leaving) until the game stream ends, an error occurs, or ctx is
// cancelled. Exactly one of these happens at the end: an EOF event is sent,
// an error is sent on errs, or ctx is done.
func StreamBoardState(ctx context.Context, events chan<- BoardEvent, errs chan<- error, gameID string) {
	send := func(ev BoardEvent) bool {
		select {
		case events <- ev:
			return true
		case <-ctx.Done():
			return false
		}
	}
	fail := func(err error) {
		select {
		case errs <- err:
		case <-ctx.Done():
		}
	}

	body, err := openStream(ctx, "GET", "/api/board/game/stream/"+url.PathEscape(gameID), nil)
	if err != nil {
		fail(err)
		return
	}
	defer body.Close()

	d := json.NewDecoder(body)
	for {
		var raw json.RawMessage
		if err := decode(d, &raw); err != nil {
			if ctx.Err() != nil {
				return
			}
			if err == io.EOF {
				send(BoardEvent{Type: EOF})
				return
			}
			fail(err)
			return
		}
		var head struct {
			Type string `json:"type"`
		}
		if err := unmarshal(raw, &head); err != nil {
			fail(err)
			return
		}
		var ev BoardEvent
		switch head.Type {
		case "gameFull":
			ev.Type = GameFull
			err = unmarshal(raw, &ev.Full)
		case "gameState":
			ev.Type = GameState
			err = unmarshal(raw, &ev.State)
		case "chatLine":
			ev.Type = ChatLine
			err = unmarshal(raw, &ev.Chat)
		case "opponentGone":
			ev.Type = OpponentGone
			err = unmarshal(raw, &ev.Gone)
		default:
			continue
		}
		if err != nil {
			fail(err)
			return
		}
		if !send(ev) {
			return
		}
	}
}

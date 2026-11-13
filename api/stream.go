package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

var (
	online        atomic.Bool
	streamStarted atomic.Bool
)

// IsOnline reports whether the account event stream is currently connected.
func IsOnline() bool { return online.Load() }

// EventStreamStarted reports whether StreamEvent has been told to start.
func EventStreamStarted() bool { return streamStarted.Load() }

// StreamEvent streams the account's events (game starts/finishes, challenges)
// once ready fires, and keeps reconnecting with backoff if the connection
// drops. It only gives up if the token is rejected.
func StreamEvent(events chan<- StreamEventType, ready <-chan struct{}) {
	<-ready
	streamStarted.Store(true)
	backoff := time.Second
	for {
		start := time.Now()
		err := streamEventsOnce(events)
		online.Store(false)
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
			return // token revoked or expired; reconnecting won't help
		}
		if time.Since(start) > time.Minute {
			backoff = time.Second // it was a healthy connection; retry quickly
		}
		time.Sleep(backoff)
		backoff = min(backoff*2, time.Minute)
	}
}

func streamEventsOnce(events chan<- StreamEventType) error {
	body, err := openStream(context.Background(), "GET", "/api/stream/event", nil)
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
		if err := d.Decode(&ev); err != nil {
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
		events <- out
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
		if err := d.Decode(&raw); err != nil {
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
		if err := json.Unmarshal(raw, &head); err != nil {
			fail(err)
			return
		}
		var ev BoardEvent
		switch head.Type {
		case "gameFull":
			ev.Type = GameFull
			err = json.Unmarshal(raw, &ev.Full)
		case "gameState":
			ev.Type = GameState
			err = json.Unmarshal(raw, &ev.State)
		case "chatLine":
			ev.Type = ChatLine
			err = json.Unmarshal(raw, &ev.Chat)
		case "opponentGone":
			ev.Type = OpponentGone
			err = json.Unmarshal(raw, &ev.Gone)
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

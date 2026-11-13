package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

// Time options used by CreateChallengeType.TimeOption.
const (
	TimeRealTime       = 0
	TimeCorrespondence = 1
	TimeUnlimited      = 2
)

// challengeForm builds the shared challenge parameters (clock, variant, colour).
func challengeForm(c CreateChallengeType) url.Values {
	f := url.Values{"variant": {c.Variant}}
	if c.Color != "" {
		f.Set("color", c.Color)
	}
	if c.Fen != "" {
		f.Set("fen", c.Fen)
	}
	switch c.TimeOption {
	case TimeRealTime:
		f.Set("clock.limit", c.ClockLimit)
		f.Set("clock.increment", c.ClockIncrement)
	case TimeCorrespondence:
		f.Set("days", c.Days)
	}
	return f
}

// Pending is an open challenge or seek. Lichess keeps it alive only while the
// connection stays open, so Cancel withdraws it.
type Pending struct {
	ID     string // challenge id (empty for real-time seeks)
	cancel context.CancelFunc
	done   chan string
}

// Cancel withdraws the challenge or seek by closing its connection.
func (p *Pending) Cancel() {
	if p != nil && p.cancel != nil {
		p.cancel()
	}
}

// Done delivers how the request ended ("accepted", "declined", "canceled", or
// "closed" if the connection dropped), then closes.
func (p *Pending) Done() <-chan string { return p.done }

// keepAlive drains a kept-alive challenge/seek stream in the background until
// Lichess reports it finished or the connection closes.
func keepAlive(body io.ReadCloser, cancel context.CancelFunc, sc *bufio.Scanner) *Pending {
	p := &Pending{cancel: cancel, done: make(chan string, 1)}
	go func() {
		defer body.Close()
		defer close(p.done)
		for sc.Scan() {
			var msg struct {
				Done string `json:"done"`
			}
			if json.Unmarshal(sc.Bytes(), &msg) == nil && msg.Done != "" {
				p.done <- msg.Done
				return
			}
		}
		p.done <- "closed"
	}()
	return p
}

// CreateChallenge challenges a user. The challenge stays open until accepted,
// declined, or Pending.Cancel is called.
func CreateChallenge(c CreateChallengeType) (*Pending, error) {
	f := challengeForm(c)
	f.Set("rated", c.Rated)
	f.Set("keepAliveStream", "true")
	ctx, cancel := context.WithCancel(context.Background())
	body, err := openStream(ctx, "POST", "/api/challenge/"+url.PathEscape(c.DestUser), f)
	if err != nil {
		cancel()
		return nil, err
	}
	// The first line is the challenge itself; later lines report completion.
	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		var first struct {
			Challenge struct {
				ID string `json:"id"`
			} `json:"challenge"`
			ID string `json:"id"`
		}
		if len(sc.Bytes()) == 0 {
			continue // keep-alive newline
		}
		if err := json.Unmarshal(sc.Bytes(), &first); err != nil {
			body.Close()
			cancel()
			return nil, fmt.Errorf("unexpected challenge response: %w", err)
		}
		p := keepAlive(body, cancel, sc)
		p.ID = first.Challenge.ID
		if p.ID == "" {
			p.ID = first.ID
		}
		return p, nil
	}
	body.Close()
	cancel()
	return nil, fmt.Errorf("challenge stream closed before a challenge was created")
}

// CreateAiChallenge starts a game against Lichess's Stockfish and returns the
// game id. The game starts immediately.
func CreateAiChallenge(c CreateChallengeType) (string, error) {
	f := challengeForm(c)
	f.Set("level", c.Level)
	var game struct {
		ID string `json:"id"`
	}
	if err := call("POST", "/api/challenge/ai", f, &game); err != nil {
		return "", err
	}
	return game.ID, nil
}

// CreateSeek looks for a random opponent. Real-time seeks stay open until an
// opponent is found or Pending.Cancel is called; correspondence seeks return
// straight away with an id.
func CreateSeek(c CreateChallengeType) (*Pending, error) {
	f := url.Values{"rated": {c.Rated}, "variant": {c.Variant}}
	if c.Color != "" && c.Color != "random" {
		f.Set("color", c.Color)
	}
	switch c.TimeOption {
	case TimeRealTime:
		// The seek endpoint takes minutes + increment, unlike challenges.
		f.Set("time", strconv.FormatFloat(c.MinTurn, 'f', -1, 64))
		f.Set("increment", c.ClockIncrement)
	case TimeCorrespondence:
		f.Set("days", c.Days)
	default:
		return nil, fmt.Errorf("seeks need a clock or days per move; unlimited games are challenge-only")
	}
	ctx, cancel := context.WithCancel(context.Background())
	body, err := openStream(ctx, "POST", "/api/board/seek", f)
	if err != nil {
		cancel()
		return nil, err
	}
	if c.TimeOption == TimeCorrespondence {
		defer body.Close()
		defer cancel()
		var seek struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(body).Decode(&seek); err != nil {
			return nil, err
		}
		done := make(chan string, 1)
		done <- "created"
		close(done)
		return &Pending{ID: seek.ID, done: done}, nil
	}
	sc := bufio.NewScanner(body)
	return keepAlive(body, cancel, sc), nil
}

// GetOngoingGames lists the user's games in progress.
func GetOngoingGames() ([]OngoingGameInfo, error) {
	var res struct {
		NowPlaying []OngoingGameInfo `json:"nowPlaying"`
	}
	if err := call("GET", "/api/account/playing?nb=50", nil, &res); err != nil {
		return nil, err
	}
	return res.NowPlaying, nil
}

// GetChallenges lists incoming and outgoing challenges.
func GetChallenges() (in, out []ChallengeInfo, err error) {
	var res struct {
		In  []ChallengeInfo `json:"in"`
		Out []ChallengeInfo `json:"out"`
	}
	if err := call("GET", "/api/challenge", nil, &res); err != nil {
		return nil, nil, err
	}
	return res.In, res.Out, nil
}

// GetUsername returns the logged-in user's username.
func GetUsername() (string, error) {
	var acct struct {
		Username string `json:"username"`
	}
	if err := call("GET", "/api/account", nil, &acct); err != nil {
		return "", err
	}
	if acct.Username == "" {
		return "", fmt.Errorf("lichess account has no username")
	}
	return acct.Username, nil
}

// GetFriends returns the usernames the user follows.
func GetFriends() ([]string, error) {
	body, err := openStream(context.Background(), "GET", "/api/rel/following", nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var friends []string
	d := json.NewDecoder(body)
	for {
		var u struct {
			Username string `json:"username"`
		}
		if err := d.Decode(&u); err == io.EOF {
			return friends, nil
		} else if err != nil {
			return friends, err
		}
		if u.Username != "" {
			friends = append(friends, u.Username)
		}
	}
}

// AcceptChallenge accepts an incoming challenge.
func AcceptChallenge(id string) error {
	return call("POST", "/api/challenge/"+url.PathEscape(id)+"/accept", nil, nil)
}

// DeclineChallenge declines an incoming challenge. reason may be empty, or one
// of Lichess's reasons: generic, later, tooFast, tooSlow, timeControl, rated,
// casual, standard, variant, noBot, onlyBot.
func DeclineChallenge(id, reason string) error {
	var f url.Values
	if reason != "" {
		f = url.Values{"reason": {reason}}
	}
	return call("POST", "/api/challenge/"+url.PathEscape(id)+"/decline", f, nil)
}

// CancelChallenge withdraws an outgoing challenge.
func CancelChallenge(id string) error {
	return call("POST", "/api/challenge/"+url.PathEscape(id)+"/cancel", nil, nil)
}

func boardAction(gameID, action string) error {
	return call("POST", "/api/board/game/"+url.PathEscape(gameID)+"/"+action, nil, nil)
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// MakeMove plays a move (UCI notation) in a Board API game.
func MakeMove(gameID, move string) error { return boardAction(gameID, "move/"+url.PathEscape(move)) }

// AbortGame aborts a game that hasn't really started yet.
func AbortGame(gameID string) error { return boardAction(gameID, "abort") }

// ResignGame resigns a game.
func ResignGame(gameID string) error { return boardAction(gameID, "resign") }

// HandleDraw offers/accepts (true) or declines (false) a draw.
func HandleDraw(gameID string, accept bool) error {
	return boardAction(gameID, "draw/"+yesNo(accept))
}

// HandleTakeback proposes/accepts (true) or declines (false) a takeback.
func HandleTakeback(gameID string, accept bool) error {
	return boardAction(gameID, "takeback/"+yesNo(accept))
}

// ClaimVictory claims the win when the opponent has left the game.
func ClaimVictory(gameID string) error { return boardAction(gameID, "claim-victory") }

// ClaimDraw claims a draw when the opponent has left the game. (Threefold
// repetition is claimed by offering a draw.)
func ClaimDraw(gameID string) error { return boardAction(gameID, "claim-draw") }

// ChatMessage is one line of game chat.
type ChatMessage struct {
	User string `json:"user"`
	Text string `json:"text"`
}

// WriteChat posts to a game's chat. room is "player" or "spectator".
func WriteChat(gameID, room, text string) error {
	return call("POST", "/api/board/game/"+url.PathEscape(gameID)+"/chat",
		url.Values{"room": {room}, "text": {text}}, nil)
}

// GetChat fetches the player chat history of a game.
func GetChat(gameID string) ([]ChatMessage, error) {
	var msgs []ChatMessage
	err := call("GET", "/api/board/game/"+url.PathEscape(gameID)+"/chat", nil, &msgs)
	return msgs, err
}

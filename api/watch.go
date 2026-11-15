package api

import (
	"bufio"
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// Watching games: Lichess TV and any ongoing game. No login needed. Lichess
// sends ongoing games 3 moves behind, so the API can't help cheaters.

// TV channels the board can show (standard-rules games), in menu order.
var TvChannels = []struct{ Key, Name string }{
	{"best", "Top rated"}, {"bullet", "Bullet"}, {"blitz", "Blitz"}, {"rapid", "Rapid"},
	{"classical", "Classical"}, {"chess960", "Chess960"}, {"bot", "Bots"}, {"computer", "Computer"},
}

// WatchPlayer is one side of a watched game.
type WatchPlayer struct {
	Name   string
	Title  string
	Rating int
}

// WatchUpdate is one message from a watch stream: a new game (Game set),
// a position, or the game ending (Status set).
type WatchUpdate struct {
	Game                   *WatchGame // a new game starts (TV changes game, or the first message)
	FEN                    string
	LastMove               string // UCI
	WhiteClock, BlackClock int    // seconds; -1 if unknown
	Status                 string // set when the game is over, e.g. "mate", "resign"
	Winner                 string
}

// WatchGame describes a watched game.
type WatchGame struct {
	ID             string
	White, Black   WatchPlayer
	Orientation    string // "white" or "black": which side TV shows at the bottom
	Speed          string
	Rated          bool
	Moves          string // SAN moves so far, when known
	InitialFEN     string
	ClockInitial   int // seconds each side started with; 0 if unknown
	ClockIncrement int // seconds added per move
	DaysPerTurn    int // correspondence
}

// StreamTv follows a TV channel until ctx is cancelled or the stream ends.
func StreamTv(ctx context.Context, channel string, fn func(WatchUpdate)) error {
	body, err := openStream(ctx, "GET", "/api/tv/"+url.PathEscape(channel)+"/feed", nil)
	if err != nil {
		return err
	}
	defer body.Close()
	return eachLine(body, func(line []byte) {
		var msg struct {
			T string `json:"t"`
			D struct {
				ID          string `json:"id"`
				Orientation string `json:"orientation"`
				FEN         string `json:"fen"`
				LM          string `json:"lm"`
				WC          *int   `json:"wc"`
				BC          *int   `json:"bc"`
				Players     []struct {
					Color   string `json:"color"`
					Rating  int    `json:"rating"`
					Seconds int    `json:"seconds"`
					User    struct {
						Name  string `json:"name"`
						Title string `json:"title"`
					} `json:"user"`
				} `json:"players"`
			} `json:"d"`
		}
		if unmarshal(line, &msg) != nil {
			return
		}
		u := WatchUpdate{FEN: msg.D.FEN, LastMove: msg.D.LM, WhiteClock: -1, BlackClock: -1}
		if msg.D.WC != nil {
			u.WhiteClock = *msg.D.WC
		}
		if msg.D.BC != nil {
			u.BlackClock = *msg.D.BC
		}
		if msg.T == "featured" {
			g := &WatchGame{ID: msg.D.ID, Orientation: msg.D.Orientation}
			for _, p := range msg.D.Players {
				wp := WatchPlayer{Name: p.User.Name, Title: p.User.Title, Rating: p.Rating}
				if p.Color == "white" {
					g.White, u.WhiteClock = wp, p.Seconds
				} else {
					g.Black, u.BlackClock = wp, p.Seconds
				}
			}
			u.Game = g
		}
		if u.FEN != "" || u.Game != nil {
			fn(u)
		}
	})
}

// gameJSON is a game as Lichess exports it (and as the game stream starts).
type gameJSON struct {
	ID         string `json:"id"`
	Speed      string `json:"speed"`
	Rated      bool   `json:"rated"`
	Status     string `json:"status"`
	Winner     string `json:"winner"`
	Moves      string `json:"moves"`
	InitialFen string `json:"initialFen"`
	FEN        string `json:"fen"`
	LastMove   string `json:"lastMove"`
	Clock      struct {
		Initial   int `json:"initial"`
		Increment int `json:"increment"`
	} `json:"clock"`
	DaysPerTurn int `json:"daysPerTurn"`
	Players     struct {
		White, Black struct {
			Rating int `json:"rating"`
			User   struct {
				Name  string `json:"name"`
				Title string `json:"title"`
			} `json:"user"`
			AiLevel int `json:"aiLevel"`
		}
	} `json:"players"`
}

func (g gameJSON) watchGame() *WatchGame {
	player := func(name, title string, rating, ai int) WatchPlayer {
		if name == "" && ai > 0 {
			name = "Stockfish level " + string(rune('0'+ai))
		}
		return WatchPlayer{Name: name, Title: title, Rating: rating}
	}
	return &WatchGame{
		ID: g.ID, Speed: g.Speed, Rated: g.Rated, Moves: g.Moves, InitialFEN: g.InitialFen, Orientation: "white",
		ClockInitial: g.Clock.Initial, ClockIncrement: g.Clock.Increment, DaysPerTurn: g.DaysPerTurn,
		White: player(g.Players.White.User.Name, g.Players.White.User.Title, g.Players.White.Rating, g.Players.White.AiLevel),
		Black: player(g.Players.Black.User.Name, g.Players.Black.User.Title, g.Players.Black.Rating, g.Players.Black.AiLevel),
	}
}

// GetGame fetches a game with its moves so far.
func GetGame(id string) (*WatchGame, error) {
	if !ValidGameID(id) {
		return nil, ErrBadGameID
	}
	var g gameJSON
	if err := callJSON("/game/export/"+url.PathEscape(id)+"?moves=true&clocks=false&evals=false", &g); err != nil {
		return nil, err
	}
	return g.watchGame(), nil
}

// GetCurrentGame fetches a player's ongoing (or last) game.
func GetCurrentGame(username string) (*WatchGame, error) {
	if !ValidUsername(username) {
		return nil, ErrBadUsername
	}
	var g gameJSON
	if err := callJSON("/api/user/"+url.PathEscape(username)+"/current-game?moves=true&clocks=false&evals=false", &g); err != nil {
		return nil, err
	}
	return g.watchGame(), nil
}

// callJSON is a GET asking for JSON (some endpoints default to PGN).
func callJSON(path string, out any) error {
	req, err := newRequest(context.Background(), "GET", path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	res, err := apiClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return responseError(res)
	}
	return decode(json.NewDecoder(res.Body), out)
}

// StreamGame follows an ongoing game until it ends or ctx is cancelled. The
// first update describes the game; the last has its Status.
func StreamGame(ctx context.Context, id string, fn func(WatchUpdate)) error {
	if !ValidGameID(id) {
		return ErrBadGameID
	}
	body, err := openStream(ctx, "GET", "/api/stream/game/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	defer body.Close()
	return eachLine(body, func(line []byte) {
		var msg struct {
			gameJSON
			LM string `json:"lm"`
			WC *int   `json:"wc"`
			BC *int   `json:"bc"`
		}
		if unmarshal(line, &msg) != nil {
			return
		}
		u := WatchUpdate{FEN: msg.FEN, LastMove: msg.LM, WhiteClock: -1, BlackClock: -1}
		if msg.WC != nil {
			u.WhiteClock = *msg.WC
		}
		if msg.BC != nil {
			u.BlackClock = *msg.BC
		}
		if msg.ID != "" { // a game description: the first message, and the last
			if msg.Status != "" && msg.Status != "started" && msg.Status != "created" {
				u.Status, u.Winner = msg.Status, msg.Winner
			} else {
				u.Game = msg.watchGame()
			}
			if u.LastMove == "" {
				u.LastMove = msg.LastMove
			}
		}
		fn(u)
	})
}

// GameIDFrom extracts a game ID from a Lichess link or a bare ID.
func GameIDFrom(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "lichess.org/"); i >= 0 {
		s = s[i+len("lichess.org/"):]
	}
	s = strings.Trim(s, "/")
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	if len(s) > 8 { // links to a player's view add 4 characters
		s = s[:8]
	}
	return s
}

// eachLine calls fn for each non-empty line of an ndjson stream.
func eachLine(r interface{ Read([]byte) (int, error) }, fn func([]byte)) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		if line := sc.Bytes(); len(line) > 0 {
			fn(line)
		}
	}
	return sc.Err()
}

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/url"
)

// Puzzle is a Lichess puzzle and the game it comes from.
type Puzzle struct {
	Game struct {
		ID      string                `json:"id"`
		PGN     string                `json:"pgn"` // SAN moves, up to and including the opponent's last move
		Clock   string                `json:"clock"`
		Rated   bool                  `json:"rated"`
		Perf    struct{ Name string } `json:"perf"`
		Players []struct {
			Name   string `json:"name"`
			Title  string `json:"title"`
			Color  string `json:"color"`
			Rating int    `json:"rating"`
		} `json:"players"`
	} `json:"game"`
	Puzzle struct {
		ID         string   `json:"id"`
		Rating     int      `json:"rating"`
		Plays      int      `json:"plays"`
		InitialPly int      `json:"initialPly"`
		Solution   []string `json:"solution"` // UCI; the player's moves are at even indexes
		Themes     []string `json:"themes"`
		FEN        string   `json:"fen"`
		LastMove   string   `json:"lastMove"`
	} `json:"puzzle"`
}

// Puzzle difficulties, relative to the player's puzzle rating (or 1500).
var PuzzleDifficulties = []string{"easiest", "easier", "normal", "harder", "hardest"}

// GetDailyPuzzle fetches today's puzzle. No login needed.
func GetDailyPuzzle() (*Puzzle, error) {
	var p Puzzle
	if err := call("GET", "/api/puzzle/daily", nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetNextPuzzle fetches a random puzzle. When logged in, Lichess only sends
// puzzles the player hasn't seen, and difficulty is relative to their rating.
func GetNextPuzzle(difficulty string) (*Puzzle, error) {
	q := url.Values{}
	if difficulty != "" {
		q.Set("difficulty", difficulty)
	}
	var p Puzzle
	if err := call("GET", "/api/puzzle/next?"+q.Encode(), nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// PuzzleRound is Lichess's answer to a submitted puzzle result.
type PuzzleRound struct {
	ID         string `json:"id"`
	Win        bool   `json:"win"`
	RatingDiff int    `json:"ratingDiff"`
}

// PuzzleResult is what Lichess says after a result is submitted: the round
// and the player's new puzzle rating.
type PuzzleResult struct {
	Rounds []PuzzleRound `json:"rounds"`
	Glicko struct {
		Rating      float64 `json:"rating"`
		Provisional bool    `json:"provisional"`
	} `json:"glicko"`
}

// SubmitPuzzle records a solved (win) or failed puzzle on the player's
// account; rated decides whether it changes their puzzle rating. Needs a
// login with the puzzle:write scope.
func SubmitPuzzle(id string, win, rated bool) (*PuzzleResult, error) {
	body, err := json.Marshal(map[string]any{
		"solutions": []map[string]any{{"id": id, "win": win, "rated": rated}},
	})
	if err != nil {
		return nil, err
	}
	req, err := newRequest(context.Background(), "POST", "/api/puzzle/batch/mix", nil)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, responseError(res)
	}
	var r PuzzleResult
	if err := decode(json.NewDecoder(res.Body), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

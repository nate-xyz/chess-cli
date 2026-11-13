package api

import (
	"time"
)

// HostUrl is the Lichess server (a variable so tests can point it at a fake).
var HostUrl = "https://lichess.org"

const (
	ClientID string = "chess-cli"

	// legacyConfigFile is where older versions kept the token (in the working
	// directory). LoadConfig migrates it to ConfigDir.
	legacyConfigFile string = "user_config.json"
)

var (
	UserInfo UserConfig
	Scopes   = []string{
		"preference:read",
		"preference:write",
		"challenge:read",
		"challenge:write",
		"challenge:bulk",
		"study:read",
		"study:write",
		"puzzle:read",
		"follow:read",
		"follow:write",
		"msg:write",
		"bot:play",
		"board:play",
	}
)

type BoardEventEnum int

const (
	GameFull     BoardEventEnum = iota // full game info, sent first and on reconnect
	GameState                          // moves/clocks/offers changed
	ChatLine                           // a chat message
	OpponentGone                       // the opponent left (or came back)
	EOF                                // the game stream ended normally
)

type BoardEvent struct {
	Type  BoardEventEnum
	Full  StreamBoardGameFull
	State StreamBoardGameState
	Chat  StreamBoardChat
	Gone  StreamOpponentGone
}

type StreamEventType struct {
	EventType string
	GameID    string
	Source    string
}

// type for storing user info into a json
type UserConfig struct {
	ApiToken            string
	TokenCreationDate   time.Time
	TokenExpirationDate time.Time
}

type CreateChallengeType struct {
	Type           int
	DestUser       string
	Variant        string
	VariantIndex   int
	TimeOption     int
	ClockLimit     string
	ClockIncrement string
	Days           string
	Rated          string
	RatedBool      bool
	Color          string
	ColorIndex     int
	MinTurn        float64
	OpenEnded      bool
	Level          string
	Fen            string // starting position, for "fromPosition" games
}

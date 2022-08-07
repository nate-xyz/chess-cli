package api

import (
	"time"
)

const (
	HostUrl   string = "https://lichess.org"
	ClientID  string = "chess-cli"
	AuthURL   string = HostUrl + "/oauth"
	TokenURL  string = HostUrl + "/api/token"
	json_path string = "user_config.json"
)

var (
	UserInfo                = UserConfig{ApiToken: "", TokenCreationDate: time.Now(), TokenExpirationDate: time.Now().AddDate(1, 0, 0)}
	StreamEventStarted bool = false
	Online             bool = false
	Scopes                  = []string{
		"preference:read",
		"preference:write",
		"email:read",
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
	GameFull BoardEventEnum = iota
	GameState
	ChatLine
	ChatLineSpectator
	GameStateResign
	EOF
)

type BoardEvent struct {
	Type   BoardEventEnum
	Full   StreamBoardGameFull
	State  StreamBoardGameState
	Chat   StreamBoardChat
	Resign StreamBoardResign
}

type StreamEventType struct {
	EventType string
	GameID    string
	Source    string
}

//type for storing user info into a json
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
}

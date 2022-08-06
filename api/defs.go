package api

import (
	"fmt"
	"time"
)

const (
	GameFull BoardEvent = iota
	GameState
	ChatLine
	ChatLineSpectator
	GameStateResign
	EOF
)

var (
	HostUrl  string = "https://lichess.org"
	ClientID string = "chess-cli"
	Scopes          = []string{
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

	UserEmail                   string
	Username                    string
	UserProfile                 map[string]interface{}
	AllFriends                  []string
	UserInfo                           = UserConfig{ApiToken: "", TokenCreationDate: time.Now(), TokenExpirationDate: time.Now().AddDate(1, 0, 0)}
	AuthURL                     string = fmt.Sprintf("%s/oauth", HostUrl)
	TokenURL                    string = fmt.Sprintf("%s/api/token", HostUrl)
	RedirectURL                 string
	redirectPort                int
	json_path                        = "user_config.json"
	StreamEventStarted          bool = false
	Ready                       chan struct{}
	Online                      bool = false
	ChallengeId                 string
	OngoingGames                []OngoingGameInfo
	IncomingChallenges          []ChallengeInfo
	OutgoingChallenges          []ChallengeInfo
	EventStreamArr              []StreamEventType
	BoardFullGame               StreamBoardGameFull
	BoardGameState              StreamBoardGameState
	BoardChatLine               StreamBoardChat
	BoardChatLineSpectator      StreamBoardChat
	BoardResign                 StreamBoardResign
	CurrentStreamEventChallenge StreamEventChallenge
	CurrentStreamEventGame      StreamEventGame
)

//type for storing user info into a json
type UserConfig struct {
	ApiToken            string
	TokenCreationDate   time.Time
	TokenExpirationDate time.Time
}

type BoardEvent int

type CreateChallengeType struct {
	Type           int
	Username       string
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

type StreamEventType struct {
	EventType string
	GameID    string
	Source    string
}

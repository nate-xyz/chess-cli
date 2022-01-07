package lichess

import "fmt"

var CurrentChallenge CreateChallengeType
var WaitingAlert chan StreamEventType

//API
var currentGameID string
var UserEmail string
var Username string
var UserProfile map[string]interface{}
var UserFriends string
var allFriends []string
var FriendsMap map[string]bool
var ChallengeId string
var streamEvent string
var OngoingGames []OngoingGameInfo
var IncomingChallenges []ChallengeInfo
var OutgoingChallenges []ChallengeInfo
var BoardStreamArr []BoardState
var EventStreamArr []StreamEventType
var gameStateChan chan BoardState
var board_state_sig chan bool
var testChallenge = CreateChallengeType{
	Type:       1,
	TimeOption: 2,
	DestUser:   "",
	Rated:      "false",
	Color:      "white",
	Variant:    "standard"}

//API TYPES
type OngoingGameInfo struct {
	FullId   string `json: "fullId"`
	GameID   string `json: "gameId"`
	FEN      string `json: "fen"`
	Color    string `json: "color"`
	LastMove string `json: "lastMove"`
	Variant  struct {
		Key  string `json: "key"`
		Name string `json: "name"`
	} `json: "variant"`
	Speed    string `json: "speed"`
	Perf     string `json: "perf"`
	Rated    bool   `json: "rated"`
	Opponent struct {
		Id       string `json: "id"`
		Username string `json: "username"`
		Rating   int    `json: "rating"`
	} `json: "opponent"`
	IsMyTurn bool `json: "isMyTurn"`
}

type ChallengeInfo struct {
	Id          string `json: "id"`
	URL         string `json: "url"`
	Color       string `json: "color"`
	Direction   string `json: "direction"`
	TimeControl struct {
		Increment int    `json: "increment"`
		Limit     int    `json: "limit"`
		Show      string `json: "show"`
		Type      string `json: "type"`
	} `json: "timeControl"`
	Variant struct {
		Key   string `json: "key"`
		Name  string `json: "name"`
		Short string `json: "short"`
	} `json: "variant"`
	Challenger struct {
		Id     string `json: "id"`
		Name   string `json: "name"`
		Rating int    `json: "rating"`
		Title  string `json: "title"`
	} `json: "challenger"`
	DestUser struct {
		Id     string `json: "id"`
		Name   string `json: "name"`
		Rating int    `json: "rating"`
		Title  string `json: "title"`
	} `json: "destUser"`
	Perf struct {
		Icon string `json: "icon"`
		Name string `json: "name"`
	} `json: "perf"`
	Rated  bool   `json: "rated"`
	Speed  string `json: "speed"`
	Status string `json: "status"`
}

type StreamEventType struct {
	Event string
	Id    string
}

type BoardState struct {
	Type   string
	Moves  string
	Status string
	Rated  bool
}

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
}

//OAUTH TYPES

//type for storing user info into a json
type UserConfig struct {
	ApiToken string
}

//OAUTH VARS
var hostUrl string = "https://lichess.org"
var ClientID string = "chess-cli"
var Scopes = []string{
	"preference:read",
	"preference:write",
	"email:read",
	"challenge:read",
	"challenge:write",
	"challenge:bulk",
	"study:read",
	"study:write",
	"puzzle:read",
	"follow:write",
	"msg:write",
	"board:play",
}

var UserInfo = UserConfig{ApiToken: ""}
var AuthURL string = fmt.Sprintf("%s/oauth", hostUrl)
var TokenURL string = fmt.Sprintf("%s/api/token", hostUrl)
var RedirectURL string
var redirectPort int
var json_path = "user_config.json"
var StreamChannel chan StreamEventType

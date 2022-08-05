package main

import (
	"fmt"
	"time"
)

//stream event variables
//https://lichess.org/api#operation/apiStreamEvent
var CurrentChallenge CreateChallengeType
var CurrentStreamEventChallenge StreamEventChallenge
var CurrentStreamEventGame StreamEventGame

//board event stream variables
//https://lichess.org/api#operation/boardGameStream
type BoardEvent int

const (
	GameFull BoardEvent = iota
	GameState
	ChatLine
	ChatLineSpectator
	GameStateResign
	EOF
)

var streamed_move_sequence chan string
var BoardFullGame StreamBoardGameFull
var BoardGameState StreamBoardGameState
var BoardChatLine StreamBoardChat
var BoardChatLineSpectator StreamBoardChat
var BoardResign StreamBoardResign

//var CurrentAiChallenge CreateChallengeType
var WaitingAlert chan StreamEventType

type StreamEventType struct {
	EventType string
	GameID    string
	Source    string
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
	Level          string
}

//API VARS
var currentGameID string
var UserEmail string
var Username string

var UserProfile map[string]interface{}
var UserFriends string
var allFriends []string
var FriendsMap map[string]bool
var ChallengeId string

//var streamEvent string
var OngoingGames []OngoingGameInfo
var IncomingChallenges []ChallengeInfo
var OutgoingChallenges []ChallengeInfo
var BoardStreamArr []BoardState

var EventStreamArr []StreamEventType
var gameStateChan chan BoardEvent
var board_state_sig chan bool

//OAUTH TYPES

//type for storing user info into a json
type UserConfig struct {
	ApiToken            string
	TokenCreationDate   time.Time
	TokenExpirationDate time.Time
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
	"follow:read",
	"follow:write",
	"msg:write",
	"bot:play",
	"board:play",
}

var UserInfo = UserConfig{ApiToken: "", TokenCreationDate: time.Now(), TokenExpirationDate: time.Now().AddDate(1, 0, 0)}
var AuthURL string = fmt.Sprintf("%s/oauth", hostUrl)
var TokenURL string = fmt.Sprintf("%s/api/token", hostUrl)
var RedirectURL string
var redirectPort int
var json_path = "user_config.json"

var StreamChannel chan StreamEventType
var StreamChannelForWaiter chan StreamEventType

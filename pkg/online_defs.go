package pkg

import (
	"github.com/nate-xyz/chess-cli/api"
)

//stream event variables
//https://lichess.org/api#operation/apiStreamEvent
var CurrentChallenge api.CreateChallengeType

//board event stream variables
//https://lichess.org/api#operation/boardGameStream

var streamed_move_sequence chan string

//var CurrentAiChallenge CreateChallengeType
var WaitingAlert chan api.StreamEventType

type BoardState struct {
	Type   string
	Moves  string
	Status string
	Rated  bool
}

//API VARS
var currentGameID string

var UserFriends string

var FriendsMap map[string]bool

//var streamEvent string

var BoardStreamArr []BoardState

var gameStateChan chan api.BoardEvent
var board_state_sig chan bool

var StreamChannel chan api.StreamEventType
var StreamChannelForWaiter chan api.StreamEventType

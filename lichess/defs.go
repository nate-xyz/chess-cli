package lichess

import "fmt"

var CurrentChallenge CreateChallengeType

//var CurrentAiChallenge CreateChallengeType
var WaitingAlert chan StreamEventType

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

type AiChallengeInfo struct {
	Id         string `json: "id"'`
	Rated      bool   `json: "rated"`
	Variant    string `json: "variant"`
	Speed      string `json: "speed"`
	Perf       string `json: "perf"`
	CreatedAt  int    `json: "createdAt"`
	LastMoveAt int    `json: "lastMoveAt"`
	Status     string `json: "status"`
	Players    struct {
		White struct {
			User struct {
				Name   string `json: "name"`
				Title  string `json: "title"`
				Patron bool   `json: "patron"`
				ID     string `json: "id"`
			}
			Rating      int    `json: "rating"`
			RatingDiff  int    `json: "ratingDiff"`
			Name        string `json: "name"`
			Provisional bool   `json: "provisional"`
			AiLevel     int    `json: "aiLevel"`
			Analysis    struct {
				Inaccuracy int `json: "inaccuracy"`
				Mistake    int `json: "mistake"`
				Blunder    int `json: "blunder"`
				Acpl       int `json: "acpl"`
			}
			Team string `json: "team"`
		}
		Black struct {
			User struct {
				Name   string `json: "name"`
				Title  string `json: "title"`
				Patron bool   `json: "patron"`
				ID     string `json: "id"`
			}
			Rating      int    `json: "rating"`
			RatingDiff  int    `json: "ratingDiff"`
			Name        string `json: "name"`
			Provisional bool   `json: "provisional"`
			AiLevel     int    `json: "aiLevel"`
			Analysis    struct {
				Inaccuracy int `json: "inaccuracy"`
				Mistake    int `json: "mistake"`
				Blunder    int `json: "blunder"`
				Acpl       int `json: "acpl"`
			}
			Team string `json: "team"`
		}
	}
	InitialFen string `json: "initialFen"`
	Winner     string `json: "winner"`
	Opening    struct {
		Eco  string `json: "eco"`
		Name string `json: "name"`
		Ply  int    `json: "ply"`
	}
	Moves       string `json: "moves"`
	Pgn         string `json: "pgn"`
	DaysPerTurn int    `json: "daysPerTurn"`
	Analysis    struct {
		Eval      int    `json: "eval"`
		Best      string `json: "best"`
		Variation string `json: "variation"`
		Judgement struct {
			Name    string `json: "name"`
			Comment string `json: "comment"`
		}
	}
	Tournament string `json: "tournament"`
	Swiss      string `json: "swiss"`
	Clock      struct {
		Initial   int `json: "initial"`
		Increment int `json: "increment"`
		TotalTime int `json: "totalTime"`
	}
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
	Level          string
}

// type CreateAiChallengeType struct {
// 	Type           int
// 	TimeOption     int
// 	Level          string
// 	ClockLimit     string
// 	ClockIncrement string
// 	DaysPerTurn    string
// 	Days           string
// 	Color          string
// 	Variant        string
// 	Fen            string
// }

var testAiChallenge = CreateChallengeType{
	Level:      "2",
	Type:       2,
	TimeOption: 2,
	Color:      "black",
	Variant:    "standard"}

//API VARS
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
	Color:      "black",
	Variant:    "standard"}

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
var StreamChannelForWaiter chan StreamEventType

//var Challenge map[string]interface{}

// type TimeInfo struct {
// 	Increment int    `json: "increment"`
// 	Limit     int    `json: "limit"`
// 	Show      string `json: "show"`
// 	Type      string `json: "type"`
// }

// type VariantInfo struct {
// 	Key   string `json: "key"`
// 	Name  string `json: "name"`
// 	Short string `json: "short"`
// }

// type ChallengerInfo struct {
// 	Id     string `json: "id"`
// 	Name   string `json: "name"`
// 	Rating int    `json: "rating"`
// 	Title  string `json: "title"`
// }

// type Perf_ struct {
// 	Icon string `json: "icon"`
// 	Name string `json: "name"`
// }

// type OngoingGameVariant struct {
// 	Key  string `json: "key"`
// 	Name string `json: "name"`
// }

// type OngoingGameOpp struct {
// 	Id       string `json: "id"`
// 	Username string `json: "username"`
// 	Rating   string `json: "rating"`
// }

// type ChallengeInfo struct {
// 	Id          string         `json: "id"`
// 	URL         string         `json: "url"`
// 	Color       string         `json: "color"`
// 	Direction   string         `json: "direction"`
// 	TimeControl TimeInfo       `json: "timeControl"`
// 	Variant     VariantInfo    `json: "variant"`
// 	Challenger  ChallengerInfo `json: "challenger"`
// 	DestUser    ChallengerInfo `json: "destUser"`
// 	Perf        Perf_          `json: "perf"`
// 	Rated       bool           `json: "rated"`
// 	Speed       string         `json: "speed"`
// 	Status      string         `json: "status"`
// }

// type ChallengeJSON struct {
// 	In  []ChallengeInfo `json: "in"`
// 	Out []ChallengeInfo `json: "out"`
// }

//var JSONresult ChallengeJSON

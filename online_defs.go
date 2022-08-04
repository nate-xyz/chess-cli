package main

import "fmt"

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

//API TYPES

//https://lichess.org/api#operation/boardGameStream
type StreamBoardResign struct {
	Type   string `json:"type"`
	Moves  string `json:"moves"`
	Wtime  int    `json:"wtime"`
	Btime  int    `json:"btime"`
	Winc   int    `json:"winc"`
	Binc   int    `json:"binc"`
	Status string `json:"status"`
	Winner string `json:"winner"`
}

//https://lichess.org/api#operation/boardGameStream
type StreamBoardChat struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Text     string `json:"text"`
	Room     string `json:"room"`
}

//https://lichess.org/api#operation/boardGameStream
type StreamBoardGameState struct {
	Type   string `json:"type"`
	Moves  string `json:"moves"`
	Wtime  int    `json:"wtime"`
	Btime  int    `json:"btime"`
	Winc   int    `json:"winc"`
	Binc   int    `json:"binc"`
	Status string `json:"status"`
}

//https://lichess.org/api#operation/boardGameStream
type StreamBoardGameFull struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Rated   bool   `json:"rated"`
	Variant struct {
		Key   string `json:"key"`
		Name  string `json:"name"`
		Short string `json:"short"`
	} `json:"variant"`
	Clock struct {
		Initial   int `json:"initial"`
		Increment int `json:"increment"`
	} `json:"clock"`
	Speed string `json:"speed"`
	Perf  struct {
		Name string `json:"name"`
	} `json:"perf"`
	CreatedAt int64 `json:"createdAt"`
	White     struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Provisional bool   `json:"provisional"`
		Rating      int    `json:"rating"`
		Title       string `json:"title"`
	} `json:"white"`
	Black struct {
		ID     string      `json:"id"`
		Name   string      `json:"name"`
		Rating int         `json:"rating"`
		Title  interface{} `json:"title"`
	} `json:"black"`
	InitialFen string `json:"initialFen"`
	State      struct {
		Type   string `json:"type"`
		Moves  string `json:"moves"`
		Wtime  int    `json:"wtime"`
		Btime  int    `json:"btime"`
		Winc   int    `json:"winc"`
		Binc   int    `json:"binc"`
		Status string `json:"status"`
	} `json:"state"`
}

type StreamEventChallenge struct {
	Type      string `json:"type"`
	Challenge struct {
		ID     string `json:"id"`
		URL    string `json:"url"`
		Status string `json:"status"`
		Compat struct {
			Bot   bool `json:"bot"`
			Board bool `json:"board"`
		} `json:"compat"`
		Challenger struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Title  string `json:"title"`
			Rating int    `json:"rating"`
			Patron bool   `json:"patron"`
			Online bool   `json:"online"`
			Lag    int    `json:"lag"`
		} `json:"challenger"`
		DestUser struct {
			ID          string      `json:"id"`
			Name        string      `json:"name"`
			Title       interface{} `json:"title"`
			Rating      int         `json:"rating"`
			Provisional bool        `json:"provisional"`
			Online      bool        `json:"online"`
			Lag         int         `json:"lag"`
		} `json:"destUser"`
		Variant struct {
			Key   string `json:"key"`
			Name  string `json:"name"`
			Short string `json:"short"`
		} `json:"variant"`
		Rated       bool `json:"rated"`
		TimeControl struct {
			Type      string `json:"type"`
			Limit     int    `json:"limit"`
			Increment int    `json:"increment"`
			Show      string `json:"show"`
		} `json:"timeControl"`
		Color string `json:"color"`
		Speed string `json:"speed"`
		Perf  struct {
			Icon string `json:"icon"`
			Name string `json:"name"`
		} `json:"perf"`
	} `json:"challenge"`
}

type StreamEventGame struct {
	Type string `json:"type"`
	Game struct {
		GameID   string `json:"gameId"`
		FullID   string `json:"fullId"`
		Color    string `json:"color"`
		Fen      string `json:"fen"`
		HasMoved bool   `json:"hasMoved"`
		IsMyTurn bool   `json:"isMyTurn"`
		LastMove string `json:"lastMove"`
		Opponent struct {
			ID       string `json:"id"`
			Rating   int    `json:"rating"`
			Username string `json:"username"`
		} `json:"opponent"`
		Perf        string `json:"perf"`
		Rated       bool   `json:"rated"`
		SecondsLeft int    `json:"secondsLeft"`
		Source      string `json:"source"`
		Speed       string `json:"speed"`
		Variant     struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"variant"`
		Compat struct {
			Bot   bool `json:"bot"`
			Board bool `json:"board"`
		} `json:"compat"`
	} `json:"game"`
}

// type OngoingGameInfo struct {
// 	FullId   string `json: "fullId"`
// 	GameID   string `json: "gameId"`
// 	FEN      string `json: "fen"`
// 	Color    string `json: "color"`
// 	LastMove string `json: "lastMove"`
// 	Variant  struct {
// 		Key  string `json: "key"`
// 		Name string `json: "name"`
// 	} `json: "variant"`
// 	Speed    string `json: "speed"`
// 	Perf     string `json: "perf"`
// 	Rated    bool   `json: "rated"`
// 	Opponent struct {
// 		Id       string `json: "id"`
// 		Username string `json: "username"`
// 		Rating   int    `json: "rating"`
// 	} `json: "opponent"`
// 	IsMyTurn bool `json: "isMyTurn"`
// }

type OngoingGameInfo struct {
	GameID   string `json:"gameId"`
	FullID   string `json:"fullId"`
	Color    string `json:"color"`
	Fen      string `json:"fen"`
	HasMoved bool   `json:"hasMoved"`
	IsMyTurn bool   `json:"isMyTurn"`
	LastMove string `json:"lastMove"`
	Opponent struct {
		ID       string `json:"id"`
		Rating   int    `json:"rating"`
		Username string `json:"username"`
	} `json:"opponent"`
	Perf        string `json:"perf"`
	Rated       bool   `json:"rated"`
	SecondsLeft int    `json:"secondsLeft"`
	Source      string `json:"source"`
	Speed       string `json:"speed"`
	Variant     struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"variant"`
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
	"follow:read",
	"follow:write",
	"msg:write",
	"bot:play",
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

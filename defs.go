package main

//TYPE DEFINITIONS

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

//OTHER TYPES

//type for piece location
type coord_pair struct {
	x_coord_ int
	y_coord_ int
}

//type for piece color
type piece_color struct {
	color int16
	piece rune
}

type windowSizePos struct {
	h int
	w int
	y int
	x int
}

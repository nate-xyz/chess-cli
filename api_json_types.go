package main

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
	Type      string `json:"type"`
	Moves     string `json:"moves"`
	Wtime     int    `json:"wtime"`
	Btime     int    `json:"btime"`
	Winc      int    `json:"winc"`
	Binc      int    `json:"binc"`
	Status    string `json:"status"`
	Winner    string `json:"winner"`
	Wdraw     bool   `json:"wdraw"`
	Bdraw     bool   `json:"bdraw"`
	Wtakeback bool   `json:"wtakeback"`
	Btakeback bool   `json:"btakeback"`
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

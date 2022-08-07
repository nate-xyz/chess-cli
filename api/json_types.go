package api

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

//https://lichess.org/api#operation/apiStreamEvent
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

//https://lichess.org/api#operation/apiStreamEvent
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

//https://lichess.org/api#operation/apiAccountPlaying
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

//https://lichess.org/api#operation/challengeList
type ChallengeInfo struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Color       string `json:"color"`
	Direction   string `json:"direction"`
	TimeControl struct {
		Increment int    `json:"increment"`
		Limit     int    `json:"limit"`
		Show      string `json:"show"`
		Type      string `json:"type"`
	} `json:"timeControl"`
	Variant struct {
		Key   string `json:"key"`
		Name  string `json:"name"`
		Short string `json:"short"`
	} `json:"variant"`
	Challenger struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Online      bool   `json:"online"`
		Provisional bool   `json:"provisional"`
		Rating      int    `json:"rating"`
		Title       string `json:"title"`
	} `json:"challenger"`
	DestUser struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Online      bool   `json:"online"`
		Provisional bool   `json:"provisional"`
		Rating      int    `json:"rating"`
		Title       string `json:"title"`
	} `json:"destUser"`
	Perf struct {
		Icon string `json:"icon"`
		Name string `json:"name"`
	} `json:"perf"`
	Rated  bool   `json:"rated"`
	Speed  string `json:"speed"`
	Status string `json:"status"`
}

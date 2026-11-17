package api

import (
	"time"
)

// HostUrl is the Lichess server (a variable so tests can point it at a fake).
var HostUrl = "https://lichess.org"

const (
	ClientID string = "chess-cli"

	// legacyConfigFile is where older versions kept the token (in the working
	// directory). LoadConfig migrates it to ConfigDir.
	legacyConfigFile string = "user_config.json"
)

var (
	UserInfo UserConfig
	Scopes   = []string{
		"preference:read",
		"preference:write",
		"challenge:read",
		"challenge:write",
		"challenge:bulk",
		"study:read",
		"study:write",
		"puzzle:read",
		"puzzle:write",
		"follow:read",
		"follow:write",
		"msg:write",
		"bot:play",
		"board:play",
	}
)

type BoardEventEnum int

const (
	GameFull     BoardEventEnum = iota // full game info, sent first and on reconnect
	GameState                          // moves/clocks/offers changed
	ChatLine                           // a chat message
	OpponentGone                       // the opponent left (or came back)
	EOF                                // the game stream ended normally
)

type BoardEvent struct {
	Type  BoardEventEnum
	Full  StreamBoardGameFull
	State StreamBoardGameState
	Chat  StreamBoardChat
	Gone  StreamOpponentGone
}

type StreamEventType struct {
	EventType string
	GameID    string
	Source    string
}

// type for storing user info into a json
type UserConfig struct {
	ApiToken            string
	TokenCreationDate   time.Time
	TokenExpirationDate time.Time
	// Pieces is how the board is drawn: "text" or "image" (empty = automatic).
	Pieces         string `json:",omitempty"`
	PieceSet       string `json:",omitempty"` // image rendering: piece set and board
	Board          string `json:",omitempty"`
	StockfishPath  string `json:",omitempty"` // engine chosen in chess-cli (else found automatically)
	StockfishLevel int    `json:",omitempty"`
	MoveBoxAlways  bool   `json:",omitempty"` // typed-move box always shown, not just on Ctrl-K
	ClockBarOff    bool   `json:",omitempty"` // clocks as plain text in the sidebar, no bar
	Zen            bool   `json:",omitempty"` // zen mode: hide names, ratings, chat and game info
	LocalClock     int    `json:",omitempty"` // local games time control (index; 0 = none)
	Sound          string `json:",omitempty"` // sound theme: "" (retro), "wood" or "off"
	SoundAsked     bool   `json:",omitempty"` // the first-run screen asked about sound
	StockfishAsked bool   `json:",omitempty"` // the first-run screen offered to get Stockfish
	NoVimKeys      bool   `json:",omitempty"` // arrow keys only: hjkl (and g/G in lists) don't move
	KeysAsked      bool   `json:",omitempty"` // the first-run screen asked about hjkl
	// Puzzle preferences
	PuzzleCasual     bool   `json:",omitempty"`
	PuzzleDifficulty string `json:",omitempty"`
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
	Fen            string // starting position, for "fromPosition" games
}

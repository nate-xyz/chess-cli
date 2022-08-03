package main

import (
	"fmt"
	"os"

	cv "code.rocketnine.space/tslocum/cview"
	"github.com/notnil/chess"
)

type State struct {
	app              *cv.Application
	nav              *cv.Panels
	shell            string
	currentLocalGame *LocalGame
	//window pointers
	//local
	Board   *cv.Table
	Status  *cv.TextView
	History *cv.TextView
	Time    *cv.TextView

	//post local
	PostStatus  *cv.TextView
	PostBoard   *cv.Table
	PostHistory *cv.TextView

	//lichess welcome
	LichessTitle *cv.TextView

	//loader
	LoaderIcon *cv.TextView
	LoaderMsg  *cv.TextView

	//online
	OnlineBoard   *cv.Table
	OnlineStatus  *cv.TextView
	OnlineHistory *cv.TextView

	//online post
	OnlinePostStatus  *cv.TextView
	OnlinePostBoard   *cv.Table
	OnlinePostHistory *cv.TextView
	OnlineTime        *cv.TextView
}

const (
	ApplicationTitle string = `[blue]chess-cli ♟️ [white]
Play locally with a [yellow]friend[white] or online with [red]Lichess!`
	LichessTitle        string = "[blue]chess-cli[white]: [red]lichess[white] client"
	LichessRibbon       string = "CHESS-CLI | LICHESS CLIENT | Press 'Ctrl-c' to quit"
	welcomeRibbonstr    string = "WELCOME TO CHESS-CLI ! | Press 'Ctrl-c' to quit"
	gameRibbonstr       string = "CHESS-CLI -> local game | Press 'Ctrl-c' to quit"
	challengeRibbonstr  string = "CHESS-CLI -> create a challenge | Press 'Ctrl-c' to quit"
	gameOnlineRibbonstr string = "CHESS-CLI -> online game | Press 'Ctrl-c' to quit"
	EmptyChar           string = "博"
)

var (
	root             = new(State)
	Sigs             chan os.Signal
	Ready            chan struct{}
	NewChessGame     *chess.Game         //used in online.go in the LichessGame() function to update the board position from new stream event
	newChallenge     CreateChallengeType //used initUI.go in initConstruct() in order to save settings from a challenge construction
	Online           bool                = false
	MoveCount        int
	LastSelectedCell PiecePosition
)

type ListSelectedFunc func() //used to easily add functions to list items in initUI.go

type PiecePosition struct {
	Row    int
	Col    int
	Alg    string
	Empty  bool
	Symbol string
}

type BothInc struct {
	wtime int64
	btime int64
}

type LocalGame struct {
	Game             *chess.Game
	NextMove         string
	MoveHistoryArray []string
	LegalMoves       []string
	Status           string
}

func (l *LocalGame) Init() {
	l.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	l.MoveHistoryArray = []string{}
}

var PiecesMap = map[rune]string{
	'K': "♔",
	'Q': "♕",
	'R': "♖",
	'B': "♗",
	'N': "♘",
	'P': "♙",
	'k': "♚",
	'q': "♛",
	'r': "♜",
	'b': "♝",
	'n': "♞",
	'p': "♙",
}

var KnightIconMap = map[int]string{
	0: " ♞ ",
	1: " 🨇 ",
	2: " 🨓 ",
	3: " 🨜 ",
	4: " 🨨 ",
	5: " 🨱 ",
	6: " 🨽 ",
	7: " 🩆 ",
}

var NormalLoaderMap = map[int64]string{
	0: "⠋",
	1: "⠙",
	2: "⠹",
	3: "⠸",
	4: "⠼",
	5: "⠴",
	6: "⠦",
	7: "⠧",
	8: "⠇",
	9: "⠏",
}

var RandQuoteMap = map[int]string{
	0:  "“I have come to the personal conclusion that while all artists are not chess players, all chess players are artists.”\n – Marcel Duchamp",
	1:  "“Unlike other games in which lucre is the end and aim, [chess] recommends itself to the wise by the fact that its mimic battles are fought for no prize but honor. It is eminently and emphatically the philosopher’s game.”\n – Paul Morphy",
	2:  "“The beauty of chess is it can be whatever you want it to be. It transcends language, age, race, religion, politics, gender, and socioeconomic background. Whatever your circumstances, anyone can enjoy a good fight to the death over the chess board.”\n – Simon Williams",
	3:  "“Chess is the struggle against the error.”\n – Johannes Zukertort",
	4:  "“Every chess master was once a beginner.”\n – Irving Chernev",
	5:  "“Avoid the crowd. Do your own thinking independently. Be the chess player, not the chess piece.”\n – Ralph Charell",
	6:  "“Chess makes men wiser and clear-sighted.”\n – Vladimir Putin",
	7:  "“Chess is the gymnasium of the mind.”\n – Blaise Pascal",
	8:  "“Chess holds its master in its own bonds, shackling the mind and brain so that the inner freedom of the very strongest must suffer.”\n – Albert Einstein",
	9:  "“Chess is a war over the board. The object is to crush the opponent’s mind.”\n – Bobby Fischer",
	10: "“I am convinced, the way one plays chess always reflects the player’s personality. If something defines his character, then it will also define his way of playing.”\n – Vladimir Kramnik",
	11: "“The game of chess is not merely an idle amusement. Several very valuable qualities of the mind, useful in the course of human life, are to be acquired or strengthened by it… Life is a kind of Chess, in which we have often points to gain, and competitors or adversaries to contend with.”\n – Benjamin Franklin",
	12: "“As proved by evidence, [chess is] more lasting in its being and presence than all books and achievements; the only game that belongs to all people and all ages; of which none knows the divinity that bestowed it on the world, to slay boredom, to sharpen the senses, to exhilarate the spirit.”\n – Stefan Zweig",
	13: "“Chess doesn’t drive people mad, it keeps mad people sane.”\n – Bill Hartston",
	14: "“In life, as in chess, one’s own pawns block one’s way.  A man’s very wealthy, ease, leisure, children, books, which should help him to win, more often checkmate him.”\n – Charles Buxton",
	15: "“Chess is life in miniature. Chess is a struggle, chess battles.”\n – Garry Kasparov",
	16: "“Chess, like love, like music, has the power to make men happy.”\n – Siegbert Tarrasch",
	17: "“For in the idea of chess and the development of the chess mind we have a picture of the intellectual struggle of mankind.”\n – Richard Réti",
	18: "“I don’t believe in psychology. I believe in good moves.”\n – Bobby Fischer",
	19: "“Play the opening like a book, the middlegame like a magician, and the endgame like a machine.”\n – Rudolph Spielmann",
	20: "“I used to attack because it was the only thing I knew. Now I attack because I know it works best.”\n – Garry Kasparov",
	21: "“It is my style to take my opponent and myself on to unknown grounds. A game of chess is not an examination of knowledge; it is a battle of nerves.”\n – David Bronstein",
	22: "“Chess is rarely a game of ideal moves. Almost always, a player faces a series of difficult consequences whichever move he makes.”\n – David Shenk",
	23: "“When you see a good move, look for a better one.”\n – Emanuel Lasker",
	24: "“After a bad opening, there is hope for the middle game. After a bad middle game, there is hope for the endgame. But once you are in the endgame, the moment of truth has arrived.”\n – Edmar Mednis",
	25: "“Give me a difficult positional game, I will play it. But totally won positions, I cannot stand them.”\n – Hein Donner",
	26: "“There is no remorse like the remorse of chess.”\n – H. G. Wells",
	27: "“Half the variations which are calculated in a tournament game turn out to be completely superfluous. Unfortunately, no one knows in advance which half.”\n – Jan Timman",
	28: "“Even a poor plan is better than no plan at all.”\n – Mikhail Chigorin",
	29: "“Tactics is knowing what to do when there is something to do; strategy is knowing what to do when there is nothing to do.”\n – Savielly Tartakower",
	30: "“In life, as in chess, forethought wins.”\n – Charles Buxton",
	31: "“You may learn much more from a game you lose than from a game you win. You will have to lose hundreds of games before becoming a good player.”\n – José Raúl Capablanca",
	32: "“Pawns are the soul of the game.”\n – François-André Danican Philidor",
	33: "“The passed pawn is a criminal, who should be kept under lock and key. Mild measures, such as police surveillance, are not sufficient.”\n – Aron Nimzowitsch",
	34: "“Modern chess is too much concerned with things like pawn structure. Forget it, checkmate ends the game.”\n – Nigel Short",
	35: "“Pawn endings are to chess what putting is to golf.”\n – Cecil Purdy",
	36: "“Nobody ever won a chess game by resigning.”\n – Savielly Tartakower",
	37: "“The blunders are all there on the board, waiting to be made.”\n – Savielly Tartakower",
	38: "“It’s always better to sacrifice your opponent’s men.”\n – Savielly Tartakower",
	39: "“One doesn’t have to play well, it’s enough to play better than your opponent.”\n – Siegbert Tarrasch",
	40: "“Up to this point, White has been following well-known analysis. But now he makes a fatal error: he begins to use his own head.”\n – Siegbert Tarrasch",
	41: "“Of chess, it has been said that life is not long enough for it, but that is the fault of life, not chess.”\n – William Napier",
	42: "“Chess is beautiful enough to waste your life for.”\n – Hans Ree",
	43: "“A chess game in progress is… a cosmos unto itself, fully insulated from an infant’s cry, an erotic invitation, or war.”\n – David Shenk",
	44: "“It will be cheering to know that many people are skillful chess players, though in many instances their brains, in a general way, compare unfavorably with the cognitive faculties of a rabbit.”\n – James Mortimer",
	45: "“The pin is mightier than the sword.”\n – Fred Reinfeld",
	46: "“The only thing chess players have in common is chess.”\n – Lodewijk Prins",
	47: "“Those who say they understand chess, understand nothing.”\n – Robert Hübner",
	48: "“One bad move nullifies forty good ones.”\n – Bernhard Horwitz",
	49: "“If your opponent offers you a draw, try to work out why he thinks he’s worse off.”\n – Nigel Short",
	50: "“A computer once beat me at chess, but it was no match for me at kick boxing.” - Emo Phillips",
	51: "“I did **** all, and it proved to be enough!” - Tony Miles",
	52: "“To win against me, you must beat me three times: in the opening, the middlegame and the endgame.”\n – Alexander Alekhine",
	53: "“There are two kinds of sacrifices; correct ones and mine.”\n – Mikhail Tal",
	54: "“After 1.e4 White”s game is in its last throes.” - Gyula Breyer.",
	55: "“The most important move in chess, as in life, is the one you just made.” - Unknown",
	56: "“I like chess.” - H.F. Witte",
}

var testChallenge = CreateChallengeType{
	Type:           1, //friend
	TimeOption:     0, //real time
	DestUser:       "",
	Rated:          "false",
	Color:          "black",
	Variant:        "standard",
	ClockLimit:     fmt.Sprintf("%v", int(3*60)), //minutes
	ClockIncrement: fmt.Sprintf("%v", 0),         //seconds
}

// var testChallenge = CreateChallengeType{
// 	Type:       1, //friend
// 	TimeOption: 1, //Correspondence
// 	DestUser:   "",
// 	Rated:      "false",
// 	Color:      "black",
// 	Variant:    "standard",
// 	Days:       "5",
// }

// var testChallenge = CreateChallengeType{
// 	Type:       1, //friend
// 	TimeOption: 2, //unlimited
// 	DestUser:   "",
// 	Rated:      "false",
// 	Color:      "black",
// 	Variant:    "standard",
// }

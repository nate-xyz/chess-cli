package pkg

import (
	cv "codeberg.org/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	Root               = new(State)
	newChallenge       api.CreateChallengeType //used initUI.go in initConstruct() in order to save settings from a challenge construction
	CurrentChallenge   api.CreateChallengeType
	currentGameID      string
	StreamChannel      chan api.StreamEventType
	GameListIDArr      []string
	OutChallengeGameID []string
	InChallengeGameID  []string
	caser              = cases.Title(language.English)
)

const (
	ApplicationTitle string = `[blue]chess-cli ♟️ [white]
Play locally with a [yellow]friend[white] or online with [red]Lichess!`
	LichessTitle        string = "[blue]chess-cli[white]: [red]lichess[white] client"
	LichessRibbon       string = " chess-cli · Lichess │ ↑↓ + Enter, or press a letter │ Ctrl-C quit"
	welcomeRibbonstr    string = " chess-cli │ ↑↓ + Enter, or press a letter │ Ctrl-C quit"
	gameRibbonstr       string = " chess-cli · local game │ Ctrl-K type a move │ click or arrows/hjkl + Enter to move │ Tab menu │ Ctrl-C quit"
	challengeRibbonstr  string = " chess-cli · new challenge │ ↑↓ + Enter, or press a letter │ Ctrl-C quit"
	gameOnlineRibbonstr string = " chess-cli · Lichess game │ Ctrl-K type a move │ click or arrows/hjkl + Enter to move │ Tab chat/menu │ Ctrl-C quit"
	postRibbonstr       string = " chess-cli · game over │ click a move, scroll or [ ] to look back │ Tab board ⇄ menu │ Ctrl-C quit"
	OngoingRibbonstr    string = " chess-cli · ongoing games │ ↑↓ or 1-9 to pick │ Tab menu │ Ctrl-C quit"
	savedRibbonstr      string = " chess-cli · saved games │ ↑↓ or 1-9 to pick │ Tab menu │ Ctrl-C quit"
	saved_path          string = "saved_games.json" // legacy location, see savedGamesPath
)

type State struct {
	App        *cv.Application
	nav        *cv.Panels
	Shell      string
	gameState  *GameState
	User       *Login
	lgame      *GameScreen
	sgame      *SavedGames
	pgame      *PostGameScreen
	wonline    *WelcomeOnline
	loader     *Loader
	ongame     *OnlineGame
	ponline    *OnlinePostGame
	ongoing    *Ongoing
	challenges *Challenges
	puzzle     *PuzzleScreen
	watch      *WatchScreen
	study      *StudyScreen
	sglist     *SavedGameList
}

type Login struct {
	Token              string
	Online             bool
	StreamStart        bool
	Name               string
	Friends            []string
	OngoingGames       []api.OngoingGameInfo
	IncomingChallenges []api.ChallengeInfo
	OutgoingChallenges []api.ChallengeInfo
}

type GameState struct {
	Game             *chess.Game
	NextMove         string
	MoveHistoryArray []string
	LegalMoves       []string
	Status           string
	MoveCount        int
	WhiteCaptured    []string
	BlackCaptured    []string
	Started          string
	Hash             string
	LastPlayed       string
}

type ListSelectedFunc func() //used to easily add functions to list items in initUI.go

type PiecePosition struct {
	Row    int
	Col    int
	Alg    string
	Empty  bool
	Symbol string
}

var (
	PiecesMap = map[rune]string{
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
		'p': "♟\uFE0E", // U+FE0E: text, not emoji
	}

	KnightIconMap = map[int]string{
		0: " ♞ ",
		1: " 🨇 ",
		2: " 🨓 ",
		3: " 🨜 ",
		4: " 🨨 ",
		5: " 🨱 ",
		6: " 🨽 ",
		7: " 🩆 ",
	}

	NormalLoaderMap = map[int64]string{
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

	RandQuoteMap = map[int]string{
		0:  "“I have come to the personal conclusion that while all artists are not chess players, all chess players are artists.”\n – Marcel Duchamp",
		1:  "“Unlike other games in which lucre is the end and aim, [chess] recommends itself to the wise by the fact that its mimic battles are fought for no prize but honor. \nIt is eminently and emphatically the philosopher’s game.”\n – Paul Morphy",
		2:  "“The beauty of chess is it can be whatever you want it to be. \nIt transcends language, age, race, religion, politics, gender, and socioeconomic background. \nWhatever your circumstances, anyone can enjoy a good fight to the death over the chess board.”\n – Simon Williams",
		3:  "“Chess is the struggle against the error.”\n – Johannes Zukertort",
		4:  "“Every chess master was once a beginner.”\n – Irving Chernev",
		5:  "“Avoid the crowd. \nDo your own thinking independently. \nBe the chess player, not the chess piece.”\n – Ralph Charell",
		6:  "“Chess makes men wiser and clear-sighted.”\n – Vladimir Putin",
		7:  "“Chess is the gymnasium of the mind.”\n – Blaise Pascal",
		8:  "“Chess holds its master in its own bonds, shackling the mind and brain so that the inner freedom of the very strongest must suffer.”\n – Albert Einstein",
		9:  "“Chess is a war over the board. \nThe object is to crush the opponent’s mind.”\n – Bobby Fischer",
		10: "“I am convinced, the way one plays chess always reflects the player’s personality. \nIf something defines his character, then it will also define his way of playing.”\n – Vladimir Kramnik",
		11: "“The game of chess is not merely an idle amusement. \nSeveral very valuable qualities of the mind, useful in the course of human life, are to be acquired or strengthened by it… Life is a kind of Chess, in which we have often points to gain, and competitors or adversaries to contend with.”\n – Benjamin Franklin",
		12: "“As proved by evidence, [chess is] more lasting in its being and presence than all books and achievements; the only game that belongs to all people and all ages; of which none knows the divinity that bestowed it on the world, to slay boredom, to sharpen the senses, to exhilarate the spirit.”\n – Stefan Zweig",
		13: "“Chess doesn’t drive people mad, it keeps mad people sane.”\n – Bill Hartston",
		14: "“In life, as in chess, one’s own pawns block one’s way. \n A man’s very wealthy, ease, leisure, children, books, which should help him to win, more often checkmate him.”\n – Charles Buxton",
		15: "“Chess is life in miniature. \nChess is a struggle, chess battles.”\n – Garry Kasparov",
		16: "“Chess, like love, like music, has the power to make men happy.”\n – Siegbert Tarrasch",
		17: "“For in the idea of chess and the development of the chess mind we have a picture of the intellectual struggle of mankind.”\n – Richard Réti",
		18: "“I don’t believe in psychology. \nI believe in good moves.”\n – Bobby Fischer",
		19: "“Play the opening like a book, the middlegame like a magician, and the endgame like a machine.”\n – Rudolph Spielmann",
		20: "“I used to attack because it was the only thing I knew. \nNow I attack because I know it works best.”\n – Garry Kasparov",
		21: "“It is my style to take my opponent and myself on to unknown grounds. \nA game of chess is not an examination of knowledge; it is a battle of nerves.”\n – David Bronstein",
		22: "“Chess is rarely a game of ideal moves. \nAlmost always, a player faces a series of difficult consequences whichever move he makes.”\n – David Shenk",
		23: "“When you see a good move, look for a better one.”\n – Emanuel Lasker",
		24: "“After a bad opening, there is hope for the middle game. \nAfter a bad middle game, there is hope for the endgame. \nBut once you are in the endgame, the moment of truth has arrived.”\n – Edmar Mednis",
		25: "“Give me a difficult positional game, I will play it. \nBut totally won positions, I cannot stand them.”\n – Hein Donner",
		26: "“There is no remorse like the remorse of chess.”\n – H. G. Wells",
		27: "“Half the variations which are calculated in a tournament game turn out to be completely superfluous. \nUnfortunately, no one knows in advance which half.”\n – Jan Timman",
		28: "“Even a poor plan is better than no plan at all.”\n – Mikhail Chigorin",
		29: "“Tactics is knowing what to do when there is something to do; strategy is knowing what to do when there is nothing to do.”\n – Savielly Tartakower",
		30: "“In life, as in chess, forethought wins.”\n – Charles Buxton",
		31: "“You may learn much more from a game you lose than from a game you win. \nYou will have to lose hundreds of games before becoming a good player.”\n – José Raúl Capablanca",
		32: "“Pawns are the soul of the game.”\n – François-André Danican Philidor",
		33: "“The passed pawn is a criminal, who should be kept under lock and key. \nMild measures, such as police surveillance, are not sufficient.”\n – Aron Nimzowitsch",
		34: "“Modern chess is too much concerned with things like pawn structure. \nForget it, checkmate ends the game.”\n – Nigel Short",
		35: "“Pawn endings are to chess what putting is to golf.”\n – Cecil Purdy",
		36: "“Nobody ever won a chess game by resigning.”\n – Savielly Tartakower",
		37: "“The blunders are all there on the board, waiting to be made.”\n – Savielly Tartakower",
		38: "“It’s always better to sacrifice your opponent’s men.”\n – Savielly Tartakower",
		39: "“One doesn’t have to play well, it’s enough to play better than your opponent.”\n – Siegbert Tarrasch",
		40: "“Up to this point, White has been following well-known analysis. \nBut now he makes a fatal error: he begins to use his own head.”\n – Siegbert Tarrasch",
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

	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

package shared

// #include <sys/ioctl.h>
import "C"

import (
	"math/rand"
	"syscall"
	"time"
	"unsafe"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func get_index(s []string, str string) int {
	for i, v := range s {
		if v == str {
			return i
		}
	}
	return -1
}

func GetRandomQuote() string {
	var rand_quote_map = map[int]string{

		0:  "'I have come to the personal conclusion that while all artists are not chess players, all chess players are artists.' – Marcel Duchamp",
		1:  "'Unlike other games in which lucre is the end and aim, [chess] recommends itself to the wise by the fact that its mimic battles are fought for no prize but honor. It is eminently and emphatically the philosopher’s CurrentGame.' – Paul Morphy",
		2:  "'The beauty of chess is it can be whatever you want it to be. It transcends language, age, race, religion, politics, gender, and socioeconomic background. Whatever your circumstances, anyone can enjoy a good fight to the death over the chess board.' – Simon Williams",
		3:  "'Chess is the struggle against the error.' – Johannes Zukertort",
		4:  "'Every chess master was once a beginner.' – Irving Chernev",
		5:  "'Avoid the crowd. Do your own thinking independently. Be the chess player, not the chess piece.' – Ralph Charell",
		6:  "'Chess makes men wiser and clear-sighted.' – Vladimir Putin",
		7:  "'Chess is the gymnasium of the mind.' – Blaise Pascal",
		8:  "'Chess holds its master in its own bonds, shackling the mind and brain so that the inner freedom of the very strongest must suffer.' – Albert Einstein",
		9:  "'Chess is a war over the board. The object is to crush the opponent’s mind.' – Bobby Fischer",
		10: "'I am convinced, the way one plays chess always reflects the player’s personality. If something defines his character, then it will also define his way of playing.' – Vladimir Kramnik",
		11: "'The game of chess is not merely an idle amusement. Several very valuable qualities of the mind, useful in the course of human life, are to be acquired or strengthened by it… Life is a kind of Chess, in which we have often points to gain, and competitors or adversaries to contend with.' – Benjamin Franklin",
		12: "'As proved by evidence, [chess is] more lasting in its being and presence than all books and achievements; the only game that belongs to all people and all ages; of which none knows the divinity that bestowed it on the world, to slay boredom, to sharpen the senses, to exhilarate the spirit.' – Stefan Zweig",
		13: "'Chess doesn’t drive people mad, it keeps mad people sane.' – Bill Hartston",
		14: "'In life, as in chess, one’s own pawns block one’s way.  A man’s very wealthy, ease, leisure, children, books, which should help him to win, more often checkmate him.' – Charles Buxton",
		15: "'Chess is life in miniature. Chess is a struggle, chess battles.' – Garry Kasparov",
		16: "'Chess, like love, like music, has the power to make men happy.' – Siegbert Tarrasch",
		17: "'For in the idea of chess and the development of the chess mind we have a picture of the intellectual struggle of mankind.' – Richard Réti",
		18: "'I don’t believe in psychology. I believe in good moves.' – Bobby Fischer",
		19: "'Play the opening like a book, the middlegame like a magician, and the endgame like a machine.' – Rudolph Spielmann",
		20: "'I used to attack because it was the only thing I knew. Now I attack because I know it works best.' – Garry Kasparov",
		21: "'It is my style to take my opponent and myself on to unknown grounds. A game of chess is not an examination of knowledge; it is a battle of nerves.' – David Bronstein",
		22: "'Chess is rarely a game of ideal moves. Almost always, a player faces a series of difficult consequences whichever move he makes.' – David Shenk",
		23: "'When you see a good move, look for a better one.' – Emanuel Lasker",
		24: "'After a bad opening, there is hope for the middle CurrentGame. After a bad middle game, there is hope for the endgame. But once you are in the endgame, the moment of truth has arrived.' – Edmar Mednis",
		25: "'Give me a difficult positional game, I will play it. But totally won positions, I cannot stand them.' – Hein Donner",
		26: "'There is no remorse like the remorse of chess.' – H. G. Wells",
		27: "'Half the variations which are calculated in a tournament game turn out to be completely superfluous. Unfortunately, no one knows in advance which half.' – Jan Timman",
		28: "'Even a poor plan is better than no plan at all.' – Mikhail Chigorin",
		29: "'Tactics is knowing what to do when there is something to do; strategy is knowing what to do when there is nothing to do.' – Savielly Tartakower",
		30: "'In life, as in chess, forethought wins.' – Charles Buxton",
		31: "'You may learn much more from a game you lose than from a game you win. You will have to lose hundreds of games before becoming a good player.' – José Raúl Capablanca",
		32: "'Pawns are the soul of the CurrentGame.' – François-André Danican Philidor",
		33: "'The passed pawn is a criminal, who should be kept under lock and key. Mild measures, such as police surveillance, are not sufficient.' – Aron Nimzowitsch",
		34: "'Modern chess is too much concerned with things like pawn structure. Forget it, checkmate ends the CurrentGame.' – Nigel Short",
		35: "'Pawn endings are to chess what putting is to golf.' – Cecil Purdy",
		36: "'Nobody ever won a chess game by resigning.' – Savielly Tartakower",
		37: "'The blunders are all there on the board, waiting to be made.' – Savielly Tartakower",
		38: "'It’s always better to sacrifice your opponent’s men.' – Savielly Tartakower",
		39: "'One doesn’t have to play well, it’s enough to play better than your opponent.' – Siegbert Tarrasch",
		40: "'Up to this point, White has been following well-known analysis. But now he makes a fatal error: he begins to use his own head.' – Siegbert Tarrasch",
		41: "'Of chess, it has been said that life is not long enough for it, but that is the fault of life, not chess.' – William Napier",
		42: "'Chess is beautiful enough to waste your life for.' – Hans Ree",
		43: "'A chess game in progress is… a cosmos unto itself, fully insulated from an infant’s cry, an erotic invitation, or war.' – David Shenk",
		44: "'It will be cheering to know that many people are skillful chess players, though in many instances their brains, in a general way, compare unfavorably with the cognitive faculties of a rabbit.' – James Mortimer",
		45: "'The pin is mightier than the sword.' – Fred Reinfeld",
		46: "'The only thing chess players have in common is chess.' – Lodewijk Prins",
		47: "'Those who say they understand chess, understand nothing.' – Robert Hübner",
		48: "'One bad move nullifies forty good ones.' – Bernhard Horwitz",
		49: "'If your opponent offers you a draw, try to work out why he thinks he’s worse off.' – Nigel Short",
		50: "'A computer once beat me at chess, but it was no match for me at kick boxing.' - Emo Phillips",
		51: "'I did **** all, and it proved to be enough!' - Tony Miles",
		52: "'To win against me, you must beat me three times: in the opening, the middlegame and the endgame.' – Alexander Alekhine",
		53: "'There are two kinds of sacrifices; correct ones and mine.' – Mikhail Tal",
		54: "'After 1.e4 White's game is in its last throes.' - Gyula Breyer.",
		55: "'The most important move in chess, as in life, is the one you just made.' - Unknown",
		56: "'I like chess.' - H.F. Witte",
	}

	rand.Seed(time.Now().UnixNano())
	minrand := 0
	maxrand := 56
	rand_quote_int := rand.Intn(maxrand-minrand+1) + minrand

	var rand_quote string = rand_quote_map[rand_quote_int]

	return rand_quote

}

func GetMaxLenStr(arr []string) int {
	max_len := 0
	for _, str := range arr {
		if max_len < len(str) {
			max_len = len(str)
		}
	}
	return max_len
}

func OsTermSize() (int, int, error) {
	w := &C.struct_winsize{}
	res, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(w)),
	)
	if int(res) == -1 {
		return 0, 0, err
	}
	return int(w.ws_row), int(w.ws_col), nil
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

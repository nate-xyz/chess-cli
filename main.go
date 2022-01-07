package main

// #include <sys/ioctl.h>
import "C"

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	ncurses "github.com/nate-xyz/goncurses_"
	"github.com/notnil/chess"
)

//#f3 e5 g4 Qh4#

var sigs chan os.Signal
var noti_message chan string
var error_message chan error
var ready chan struct{}
var curChallenge CreateChallengeType
var waiting_alert chan StreamEventType

// var quit_stream chan bool
// var stream_done chan struct{}

func main() {
	//init channels
	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)
	noti_message = make(chan string, 100)
	error_message = make(chan error, 10)
	ready = make(chan struct{})
	stream_channel = make(chan StreamEventType, 1)
	// Initialize ncurses. It's essential End() is called to ensure the
	// terminal isn't altered after the program ends
	stdscr, err := ncurses.Init()
	if err != nil {
		log.Fatal("init", err)
	}
	defer ncurses.End()
	stdscr.Timeout(0)

	go StreamEvent(stream_channel, ready)
	go StreamConsumer(stream_channel, noti_message)
	go notifier(stdscr, noti_message)
	go ncurses_print_error(stdscr, error_message)

	// go func(screen *ncurses.Window) {
	// 	rand.Seed(time.Now().UnixNano())
	// 	for {
	// 		rando := rand.Intn(100)
	// 		go screen.MovePrint(10, 100, fmt.Sprintf(" verification %v ", rando))
	// 		noti_message <- fmt.Sprintf("this is a test %v", rando)
	// 		//error_message <- fmt.Errorf("this is a test %v", rando)
	// 		screen.MovePrint(10, 100, fmt.Sprintf(" verification %v ", rando))
	// 		time.Sleep(time.Millisecond * 500)
	// 	}
	// }(stdscr)

	// ticker := time.NewTicker(time.Second)

	// func(screen *ncurses.Window) {
	// 	for {

	// 		select {

	// 		case <-ticker.C:
	// 			error_message <- fmt.Errorf("error test")
	// 			screen.Clear()
	// 		default:
	// 			screen.MovePrint(1, 1, "not looooooading")
	// 			screen.Refresh()
	// 		}

	// 	}
	// }(stdscr)
	//<-ready

	//necessary for mouse input, start keypad, read all mouse events
	stdscr.Keypad(true)
	mouse_int := ncurses.M_ALL | ncurses.M_POSITION
	_ = mouse_int
	//ncurses.MouseMask(mouse_int, nil)
	//fmt.Printf("\033[?1003h")

	// allow input, Start colors in goncurses
	ncurses.Echo(true)   //allow input
	ncurses.Cursor(0)    //set cursor visibility to hidden
	ncurses.StartColor() //allow color to be displayed

	//ncurses.use_default_colors()
	ncurses.InitPair(1, ncurses.C_CYAN, ncurses.C_BLACK)
	ncurses.InitPair(2, ncurses.C_RED, ncurses.C_BLACK)
	ncurses.InitPair(3, ncurses.C_BLACK, ncurses.C_WHITE)

	//piece and square colors
	if ncurses.CanChangeColor() {
		var light_square int16 = 215 //SandyBrown
		var dark_square int16 = 94   //Orange4
		var light_piece int16 = 230  //Cornsilk1
		var dark_piece int16 = 233   //Grey7
		ncurses.InitPair(4, light_piece, light_square)
		ncurses.InitPair(5, light_piece, dark_square)
		ncurses.InitPair(6, dark_piece, light_square)
		ncurses.InitPair(7, dark_piece, dark_square)

		//floating piece colors
		ncurses.InitPair(10, light_piece, dark_piece)
		ncurses.InitPair(11, dark_piece, light_piece)
		ncurses.InitPair(12, ncurses.C_RED, ncurses.C_WHITE)
		ncurses.InitPair(13, ncurses.C_RED, ncurses.C_BLACK)
		ncurses.InitPair(14, ncurses.C_BLUE, ncurses.C_WHITE)
		ncurses.InitPair(15, ncurses.C_BLUE, ncurses.C_BLACK)
		ncurses.InitPair(16, dark_piece, ncurses.C_BLACK)
		ncurses.InitPair(17, light_piece, ncurses.C_BLACK)
	} else {
		ncurses.InitPair(4, ncurses.C_RED, ncurses.C_WHITE)
		ncurses.InitPair(5, ncurses.C_RED, ncurses.C_BLACK)
		ncurses.InitPair(6, ncurses.C_BLUE, ncurses.C_WHITE)
		ncurses.InitPair(7, ncurses.C_BLUE, ncurses.C_BLACK)
	}
	//move legality colors
	ncurses.InitPair(8, ncurses.C_BLACK, ncurses.C_GREEN)
	ncurses.InitPair(9, ncurses.C_WHITE, ncurses.C_RED)

	var key ncurses.Key = one_key
	if !dev_mode {
		key = zero_key
	}
	mainScreenHandler(stdscr, key)
	ncurses.FlushInput()
	ncurses.Echo(false) //turn off input
	ncurses.End()
}

//screen handlers
func mainScreenHandler(stdscr *ncurses.Window, key ncurses.Key) {
	switch key {
	case zero_key:
		key = welcome_screen(stdscr) //go back to welcome screen
	case one_key:
		_, key = localGameHandler(stdscr, local_game_screen(stdscr)) //go to local game screen, two player with chess lib
	case two_key:
		_, key = lichessScreenHandler(stdscr, lichess_welcome(stdscr)) //go to lichess welcome screen, login w oauth
	case three_key:
		return //go to stockfish ai screen, todo
	case control_o_key:
		return //quit game
	}
	mainScreenHandler(stdscr, key)
}

func localGameHandler(stdscr *ncurses.Window, option int) (int, ncurses.Key) {
	switch option {
	case 0:
		return -1, zero_key //go back to welcome screen
	case 1:
		option = local_game_screen(stdscr) //go to game screen
	case 2:
		option = post_screen(stdscr)
	case 3:
		return -1, control_o_key //quit game
	}
	return localGameHandler(stdscr, option)
}

func lichessScreenHandler(stdscr *ncurses.Window, option int) (int, ncurses.Key) {
	switch option {
	case 0:
		return -1, zero_key //go back to welcome screen
	case 1:
		option = lichess_welcome(stdscr)
	case 2:
		option = lichess_challenges(stdscr) //go to challenge screen
	case 3:
		option = create_game(stdscr)
	case 4:
		option = lichess_game_wait(stdscr)
	case 5:
		return -1, control_o_key //quit game
	case 6:
		option = lichess_game(stdscr)
	}
	return lichessScreenHandler(stdscr, option)
}

func game_logic(board_window *ncurses.Window) bool {
	//inputted_str = inputted_str.strip(' ').strip('\0').strip('^@')
	inputted_str = strings.Trim(inputted_str, " ^@")
	board_window.MovePrint(1, 1, inputted_str)
	legal_moves := game.ValidMoves()
	legal_move_str_array = []string{}

	for _, move := range legal_moves {
		legal_move_str_array = append(legal_move_str_array, move.String())
	}

	if entered_move {
		entered_move = false

		if err := game.MoveStr(inputted_str); err != nil {
			status_str = "last input is invalid"
			inputted_str = ""
		} else {
			status_str = "move is legal!"
			last_move_str = inputted_str //set the last move string to be displayed in the info window
			history_arr = append([]string{inputted_str}, history_arr...)
			move_amount++ //increment the global move amount for the history window
			ncurses.Flash()
			ncurses.Beep()

			if game.Outcome() != chess.NoOutcome { //check if the game is over
				status_str = game.Method().String()
				final_position = game.Position().Board().Draw()
				return true
			}
		}

	}

	legal_moves = game.ValidMoves()
	return false
}

func lichess_game_logic(board_window *ncurses.Window) bool {
	//inputted_str = inputted_str.strip(' ').strip('\0').strip('^@')
	inputted_str = strings.Trim(inputted_str, " ^@")
	board_window.MovePrint(1, 1, inputted_str)
	legal_moves := game.ValidMoves()
	legal_move_str_array = []string{}

	for _, move := range legal_moves {
		legal_move_str_array = append(legal_move_str_array, move.String())
	}

	if entered_move {
		entered_move = false

		if err := game.MoveStr(inputted_str); err != nil {
			status_str = "last input is invalid"
			inputted_str = ""
		} else {
			status_str = "move is legal!"
			last_move_str = inputted_str //set the last move string to be displayed in the info window
			history_arr = append([]string{inputted_str}, history_arr...)
			move_amount++ //increment the global move amount for the history window
			ncurses.Flash()
			ncurses.Beep()

			if game.Outcome() != chess.NoOutcome { //check if the game is over
				status_str = game.Method().String()
				final_position = game.Position().Board().Draw()
				return true
			}
		}

	}

	legal_moves = game.ValidMoves()
	return false
}

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
		1:  "'Unlike other games in which lucre is the end and aim, [chess] recommends itself to the wise by the fact that its mimic battles are fought for no prize but honor. It is eminently and emphatically the philosopher’s game.' – Paul Morphy",
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
		24: "'After a bad opening, there is hope for the middle game. After a bad middle game, there is hope for the endgame. But once you are in the endgame, the moment of truth has arrived.' – Edmar Mednis",
		25: "'Give me a difficult positional game, I will play it. But totally won positions, I cannot stand them.' – Hein Donner",
		26: "'There is no remorse like the remorse of chess.' – H. G. Wells",
		27: "'Half the variations which are calculated in a tournament game turn out to be completely superfluous. Unfortunately, no one knows in advance which half.' – Jan Timman",
		28: "'Even a poor plan is better than no plan at all.' – Mikhail Chigorin",
		29: "'Tactics is knowing what to do when there is something to do; strategy is knowing what to do when there is nothing to do.' – Savielly Tartakower",
		30: "'In life, as in chess, forethought wins.' – Charles Buxton",
		31: "'You may learn much more from a game you lose than from a game you win. You will have to lose hundreds of games before becoming a good player.' – José Raúl Capablanca",
		32: "'Pawns are the soul of the game.' – François-André Danican Philidor",
		33: "'The passed pawn is a criminal, who should be kept under lock and key. Mild measures, such as police surveillance, are not sufficient.' – Aron Nimzowitsch",
		34: "'Modern chess is too much concerned with things like pawn structure. Forget it, checkmate ends the game.' – Nigel Short",
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

func notifier(screen *ncurses.Window, message <-chan string) {
	for {
		select {
		case m := <-message:
			title := "notification"
			_, s_width := screen.MaxYX()
			//x := rand.Intn(width) + 1
			//y := rand.Intn(height) + 1

			w := getMaxLenStr([]string{m, title}) + 2
			x := s_width - w - 1
			//y := rand.Intn(height) + 1

			//win, _ := ncurses.NewWindow(3, 20, 1, width-20)

			timeout := time.After(time.Second * 2)
		loop:
			for tick := range time.Tick(time.Millisecond * 10) {
				_ = tick
				select {
				// case <-sigs:
				// 	_, s_width := screen.MaxYX()
				// 	x := s_width - w - 1
				// 	win.Clear()
				// 	win.MoveWindow(1, x) //move windows to appropriate locations
				// 	win.Box('|', '-')
				// 	win.AttrOn(ncurses.ColorPair(2))
				// 	win.AttrOn(ncurses.A_BOLD)
				// 	win.MovePrint(0, 1, title)
				// 	win.AttrOff(ncurses.ColorPair(2))
				// 	win.AttrOff(ncurses.A_BOLD)
				// 	win.MovePrint(1, 1, m)
				// 	win.Refresh()
				case <-timeout:
					break loop
				default:
					win, _ := ncurses.NewWindow(3, w, 1, x)
					win.Box('|', '-')
					win.AttrOn(ncurses.ColorPair(2))
					win.AttrOn(ncurses.A_BOLD)
					win.MovePrint(0, 1, title)
					win.AttrOff(ncurses.ColorPair(2))
					win.AttrOff(ncurses.A_BOLD)
					win.MovePrint(1, 1, m)
					win.NoutRefresh()
				}
			}
			//time.Sleep(time.Second * 5)
			screen.Clear()
			sigs <- syscall.SIGWINCH
			//time.Sleep(time.Second * 1)
			// default:
			// 	height, width := screen.MaxYX()
			// 	rand.Seed(time.Now().UnixNano())
			// 	x := rand.Intn(width) + 1
			// 	y := rand.Intn(height) + 1
			// 	screen.MovePrint(y, x, "no message!")
			// 	time.Sleep(time.Millisecond * 10)

		}
	}
}

func ncurses_print_error(screen *ncurses.Window, message <-chan error) {
	for {
		select {
		case m := <-message:
			title := "error"
			h, s_width := screen.MaxYX()

			w := getMaxLenStr([]string{fmt.Sprintf("%v", m), title}) + 2
			x := (s_width/2 - w/2 - w%2)
			y := (h / 2)

			timeout := time.After(time.Second * 5)
		loop:
			for tick := range time.Tick(time.Millisecond * 10) {
				_ = tick
				select {
				case <-timeout:
					break loop
				default:
					win, _ := ncurses.NewWindow(3, w, y, x)
					win.Box('|', '-')
					win.AttrOn(ncurses.ColorPair(2))
					win.AttrOn(ncurses.A_BOLD)
					win.MovePrint(0, 1, title)
					win.AttrOff(ncurses.ColorPair(2))
					win.AttrOff(ncurses.A_BOLD)
					win.MovePrint(1, 1, fmt.Sprintf("%v", m))
					win.NoutRefresh()
				}
			}
			screen.Clear()
			sigs <- syscall.SIGWINCH
			time.Sleep(time.Second * 1)
			//os.Exit(1)
		}
	}
}

func getMaxLenStr(arr []string) int {
	max_len := 0
	for _, str := range arr {
		if max_len < len(str) {
			max_len = len(str)
		}
	}
	return max_len
}

func osTermSize() (int, int, error) {
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

func containedInEventStream(a []StreamEventType, gameid string) (string, bool) {
	for _, e := range a {
		if e.Id == gameid {
			return e.Event, true
		}
	}
	return "", false
}

func containedInOngoingGames(a []OngoingGameInfo, gameid string) bool {
	for _, g := range a {
		if g.GameID == gameid {
			return true
		}
	}
	return false
}

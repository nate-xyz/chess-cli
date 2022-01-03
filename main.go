package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	ncurses "github.com/nate-xyz/goncurses"
	"github.com/notnil/chess"

	"log"
)

//#f3 e5 g4 Qh4#
var sigs chan os.Signal

func main() {
	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)
	// Initialize ncurses. It's essential End() is called to ensure the
	// terminal isn't altered after the program ends
	stdscr, err := ncurses.Init()
	if err != nil {
		log.Fatal("init", err)
	}

	defer ncurses.End()
	stdscr.Timeout(0)

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
		key = lichessScreenHandler(stdscr, lichess_welcome(stdscr)) //go to lichess welcome screen, login w oauth
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

func lichessScreenHandler(stdscr *ncurses.Window, key ncurses.Key) ncurses.Key {
	switch key {
	case zero_key:
		return key //go to welcome screen
	case one_key:
		key = lichess_welcome(stdscr)
	case two_key:
		key = lichess_challenges(stdscr) //go to challenge screen
	case three_key:
		return zero_key
		//key = lichess_games(stdscr) //see ongoing games
	case four_key:
		return zero_key //puzzles?
	case control_o_key:
		return key //quit game
	}
	return lichessScreenHandler(stdscr, key)
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

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

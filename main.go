package main

import (
	"fmt"
	"strings"

	"github.com/nate-xyz/goncurses"
	"github.com/notnil/chess"

	// "net/http"
	"log"
	// "io/ioutil"
	// "reflect"
)

//#f3 e5 g4 Qh4#

func main() {

	// Initialize goncurses. It's essential End() is called to ensure the
	// terminal isn't altered after the program ends
	stdscr, err := goncurses.Init()
	if err != nil {
		log.Fatal("init", err)
	}

	defer goncurses.End()

	//necessary for mouse input, start keypad, read all mouse events
	stdscr.Keypad(true)
	mouse_int := goncurses.M_ALL | goncurses.M_POSITION
	_ = mouse_int
	goncurses.MouseMask(mouse_int, nil)
	fmt.Printf("\033[?1003h")

	// allow input, Start colors in goncurses
	goncurses.Echo(true)   //allow input
	goncurses.Cursor(0)    //set cursor visibility to hidden
	goncurses.StartColor() //allow color to be displayed

	//goncurses.use_default_colors()
	goncurses.InitPair(1, goncurses.C_CYAN, goncurses.C_BLACK)
	goncurses.InitPair(2, goncurses.C_RED, goncurses.C_BLACK)
	goncurses.InitPair(3, goncurses.C_BLACK, goncurses.C_WHITE)

	//piece and square colors
	if goncurses.CanChangeColor() {
		var light_square int16 = 215 //SandyBrown
		var dark_square int16 = 94   //Orange4
		var light_piece int16 = 230  //Cornsilk1
		var dark_piece int16 = 233   //Grey7
		goncurses.InitPair(4, light_piece, light_square)
		goncurses.InitPair(5, light_piece, dark_square)
		goncurses.InitPair(6, dark_piece, light_square)
		goncurses.InitPair(7, dark_piece, dark_square)

		//floating piece colors
		goncurses.InitPair(10, light_piece, dark_piece)
		goncurses.InitPair(11, dark_piece, light_piece)
	} else {
		goncurses.InitPair(4, goncurses.C_RED, goncurses.C_WHITE)
		goncurses.InitPair(5, goncurses.C_RED, goncurses.C_BLACK)
		goncurses.InitPair(6, goncurses.C_BLUE, goncurses.C_WHITE)
		goncurses.InitPair(7, goncurses.C_BLUE, goncurses.C_BLACK)
	}
	//move legality colors
	goncurses.InitPair(8, goncurses.C_BLACK, goncurses.C_GREEN)
	goncurses.InitPair(9, goncurses.C_WHITE, goncurses.C_RED)

	if !dev_mode {
		welcome_screen(stdscr)
	}
	local_game_screen(stdscr)
	goncurses.FlushInput()
	goncurses.Echo(false) //turn off input
	goncurses.End()
}

//
// //                                                  888                   d8b
// //                                                  888                   Y8P
// //                                                  888
// //  .d88b.   8888b.  88888b.d88b.   .d88b.          888  .d88b.   .d88b.  888  .d8888b
// // d88P"88b     "88b 888 "888 "88b d8P  Y8b         888 d88""88b d88P"88b 888 d88P"
// // 888  888 .d888888 888  888  888 88888888         888 888  888 888  888 888 888
// // Y88b 888 888  888 888  888  888 Y8b.             888 Y88..88P Y88b 888 888 Y88b.
// //  "Y88888 "Y888888 888  888  888  "Y8888 88888888 888  "Y88P"   "Y88888 888  "Y8888P
// //      888                                                           888
// // Y8b d88P                                                      Y8b d88P
// //  "Y88P"                                                        "Y88P"

func game_logic(board_window *goncurses.Window) {
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
			goncurses.Flash()
			goncurses.Beep()

			if game.Outcome() != chess.NoOutcome { //check if the game is over
				status_str = game.Method().String()
				final_position = game.Position().Board().Draw()
				post_screen_toggle = true
			}
		}

	}

	//draw board
	draw_board(board_window)
	legal_moves = game.ValidMoves()
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

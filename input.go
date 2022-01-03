package main

import (
	"fmt"
	"strings"
	"unicode"

	ncurses "github.com/nate-xyz/goncurses"
)

func update_input(prompt_window *ncurses.Window, key ncurses.Key) {
	height, width := prompt_window.MaxYX()
	padding := fmt.Sprintf("%s", strings.Repeat(" ", (width-1)))
	var currentPoint string = string(rune(8248))

	if key == ncurses.KEY_MOUSE || key == ncurses.KEY_RESIZE { //dont do any input for mouse event
		return
	}
	if key == delete_key || key == ncurses.KEY_BACKSPACE { //delete key
		var delete_x int
		if prompt_x_coord-1 <= 0 {
			delete_x = 1
		} else {
			delete_x = prompt_x_coord - 1
		}

		prompt_window.MovePrint(prompt_y_coord, delete_x, currentPoint)
		prompt_window.MoveAddChar(prompt_y_coord, delete_x+1, ' ') //clear last char printed
		prompt_x_coord--                                           //decrement char position
		user_input_string = removeLastRune(user_input_string)

	}
	if key == ncurses.KEY_ENTER || key == ncurses.KEY_RETURN || key == enter_key { //enter key
		entered_move = true
		inputted_str = user_input_string //set global string to check if move is legal
		user_input_string = ""           //reset input buffer
		prompt_x_coord = 1               //reset char coordinates
		prompt_y_coord = 1               //reset char coordinates
		prompt_window.MoveAddChar(prompt_y_coord, 0, '|')
		prompt_window.MoveAddChar(prompt_y_coord, 0, '>')

		for i := 1; i < height-1; i++ { //clear window
			prompt_window.MovePrint(i, prompt_x_coord, padding)
		}
		return
	}
	//if the key entered is an input char:
	if unicode.IsLetter(rune(key)) || unicode.IsDigit(rune(key)) || key == octothorpe || key == plus_sign {
		prompt_window.MovePrint(prompt_y_coord, prompt_x_coord+1, currentPoint) //indicate char youre on
		prompt_window.MoveAddChar(prompt_y_coord, prompt_x_coord, ncurses.Char(key))
		prompt_x_coord++ //increment char position
	}

	//adjust coordinates
	if prompt_x_coord <= 0 {
		prompt_window.MoveAddChar(prompt_y_coord, 1, ' ') //clear last char pointer
		prompt_x_coord = width - 2
		prompt_y_coord--
	}
	if prompt_y_coord <= 0 {
		prompt_x_coord = 1
		prompt_y_coord = 1
	}
	if prompt_x_coord >= width-1 {
		prompt_x_coord = 1
		prompt_y_coord++
	}
	if prompt_y_coord >= height-1 {
		prompt_x_coord = width - 2
		prompt_y_coord = height - 2
		status_str = "char limit reached"
		return
	}

	//add to the current input buffer
	if key != ncurses.KEY_ENTER && key != ncurses.KEY_RETURN && key != delete_key && (unicode.IsLetter(rune(key)) || unicode.IsDigit(rune(key)) || key == octothorpe || key == plus_sign) { //not enter and not delete
		user_input_string += string(rune(key))
	}
	//redraw border in case it was painted over
	prompt_window.Box('|', '-')
	prompt_window.MoveAddChar(prompt_y_coord, 0, '>') //indicate the line that you're on.
}

// func board_window_mouse_input(screen, key, screen_width, screen_height) {
//     height, width = screen.MaxYX()

//     if key != ncurses.KEY_MOUSE: //input needs to be mouse input
//         return

//     try:
//         _, mouse_x, mouse_y, _, button_state =  ncurses.getmouse()
//         bs_str = "none"
//         if button_state & ncurses.BUTTON1_PRESSED != 0:
//             bs_str = "b1 pressed"
//             mouse_pressed = True
//         if button_state & ncurses.BUTTON1_RELEASED != 0:
//             bs_str = "b1 released"
//             mouse_pressed = False
//             floating = False

//         screen.MovePrint(2, 2, "mouse_x: {} mouse_y: {} button_state: {}".format( str(mouse_x), str(mouse_y), bs_str))
//         key_tuple = (mouse_x, mouse_y)
//         if key_tuple in board_square_coord.keys() and mouse_pressed:
//             screen.MovePrint(6, 2, "has key")
//             piece_str = board_square_coord[key_tuple][1]
//             if piece_str != None and not floating:
//                 floating = True
//                 floating_piece = board_square_coord[key_tuple]
//                 screen.MovePrint(5, 2, "piece is {}".format(piece_str ))
//         if mouse_pressed:
//             color_pair = floating_piece[0]
//             screen.AttrOn(ncurses.ColorPair(color_pair))
//             screen.AttrOn(ncurses.A_BOLD)
//             screen.MovePrint(mouse_y, mouse_x, floating_piece[1]+" ")
//             screen.AttrOn(ncurses.ColorPair(color_pair))
//             screen.AttrOn(ncurses.A_BOLD)
//     except:
//         screen.MovePrint(7, 2, "error")

func removeLastRune(s string) string {
	if len(s) <= 0 {
		return s
	}
	r := []rune(s)
	return string(r[:len(r)-1])
}

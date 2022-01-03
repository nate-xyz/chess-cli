package main

import (
	"fmt"
	"strings"
	"unicode"

	ncurses "github.com/nate-xyz/goncurses"
)

func options_input(window *ncurses.Window, key ncurses.Key, options []string, selected_index int) (int, bool) {
	_, width := window.MaxYX()

	//determine what option to choose based on input
	switch key {
	case ncurses.KEY_MOUSE, ncurses.KEY_RESIZE: //dont do any input for mouse event
		return selected_index, false
	case ncurses.KEY_ENTER, ncurses.KEY_RETURN, ncurses.KEY_RIGHT:
		return selected_index, true
	case ncurses.KEY_UP:
		selected_index--
		if selected_index < 0 {
			selected_index = len(options) - 1
		}
	case ncurses.KEY_DOWN:
		selected_index++
		if selected_index >= len(options) {
			selected_index = 0
		}
	case ncurses.KEY_LEFT:
		if options[selected_index] == "back" || options[selected_index] == "quit" {
			return selected_index, true
		}
		var quit_i int = -1
		for i, str := range options {
			if str == "back" {
				return i, true
			}
			if str == "quit" {
				quit_i = i
			}
		}
		if quit_i != -1 {
			return quit_i, true
		}
	}

	//draw standout for currently selected option
	for i, str := range options {
		if i == selected_index {
			window.AttrOn(ncurses.ColorPair(3))
			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
			window.AttrOff(ncurses.ColorPair(3))
		} else {
			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
		}
	}

	//redraw border in case it was painted over
	window.Box('|', '-')
	window.Refresh()
	return selected_index, false
}

func update_input(window *ncurses.Window, key ncurses.Key) {
	height, width := window.MaxYX()
	padding := fmt.Sprintf("%s", strings.Repeat(" ", (width-1)))
	var currentPoint string = string(rune(8248))

	//adjust coordinates
	if prompt_x_coord <= 0 {
		window.MoveAddChar(prompt_y_coord, 1, ' ') //clear last char pointer
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
	}

	if key == ncurses.KEY_MOUSE || key == ncurses.KEY_RESIZE { //dont do any input for mouse event
		return
	}

	if key == delete_key || key == ncurses.KEY_BACKSPACE { //delete key
		if prompt_x_coord <= 1 {
			window.MoveAddChar(prompt_y_coord, prompt_x_coord-1, '|')
		}
		window.MoveAddChar(prompt_y_coord, prompt_x_coord+1, ' ') //clear last char printed

		window.AttrOn(ncurses.A_BLINK)
		window.MovePrint(prompt_y_coord, prompt_x_coord, currentPoint)
		window.AttrOff(ncurses.A_BLINK)

		prompt_x_coord-- //decrement char position
		user_input_string = removeLastRune(user_input_string)

	}
	if key == ncurses.KEY_ENTER || key == ncurses.KEY_RETURN || key == enter_key { //enter key
		entered_move = true
		inputted_str = user_input_string //set global string to check if move is legal
		user_input_string = ""           //reset input buffer
		prompt_x_coord = 1               //reset char coordinates
		prompt_y_coord = 1               //reset char coordinates
		window.MoveAddChar(prompt_y_coord, 0, '|')
		window.MoveAddChar(prompt_y_coord, 0, '>')

		for i := 1; i < height-1; i++ { //clear window
			window.MovePrint(i, prompt_x_coord, padding)
		}
		return
	}
	//if the key entered is an input char:
	if !unicode.IsLetter(rune(key)) && !unicode.IsDigit(rune(key)) && key != octothorpe && key != plus_sign {
		return
	} else {
		window.MoveAddChar(prompt_y_coord, prompt_x_coord, ncurses.Char(key))

		if prompt_x_coord+3 <= width {
			window.AttrOn(ncurses.A_BLINK)
			window.MovePrint(prompt_y_coord, prompt_x_coord+1, currentPoint) //indicate char youre on
			window.AttrOff(ncurses.A_BLINK)
			window.MoveAddChar(prompt_y_coord, 0, '>') //indicate the line that you're on.
		} else {
			window.MoveAddChar(prompt_y_coord, 0, '|')
		}
		prompt_x_coord++ //increment char position
	}

	//add to the current input buffer
	if key != ncurses.KEY_ENTER && key != ncurses.KEY_RETURN && key != delete_key && (unicode.IsLetter(rune(key)) || unicode.IsDigit(rune(key)) || key == octothorpe || key == plus_sign) { //not enter and not delete
		user_input_string += string(rune(key))
	}
	//redraw border in case it was painted over
	//window.Box('|', '-')

	window.NoutRefresh()
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

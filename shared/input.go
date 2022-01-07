package shared

import (
	"fmt"
	"strings"
	"unicode"

	ncurses "github.com/nate-xyz/goncurses_"
)

func OptionsInput(window *ncurses.Window, key ncurses.Key, options []string, selected_index int) (int, bool) {

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
		back_i := get_index(options, "back")
		if back_i != -1 {
			return back_i, true
		} else {
			quit_i := get_index(options, "quit")
			if quit_i != -1 {
				return quit_i, true
			}
		}
		return -1, true
	}

	draw_options_input(window, options, selected_index)
	//draw standout for currently selected option
	// for i, str := range options {
	// 	if i == selected_index {
	// 		window.AttrOn(ncurses.ColorPair(3))
	// 		window.MovePrint(i+1, (width/2)-(len(str)/2), str)
	// 		window.AttrOff(ncurses.ColorPair(3))
	// 	} else {
	// 		window.MovePrint(i+1, (width/2)-(len(str)/2), str)
	// 	}
	// }

	//redraw border in case it was painted over
	//window.Box('|', '-')
	//window.Refresh()
	return selected_index, false
}

func SliderInput(win *ncurses.Window, key ncurses.Key, t int, tic_index []int, slider_index int) ([]int, int, float64, bool) {
	height, width := win.MaxYX()
	_ = height
	//slider arrays
	//second array
	lo, hi := 0, 180
	s := make([]int, hi-lo+1)
	for i := range s {
		s[i] = i + lo
	}
	//minute array
	m := []float64{0.25, 0.5, 0.75, 1.0, 1.5}
	lof, hif := 2, 180
	m_ := make([]float64, hif-lof+1)
	for i := range m_ {
		m_[i] = float64(i + lo)
	}
	m = append(m, m_...)
	//day array
	lod, hid := 1, 14
	d := make([]int, hid-lod+1)
	for i := range d {
		d[i] = i + lod
	}

	//determine what option to choose based on input
	switch t { // real time slider: one for minute, one for seconds
	case 0:
		switch key {
		case ncurses.KEY_MOUSE, ncurses.KEY_RESIZE: //no input
			return tic_index, slider_index, -1.0, false
		case ncurses.KEY_ENTER, ncurses.KEY_RETURN: //selected input
			return tic_index, slider_index, m[tic_index[0]], true
		case ncurses.KEY_LEFT:

			switch slider_index {
			case 0: //minute slider
				tic_index[0]--
				if tic_index[0] < 0 {
					tic_index[0] = len(m) - 1
				}
			case 1: //second slider
				tic_index[1]--
				if tic_index[1] < 0 {
					tic_index[1] = len(s) - 1
				}
			}

		case ncurses.KEY_RIGHT:
			switch slider_index {
			case 0: //minute slider
				tic_index[0] += 1
				if tic_index[0] >= len(m) {
					tic_index[0] = 0
				}
			case 1: //second slider
				tic_index[1] += 1
				if tic_index[1] >= len(s) {
					tic_index[1] = 0
				}
			}
		case ncurses.KEY_UP:
			slider_index--
			if slider_index < 0 {
				slider_index = 3
			}
		case ncurses.KEY_DOWN:
			slider_index++
			if slider_index > 3 {
				slider_index = 0
			}
		}

	case 1: // correspondence slider, days betwenn 1 - 14
		switch key {
		case ncurses.KEY_MOUSE, ncurses.KEY_RESIZE:
			return tic_index, slider_index, -1.0, false
		case ncurses.KEY_ENTER, ncurses.KEY_RETURN:
			return tic_index, slider_index, -1.0, true
		case ncurses.KEY_LEFT:
			if slider_index == 0 {
				tic_index[0]--
				if tic_index[0] < 0 {
					tic_index[0] = len(d) - 1
				}
			}
		case ncurses.KEY_RIGHT:
			if slider_index == 0 {
				tic_index[0]++
				if tic_index[0] >= len(d) {
					tic_index[0] = 0
				}
			}
		case ncurses.KEY_UP:
			slider_index--
			if slider_index < 0 {
				slider_index = 2
			}
		case ncurses.KEY_DOWN:
			slider_index++
			if slider_index > 2 {
				slider_index = 0
			}
		}
	}

	// func replaceAtIndex(in string, r rune, i int) string {
	// 	out := []rune(in)
	// 	out[i] = r
	// 	return string(out)
	// }

	//draw logic
	tic_line := fmt.Sprintf("%s", strings.Repeat("-", (width-4)))
	switch t {
	case 0: // real time slider: one for minute, one for seconds

		mes_one := fmt.Sprintf("Minutes per side: %v", m[tic_index[0]])
		win.MovePrint(1, 1, fmt.Sprintf("%s", strings.Repeat(" ", (width-2))))

		if slider_index == 0 {
			win.AttrOn(ncurses.ColorPair(3))
			win.MovePrint(1, ((width / 2) - (len(mes_one) / 2) - len(mes_one)%2), mes_one) //print the day the tick is at

			win.AttrOff(ncurses.ColorPair(3))
		} else {
			win.MovePrint(1, ((width / 2) - (len(mes_one) / 2) - len(mes_one)%2), mes_one) //print the day the tick is at

		}
		tic_loc := int((float64(len(tic_line))/float64(len(m)))*float64(tic_index[0])) + 2
		win.MovePrint(height/3, 2, tic_line) //print the tic_line
		win.MovePrint(height/3, tic_loc, "|")

		mes_two := fmt.Sprintf("Increment in seconds: %v", s[tic_index[1]])
		win.MovePrint((height/3)+2, 1, fmt.Sprintf("%s", strings.Repeat(" ", (width-2))))

		if slider_index == 1 {
			win.AttrOn(ncurses.ColorPair(3))
			win.MovePrint((height/3)+2, ((width / 2) - (len(mes_two) / 2) - len(mes_two)%2), mes_two) //print the day the tick is at

			win.AttrOff(ncurses.ColorPair(3))
		} else {
			win.MovePrint((height/3)+2, ((width / 2) - (len(mes_two) / 2) - len(mes_two)%2), mes_two) //print the day the tick is at

		}
		tic_loc = int((float64(len(tic_line))/float64(len(m)))*float64(tic_index[1])) + 2
		win.MovePrint((height/3)+4, 2, tic_line) //print the tic_line
		win.MovePrint((height/3)+4, tic_loc, "|")
		if slider_index == 2 {
			win.AttrOn(ncurses.ColorPair(3))
			win.MovePrint((height/3)+5, (width/2)-(len("submit")/2), "submit")
			win.AttrOff(ncurses.ColorPair(3))
		} else {
			win.MovePrint((height/3)+5, (width/2)-(len("submit")/2), "submit")
		}
		if slider_index == 3 {
			win.AttrOn(ncurses.ColorPair(3))
			win.MovePrint((height/3)+6, (width/2)-(len("back")/2), "back")
			win.AttrOff(ncurses.ColorPair(3))
		} else {
			win.MovePrint((height/3)+6, (width/2)-(len("back")/2), "back")
		}

	case 1: // correspondence slider, days betwenn 1 - 14
		mes := fmt.Sprintf("Days per turn: %v", d[tic_index[0]])
		win.MovePrint(1, 1, fmt.Sprintf("%s", strings.Repeat(" ", (width-2))))

		if slider_index == 0 {
			win.AttrOn(ncurses.ColorPair(3))
			win.MovePrint(1, ((width / 2) - (len(mes) / 2) - len(mes)%2), mes) //print the day the tick is at
			win.AttrOff(ncurses.ColorPair(3))
		} else {
			win.MovePrint(1, ((width / 2) - (len(mes) / 2) - len(mes)%2), mes) //print the day the tick is at
		}
		win.MovePrint(height/2, 1, tic_line) //print the tic_line
		tic_loc := int((float64(len(tic_line))/float64(len(d)))*float64(tic_index[0])) + 2
		win.MovePrint(height/2, tic_loc, "|")
		if slider_index == 1 {
			win.AttrOn(ncurses.ColorPair(3))
			win.MovePrint((height/3)+5, (width/2)-(len("submit")/2), "submit")
			win.AttrOff(ncurses.ColorPair(3))
		} else {
			win.MovePrint((height/3)+5, (width/2)-(len("submit")/2), "submit")
		}
		if slider_index == 2 {
			win.AttrOn(ncurses.ColorPair(3))
			win.MovePrint((height/3)+6, (width/2)-(len("back")/2), "back")
			win.AttrOff(ncurses.ColorPair(3))
		} else {
			win.MovePrint((height/3)+6, (width/2)-(len("back")/2), "back")
		}

	}

	win.Refresh()
	//win.NoutRefresh()
	//ncurses.Update()
	return tic_index, slider_index, -1.0, false
}

func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

func UpdateInput(window *ncurses.Window, key ncurses.Key) {
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
		StatusMessage = "char limit reached"
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
		UserInputString = removeLastRune(UserInputString)

	}
	if key == ncurses.KEY_ENTER || key == ncurses.KEY_RETURN || key == enter_key { //enter key
		HasEnteredMove = true
		EnteredPromptStr = UserInputString //set global string to check if move is legal
		UserInputString = ""               //reset input buffer
		prompt_x_coord = 1                 //reset char coordinates
		prompt_y_coord = 1                 //reset char coordinates
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
		UserInputString += string(rune(key))
	}
	//redraw border in case it was painted over
	//window.Box('|', '-')

	window.NoutRefresh()
}

func removeLastRune(s string) string {
	if len(s) <= 0 {
		return s
	}
	r := []rune(s)
	return string(r[:len(r)-1])
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

package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	ncurses "github.com/nate-xyz/goncurses"
	"github.com/notnil/chess"
)

// //      888 d8b                   888                            d8b           .d888
// //      888 Y8P                   888                            Y8P          d88P"
// //      888                       888                                         888
// //  .d88888 888 .d8888b  88888b.  888  8888b.  888  888          888 88888b.  888888 .d88b.
// // d88" 888 888 88K      888 "88b 888     "88b 888  888          888 888 "88b 888   d88""88b
// // 888  888 888 "Y8888b. 888  888 888 .d888888 888  888          888 888  888 888   888  888
// // Y88b 888 888      X88 888 d88P 888 888  888 Y88b 888          888 888  888 888   Y88..88P
// //  "Y88888 888  88888P' 88888P"  888 "Y888888  "Y88888 88888888 888 888  888 888    "Y88P"
// //                       888                        888
// //                       888                   Y8b d88P
// //                       888                    "Y88P"

func display_info(info_window *ncurses.Window) {
	height, width := info_window.MaxYX()

	info_window.AttrOn(ncurses.ColorPair(3))
	if game.Position().Turn() == chess.White {
		info_window.MovePrint(1, 1, "white to move")
	} else if game.Position().Turn() == chess.Black {
		//info_window.AttrOn(ncurses.A_REVERSE)
		info_window.MovePrint(1, 1, "black to move")
		//info_window.AttrOff(ncurses.A_REVERSE)
	}
	info_window.MovePrint(2, 1, fmt.Sprintf("last move: %s", last_move_str))
	info_window.AttrOff(ncurses.ColorPair(3))

	var text_colour int16
	if status_str == "move is legal!" {
		text_colour = 8
	} else {
		text_colour = 9
	}
	info_window.AttrOn(ncurses.ColorPair(text_colour))
	info_window.MovePrint(3, 1, status_str)
	info_window.AttrOff(ncurses.ColorPair(text_colour))

	info_window.MovePrint(4, 1, fmt.Sprintf("input: %s", inputted_str))

	info_window.AttrOn(ncurses.ColorPair(8))

	//wrap_y := 0
	san_move_str := fmt.Sprintf("legal moves: %s", strings.Join(legal_move_str_array[:], ", "))

	for y := 5; y < height-1; y++ {
		//wrap_y = y
		if len(san_move_str) > width-2 {
			info_window.MovePrint(y, 1, san_move_str[:width-2])
			san_move_str = san_move_str[width-2:]
		} else {
			info_window.MovePrint(y, 1, san_move_str)
			break
		}

	}
	// legal_move_str := fmt.Sprintf("legal moves (uci): %s", legal_move_str)

	// for y := wrap_y + 2; y < height-1; y++ {
	// 	if len(legal_move_str) > width-2 {
	// 		info_window.MovePrint(y, 1, legal_move_str[:width-2])
	// 		legal_move_str = legal_move_str[width-2:]
	// 	} else {
	// 		info_window.MovePrint(y, 1, legal_move_str)
	// 		break
	// 	}
	// }
	//info_window.MovePrint(7, 1, "{}: {}".format("legal moves (uci)", legal_move_str))
	info_window.AttrOff(ncurses.ColorPair(8))

	status_str = ""
}

// //          88  88                          88                                     88           88
// //          88  ""                          88                                     88           ""               ,d
// //          88                              88                                     88                            88
// //  ,adPPYb,88  88  ,adPPYba,  8b,dPPYba,   88  ,adPPYYba,  8b       d8            88,dPPYba,   88  ,adPPYba,  MM88MMM  ,adPPYba,   8b,dPPYba,  8b       d8
// // a8"    `Y88  88  I8[    ""  88P'    "8a  88  ""     `Y8  `8b     d8'            88P'    "8a  88  I8[    ""    88    a8"     "8a  88P'   "Y8  `8b     d8'
// // 8b       88  88   `"Y8ba,   88       d8  88  ,adPPPPP88   `8b   d8'             88       88  88   `"Y8ba,     88    8b       d8  88           `8b   d8'
// // "8a,   ,d88  88  aa    ]8I  88b,   ,a8"  88  88,    ,88    `8b,d8'              88       88  88  aa    ]8I    88,   "8a,   ,a8"  88            `8b,d8'
// //  `"8bbdP"Y8  88  `"YbbdP"'  88`YbbdP"'   88  `"8bbdP"Y8      Y88'               88       88  88  `"YbbdP"'    "Y888  `"YbbdP"'   88              Y88'
// //                             88                               d8'                                                                                 d8'
// //                             88                              d8'     888888888888                                                                d8'
func display_history(history_window *ncurses.Window) {
	height, width := history_window.MaxYX()

	history_str_i := 0
	if len(history_arr) <= 0 {
		history_window.MovePrint(1, 1, "no moves yet")
	}
	for y := 1; y < height-1; y++ {
		if y >= len(history_arr) {
			break
		}
		hist_str := history_arr[history_str_i]
		piece_str := pieces['p']
		if unicode.IsUpper(rune(hist_str[0])) {
			piece_str = pieces[rune(hist_str[0])]
		}
		hist_str = "move " + strconv.Itoa(move_amount-history_str_i) + ": " + hist_str + " " + string(piece_str)
		if len(hist_str) > width-2 {
			history_window.MovePrint(y, 1, hist_str[:width-2])
			//hist_str = hist_str[width-2:]
		} else {
			history_window.MovePrint(y, 1, hist_str)
		}
		history_str_i += 1
	}
}

// //      888                                       888                                    888
// //      888                                       888                                    888
// //      888                                       888                                    888
// //  .d88888 888d888 8888b.  888  888  888         88888b.   .d88b.   8888b.  888d888 .d88888
// // d88" 888 888P"      "88b 888  888  888         888 "88b d88""88b     "88b 888P"  d88" 888
// // 888  888 888    .d888888 888  888  888         888  888 888  888 .d888888 888    888  888
// // Y88b 888 888    888  888 Y88b 888 d88P         888 d88P Y88..88P 888  888 888    Y88b 888
// //  "Y88888 888    "Y888888  "Y8888888P" 88888888 88888P"   "Y88P"  "Y888888 888     "Y88888
func draw_board(board_window *ncurses.Window) {
	height, width := board_window.MaxYX()
	board_FEN := game.Position().String()
	// board_square_coord = {}
	x_notation_string := "abcdefgh"
	y_notation_string := "87654321"
	_ = y_notation_string
	// 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR'
	// rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
	x_inc := 2
	y_inc := 1

	x_coord := width/2 - 4*x_inc  //increment by 2
	y_coord := height/2 - 4*y_inc //increment by 2

	og_xcoord := x_coord
	og_ycoord := y_coord

	square_count := 0
	var color_pair int16

	for _, current_piece := range board_FEN { //loop to parse the FEN string
		var key_tuple coord_pair
		key_tuple.x_coord_ = x_coord
		key_tuple.y_coord_ = y_coord

		if current_piece == ' ' {
			board_window.MovePrint(y_coord, x_coord, "\t")
			break
		} else if current_piece == '/' {
			board_window.MovePrint(y_coord, x_coord, "\t")
			x_coord = og_xcoord //set x_coord to first in the row
			y_coord += y_inc    //incremen
			square_count++
			continue
		} else if unicode.IsDigit(current_piece) {
			int_, _ := strconv.Atoi(string(current_piece))
			for j := 0; j < int_; j++ {
				if square_count%2 == 0 {
					color_pair = 4
				} else {
					color_pair = 5
				}
				board_window.AttrOn(ncurses.ColorPair(color_pair))
				board_window.MovePrint(y_coord, x_coord, " \t") //add a space+tab character for an empty square
				var pair_temp piece_color
				pair_temp.color = color_pair
				pair_temp.piece = -1
				board_square_coord[key_tuple] = pair_temp
				board_window.AttrOff(ncurses.ColorPair(color_pair))
				square_count += 1
				x_coord += x_inc
			}
			continue
		} else if !unicode.IsDigit(current_piece) {
			//determine proper color pair
			var floating_color int16 = 11
			if unicode.IsUpper(current_piece) {
				floating_color = 10
				if square_count%2 == 0 {
					color_pair = 4
				} else {
					color_pair = 5
				}
			} else {
				floating_color = 11
				if square_count%2 == 0 {
					color_pair = 6
				} else {
					color_pair = 7
				}
			}

			board_window.AttrOn(ncurses.ColorPair(color_pair))
			board_window.AttrOn(ncurses.A_BOLD)

			board_window.MovePrint(y_coord, x_coord, string(pieces[unicode.ToLower(current_piece)])+"\t")
			//board_window.MovePrint(y_coord, x_coord, string(pieces[unicode.ToUpper(current_piece)])+" ")

			//board_window.MoveAddChar(y_coord, x_coord, ncurses.Char(pieces[unicode.ToUpper(current_piece)]))

			//color, piece
			var pair_temp piece_color
			pair_temp.color = floating_color
			pair_temp.piece = pieces[unicode.ToUpper(current_piece)]
			board_square_coord[key_tuple] = pair_temp

			board_window.AttrOff(ncurses.ColorPair(color_pair))
			board_window.AttrOff(ncurses.A_BOLD)

			square_count += 1
			x_coord += x_inc
			continue
		} else {
			print("error parsing starting FEN")
			break
		}
	}

	for i := 0; i < 8; i++ {
		board_window.MovePrint(og_ycoord-y_inc-1, og_xcoord+x_inc*i, string(x_notation_string[i]))
		board_window.MovePrint(og_ycoord+8*y_inc+1, og_xcoord+x_inc*i, string(x_notation_string[i]))
		board_window.MovePrint(og_ycoord+y_inc*i, og_xcoord-x_inc-1, string(y_notation_string[i]))
		board_window.MovePrint(og_ycoord+y_inc*i, og_xcoord+8*x_inc+1, string(y_notation_string[i]))
	}
}

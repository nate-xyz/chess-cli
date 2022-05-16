package shared

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	ncurses "github.com/nate-xyz/goncurses_"
	"github.com/notnil/chess"
)

func DisplayInfoWindow(info_window *ncurses.Window) {
	height, width := info_window.MaxYX()

	info_window.AttrOn(ncurses.ColorPair(3))
	if CurrentGame.Position().Turn() == chess.White {
		info_window.MovePrint(1, 1, "white to move")
	} else {
		//info_window.AttrOn(ncurses.A_REVERSE)
		info_window.MovePrint(1, 1, "black to move")
		//info_window.AttrOff(ncurses.A_REVERSE)
	}
	info_window.MovePrint(2, 1, fmt.Sprintf("last move: %s", LastMoveString))
	info_window.AttrOff(ncurses.ColorPair(3))

	var text_colour int16
	if StatusMessage == "move is legal!" {
		text_colour = 8
	} else {
		text_colour = 9
	}
	info_window.AttrOn(ncurses.ColorPair(text_colour))
	info_window.MovePrint(3, 1, StatusMessage)
	info_window.AttrOff(ncurses.ColorPair(text_colour))

	info_window.MovePrint(4, 1, fmt.Sprintf("input: %s", EnteredPromptStr))

	info_window.AttrOn(ncurses.ColorPair(8))

	//wrap_y := 0
	san_move_str := fmt.Sprintf("legal moves: %s", strings.Join(LegalMoveStrArray[:], ", "))

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

	StatusMessage = ""
}

func DisplayHistoryWindow(history_window *ncurses.Window) {
	height, width := history_window.MaxYX()

	history_str_i := 0
	if len(MoveHistoryArray) <= 0 {
		history_window.MovePrint(1, 1, "no moves yet")
	}
	for y := 1; y < height-1; y++ {
		if y >= len(MoveHistoryArray) {
			break
		}
		hist_str := MoveHistoryArray[history_str_i]
		piece_str := pieces['p']
		if unicode.IsUpper(rune(hist_str[0])) {
			piece_str = pieces[rune(hist_str[0])]
		}
		hist_str = "move " + strconv.Itoa(MoveAmount-history_str_i) + ": " + hist_str + " " + string(piece_str)
		if len(hist_str) > width-2 {
			history_window.MovePrint(y, 1, hist_str[:width-2])
			//hist_str = hist_str[width-2:]
		} else {
			history_window.MovePrint(y, 1, hist_str)
		}
		history_str_i += 1
	}
}

func DrawBoardWindow(board_window *ncurses.Window, FEN string) {
	height, width := board_window.MaxYX()

	board_window.MovePrint(1, 1, FEN)
	//board_FEN := CurrentGame.Position().String()
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

	for _, current_piece := range FEN { //loop to parse the FEN string
		var key_tuple CoordPair
		key_tuple.X = x_coord
		key_tuple.Y = y_coord

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
				var pair_temp PieceColor
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
			var pair_temp PieceColor
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

func LoadingScreen(screen *ncurses.Window, message string) {
	height, width := screen.MaxYX()
	screen.MovePrint(height/2, width/2-len(message)/2, message)
	if UnicodeSupport {
		dt := time.Now().Unix() % 8
		screen.MovePrint((height/2)+1, width/2, fmt.Sprintf("%v", knight_loader[dt]))
	} else {
		dt := time.Now().Unix() % 10
		screen.MovePrint((height/2)+1, width/2, fmt.Sprintf("%v", loader[dt]))
	}
	screen.Refresh()
}

func draw_options_input(window *ncurses.Window, options []string, selected_index int) {
	_, width := window.MaxYX()

	piece := "♟︎ "

	//draw standout for currently selected option
	for i, str := range options {
		if i == selected_index {
			window.AttrOn(ncurses.ColorPair(3))
			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
			window.AttrOff(ncurses.ColorPair(3))
			if UnicodeSupport {
				window.AttrOn(ncurses.A_DIM)
				window.AttrOn(ncurses.A_BLINK)
				window.MovePrint(i+1, 1, piece)
				window.MovePrint(i+1, width-3, piece)
				window.AttrOff(ncurses.A_BLINK)
				window.AttrOff(ncurses.A_DIM)
			}
		} else {
			window.MovePrint(i+1, (width/2)-(len(str)/2), str)
			if UnicodeSupport {
				window.MovePrint(i+1, 1, " ")
				window.MovePrint(i+1, width-3, " ")
			}
		}
	}
	window.Refresh()
}

package main

import (
	"regexp"

	ncurses "github.com/nate-xyz/goncurses_"
	"github.com/notnil/chess"
)

// GLOBAL VAR DECLARATIONS

var currentGameID string
var control_l_key ncurses.Key = 12
var control_o_key ncurses.Key = 15
var q_key ncurses.Key = 113
var zero_key ncurses.Key = 48
var one_key ncurses.Key = 49
var two_key ncurses.Key = 50
var three_key ncurses.Key = 51
var four_key ncurses.Key = 52

// set to true to skip welcome screen
var dev_mode bool = false

// Set true to disable post screen
var post_screen_toggle bool = false

// prompt vars
var prompt_x_coord int = 1
var prompt_y_coord int = 1

// global strings
var last_move_str string = "no move yet"
var user_input_string string = ""
var inputted_str string = ""
var status_str string = ""
var legal_move_str string = ""
var san_move_str string = ""

var final_position string = ""
var legal_move_str_array []string

var move_amount int = 0
var game_outcome_enum int = 0

// true if user hits enter key
var entered_move bool = false
var quit_game bool = false
var mouse_pressed bool = false
var floating_piece string = ""
var floating bool = false
var mouse_event_bool bool = false

var board_square_coord = make(map[coord_pair]piece_color)

var history_arr = []string{"init"}

// var outcome_tuple = []string{
// 	"Good luck.",
// 	"Checkmate!",
// 	"Stalemate.",
// 	"Draw - insufficient material.",
// 	"Draw - 75 move rule.",
// 	"Draw - fivefold repetition.",
// 	"Draw - 50 move rule.",
// 	"Draw by threefold repetition has been claimed",
// }

//convert to map
var file = map[rune]int{
	'a': 0,
	'b': 1,
	'c': 2,
	'd': 3,
	'e': 4,
	'f': 5,
	'g': 6,
	'h': 7,
}

//convert to map
var pieces = map[rune]rune{
	'K': '♔',
	'Q': '♕',
	'R': '♖',
	'B': '♗',
	'N': '♘',
	'P': '♙',
	'k': '♚',
	'q': '♛',
	'r': '♜',
	'b': '♝',
	'n': '♞',
	//'p': '♟︎',
	'p': '♙',
}

var knight_loader = map[int64]string{
	0: "♞ ",
	1: "🨇 ",
	2: "🨓 ",
	3: "🨜 ",
	4: "🨨 ",
	5: "🨱 ",
	6: "🨽 ",
	7: "🩆 ",
}

var loader = map[int64]string{
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

var game *chess.Game = chess.NewGame()

var alphanumeric *regexp.Regexp = regexp.MustCompile("^[a-zA-Z0-9]*$")

// var isletter *regexp.Regexp = regexp.MustCompile("^[a-zA-Z]*$")
var isupper *regexp.Regexp = regexp.MustCompile("^[A-Z]*$")
var isdigit *regexp.Regexp = regexp.MustCompile("^[0-9]*$")

//ascii key codes
var enter_key ncurses.Key = 10
var space ncurses.Key = 32
var octothorpe ncurses.Key = 35 // # key
var plus_sign ncurses.Key = 43  // + key
var delete_key ncurses.Key = 127
var up_arrow ncurses.Key = 259
var down_arrow ncurses.Key = 258
var left_arrow ncurses.Key = 260
var right_arrow ncurses.Key = 261

var lichess_bg string = `::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
:         :::::::::         :::www:::   _+_   :::::::::         ::::::::::
:  |_|_|  :: _,,:::   (/)   :::)@(:::   )@(   :::(/):::   _,,   ::|_|_|:::
:   |@|   ::"- \~::   |@|   :::|@|:::   |@|   :::|@|:::  "- \~  :::|@|::::
:   |@|   :::|@|:::   |@|   :::|@|:::   |@|   :::|@|:::   |@|   :::|@|::::
:  /@@@\  ::/@@@\::  /@@@\  ::/@@@\::  /@@@\  ::/@@@\::  /@@@\  ::/@@@\:::
::::::::::         :::::::::         :::::::::         :::::::::         :
:::::():::    ()   ::::():::    ()   ::::():::    ()   ::::():::    ()   :
:::::)(:::    )(   ::::)(:::    )(   ::::)(:::    )(   ::::)(:::    )(   :
::::/@@\::   /@@\  :::/@@\::   /@@\  :::/@@\::   /@@\  :::/@@\::   /@@\  :
::::::::::         :::::::::         :::::::::         :::::::::         :
:         :::::::::         :::::::::         :::::::::         ::::::::::
:         :::::::::         :::::::::         :::::::::         ::::::::::
:         :::::::::         :::::::::         :::::::::         ::::::::::
:         :::::::::         :::::::::         :::::::::         ::::::::::
:         :::::::::         :::::::::         :::::::::         ::::::::::
::::::::::         :::::::::         :::::::::         :::::::::         :
::::::::::         :::::::::         :::::::::         :::::::::         :
::::::::::         :::::::::         :::::::::         :::::::::         :
::::::::::         :::::::::         :::::::::         :::::::::         :
::::::::::         :::::::::         :::::::::         :::::::::         :
:         :::::::::         :::::::::         :::::::::         ::::::::::
:         :::::::::         :::::::::         :::::::::         ::::::::::
:         :::::::::         :::::::::         :::::::::         ::::::::::
:         :::::::::         :::::::::         :::::::::         ::::::::::
:         :::::::::         :::::::::         :::::::::         ::::::::::
::::::::::         :::::::::         :::::::::         :::::::::         :
::::::::::         :::::::::         :::::::::         :::::::::         :
::::::::::         :::::::::         :::::::::         :::::::::         :
::::::::::         :::::::::         :::::::::         :::::::::         :
::::::::::         :::::::::         :::::::::         :::::::::         :
:         :::::::::         :::::::::         :::::::::         ::::::::::
:    ()   ::::():::    ()   ::::():::    ()   ::::():::    ()   ::::()::::
:    )(   ::::)(:::    )(   ::::)(:::    )(   ::::)(:::    )(   ::::)(::::
:   /__\  :::/__\::   /__\  :::/__\::   /__\  :::/__\::   /__\  :::/__\:::
:         :::::::::         :::::::::         :::::::::         ::::::::::
::::::::::         :::::::::   www   :::_+_:::         :::::::::         :
:::|_|_|::   _,,   :::(/):::   ) (   :::) (:::   (/)   :::_,,:::  |_|_|  :
::::| |:::  "- \~  :::| |:::   | |   :::| |:::   | |   ::"- \~::   | |   :
::::| |:::   | |   :::| |:::   | |   :::| |:::   | |   :::| |:::   | |   :
:::/___\::  /___\  ::/___\::  /___\  ::/___\::  /___\  ::/___\::  /___\  :
::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::`

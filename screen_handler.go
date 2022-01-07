package main

import ncurses "github.com/nate-xyz/goncurses_"

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

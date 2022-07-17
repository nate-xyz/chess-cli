package main

import (
	. "github.com/nate-xyz/chess-cli/local"
	. "github.com/nate-xyz/chess-cli/online"
	. "github.com/nate-xyz/chess-cli/shared"
	ncurses "github.com/nate-xyz/goncurses_"
	//	ai "github.com/nate-xyz/chess-cli/stockfish"
)

//screen handlers
func mainScreenHandler(stdscr *ncurses.Window, key ncurses.Key) {
	switch key {
	case ZeroKey:
		key = WelcomeScreen(stdscr) //go back to welcome screen
	case OneKey:
		_, key = localGameHandler(stdscr, LocalGameScreen(stdscr)) //go to local game screen, two player with chess lib
	case TwoKey:
		_, key = LichessScreenHandler(stdscr, LichessWelcome(stdscr)) //go to lichess welcome screen, login w oauth
	case ThreeKey:
		return //go to stockfish ai screen, todo
	case CtrlO_Key:
		return //quit game
	}
	mainScreenHandler(stdscr, key)
}

func localGameHandler(stdscr *ncurses.Window, option int) (int, ncurses.Key) {
	switch option {
	case 0:
		return -1, ZeroKey //go back to welcome screen
	case 1:
		option = LocalGameScreen(stdscr) //go to game screen
	case 2:
		option = PostScreen(stdscr)
	case 3:
		return -1, CtrlO_Key //quit game
	}
	return localGameHandler(stdscr, option)
}

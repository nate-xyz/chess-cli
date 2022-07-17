package online

import (
	. "github.com/nate-xyz/chess-cli/shared"
	ncurses "github.com/nate-xyz/goncurses_"
	//	ai "github.com/nate-xyz/chess-cli/stockfish"
)

//screen handlers
func LichessScreenHandler(stdscr *ncurses.Window, option int) (int, ncurses.Key) {
	switch option {
	case 0:
		return -1, ZeroKey //go back to welcome screen
	case 1:
		option = LichessWelcome(stdscr)
	case 2:
		option = LichessChallenges(stdscr) //go to challenge screen
	case 3:
		option = CreateLichessGame(stdscr)
	case 4:
		option = WaitForLichessGameResponse(stdscr)
	case 5:
		return -1, CtrlO_Key //quit game
	case 6:
		option = LichessGameScreen(stdscr, currentGameID)
	}
	return LichessScreenHandler(stdscr, option)
}

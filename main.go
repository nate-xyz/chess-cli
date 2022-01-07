package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	ncurses "github.com/nate-xyz/goncurses_"
)

//#f3 e5 g4 Qh4#

var sigs chan os.Signal
var noti_message chan string
var error_message chan error
var ready chan struct{}
var curChallenge CreateChallengeType
var waiting_alert chan StreamEventType

// var quit_stream chan bool
// var stream_done chan struct{}

func main() {
	//init channels
	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)
	noti_message = make(chan string, 100)
	error_message = make(chan error, 10)
	ready = make(chan struct{})
	stream_channel = make(chan StreamEventType, 1)
	// Initialize ncurses. It's essential End() is called to ensure the
	// terminal isn't altered after the program ends
	stdscr, err := ncurses.Init()
	if err != nil {
		log.Fatal("init", err)
	}
	defer ncurses.End()
	stdscr.Timeout(0)

	go StreamEvent(stream_channel, ready)
	go StreamConsumer(stream_channel, noti_message)
	go notifier(stdscr, noti_message)
	go ncurses_print_error(stdscr, error_message)

	// go func(screen *ncurses.Window) {
	// 	rand.Seed(time.Now().UnixNano())
	// 	for {
	// 		rando := rand.Intn(100)
	// 		go screen.MovePrint(10, 100, fmt.Sprintf(" verification %v ", rando))
	// 		noti_message <- fmt.Sprintf("this is a test %v", rando)
	// 		//error_message <- fmt.Errorf("this is a test %v", rando)
	// 		screen.MovePrint(10, 100, fmt.Sprintf(" verification %v ", rando))
	// 		time.Sleep(time.Millisecond * 500)
	// 	}
	// }(stdscr)

	// ticker := time.NewTicker(time.Second)

	// func(screen *ncurses.Window) {
	// 	for {

	// 		select {

	// 		case <-ticker.C:
	// 			error_message <- fmt.Errorf("error test")
	// 			screen.Clear()
	// 		default:
	// 			screen.MovePrint(1, 1, "not looooooading")
	// 			screen.Refresh()
	// 		}

	// 	}
	// }(stdscr)
	//<-ready

	//necessary for mouse input, start keypad, read all mouse events
	stdscr.Keypad(true)
	mouse_int := ncurses.M_ALL | ncurses.M_POSITION
	_ = mouse_int
	//ncurses.MouseMask(mouse_int, nil)
	//fmt.Printf("\033[?1003h")

	// allow input, Start colors in goncurses
	ncurses.Echo(true)   //allow input
	ncurses.Cursor(0)    //set cursor visibility to hidden
	ncurses.StartColor() //allow color to be displayed

	//ncurses.use_default_colors()
	ncurses.InitPair(1, ncurses.C_CYAN, ncurses.C_BLACK)
	ncurses.InitPair(2, ncurses.C_RED, ncurses.C_BLACK)
	ncurses.InitPair(3, ncurses.C_BLACK, ncurses.C_WHITE)

	//piece and square colors
	if ncurses.CanChangeColor() {
		var light_square int16 = 215 //SandyBrown
		var dark_square int16 = 94   //Orange4
		var light_piece int16 = 230  //Cornsilk1
		var dark_piece int16 = 233   //Grey7
		ncurses.InitPair(4, light_piece, light_square)
		ncurses.InitPair(5, light_piece, dark_square)
		ncurses.InitPair(6, dark_piece, light_square)
		ncurses.InitPair(7, dark_piece, dark_square)

		//floating piece colors
		ncurses.InitPair(10, light_piece, dark_piece)
		ncurses.InitPair(11, dark_piece, light_piece)
		ncurses.InitPair(12, ncurses.C_RED, ncurses.C_WHITE)
		ncurses.InitPair(13, ncurses.C_RED, ncurses.C_BLACK)
		ncurses.InitPair(14, ncurses.C_BLUE, ncurses.C_WHITE)
		ncurses.InitPair(15, ncurses.C_BLUE, ncurses.C_BLACK)
		ncurses.InitPair(16, dark_piece, ncurses.C_BLACK)
		ncurses.InitPair(17, light_piece, ncurses.C_BLACK)
	} else {
		ncurses.InitPair(4, ncurses.C_RED, ncurses.C_WHITE)
		ncurses.InitPair(5, ncurses.C_RED, ncurses.C_BLACK)
		ncurses.InitPair(6, ncurses.C_BLUE, ncurses.C_WHITE)
		ncurses.InitPair(7, ncurses.C_BLUE, ncurses.C_BLACK)
	}
	//move legality colors
	ncurses.InitPair(8, ncurses.C_BLACK, ncurses.C_GREEN)
	ncurses.InitPair(9, ncurses.C_WHITE, ncurses.C_RED)

	var key ncurses.Key = one_key
	if !dev_mode {
		key = zero_key
	}
	mainScreenHandler(stdscr, key)
	ncurses.FlushInput()
	ncurses.Echo(false) //turn off input
	ncurses.End()
}

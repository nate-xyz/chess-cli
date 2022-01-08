package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	. "github.com/nate-xyz/chess-cli/lichess"
	//	. "github.com/nate-xyz/chess-cli/local"
	. "github.com/nate-xyz/chess-cli/shared"
	ncurses "github.com/nate-xyz/goncurses_"
)

//#f3 e5 g4 Qh4#

// var quit_stream chan bool
// var stream_done chan struct{}

func main() {
	//init channels
	Sigs = make(chan os.Signal, 1)
	signal.Notify(Sigs, syscall.SIGWINCH)
	NotiMessage = make(chan string, 100)
	ErrorMessage = make(chan error, 10)
	Ready = make(chan struct{})
	StreamChannel = make(chan StreamEventType, 1)
	StreamChannelForWaiter = make(chan StreamEventType, 1000)

	// Initialize ncurses. It's essential End() is called to ensure the
	// terminal isn't altered after the program ends
	stdscr, err := ncurses.Init()
	if err != nil {
		log.Fatal("init", err)
	}
	defer ncurses.End()
	stdscr.Timeout(0)

	go StreamEvent(StreamChannel, Ready)
	go StreamConsumer(StreamChannel, NotiMessage)
	go notifier(stdscr, NotiMessage)
	go ncurses_print_error(stdscr, ErrorMessage)

	// go func(screen *ncurses.Window) {
	// 	rand.Seed(time.Now().UnixNano())
	// 	for {
	// 		rando := rand.Intn(100)
	// 		go screen.MovePrint(10, 100, fmt.Sprintf(" verification %v ", rando))
	// 		NotiMessage <- fmt.Sprintf("this is a test %v", rando)
	// 		//ErrorMessage <- fmt.Errorf("this is a test %v", rando)
	// 		screen.MovePrint(10, 100, fmt.Sprintf(" verification %v ", rando))
	// 		time.Sleep(time.Millisecond * 500)
	// 	}
	// }(stdscr)

	// ticker := time.NewTicker(time.Second)

	// func(screen *ncurses.Window) {
	// 	for {

	// 		select {

	// 		case <-ticker.C:
	// 			ErrorMessage <- fmt.Errorf("error test")
	// 			screen.Clear()
	// 		default:
	// 			screen.MovePrint(1, 1, "not looooooading")
	// 			screen.Refresh()
	// 		}

	// 	}
	// }(stdscr)
	//<-Ready

	//determing unicode support
	// _, x := stdscr.CursorYX()
	// test_char := "♟︎"
	// stdscr.MovePrint(1, 1, test_char)
	// _, nx := stdscr.CursorYX()
	// diff := (nx - x)
	// NotiMessage <- fmt.Sprintf("%v", stdscr.MoveInChar(1, 1))
	// if diff != 1 {
	// 	UnicodeSupport = false
	// }

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

	var key ncurses.Key = OneKey
	if !DevMode {
		key = ZeroKey
	}
	mainScreenHandler(stdscr, key)
	ncurses.FlushInput()
	ncurses.Echo(false) //turn off input
	ncurses.End()
}

func notifier(screen *ncurses.Window, message <-chan string) {
	for {
		select {
		case m := <-message:
			title := "notification"
			_, s_width := screen.MaxYX()
			//x := rand.Intn(width) + 1
			//y := rand.Intn(height) + 1

			w := GetMaxLenStr([]string{m, title}) + 2
			x := s_width - w - 1
			//y := rand.Intn(height) + 1

			//win, _ := ncurses.NewWindow(3, 20, 1, width-20)

			timeout := time.After(time.Second * 1)
		loop:
			for tick := range time.Tick(time.Millisecond * 10) {
				_ = tick
				select {
				// case <-Sigs:
				// 	_, s_width := screen.MaxYX()
				// 	x := s_width - w - 1
				// 	win.Clear()
				// 	win.MoveWindow(1, x) //move windows to appropriate locations
				// 	win.Box('|', '-')
				// 	win.AttrOn(ncurses.ColorPair(2))
				// 	win.AttrOn(ncurses.A_BOLD)
				// 	win.MovePrint(0, 1, title)
				// 	win.AttrOff(ncurses.ColorPair(2))
				// 	win.AttrOff(ncurses.A_BOLD)
				// 	win.MovePrint(1, 1, m)
				// 	win.Refresh()
				case <-timeout:
					break loop
				default:
					win, _ := ncurses.NewWindow(3, w, 1, x)
					win.Box('|', '-')
					win.AttrOn(ncurses.ColorPair(2))
					win.AttrOn(ncurses.A_BOLD)
					win.MovePrint(0, 1, title)
					win.AttrOff(ncurses.ColorPair(2))
					win.AttrOff(ncurses.A_BOLD)
					win.MovePrint(1, 1, m)
					win.NoutRefresh()
				}
			}
			//time.Sleep(time.Second * 5)
			screen.Clear()
			Sigs <- syscall.SIGWINCH
			//time.Sleep(time.Second * 1)
			// default:
			// 	height, width := screen.MaxYX()
			// 	rand.Seed(time.Now().UnixNano())
			// 	x := rand.Intn(width) + 1
			// 	y := rand.Intn(height) + 1
			// 	screen.MovePrint(y, x, "no message!")
			// 	time.Sleep(time.Millisecond * 10)

		}
	}
}

func ncurses_print_error(screen *ncurses.Window, message <-chan error) {
	for {
		select {
		case m := <-message:
			title := "error"
			h, s_width := screen.MaxYX()

			w := GetMaxLenStr([]string{fmt.Sprintf("%v", m), title}) + 2
			x := (s_width/2 - w/2 - w%2)
			y := (h / 2)

			timeout := time.After(time.Second * 5)
		loop:
			for tick := range time.Tick(time.Millisecond * 10) {
				_ = tick
				select {
				case <-timeout:
					break loop
				default:
					win, _ := ncurses.NewWindow(3, w, y, x)
					win.Box('|', '-')
					win.AttrOn(ncurses.ColorPair(2))
					win.AttrOn(ncurses.A_BOLD)
					win.MovePrint(0, 1, title)
					win.AttrOff(ncurses.ColorPair(2))
					win.AttrOff(ncurses.A_BOLD)
					win.MovePrint(1, 1, fmt.Sprintf("%v", m))
					win.NoutRefresh()
				}
			}
			screen.Clear()
			Sigs <- syscall.SIGWINCH
			time.Sleep(time.Second * 1)
			//os.Exit(1)
		}
	}
}

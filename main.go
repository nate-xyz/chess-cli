package main

import (
	"fmt"
	"os"
	"runtime"

	cv "codeberg.org/tslocum/cview"
	_ "github.com/gdamore/tcell/v3/encoding" // support non-UTF-8 terminals
	api "github.com/nate-xyz/chess-cli/api"
	pkg "github.com/nate-xyz/chess-cli/pkg"
)

// version is set at release build time (see .goreleaser.yaml).
var version = "dev"

const usage = `chess-cli: play chess in the terminal, locally or on Lichess.

Usage: chess-cli [--version | --help]

Files are kept in %s
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println("chess-cli", version)
			return
		default:
			dir, err := api.ConfigDir()
			if err != nil {
				dir = "your user config directory"
			}
			fmt.Printf(usage, dir)
			if os.Args[1] != "-h" && os.Args[1] != "--help" && os.Args[1] != "help" {
				os.Exit(2)
			}
			return
		}
	}
	pkg.SetupTextSizing() // must run before the UI takes over the terminal

	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "CMD.EXE"
		} else {
			shell = "/bin/sh"
		}
	}
	pkg.Root.Shell = shell

	App := cv.NewApplication()
	defer App.HandlePanic()
	App.EnableMouse(true)

	pkg.Root.App = App

	pkg.StreamChannel = make(chan api.StreamEventType, 1)
	pkg.Ready = make(chan struct{})
	go api.StreamEvent(pkg.StreamChannel, pkg.Ready)
	go pkg.StreamConsumer(pkg.StreamChannel)

	pkg.InitUI()

	if err := pkg.Root.App.Run(); err != nil {
		pkg.Root.App.Stop()
		panic(err)
	}
}

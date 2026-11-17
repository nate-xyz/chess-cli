package main

import (
	"errors"
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

Usage: chess-cli [--version | --help | --reset]

  --reset    put all settings back to their defaults; the first-run setup
             runs next time (your Lichess login and saved games are kept)

Files are kept in %s
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println("chess-cli", version)
			return
		case "--reset":
			if dir, err := api.ConfigDir(); err == nil {
				if err := pkg.LockInstance(dir); errors.Is(err, pkg.ErrAlreadyRunning) {
					fmt.Fprintf(os.Stderr, "%v: quit it first, then reset.\n", err)
					os.Exit(1)
				}
			}
			if err := api.LoadConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "Couldn't read the settings (%v); writing defaults.\n", err)
			}
			if err := api.ResetPreferences(); err != nil {
				fmt.Fprintln(os.Stderr, "Couldn't reset:", err)
				os.Exit(1)
			}
			fmt.Println("Settings reset to defaults. Your Lichess login and saved games are kept.")
			fmt.Println("Run chess-cli to go through the first-run setup again.")
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
	if dir, err := api.ConfigDir(); err == nil {
		if err := pkg.LockInstance(dir); errors.Is(err, pkg.ErrAlreadyRunning) {
			fmt.Fprintf(os.Stderr, "%v.\nSwitch to that window, or quit it first.\n", err)
			os.Exit(1)
		}
	}
	_ = api.LoadConfig() // for saved preferences; a broken file is reported at login
	pkg.SetupRendering() // must run before the UI takes over the terminal

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
	go pkg.StreamConsumer(pkg.StreamChannel)

	pkg.InitUI()

	if err := pkg.Root.App.Run(); err != nil {
		pkg.Root.App.Stop()
		panic(err)
	}
}

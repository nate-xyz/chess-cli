package main

import (
	"os"
	"runtime"

	cv "code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/encoding"
	api "github.com/nate-xyz/chess-cli/api"
	pkg "github.com/nate-xyz/chess-cli/pkg"
)

func main() {
	encoding.Register()
	// Sigs = make(chan os.Signal, 1)
	// signal.Notify(Sigs, syscall.SIGWINCH)
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
	api.Ready = make(chan struct{})
	go api.StreamEvent(pkg.StreamChannel, api.Ready)
	go api.StreamConsumer(pkg.StreamChannel)

	pkg.InitUI()

	if err := pkg.Root.App.Run(); err != nil {
		pkg.Root.App.Stop()
		panic(err)
	}
}

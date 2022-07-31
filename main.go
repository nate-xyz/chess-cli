package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gdamore/tcell/encoding"

	cv "code.rocketnine.space/tslocum/cview"
)

func main() {
	encoding.Register()
	Sigs = make(chan os.Signal, 1)
	signal.Notify(Sigs, syscall.SIGWINCH)
	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "CMD.EXE"
		} else {
			shell = "/bin/sh"
		}
	}
	root.shell = shell

	app := cv.NewApplication()
	defer app.HandlePanic()
	app.EnableMouse(true)

	root.app = app

	StreamChannel = make(chan StreamEventType, 1)
	Ready = make(chan struct{})
	go StreamEvent(StreamChannel, Ready)
	go StreamConsumer(StreamChannel)

	InitUI()

	if err := root.app.Run(); err != nil {
		root.app.Stop()
		panic(err)
	}
}

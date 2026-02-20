module github.com/nate-xyz/chess-cli

go 1.24.0

require (
	code.rocketnine.space/tslocum/cview v1.5.7
	github.com/gdamore/tcell v1.4.0
	github.com/gdamore/tcell/v3 v3.1.2
	github.com/jimlambrt/go-oauth-pkce-code-verifier v0.0.0-20201220003123-6363600dffda
	github.com/notnil/chess v1.7.2
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	golang.org/x/text v0.33.0
)

require (
	codeberg.org/tslocum/cbind v0.1.8 // indirect
	github.com/clipperhouse/stringish v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.3.0 // indirect
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/term v0.39.0 // indirect
)

replace (
	code.rocketnine.space/tslocum/cbind => codeberg.org/tslocum/cbind v0.1.8
	code.rocketnine.space/tslocum/cview => codeberg.org/tslocum/cview v1.6.3
)

package pkg

import (
	"fmt"

	"github.com/nate-xyz/chess-cli/api"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

type optionsState struct {
	list    *cv.List
	preview *BoardView
	note    *cv.TextView
	sub     string // "" = main list, or "pieces" / "boards" submenu
	before  string // setting when the submenu opened, restored by Esc
}

var optionsScreen optionsState

// optionsInit builds the Options screen: piece style, with a live preview.
func optionsInit() cv.Primitive {
	list := cv.NewList()
	list.SetHover(true)
	list.SetWrapAround(true)

	preview := newPreviewBoard()
	forwardKeys(preview, list)

	note := textBox(cv.AlignLeft)
	title := titlePrimitive("[blue]Options[white]", list)

	right := cv.NewFlex()
	right.SetDirection(cv.FlexRow)
	right.AddItem(list, 0, 3, true)
	right.AddItem(note, 0, 1, false)

	right.SetPadding(0, 0, 2, 0)
	body := newPreviewLayout(preview, right)

	root := cv.NewFlex()
	root.SetDirection(cv.FlexRow)
	root.AddItem(title, 3, 0, false)
	root.AddItem(body, 0, 1, true)
	root.AddItem(ribbonPrimitive(" chess-cli · options │ ↑↓ preview │ Enter to choose │ Esc back │ Ctrl-C quit", list), 1, 0, false)

	optionsScreen.list, optionsScreen.preview, optionsScreen.note = list, preview, note
	registerMenus("options", false, list)
	return root
}

// newPreviewBoard shows a sample position with a last move, a selected
// piece and its moves, to preview how boards look.
// previewLayout is a sample board on the left and text on the right (first
// run, Options): the board gets the width it needs and the text the rest,
// so explanations aren't cut off in a big terminal.
type previewLayout struct {
	*cv.Flex
	board *BoardView
	text  cv.Primitive
	last  int
}

// The text column's width: wide enough for every explanation when there's
// room (the board is only a sample here), never less than the minimum.
const (
	textWidth       = 76
	minTextWidth    = 50
	minPreviewWidth = 40 // the sample board keeps at least this
)

func newPreviewLayout(board *BoardView, text cv.Primitive) *previewLayout {
	l := &previewLayout{Flex: cv.NewFlex(), board: board, text: text, last: -1}
	l.AddItem(board, 0, 3, false)
	l.AddItem(text, 0, 2, true)
	return l
}

func (l *previewLayout) Draw(screen tc.Screen) {
	_, _, w, h := l.GetRect()
	text := max(minTextWidth, min(textWidth, w-minPreviewWidth)) // small terminal: share the room
	_, bw, _ := l.board.fit(screen, max(0, w-text), h)
	bw = min(w, bw+4) // a little room around the board
	if bw != l.last {
		l.last = bw
		l.ResizeItem(l.board, bw, 0)
		l.ResizeItem(l.text, 0, 1)
	}
	l.Flex.Draw(screen)
}

func newPreviewBoard() *BoardView {
	preview := NewBoardView(nil)
	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for _, m := range []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6"} {
		_ = game.MoveStr(m)
	}
	preview.SetPosition(game.Position(), true)
	preview.SetLastMove(chess.G8, chess.F6)
	preview.SetSelection(chess.F3, map[chess.Square]bool{chess.G5: true, chess.E5: true, chess.H4: true, chess.D4: true})
	return preview
}

func gotoOptions() {
	optionsScreen.sub = ""
	updateOptions()
	Root.Switch("options")
}

// optionsBack leaves a submenu (undoing its preview) or the screen.
func optionsBack() {
	o := &optionsScreen
	switch o.sub {
	case "pieces":
		api.SetConfig(func(c *api.UserConfig) { c.PieceSet = o.before })
	case "boards":
		api.SetConfig(func(c *api.UserConfig) { c.Board = o.before })
	default:
		gotoWelcome()
		return
	}
	o.sub = ""
	updateOptions()
}

func updateOptions() {
	o := &optionsScreen
	switch o.sub {
	case "pieces":
		o.showChoices(pieceSets, currentPieceSet(), func(v string) { api.SetConfig(func(c *api.UserConfig) { c.PieceSet = v }) })
		return
	case "boards":
		names := make([]string, len(boardThemes))
		for i, t := range boardThemes {
			names[i] = t.name
		}
		o.showChoices(names, currentBoard().name, func(v string) { api.SetConfig(func(c *api.UserConfig) { c.Board = v }) })
		return
	}

	current := PiecesSetting()
	mark := func(style string) string {
		if style == current {
			return "[green]●[-] "
		}
		return "○ "
	}
	sel := o.list.GetCurrentItemIndex()
	o.list.SetChangedFunc(nil)
	o.list.Clear()
	o.list.ShowSecondaryText(true)
	add := func(text, explain string, key rune, f func()) {
		item := cv.NewListItem(text)
		item.SetSecondaryText("  " + explain)
		item.SetShortcut(key)
		item.SetSelectedFunc(f)
		o.list.AddItem(item)
	}
	add(mark(PiecesText)+"Advanced rendering", "Unicode chess symbols; big in kitty", 'c', func() { choosePieces(PiecesText) })
	add(mark(PiecesImage)+"Image rendering", "Lichess boards and piece sets", 'i', func() { choosePieces(PiecesImage) })
	if current == PiecesImage {
		add("    Piece set: [::b]"+currentPieceSet()+"[::-] ›", "  Choose from "+fmt.Sprint(len(pieceSets))+" sets", 'p', func() { o.openSub("pieces", currentPieceSet()) })
		add("    Board: [::b]"+currentBoard().name+"[::-] ›", "  Choose the board colours", 'o', func() { o.openSub("boards", currentBoard().name) })
	}
	add(mark(PiecesBasic)+"Basic rendering", "Plain symbols and colours, for any terminal", 'a', func() { choosePieces(PiecesBasic) })
	moveBox, other := "Ctrl-K only", "always shown"
	if moveBoxAlways() {
		moveBox, other = "always shown", "Ctrl-K only"
	}
	add("Move box: [::b]"+moveBox+"[::-]", "Typing moves: Enter to switch to "+other, 'm', func() {
		setMoveBoxAlways(!moveBoxAlways())
		updateOptions()
	})
	onOff := func(b bool) string {
		if b {
			return "on"
		}
		return "off"
	}
	add("Clock bar: [::b]"+onOff(clockBarsOn())+"[::-]", "Clocks with a bar beside the board, or plain in the sidebar", 'k', func() {
		setClockBars(!clockBarsOn())
		updateOptions()
	})
	add("Zen mode: [::b]"+onOff(zenMode())+"[::-]", "Hide names, ratings, chat and game info during games", 'z', func() {
		setZen(!zenMode())
		updateOptions()
	})
	keys := "arrows and hjkl"
	if !vimKeys() {
		keys = "arrows only"
	}
	add("Keys: [::b]"+keys+"[::-]", "Move with the arrow keys, or with vim's hjkl too", 'y', func() {
		setVimKeys(!vimKeys())
		updateOptions()
	})
	soundExplain := "Retro chiptune, Wood, or off; Enter for the next"
	if !soundAvailable() {
		soundExplain = "Can't play sound here: " + soundHelp()
	}
	add("Sound: [::b]"+soundThemeNames[soundTheme()]+"[::-]", soundExplain, 'u', func() {
		next := soundThemes[0]
		for i, t := range soundThemes {
			if t == soundTheme() {
				next = soundThemes[(i+1)%len(soundThemes)]
			}
		}
		setSoundTheme(next)
		playSound(soundMove) // a sample of the new theme
		updateOptions()
	})
	reset := cv.NewListItem("Reset all settings")
	reset.SetSecondaryText("  Defaults, then first-run setup again (keeps login and games)")
	reset.SetShortcut('r')
	confirmItem(reset, "Reset all settings", "Reset all settings? Press again", resetSettings)
	o.list.AddItem(reset)
	add("Back", "Return to the home screen", 'b', gotoWelcome)
	o.list.SetCurrentItem(sel)

	note := "[::b]Pieces[::-]\n\n"
	if GraphicsSupported() {
		note += "Image rendering draws Lichess's boards and piece sets as pictures, scaled to the board. Only sets with licences that allow them in chess-cli are included.\n"
	} else {
		note += "[yellow]This terminal can't show images[-], so image rendering falls back to advanced rendering. " +
			"Images work in terminals with the kitty graphics protocol, such as kitty, Ghostty and WezTerm (not inside tmux).\n"
	}
	if TextSizingMode() {
		note += "\nWith advanced rendering, pieces are drawn big using kitty's text sizing.\n"
	}
	note += "\nBasic rendering uses only the terminal's 8 standard colours and plain symbols, like the original chess-cli. Use it if the board looks wrong in your terminal.\n"
	o.note.SetText(note)
}

func (o *optionsState) openSub(sub, current string) {
	o.sub, o.before = sub, current
	o.list.SetCurrentItem(0)
	updateOptions()
}

// showChoices lists the values of a submenu. Moving through the list
// previews each one; Enter keeps it, Esc goes back to the previous one.
func (o *optionsState) showChoices(values []string, current string, apply func(string)) {
	o.list.SetChangedFunc(nil)
	o.list.Clear()
	o.list.ShowSecondaryText(false)
	sel := 0
	for i, v := range values {
		v := v
		item := cv.NewListItem("  " + v)
		item.SetShortcut(numberKey(i))
		if v == current {
			item.SetMainText("[green]●[-] " + v)
			sel = i
		}
		item.SetSelectedFunc(func() {
			apply(v)
			_ = api.SaveConfig()
			o.sub = ""
			updateOptions()
		})
		o.list.AddItem(item)
	}
	back := cv.NewListItem("Back")
	back.SetShortcut('b')
	back.SetSelectedFunc(optionsBack)
	o.list.AddItem(back)
	o.list.SetCurrentItem(sel)
	o.list.SetChangedFunc(func(i int, _ *cv.ListItem) {
		if i < len(values) {
			apply(values[i])
		} else {
			apply(o.before)
		}
	})
	what := "piece set"
	if o.sub == "boards" {
		what = "board"
	}
	o.note.SetText(fmt.Sprintf("[::b]Choose a %s[::-]\n\nMove through the list to preview; Enter to keep, Esc to go back.", what))
}

func choosePieces(style string) {
	if err := SetPiecesSetting(style); err != nil {
		optionsScreen.note.SetText(fmt.Sprintf("[red]Couldn't save the setting: %v[-]", err))
		return
	}
	updateOptions()
}

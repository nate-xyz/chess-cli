package pkg

import (
	"fmt"

	cv "codeberg.org/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
)

// The first-run screen: says what the terminal supports, recommends a piece
// style and lets you try each one on a sample board. Shown until a style has
// been saved; it can be changed later in Options.

// needsOnboarding: no rendering style chosen yet, or never asked about
// sound, or about Stockfish when there's none (people upgrading get just the
// new questions).
func needsOnboarding() bool {
	return api.UserInfo.Pieces == "" || !api.UserInfo.SoundAsked || !api.UserInfo.KeysAsked || needsStockfishStep()
}

// keysStep asks about hjkl (if not asked yet), then Stockfish.
func keysStep(list *cv.List, info *cv.TextView) {
	if api.UserInfo.KeysAsked {
		onboardStockfish(list, info)
		return
	}
	onboardKeys(list, info, func() { onboardStockfish(list, info) })
}

func needsStockfishStep() bool { return !api.UserInfo.StockfishAsked && findEngine() == "" }

// onboardStockfish offers to download Stockfish, for games against the
// computer and analysis, when there's none on this computer.
func onboardStockfish(list *cv.List, info *cv.TextView) {
	if !needsStockfishStep() {
		gotoWelcome()
		return
	}
	finish := func() {
		_ = api.UpdateConfig(func(c *api.UserConfig) { c.StockfishAsked = true })
		gotoWelcome()
	}
	list.SetChangedFunc(nil)
	list.Clear()
	add := func(text, explain string, f func()) {
		item := cv.NewListItem(text)
		item.SetShortcut(numberKey(list.GetItemCount()))
		item.SetSecondaryText("  " + explain)
		item.SetSelectedFunc(f)
		list.AddItem(item)
	}
	add("Download Stockfish [green](recommended)[-]", "the official release from GitHub, about 80 MB", func() {
		if !downloading {
			downloadEngine(list, finish)
		}
	})
	add("I'll install it myself", installHint(), finish)
	add("Not now", "later: Home → Local → Play local Stockfish", finish)
	list.SetCurrentItem(0)
	info.SetText("[::b]Stockfish?[::-]\n\nStockfish is the chess engine chess-cli uses to play against you offline " +
		"and to analyse your games when they end: accuracy, mistakes and an evaluation graph.\n\n" +
		"It isn't on this computer yet. chess-cli can download it for you, or you can install it yourself.\n\n[#a0a0a0]Stockfish is free software (GPLv3), from stockfishchess.org.[-]")
	Root.App.SetFocus(list)
}

var onboardList *cv.List

// resetSettings puts all settings back to their defaults, applies them and
// runs the first-run setup again.
func resetSettings() {
	if err := api.ResetPreferences(); err != nil {
		optionsScreen.note.SetText("[red]Couldn't reset: " + cv.Escape(err.Error()) + "[-]")
		return
	}
	setMoveBoxAlways(moveBoxAlways())
	applyLayoutPrefs()
	applyKeyPrefs()
	previewPieces("")
	runOnboarding()
}

// runOnboarding shows the first-run setup (again).
func runOnboarding() {
	if Root.nav.HasPanel("onboard") {
		Root.nav.RemovePanel("onboard")
	}
	Root.nav.AddPanel("onboard", onboardInit(), true, false)
	Root.Switch("onboard")
	Root.App.SetFocus(onboardList)
}

// soundStep moves the first-run screen on to the sound question.
var soundStep = func() {}

// onboardSound asks whether to have sound, playing a sample of each
// choice as it's highlighted.
func onboardSound(list *cv.List, info *cv.TextView) {
	list.SetChangedFunc(nil)
	list.Clear()
	choices := []struct{ theme, text, explain string }{
		{SoundRetro, "Yes, retro sounds [green](recommended)[-]", "  chiptune blips, like an old console"},
		{SoundWood, "Yes, wood sounds", "  a soft click of pieces on a board"},
		{SoundOff, "No sound", "  chess-cli stays quiet"},
	}
	for i, c := range choices {
		c := c
		item := cv.NewListItem(c.text)
		item.SetShortcut(numberKey(i))
		item.SetSecondaryText(c.explain)
		item.SetSelectedFunc(func() {
			_ = api.UpdateConfig(func(cfg *api.UserConfig) { cfg.Sound, cfg.SoundAsked = c.theme, true })
			keysStep(list, info)
		})
		list.AddItem(item)
	}
	list.SetCurrentItem(0)
	list.SetChangedFunc(func(i int, _ *cv.ListItem) { playSoundIn(choices[i].theme, soundMove) })
	note := ""
	if !soundAvailable() {
		note = "\n[yellow]" + cv.Escape(soundHelp()) + "[-]: you can still choose now and hear it later.\n"
	}
	info.SetText("[::b]Sound?[::-]\n\nchess-cli can play sounds for moves, captures, checks, low time and " +
		"the end of a game, made for the terminal. Use ↑↓ to hear each, Enter to choose.\n" + note +
		"\n[#a0a0a0]This can be changed in Options.[-]")
	Root.App.SetFocus(list)
}

func onboardInit() cv.Primitive {
	preview := newPreviewBoard()
	list := cv.NewList()
	onboardList = list
	list.SetHover(true)
	list.SetWrapAround(true)
	info := textBox(cv.AlignLeft)
	forwardKeys(preview, list)

	rec := RecommendedPieces()
	styles := []string{PiecesImage, PiecesText, PiecesBasic}
	names := map[string]string{PiecesImage: "Image rendering", PiecesText: "Advanced rendering", PiecesBasic: "Basic rendering"}
	explain := map[string]string{
		PiecesImage: "Lichess boards and pieces as pictures",
		PiecesText:  "Unicode chess symbols",
		PiecesBasic: "plain symbols and 8 colours, for any terminal",
	}
	if TextSizingMode() {
		explain[PiecesText] = "Unicode chess symbols, drawn big"
	}
	if !GraphicsSupported() {
		explain[PiecesImage] = "not available in this terminal"
	}
	for _, st := range styles {
		st := st
		text := names[st]
		if st == rec {
			text += " [green](recommended)[-]"
		}
		item := cv.NewListItem(text)
		item.SetShortcut(numberKey(len(list.GetItems())))
		item.SetSecondaryText("  " + explain[st])
		item.SetSelectedFunc(func() {
			_ = SetPiecesSetting(st)
			soundStep()
		})
		list.AddItem(item)
	}
	for i, st := range styles {
		if st == rec {
			list.SetCurrentItem(i)
		}
	}
	list.SetChangedFunc(func(i int, _ *cv.ListItem) { previewPieces(styles[i]) })
	previewPieces(rec)

	supports := "[red]✗[-] images"
	if GraphicsSupported() {
		supports = "[green]✓[-] images"
	}
	if TextSizingMode() {
		supports += "   [green]✓[-] big text"
	} else {
		supports += "   [red]✗[-] big text"
	}
	if trueColor() {
		supports += "   [green]✓[-] true colour"
	} else {
		supports += "   [yellow]?[-] true colour"
	}
	soundStep = func() {
		if api.UserInfo.SoundAsked {
			keysStep(list, info)
			return
		}
		onboardSound(list, info)
	}
	if api.UserInfo.Pieces != "" { // only the later questions are new
		defer soundStep()
	}
	info.SetText(fmt.Sprintf("[::b]Welcome to chess-cli![::-]\n\nYou're using [::b]%s[::-].\n%s\n\n"+
		"We recommend [green]%s[-] here. Use ↑↓ to preview each style, Enter to choose.\n\n"+
		"[#a0a0a0]This can be changed in Options.[-]", cv.Escape(terminalName()), supports, names[rec]))

	right := cv.NewFlex()
	right.SetDirection(cv.FlexRow)
	right.AddItem(info, 0, 1, false)
	right.AddItem(list, 7, 0, true)
	right.AddItem(cv.NewBox(), 0, 1, false)
	right.SetPadding(1, 0, 2, 2)

	body := newPreviewLayout(preview, right)

	root := cv.NewFlex()
	root.SetDirection(cv.FlexRow)
	root.AddItem(body, 0, 1, true)
	root.AddItem(ribbonPrimitive(" chess-cli · first run │ ↑↓ preview │ Enter to choose │ Ctrl-C quit", list), 1, 0, false)
	registerMenus("onboard", false, list)
	return root
}

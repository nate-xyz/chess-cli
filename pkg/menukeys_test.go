package pkg

import (
	"strings"
	"testing"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
)

// press sends a key through the whole app, as the terminal would.
func press(s tc.Screen, k tc.Key, str string) {
	s.EventQ() <- tc.NewEventKey(k, str, tc.ModNone)
}

func shownPanel() string {
	name, _ := Root.nav.GetFrontPanel()
	return name
}

// Menu shortcuts work wherever the keyboard is on the screen; a playable
// board keeps its own keys, typing in a text box keeps letters, and screens
// with several menus cycle with Tab.
func TestMenuKeysEverywhere(t *testing.T) {
	scr := startTestApp(t, 160, 50)

	// Study: the board has the keyboard; x is the menu's Leave.
	onUI(func() {
		Root.Switch("study")
		Root.study.show(parseStudyPGN(testStudyPGN), "")
		Root.App.SetFocus(Root.study.Board)
	})
	press(scr, tc.KeyRune, "x")
	waitFor(t, "study x leaves", func() bool { return shownPanel() == "welcome" })

	// Local game: hjkl stay with the board (l moves the cursor, it doesn't
	// leave); x leaves even with the board focused.
	onUI(func() { startNewLocalGame() })
	waitFor(t, "local game", func() bool { return shownPanel() == "localgame" })
	onUI(func() { Root.App.SetFocus(Root.lgame.Board) })
	press(scr, tc.KeyRune, "l")
	press(scr, tc.KeyRune, "l")
	waitFor(t, "cursor moved by l", func() bool {
		b := Root.lgame.Board
		return b.keyboard && b.cursor.String() == "f2"
	})
	if shownPanel() != "localgame" {
		t.Fatalf("l left the game (now on %s)", shownPanel())
	}
	// typing a move: letters go to the box, not the menu
	onUI(func() { Root.lgame.Move.show() })
	press(scr, tc.KeyRune, "x")
	waitFor(t, "x typed in the move box", func() bool { return Root.lgame.Move.input.GetText() == "x" })
	onUI(func() {
		Root.lgame.Move.input.SetText("")
		Root.lgame.Move.hide()
		Root.App.SetFocus(Root.lgame.Board)
	})
	press(scr, tc.KeyRune, "x")
	waitFor(t, "x leaves the local game", func() bool { return shownPanel() == "welcome" })

	// Saved games: Tab moves from the game list to the menu.
	onUI(func() { gotoSaved() })
	waitFor(t, "saved", func() bool { return shownPanel() == "saved" })
	var gameList, options *cv.List
	onUI(func() {
		sm := menusByScreen["saved"]
		gameList, options = sm.lists[0], sm.lists[1]
		Root.App.SetFocus(gameList)
	})
	press(scr, tc.KeyTab, "")
	waitFor(t, "Tab to the menu", func() bool { return Root.App.GetFocus() == options })
	press(scr, tc.KeyRune, "b")
	waitFor(t, "b goes back", func() bool { return shownPanel() != "saved" })
}

// The first key on a title or other box that hands the keyboard to a menu
// isn't lost: ↓ moves the menu straight away.
func TestForwardKeysKeepsTheKey(t *testing.T) {
	scr := startTestApp(t, 160, 50)
	onUI(func() { gotoWelcome() })
	waitFor(t, "welcome", func() bool { return shownPanel() == "welcome" })
	var list *cv.List
	onUI(func() {
		list = menusByScreen["welcome"].lists[0]
		list.SetCurrentItem(0)
		Root.App.SetFocus(newsBox)
	})
	press(scr, tc.KeyDown, "")
	waitFor(t, "↓ moved the menu", func() bool { return Root.App.GetFocus() == list && list.GetCurrentItemIndex() == 1 })
}

// Resign and abort need a second press.
func TestConfirmItem(t *testing.T) {
	startTestApp(t, 160, 50)
	done := 0
	item := cv.NewListItem("")
	confirmItem(item, "Resign", "Resign? Press again", func() { done++ })
	l := cv.NewList()
	l.AddItem(item)
	enter := func() { l.InputHandler()(tc.NewEventKey(tc.KeyEnter, "", tc.ModNone), func(cv.Primitive) {}) }
	onUI(func() {
		enter()
		if done != 0 || item.GetMainText() != "[red::b]Resign? Press again[-::-]" {
			t.Errorf("first press: done %d, text %q", done, item.GetMainText())
		}
		enter()
		if done != 1 || item.GetMainText() != "Resign" {
			t.Errorf("second press: done %d, text %q", done, item.GetMainText())
		}
	})
}

// The home menu's column is as wide as its longest description, so nothing
// is cut off, at the size it used to cut "rated or casual".
func TestMenuColumnFits(t *testing.T) {
	startTestApp(t, 160, 50)
	defer onUI(func() { Root.nav.SetRect(0, 0, 160, 50) })
	onUI(func() {
		gotoWelcome()
		list := menusByScreen["welcome"].lists[0]
		_, p := Root.nav.GetFrontPanel()
		scr := newMockScreen(t, 143, 45)
		p.SetRect(0, 0, 143, 45)
		p.Draw(scr)
		_, _, w, _ := list.GetInnerRect()
		if need := menuWidth(list); w < need {
			t.Errorf("home menu %d wide, needs %d", w, need)
		}
	})
}

// The challenge maker's time step swaps its menu for sliders: they get the
// keyboard and room (not the 2 columns an empty menu needs), and Submit goes
// on to the next step.
func TestChallengeIntervalStep(t *testing.T) {
	scr := startTestApp(t, 160, 50)
	var list *cv.List
	onUI(func() {
		gotoChallengeConstruction()
		list = menusByScreen["challenge"].lists[0]
		Root.App.SetFocus(list)
	})
	for _, k := range []string{"a", "a", "a"} { // Random, standard, real time
		press(scr, tc.KeyRune, k)
	}
	var slider *cv.Slider
	waitFor(t, "the minutes slider has the keyboard", func() bool {
		slider, _ = Root.App.GetFocus().(*cv.Slider)
		return slider != nil
	})
	onUI(func() {
		_, p := Root.nav.GetFrontPanel()
		p.SetRect(0, 0, 160, 50)
		p.Draw(newMockScreen(t, 160, 50))
		if _, _, w, _ := slider.GetRect(); w < 40 {
			t.Errorf("slider %d wide", w)
		}
	})
	press(scr, tc.KeyRight, "")
	waitFor(t, "right moves the slider", func() bool {
		return strings.Contains(slider.GetLabel(), "11")
	})
	// Tab past the increment slider to Submit
	press(scr, tc.KeyTab, "")
	press(scr, tc.KeyTab, "")
	press(scr, tc.KeyEnter, "")
	waitFor(t, "the rated step", func() bool {
		return Root.App.GetFocus() == list && list.GetItemCount() > 0 && list.GetItem(0).GetMainText() == "casual"
	})
	onUI(func() { gotoWelcome() })
}

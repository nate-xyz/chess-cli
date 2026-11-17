package pkg

import (
	"time"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
)

// Menu keys that work the same everywhere: a menu's letter shortcuts work
// wherever the keyboard is on its screen (not only when the menu itself has
// focus), unless you're typing in a text box or the board needs that key.
// Screens with more than one menu cycle through them with Tab.

type screenMenus struct {
	lists []*cv.List
	tab   bool // Tab moves between the menus (screens without their own Tab)
}

var menusByScreen = map[string]*screenMenus{}

// registerMenus records the menus on a screen (a panel name).
func registerMenus(screen string, tab bool, lists ...*cv.List) {
	menusByScreen[screen] = &screenMenus{lists: lists, tab: tab}
}

// menuKeys is part of the global key handler. It returns nil when it used
// the key.
func menuKeys(ev *tc.EventKey) *tc.EventKey {
	name, _ := Root.nav.GetFrontPanel()
	sm := menusByScreen[name]
	if sm == nil {
		return ev
	}
	focus := Root.App.GetFocus()
	switch f := focus.(type) {
	case *cv.InputField:
		return ev // typing
	case *BoardView:
		if ev.Key() == tc.KeyRune && f.takesKeys() {
			if f.PromotionOpen() || containsRune(boardKeys(), ev.Str()) {
				return ev
			}
		}
	}

	if sm.tab && (ev.Key() == tc.KeyTab || ev.Key() == tc.KeyBacktab) {
		step := 1
		if ev.Key() == tc.KeyBacktab {
			step = -1
		}
		cur := -1
		for i, l := range sm.lists {
			if focus == l {
				cur = i
			}
		}
		for n := 1; n <= len(sm.lists); n++ {
			l := sm.lists[((cur+step*n)%len(sm.lists)+len(sm.lists))%len(sm.lists)]
			if l.GetItemCount() > 0 {
				Root.App.SetFocus(l)
				return nil
			}
		}
		return ev
	}

	if ev.Key() != tc.KeyRune || ev.Str() == " " {
		return ev
	}
	// The focused menu gets its own shortcuts first.
	for _, l := range sm.lists {
		if focus == l && hasShortcut(l, ev.Str()) {
			return ev
		}
	}
	for _, l := range sm.lists {
		if focus != l && hasShortcut(l, ev.Str()) {
			Root.App.SetFocus(l)
			l.InputHandler()(ev, Root.App.SetFocus)
			return nil
		}
	}
	return ev
}

// numberKey is the shortcut for the i-th entry of a list of games or
// challenges: 1-9, then none. Letters are for menu actions.
func numberKey(i int) rune {
	if i < 9 {
		return rune('1' + i)
	}
	return 0
}

func hasShortcut(l *cv.List, key string) bool {
	for _, it := range l.GetItems() {
		if it.GetShortcut() != 0 && string(it.GetShortcut()) == key {
			return true
		}
	}
	return false
}

func containsRune(set, s string) bool {
	for _, r := range set {
		if string(r) == s {
			return true
		}
	}
	return false
}

// forwardKeys makes a decorative box hand the keyboard to target, passing
// the key on so it isn't lost (pressing ↓ on a title moves the menu).
func forwardKeys(box interface {
	SetInputCapture(func(*tc.EventKey) *tc.EventKey)
}, target cv.Primitive) {
	box.SetInputCapture(func(ev *tc.EventKey) *tc.EventKey {
		if target == nil {
			return ev
		}
		Root.App.SetFocus(target)
		if h := target.InputHandler(); h != nil {
			h(ev, Root.App.SetFocus)
		}
		return nil
	})
}

// confirmItem makes a menu item ask for a second press before doing f
// (resign, abort): the first press changes its text to ask, for a few
// seconds.
func confirmItem(item *cv.ListItem, text, ask string, f func()) {
	var armed time.Time
	item.SetMainText(text)
	item.SetSelectedFunc(func() {
		if time.Since(armed) < 3*time.Second {
			armed = time.Time{}
			item.SetMainText(text)
			f()
			return
		}
		armed = time.Now()
		item.SetMainText("[red::b]" + ask + "[-::-]")
		time.AfterFunc(3*time.Second, func() {
			Root.App.QueueUpdateDraw(func() {
				if !armed.IsZero() && time.Since(armed) >= 3*time.Second {
					item.SetMainText(text)
				}
			})
		})
	})
}

// fitMenuColumn makes column col of a grid as wide as its menu needs (the
// widest entry or description, with the shortcut), so descriptions aren't
// cut off; the other columns share the rest. cols are the grid's columns
// otherwise (used when the menu wouldn't leave room for the rest, or the
// menu is empty because something else has its place for now).
func fitMenuColumn(g *cv.Grid, list *cv.List, cols []int, col int) {
	const minRest = 30 // columns kept for the rest of the screen
	g.SetDrawFunc(func(_ tc.Screen, x, y, w, h int) (int, int, int, int) {
		c := append([]int(nil), cols...)
		if need := menuWidth(list); list.GetItemCount() > 0 && need <= w-minRest {
			c[col] = need
		}
		g.SetColumns(c...) // the grid's own lock, not the box's held here
		return x, y, w, h
	})
}

// menuWidth is the columns a list needs to show all of every entry.
func menuWidth(l *cv.List) int {
	widest, shortcut := 0, 0
	for _, it := range l.GetItems() {
		widest = max(widest, cv.TaggedStringWidth(it.GetMainText()), cv.TaggedStringWidth(it.GetSecondaryText()))
		if it.GetShortcut() != 0 {
			shortcut = 4 // "(x) "
		}
	}
	return shortcut + widest + 2
}

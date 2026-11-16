package pkg

import (
	"strings"

	cv "codeberg.org/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
)

// Moving around with the arrow keys, or also with vim's hjkl (on the board
// and in lists, where g and G go to the first and last item too). Chosen at
// first run and in Options; on unless turned off.

func vimKeys() bool { return !api.UserInfo.NoVimKeys }

func setVimKeys(on bool) {
	_ = api.UpdateConfig(func(c *api.UserConfig) { c.NoVimKeys, c.KeysAsked = !on, true })
	applyKeyPrefs()
}

// keyRibbons are the hint ribbons that mention the movement keys; their
// text has "arrows/hjkl", shown as "arrows" when hjkl is off.
var keyRibbons = map[*cv.TextView]string{}

// applyKeyPrefs sets the lists' keys and the ribbons for the setting.
func applyKeyPrefs() {
	if vimKeys() {
		cv.Keys.MoveUp2, cv.Keys.MoveDown2 = []string{"k"}, []string{"j"}
		cv.Keys.MoveLeft2, cv.Keys.MoveRight2 = []string{"h"}, []string{"l"}
		cv.Keys.MoveFirst2, cv.Keys.MoveLast2 = []string{"g"}, []string{"G"}
	} else {
		cv.Keys.MoveUp2, cv.Keys.MoveDown2, cv.Keys.MoveLeft2, cv.Keys.MoveRight2 = nil, nil, nil, nil
		cv.Keys.MoveFirst2, cv.Keys.MoveLast2 = nil, nil
	}
	for tv, text := range keyRibbons {
		tv.SetText(keysText(text))
	}
}

func keysText(text string) string {
	if vimKeys() {
		return text
	}
	return strings.ReplaceAll(text, "arrows/hjkl", "arrows")
}

// boardKeys are the keys a playable board uses itself.
func boardKeys() string {
	if vimKeys() {
		return "hjkl[] "
	}
	return "[] "
}

// onboardKeys asks arrows or arrows and hjkl.
func onboardKeys(list *cv.List, info *cv.TextView, next func()) {
	list.SetChangedFunc(nil)
	list.Clear()
	for i, c := range []struct {
		text, explain string
		vim           bool
	}{
		{"Arrow keys and hjkl [green](recommended)[-]", "  hjkl move too, like vim (g and G in lists)", true},
		{"Arrow keys only", "  h, j, k and l are just letters", false},
	} {
		c := c
		item := cv.NewListItem(c.text)
		item.SetShortcut(numberKey(i))
		item.SetSecondaryText(c.explain)
		item.SetSelectedFunc(func() {
			setVimKeys(c.vim)
			next()
		})
		list.AddItem(item)
	}
	list.SetCurrentItem(0)
	info.SetText("[::b]Moving around[::-]\n\nThe arrow keys move the cursor on the board and through menus. " +
		"chess-cli can also take [::b]h j k l[::-] (left, down, up, right), as in vim and many terminal programs.\n\n" +
		"Menu letters work either way; on the board you can always click, or type a move with Ctrl-K.\n\n" +
		"[#a0a0a0]This can be changed in Options.[-]")
	Root.App.SetFocus(list)
}

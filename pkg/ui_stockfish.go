package pkg

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	cv "codeberg.org/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// The "Play local Stockfish" menu on the home screen: level and colour, or,
// when no Stockfish is found, ways to get one.

var (
	stockfishColor = "random" // "white", "black" or "random"
	downloading    bool
)

// expandHome turns a leading ~ into the home folder.
func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") || p == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

func stockfishLevel() int {
	if l := api.UserInfo.StockfishLevel; l >= 1 && l <= 8 {
		return l
	}
	return 3
}

// stockfishList fills the home menu list; back returns to the previous menu.
func stockfishList(list *cv.List, back func()) {
	again := func() { stockfishList(list, back) }
	cur := list.GetCurrentItemIndex()
	list.Clear()
	add := func(text, explain string, key rune, f func()) *cv.ListItem {
		item := cv.NewListItem(text)
		item.SetSecondaryText(explain)
		if key != 0 {
			item.SetShortcut(key)
		}
		item.SetSelectedFunc(f)
		list.AddItem(item)
		return item
	}
	defer list.SetCurrentItem(cur)

	path := findEngine()
	if path == "" {
		add("Download Stockfish for me", "Official release from GitHub (~80 MB), kept in chess-cli's folder", 'd', func() {
			if !downloading {
				downloadEngine(list, again)
			}
		})
		add("Use my Stockfish…", "Enter the path of a Stockfish you already have", 'u', func() { askEnginePath(again) })
		add("Install it yourself", installHint()+"  (then come back here)", 'i', again)
		add("Back", "Go back", 'b', back)
		return
	}

	level := stockfishLevel()
	add(fmt.Sprintf("Start game · level %d · %s", level, stockfishColor), "Using "+filepath.Base(path), 's', func() {
		player := chess.White
		switch stockfishColor {
		case "black":
			player = chess.Black
		case "random":
			if rand.Intn(2) == 1 {
				player = chess.Black
			}
		}
		startEngineGame(path, level, player)
	})
	add(fmt.Sprintf("Level: %d", level), "1 (beginner) to 8 (full strength); Enter for the next", 'l', func() {
		_ = api.UpdateConfig(func(c *api.UserConfig) { c.StockfishLevel = level%8 + 1 })
		again()
	})
	add("Colour: "+stockfishColor, "Play white, black or random; Enter to change", 'c', func() {
		stockfishColor = map[string]string{"random": "white", "white": "black", "black": "random"}[stockfishColor]
		again()
	})
	add("Clock: "+localClockName(localClockIndex()), "Time control; Enter for the next", 'k', func() {
		nextLocalClock()
		again()
	})
	add("Use a different Stockfish…", path, 'u', func() { askEnginePath(again) })
	add("Back", "Go back", 'b', back)
}

// askEnginePath asks where Stockfish is, checks it, and remembers it.
func askEnginePath(then func()) {
	const q = "Where is Stockfish? Enter the full path of the program."
	askText(q, "Path: ", func(text string) {
		p, err := filepath.Abs(expandHome(text))
		if err == nil {
			err = checkEngine(p)
		}
		if err != nil {
			promptError(q, "That didn't work: "+err.Error())
			return
		}
		_ = api.UpdateConfig(func(c *api.UserConfig) { c.StockfishPath = p })
		gotoWelcome()
		then()
	})
}

// downloadEngine downloads Stockfish, showing progress in the menu item.
func downloadEngine(list *cv.List, then func()) {
	downloading = true
	item := list.GetItem(list.GetCurrentItemIndex())
	item.SetSecondaryText("Starting download…")
	go func() {
		last := -1
		path, err := DownloadStockfish(func(done, total int64) {
			if total <= 0 {
				return
			}
			pct := int(done * 100 / total)
			if pct == last {
				return
			}
			last = pct
			Root.App.QueueUpdateDraw(func() {
				item.SetSecondaryText(fmt.Sprintf("Downloading… %d%% of %d MB", pct, total>>20))
			})
		})
		if err == nil {
			err = checkEngine(path)
		}
		Root.App.QueueUpdateDraw(func() {
			downloading = false
			if err != nil {
				item.SetSecondaryText(fmt.Sprintf("[red]Download failed: %v[-]", err))
				return
			}
			then()
		})
	}()
}

package pkg

import (
	"sync/atomic"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// Game screen layout: a sidebar on the left (players, clocks, moves, menu)
// and the board as big as the height allows, against the right edge, with
// the shortcut ribbon along the bottom.
//
//	┌──────────────┬───────────────────────┐
//	│ opponent     │                       │
//	│ status       │                       │
//	│ moves        │         board         │
//	│ menu         │                       │
//	│ you          │                       │
//	├──────────────┴───────────────────────┤
//	│ ribbon: shortcuts                    │
//	└──────────────────────────────────────┘

const (
	minSidebar  = 30 // columns kept for the sidebar
	boardMargin = 2  // columns between the board and the right edge (~ one row's height)
)

// boardLayout puts side and board next to each other, giving the board
// exactly the width it needs and the sidebar the rest.
type boardLayout struct {
	*cv.Flex
	board *BoardView
}

func (l *boardLayout) Draw(screen tc.Screen) {
	_, _, w, h := l.GetRect()
	bw := l.board.WidthFor(screen, w-minSidebar-boardMargin, h) + boardMargin
	l.ResizeItem(l.board, bw, 0)
	l.Flex.Draw(screen)
}

// newGameLayout builds a game screen around a board and its sidebar. The
// board starts with keyboard focus.
func newGameLayout(board *BoardView, side *cv.Flex, ribbon string) *cv.Flex {
	board.right = true
	side.SetPadding(1, 0, 2, 2) // breathing room from the left edge and the board
	body := &boardLayout{Flex: cv.NewFlex(), board: board}
	body.AddItem(side, 0, 1, false)
	body.AddItem(board, 0, 0, true)

	root := cv.NewFlex()
	root.SetDirection(cv.FlexRow)
	root.AddItem(body, 0, 1, true)
	root.AddItem(ribbonPrimitive(ribbon, board), 1, 0, false)
	return root
}

// moveBox is the typed-move field: hidden until Ctrl-K, closed again by Esc
// or after a move is entered.
type moveBox struct {
	side  *cv.Flex
	input *cv.InputField
	board *BoardView
}

func newMoveBox(side *cv.Flex, board *BoardView, submit func(text string)) *moveBox {
	m := &moveBox{side: side, board: board, input: cv.NewInputField()}
	m.input.SetLabel("Move: ")
	m.input.SetPlaceholder("e4, Nf3, e7e8=Q…")
	m.input.SetDoneFunc(func(key tc.Key) {
		text := m.input.GetText()
		m.input.SetText("")
		if key == tc.KeyEnter && moveBoxAlways() {
			// always-on: stay in the box for the next move
		} else {
			m.hide()
		}
		if key == tc.KeyEnter && text != "" {
			submit(text)
		}
	})
	moveBoxes = append(moveBoxes, m)
	return m
}

// moveBoxes are all the screens' move boxes, for the always-on setting.
var moveBoxes []*moveBox

// moveBoxAlways reports whether the move box is always shown (Options),
// rather than opened with Ctrl-K.
func moveBoxAlways() bool { return api.UserInfo.MoveBoxAlways }

// setMoveBoxAlways changes the setting and shows or hides the boxes.
func setMoveBoxAlways(on bool) {
	_ = api.UpdateConfig(func(c *api.UserConfig) { c.MoveBoxAlways = on })
	size := 0
	if on {
		size = 1
	}
	for _, m := range moveBoxes {
		m.side.ResizeItem(m.input, size, 0)
	}
}

// focusMoveBox puts the keyboard in the move box of a game screen being
// opened, when the box is always on. Returns false if it isn't.
func focusMoveBox(panel string) bool {
	if !moveBoxAlways() {
		return false
	}
	var m *moveBox
	switch panel {
	case "localgame":
		m = Root.lgame.Move
	case "onlinegame":
		m = Root.ongame.Move
	case "puzzle":
		m = Root.puzzle.Move
	}
	if m == nil {
		return false
	}
	m.show()
	return true
}

func (m *moveBox) show() {
	m.side.ResizeItem(m.input, 1, 0)
	Root.App.SetFocus(m.input)
}

// hide closes the box (it stays visible when always on) and gives the
// keyboard back to the board.
func (m *moveBox) hide() {
	if !moveBoxAlways() {
		m.side.ResizeItem(m.input, 0, 0)
	}
	Root.App.SetFocus(m.board)
}

// isCtrlK matches Ctrl-K however the terminal reports it (as a control code,
// or as "k" with the Ctrl modifier under the kitty keyboard protocol).
func isCtrlK(ev *tc.EventKey) bool {
	return ev.Key() == tc.KeyCtrlK ||
		(ev.Key() == tc.KeyRune && (ev.Str() == "k" || ev.Str() == "K") && ev.Modifiers()&tc.ModCtrl != 0)
}

// globalKeys is the application's key capture (cview sends keys straight to
// the focused widget, so screen-wide shortcuts live here): Ctrl-K opens the
// move box on game screens, Esc leaves the Options screen.
func globalKeys(ev *tc.EventKey) *tc.EventKey {
	switch name, _ := Root.nav.GetFrontPanel(); name {
	case "localgame":
		if isCtrlK(ev) {
			Root.lgame.Move.show()
			return nil
		}
	case "onlinegame":
		if isCtrlK(ev) {
			Root.ongame.Move.show()
			return nil
		}
	case "puzzle":
		if isCtrlK(ev) {
			Root.puzzle.Move.show()
			return nil
		}
	case "watch":
		if ev.Key() == tc.KeyRune && ev.Str() == "f" {
			Root.watch.flip()
			return nil
		}
	case "study":
		if ev.Key() == tc.KeyRune && ev.Str() == "f" {
			Root.study.flip()
			return nil
		}
	case "options":
		if ev.Key() == tc.KeyEscape {
			optionsBack()
			return nil
		}
	}
	return ev
}

// textBox is a read-only text view for the sidebar.
func textBox(align int) *cv.TextView {
	tv := cv.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetWordWrap(true)
	tv.SetTextAlign(align)
	tv.SetScrollBarVisibility(cv.ScrollBarNever) // a bar beside the board looks like an eval bar
	return tv
}

// Menus highlight the item under the mouse pointer. cview's List only moves
// its selection on hover; the selection is hidden while the list isn't
// focused, so it's shown while the pointer is over the list.
var hoverLists []*hoverList

type hoverList struct {
	*cv.List
	over    bool   // pointer is over the list
	onLeave func() // when the pointer leaves the list
}

// onHoverLeave runs f when the pointer leaves the menu l (made by menuList).
// Hovering moves a list's selection; f can put it back.
func onHoverLeave(l *cv.List, f func()) {
	for _, h := range hoverLists {
		if h.List == l {
			h.onLeave = f
		}
	}
}

// menuList makes a compact list (no secondary text) that highlights on hover.
func menuList() *cv.List {
	l := cv.NewList()
	l.ShowSecondaryText(false)
	l.SetHover(true)
	l.SetWrapAround(true)
	l.SetSelectedFocusOnly(true)
	hoverLists = append(hoverLists, &hoverList{List: l})
	return l
}

// trackHover is the application's mouse capture: it shows a menu's
// highlight while the pointer is over it.
func trackHover(ev *tc.EventMouse, action cv.MouseAction) (*tc.EventMouse, cv.MouseAction) {
	if ev == nil {
		return ev, action
	}
	x, y := ev.Position()
	changed := false
	for _, l := range hoverLists {
		if over := l.InRect(x, y); over != l.over {
			l.over = over
			l.SetSelectedFocusOnly(!over)
			changed = true
			if !over && l.onLeave != nil && Root.App.GetFocus() != l.List {
				l.onLeave()
			}
		}
	}
	if changed {
		go Root.App.QueueUpdateDraw(func() {})
	}
	return ev, action
}

// postGame is a game-over screen in the game layout: the final board on the
// right (step back through it as usual); the players, the result, the moves
// and a menu in the sidebar. The menu has the focus; Tab goes to the board.
type postGame struct {
	Board   *BoardView
	Top     *cv.TextView // the player at the top of the board
	Result  *cv.TextView
	History *cv.TextView
	List    *cv.List
	Bottom  *cv.TextView
}

func newPostGame(items []*cv.ListItem, ribbon string) (*postGame, cv.Primitive) {
	p := &postGame{
		Board:   NewBoardView(nil),
		Top:     textBox(cv.AlignLeft),
		Result:  textBox(cv.AlignLeft),
		History: textBox(cv.AlignLeft),
		List:    menuList(),
		Bottom:  textBox(cv.AlignLeft),
	}
	for _, item := range items {
		p.List.AddItem(item)
	}
	linkMoveList(p.History, p.Board)
	side := cv.NewFlex()
	side.SetDirection(cv.FlexRow)
	side.AddItem(p.Top, 2, 0, false)
	side.AddItem(p.Result, 0, 1, false)
	side.AddItem(p.History, 0, 2, false)
	side.AddItem(p.List, len(items)+1, 0, true)
	side.AddItem(p.Bottom, 2, 0, false)

	p.Board.SetTabFunc(func() { Root.App.SetFocus(p.List) })
	p.List.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		if event.Key() == tc.KeyTab {
			Root.App.SetFocus(p.Board)
			return nil
		}
		return event
	})
	return p, newGameLayout(p.Board, side, ribbon)
}

// menuItem is a menu entry: text, shortcut key and what it does.
func menuItem(text string, key rune, f func()) *cv.ListItem {
	item := cv.NewListItem(text)
	item.SetShortcut(key)
	item.SetSelectedFunc(f)
	return item
}

// show fills the screen from the game that just ended (Root.gameState, its
// board and player boxes) and switches to it.
func (p *postGame) show(panel string, from *BoardView, top, bottom *cv.TextView, white bool, winner chess.Color, method string) {
	p.Top.SetText(top.GetText(false))
	p.Bottom.SetText(bottom.GetText(false))
	p.Result.SetText(Root.gameState.Status)
	Root.gameState.Status = ""
	DrawMoveHistory(p.History)
	DrawBoard(p.Board, white)
	p.Board.copyClockBars(from)
	if method != "" {
		p.Board.SetResult(winner, method)
	}
	Root.Switch(panel)
	Root.App.SetFocus(p.List)
}

// copyClockBars shows another board's clocks, stopped: the final times.
func (b *BoardView) copyClockBars(from *BoardView) {
	from.mu.Lock()
	top, bottom := from.topBar, from.bottomBar
	from.mu.Unlock()
	top.active, bottom.active = false, false
	b.SetClockBars(top, bottom)
}

// Sidebar boxes: a thin muted-grey frame with a title, a little brighter
// while focused.
var (
	colBoxBorder  = tc.PaletteColor(240)
	colBoxFocused = tc.PaletteColor(246)
	colBoxTitle   = tc.PaletteColor(246)
)

// boxed frames a sidebar box and returns a function to change its title.
// The title is drawn here rather than by cview, whose title ends in a reset
// that some terminals show as a grey block.
func boxed(b *cv.Box, title string) func(string) {
	sidebarBoxes = append(sidebarBoxes, b)
	b.SetBorder(true)
	b.SetBorderColor(colBoxBorder)
	b.SetBorderColorFocused(colBoxFocused)
	b.SetDrawFunc(func(screen tc.Screen, x, y, w, h int) (int, int, int, int) {
		_, style, _ := screen.Get(x+1, y) // the border's style
		for i, r := range []rune(" " + title + " ") {
			if 2+i < w-1 {
				screen.SetContent(x+1+i, y, r, nil, style.Foreground(colBoxTitle))
			}
		}
		return x + 2, y + 1, max(0, w-4), max(0, h-2) // inside the frame, a space each side
	})
	return func(t string) { title = t }
}

// The UI is "limited" with basic rendering, or on a terminal with fewer than
// 256 colours: then only the standard colours are used, with no dark tints
// (they'd all be black) and plain underlines (old terminals can misread the
// styled ones). screenColors is the terminal's colour count, kept after each
// frame (afterDraw), from the app's own screen only.
var (
	screenColors atomic.Int32
	wasLimited   atomic.Bool
)

func uiLimited() bool {
	n := screenColors.Load()
	return PiecesSetting() == PiecesBasic || (n > 0 && n < 256)
}

// Text views and boxes whose colours depend on uiLimited.
var (
	moveLists    []*cv.TextView
	sidebarBoxes []*cv.Box
)

// colUnderline is the side lines' underline: dim, so it doesn't crowd the text.
var colUnderline = tc.PaletteColor(240)

// applyUIColours sets the move lists' highlight and the boxes' borders for
// a full or a limited terminal. UI goroutine.
func applyUIColours() {
	lim := uiLimited()
	for _, tv := range moveLists {
		if lim {
			tv.SetHighlightBackgroundColor(tc.ColorSilver)
			tv.SetHighlightForegroundColor(tc.ColorBlack)
		} else {
			tv.SetHighlightBackgroundColor(colCurrentMove)
			tv.SetHighlightForegroundColor(tc.ColorWhite)
		}
	}
	for _, b := range sidebarBoxes {
		if lim {
			b.SetBorderColor(tc.ColorSilver)
			b.SetBorderColorFocused(tc.ColorWhite)
		} else {
			b.SetBorderColor(colBoxBorder)
			b.SetBorderColorFocused(colBoxFocused)
		}
	}
	if Root.study != nil {
		Root.study.shownAt = nil // redraw the move list's tints
		Root.study.showNode()
	}
}

// checkLimited re-applies the colours when the rendering mode or terminal
// changes between full and limited. Called after each frame.
func checkLimited(screen tc.Screen) {
	screenColors.Store(int32(screen.Colors()))
	if lim := uiLimited(); wasLimited.Swap(lim) != lim && Root.App != nil {
		go Root.App.QueueUpdateDraw(applyUIColours)
	}
}

// underlineScreen draws a text view with the terminal's underline styles,
// which cview's markup can't express: an underline with blink (which
// chess-cli never uses otherwise) is dashed, with strikethrough too dotted,
// with italic double, with both curly (see underlineTags).
type underlineScreen struct{ tc.Screen }

func (s underlineScreen) Put(x, y int, str string, style tc.Style) (string, int) {
	switch {
	case !style.HasUnderline():
	case uiLimited(): // a plain underline only
		style = style.Blink(false).StrikeThrough(false).Italic(false)
	case style.HasBlink():
		u := tc.UnderlineStyleDashed
		switch strike, italic := style.HasStrikeThrough(), style.HasItalic(); {
		case strike && italic:
			u = tc.UnderlineStyleCurly
		case strike:
			u = tc.UnderlineStyleDotted
		case italic:
			u = tc.UnderlineStyleDouble
		}
		style = style.Blink(false).StrikeThrough(false).Italic(false).Underline(u, colUnderline)
	default:
		style = style.Underline(tc.UnderlineStyleSolid, colUnderline)
	}
	return s.Screen.Put(x, y, str, style)
}

// styledText is a text view drawn through underlineScreen.
type styledText struct{ *cv.TextView }

func (t styledText) Draw(screen tc.Screen) { t.TextView.Draw(underlineScreen{screen}) }

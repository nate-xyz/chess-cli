package pkg

import (
	"fmt"
	"strings"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// Studies: reading a Lichess study, chapter by chapter. The board shows the
// current line from the chapter's start to its end; ← → (or the wheel, [ ])
// step along it, and clicking a move, in a variation too, switches to that
// line. Comments and symbols are in the move list; the study's arrows and
// circles are drawn on the board.

type StudyScreen struct {
	Board    *BoardView
	Info     *cv.TextView
	Moves    *cv.TextView
	Chapters *cv.List
	List     *cv.List
	Side     *cv.Flex

	gen          int // bumped for each study opened; stale downloads are dropped
	name         string
	chapters     []*studyChapter
	ch           *studyChapter
	cur          *studyNode
	line         []*studyNode // the line on the board: root ... end
	white        bool
	syncing      bool       // highlighting a move ourselves, not a click
	shownAt      *studyNode // the node the move list was last drawn for
	chapterTitle func(string)
	infoW        int        // the width the top box was sized for
	soundAt      *studyNode // the move whose sound last played
}

const studyRibbon = " chess-cli · study │ ← → moves, ↑ ↓ start/end │ click a move to go there │ f flip │ Tab menu │ Ctrl-C quit"

func (ss *StudyScreen) Init() cv.Primitive {
	ss.Board = NewBoardView(func(chess.Square) {}) // clicks: drawing only
	ss.Board.browse = true
	ss.Info = textBox(cv.AlignLeft)
	ss.Moves = textBox(cv.AlignLeft)
	ss.Moves.SetRegions(true)
	boxed(ss.Moves.Box, "Moves")
	moveLists = append(moveLists, ss.Moves)
	applyUIColours()
	ss.Chapters = menuList()
	ss.Chapters.SetScrollBarVisibility(cv.ScrollBarNever)
	ss.chapterTitle = boxed(ss.Chapters.Box, "Chapters")
	onHoverLeave(ss.Chapters, func() { // hovering moved the selection: back to the chapter showing
		for i, c := range ss.chapters {
			if c == ss.ch {
				ss.Chapters.SetCurrentItem(i)
			}
		}
	})
	ss.List = menuList()
	for _, it := range []struct {
		text string
		key  rune
		f    func()
	}{{"Flip board", 'f', ss.flip}, {"Leave", 'x', gotoWelcome}} {
		ss.List.AddItem(menuItem(it.text, it.key, it.f))
	}

	ss.Moves.SetHighlightedFunc(func(added, _, _ []string) {
		var id int
		if ss.syncing || len(added) == 0 || ss.ch == nil {
			return
		}
		if _, err := fmt.Sscanf(added[0], "n%d", &id); err != nil || id < 0 || id >= len(ss.ch.nodes) {
			return
		}
		ss.goTo(ss.ch.nodes[id])
		// the click focused the move list; keep the focus where it was
		if prev := Root.App.GetFocus(); prev != nil && prev != ss.Moves {
			go Root.App.QueueUpdateDraw(func() { Root.App.SetFocus(prev) })
		}
	})
	ss.Board.mu.Lock()
	ss.Board.onShown = func(ply int) { // the wheel or [ ] moved along the line
		if ply >= 0 && ply < len(ss.line) && ss.line[ply] != ss.cur {
			ss.cur = ss.line[ply]
			ss.showNode()
		}
	}
	ss.Board.mu.Unlock()

	side := cv.NewFlex()
	side.SetDirection(cv.FlexRow)
	side.AddItem(ss.Info, 4, 0, false)
	ss.Info.SetDrawFunc(func(_ tc.Screen, x, y, w, h int) (int, int, int, int) {
		if ss.infoW != 0 && w != ss.infoW { // the window was resized: refit the tags
			ss.infoW = w
			go Root.App.QueueUpdateDraw(ss.fitInfo)
		}
		return x, y, w, h
	})
	side.AddItem(styledText{ss.Moves}, 0, 1, false) // nested side lines get styled underlines
	side.AddItem(ss.Chapters, 0, 0, false)
	side.AddItem(ss.List, 2, 0, false)
	ss.Side = side

	// keys on the board: arrows walk the line (Lichess's keys)
	ss.Board.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		switch event.Key() {
		case tc.KeyLeft:
			ss.step(-1)
		case tc.KeyRight:
			ss.step(1)
		case tc.KeyUp, tc.KeyHome:
			ss.step(-len(ss.line))
		case tc.KeyDown, tc.KeyEnd:
			ss.step(len(ss.line))
		case tc.KeyTab:
			if ss.Chapters.GetItemCount() > 1 {
				Root.App.SetFocus(ss.Chapters)
			} else {
				Root.App.SetFocus(ss.List)
			}
		default:
			if event.Key() == tc.KeyRune && (event.Str() == "[" || event.Str() == "]") {
				ss.step(map[string]int{"[": -1, "]": 1}[event.Str()])
				return nil
			}
			return event
		}
		return nil
	})
	ss.Chapters.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		switch event.Key() {
		case tc.KeyTab:
			Root.App.SetFocus(ss.List)
			return nil
		case tc.KeyEscape:
			Root.App.SetFocus(ss.Board)
			return nil
		}
		return event
	})
	ss.List.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		if event.Key() == tc.KeyTab || event.Key() == tc.KeyEscape {
			Root.App.SetFocus(ss.Board)
			return nil
		}
		return event
	})
	registerMenus("study", false, ss.List, ss.Chapters)
	return newGameLayout(ss.Board, side, studyRibbon)
}

// openStudy downloads a study and shows it, starting at chapter (an ID, or
// "" for the first).
func openStudy(id, chapter string) {
	ss := Root.study
	ss.gen++
	gen := ss.gen
	ss.name, ss.chapters, ss.ch, ss.cur, ss.line = "", nil, nil, nil, nil
	ss.Chapters.Clear()
	ss.Side.ResizeItem(ss.Chapters, 0, 0)
	ss.Moves.SetText("")
	ss.Info.SetText("[#a0a0a0]Loading the study…[-]")
	ss.Board.SetPosition(chess.StartingPosition(), true)
	ss.Board.SetHistory(nil, nil)
	ss.Board.SetLastMove(chess.NoSquare, chess.NoSquare)
	ss.Board.HideClockBars()
	ss.Board.SetGlyph(chess.NoSquare, 0)
	Root.Switch("study")
	Root.App.SetFocus(ss.Board)
	go func() {
		pgn, err := api.GetStudyPGN(id)
		Root.App.QueueUpdateDraw(func() {
			if gen != ss.gen {
				return
			}
			if err != nil {
				ss.Info.SetText(fmt.Sprintf("[red]Couldn't open the study: %s[-]", cv.Escape(err.Error())))
				return
			}
			ss.show(parseStudyPGN(pgn), chapter)
		})
	}()
}

// show displays a downloaded study.
func (ss *StudyScreen) show(chapters []*studyChapter, chapter string) {
	if len(chapters) == 0 {
		ss.Info.SetText("[yellow]This study has no chapters.[-]")
		return
	}
	ss.chapters = chapters
	ss.name = chapters[0].studyName()
	start := 0
	ss.Chapters.Clear()
	for i, c := range chapters {
		i := i
		item := cv.NewListItem("")
		item.SetSelectedFunc(func() { ss.selectChapter(i) })
		ss.Chapters.AddItem(item)
		if chapter != "" && c.chapterID() == chapter {
			start = i
		}
	}
	rows := 0
	if len(chapters) > 1 {
		rows = min(len(chapters), 10) + 2 // and the box
	}
	ss.Side.ResizeItem(ss.Chapters, rows, 0)
	ss.selectChapter(start)
}

func (ss *StudyScreen) selectChapter(i int) {
	if i < 0 || i >= len(ss.chapters) {
		return
	}
	c := ss.chapters[i]
	ss.ch, ss.white = c, c.white
	ss.Chapters.SetCurrentItem(i)
	width := len(fmt.Sprint(len(ss.chapters)))
	for j := range ss.chapters { // numbered, and the one showing marked
		mark := "  "
		if j == i {
			mark = "▸ "
		}
		ss.Chapters.GetItem(j).SetMainText(mark + ss.chapterLabel(j, width))
	}
	ss.chapterTitle(fmt.Sprintf("Chapter %d of %d", i+1, len(ss.chapters)))
	info := "[::b]" + ss.chapterLabel(i, 0) + "[::-]"
	if ss.name != "" {
		info = "[#a0a0a0]" + escapeText(ss.name) + "[-]\n" + info
	}
	if tags := chapterTags(c); tags != "" {
		info += "\n[#a0a0a0]" + escapeText(tags) + "[-]"
	}
	ss.Info.SetText(info)
	ss.fitInfo()
	ss.shownAt = nil
	ss.Moves.ScrollToBeginning()
	ss.goTo(c.root)
}

// fitInfo sizes the top box to its text, wrapped to the sidebar's width.
func (ss *StudyScreen) fitInfo() {
	_, _, w, _ := ss.Side.GetInnerRect()
	if w <= 0 {
		w = 40
	}
	rows := 0
	for _, line := range strings.Split(ss.Info.GetText(false), "\n") {
		rows += max(1, len(cv.WordWrap(line, w)))
	}
	ss.infoW = w
	ss.Side.ResizeItem(ss.Info, min(rows, 9)+1, 0) // and a blank line under it
	ss.Info.ScrollToBeginning()
}

// chapterLabel is "12. Fischer - Sherwin: …", the number padded to width.
func (ss *StudyScreen) chapterLabel(i, width int) string {
	title := escapeText(ss.chapters[i].name())
	if len(ss.chapters) == 1 {
		return title
	}
	return fmt.Sprintf("%*d. %s", width, i+1, title)
}

// goTo shows node n: the board gets n's line and steps to it.
func (ss *StudyScreen) goTo(n *studyNode) {
	ss.cur = n
	ss.line = n.line()
	positions := make([]*chess.Position, len(ss.line))
	moves := make([]*chess.Move, 0, len(ss.line)-1)
	for i, x := range ss.line {
		positions[i] = x.pos
		if x.move != nil {
			moves = append(moves, x.move)
		}
	}
	end := ss.line[len(ss.line)-1]
	b := ss.Board
	b.SetPosition(end.pos, ss.white)
	b.SetHistory(nil, nil) // a new line, even if it's as long as the last
	b.SetHistory(positions, moves)
	if end.move != nil {
		b.SetLastMove(end.move.S1(), end.move.S2())
	} else {
		b.SetLastMove(chess.NoSquare, chess.NoSquare)
	}
	b.ShowPly(n.ply)
	top, bottom := playerBar(ss.ch, chess.Black), playerBar(ss.ch, chess.White)
	if !ss.white {
		top, bottom = bottom, top
	}
	b.SetClockBars(top, bottom)
	b.mu.Lock()
	b.shownPly = n.ply // already showing it
	b.mu.Unlock()
	ss.showNode()
}

// showNode highlights the current move and draws its arrows and circles.
func (ss *StudyScreen) showNode() {
	n := ss.cur
	if n == nil || ss.ch == nil {
		return
	}
	if n != ss.shownAt { // redraw the list with this move's comment and side lines tinted
		ss.shownAt = n
		text := studyMoveText(ss.ch, n)
		if ss.ch.problem != "" {
			text += "\n[yellow]Stopped reading part of this chapter: " + escapeText(ss.ch.problem) + "[-]\n"
		}
		ss.Moves.SetText(text)
	}
	ss.syncing = true
	if n.move == nil {
		ss.Moves.Highlight()
	} else {
		ss.Moves.Highlight(fmt.Sprintf("n%d", n.id))
		ss.Moves.ScrollToHighlight()
	}
	ss.syncing = false
	ss.Board.SetShapes(n.shapes)
	if n.move != nil && n != ss.soundAt && Root.nav != nil {
		if name, _ := Root.nav.GetFrontPanel(); name == "study" {
			playSound(moveSound(n.move))
		}
	}
	ss.soundAt = n
	if n.move != nil {
		ss.Board.SetGlyph(n.move.S2(), glyphOf(n.nags))
	} else {
		ss.Board.SetGlyph(chess.NoSquare, 0)
	}
}

// step moves along the line: positive = forward.
func (ss *StudyScreen) step(n int) {
	if ss.cur == nil || len(ss.line) == 0 {
		return
	}
	i := max(0, min(len(ss.line)-1, ss.cur.ply-ss.line[0].ply+n))
	ss.cur = ss.line[i]
	ss.Board.ShowPly(ss.cur.ply - ss.line[0].ply)
	ss.Board.mu.Lock()
	ss.Board.shownPly = ss.cur.ply - ss.line[0].ply
	ss.Board.mu.Unlock()
	ss.showNode()
}

func (ss *StudyScreen) flip() {
	ss.white = !ss.white
	if ss.cur != nil {
		ss.goTo(ss.cur)
	}
}

// loadMyStudies fills a menu with the user's studies (loaded in the
// background), each opening the study.
func loadMyStudies(list *cv.List, back func()) {
	list.Clear()
	list.AddItem(cv.NewListItem("[#a0a0a0]Loading your studies…[-]"))
	name := ""
	if loggedIn() {
		name = Root.User.Name
	}
	go func() {
		var err error
		if name == "" {
			name, err = api.GetUsername()
		}
		var studies []api.StudyInfo
		if err == nil {
			studies, err = api.ListStudies(name)
		}
		Root.App.QueueUpdateDraw(func() {
			list.Clear()
			add := func(text, explain string, f func()) {
				item := cv.NewListItem(text)
				item.SetSecondaryText(explain)
				item.SetSelectedFunc(f)
				list.AddItem(item)
			}
			switch {
			case err != nil:
				add("[red]Couldn't load your studies[-]", err.Error(), back)
			case len(studies) == 0:
				add("No studies yet", "Make one on lichess.org/study", back)
			}
			for _, s := range studies {
				s := s
				add(cv.Escape(s.Name), "Updated "+s.Updated.Format("2 Jan 2006"), func() { openStudy(s.ID, "") })
			}
			item := cv.NewListItem("Back")
			item.SetShortcut('b')
			item.SetSecondaryText("Go back to studies")
			item.SetSelectedFunc(back)
			list.AddItem(item)
		})
	}()
}

// askStudy asks for a study link or ID and opens it.
func askStudy() {
	const q = "Which study?"
	askText(q, "Link or ID: ", func(text string) {
		id, chapter := api.StudyIDFrom(text)
		if !api.ValidStudyID(id) {
			promptError(q, api.ErrBadStudyID.Error())
			return
		}
		openStudy(id, chapter)
	})
}

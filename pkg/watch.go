package pkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// Watching games: Lichess TV channels, a player's current game, or a game by
// link/ID. The board follows the game live; history stepping and drawing
// work as in games.

type WatchScreen struct {
	Board    *BoardView
	Side     *cv.Flex
	Top      *cv.TextView // player at the top of the board, and their clock
	TopClock *cv.TextView // clocks as text, when the clock bar is off
	BotClock *cv.TextView
	Status   *cv.TextView
	Result   *cv.TextView // the result, right beside the board, once it's over
	History  *cv.TextView
	Bottom   *cv.TextView
	List     *cv.List

	cancel   context.CancelFunc
	gen      int
	title    string // what's being watched
	tv       string // TV channel key, if watching TV
	game     *api.WatchGame
	chess    *chess.Game // the game as followed (history from when we joined, or all of it)
	white    bool        // white at the bottom
	clockW   int         // seconds, as of clockAt; -1 = unknown
	clockB   int
	clockAt  time.Time
	clockMax int    // seconds each side started with (or the most seen, until known)
	tc       string // time control, e.g. "3+2", once known
	over     string
}

func (ws *WatchScreen) Init() cv.Primitive {
	ws.Board = NewBoardView(func(chess.Square) {}) // clicks: drawing and history only
	ws.Board.EnableSounds()
	ws.Top = textBox(cv.AlignLeft)
	ws.TopClock = textBox(cv.AlignLeft)
	ws.BotClock = textBox(cv.AlignLeft)
	ws.Status = textBox(cv.AlignLeft)
	ws.Result = textBox(cv.AlignRight)
	ws.History = textBox(cv.AlignLeft)
	linkMoveList(ws.History, ws.Board)
	ws.Bottom = textBox(cv.AlignLeft)
	ws.List = menuList()
	for _, it := range []struct {
		text string
		f    func()
	}{{"Flip board", ws.flip}, {"Leave", ws.leave}} {
		item := cv.NewListItem(it.text)
		item.SetSelectedFunc(it.f)
		ws.List.AddItem(item)
	}

	side := cv.NewFlex()
	side.SetDirection(cv.FlexRow)
	side.AddItem(ws.Top, 1, 0, false)
	side.AddItem(ws.TopClock, 0, 0, false)
	side.AddItem(ws.Status, 0, 1, false)
	side.AddItem(ws.Result, 0, 0, false)
	side.AddItem(ws.History, 0, 2, false)
	side.AddItem(ws.List, 2, 0, false)
	side.AddItem(ws.BotClock, 0, 0, false)
	side.AddItem(ws.Bottom, 1, 0, false)
	ws.Side = side

	ws.Board.SetTabFunc(func() { Root.App.SetFocus(ws.List) })
	ws.List.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		if event.Key() == tc.KeyTab || event.Key() == tc.KeyEscape {
			Root.App.SetFocus(ws.Board)
			return nil
		}
		return event
	})
	go ws.ticker()
	return newGameLayout(ws.Board, side, " chess-cli · watching │ scroll or [ ] to step through moves │ right-drag to draw │ f flip │ Tab menu │ Ctrl-C quit")
}

// start stops whatever was being watched and resets the screen.
func (ws *WatchScreen) start(title string) context.Context {
	if ws.cancel != nil {
		ws.cancel()
	}
	ws.gen++
	ctx, cancel := context.WithCancel(context.Background())
	ws.cancel = cancel
	ws.title, ws.tv, ws.game, ws.chess, ws.over = title, "", nil, nil, ""
	ws.clockW, ws.clockB, ws.clockMax, ws.tc = -1, -1, 0, ""
	ws.Board.HideClockBars()
	ws.white = true
	ws.Top.SetText("")
	ws.Bottom.SetText("")
	ws.History.SetText("")
	ws.showResult(chess.NoColor, "")
	ws.Board.SetPosition(chess.StartingPosition(), true)
	ws.Board.SetHistory(nil, nil)
	ws.Board.SetLastMove(chess.NoSquare, chess.NoSquare)
	ws.setStatus("[#a0a0a0]Connecting…[-]")
	Root.Switch("watch")
	return ctx
}

func (ws *WatchScreen) leave() {
	if ws.cancel != nil {
		ws.cancel()
		ws.cancel = nil
	}
	gotoWelcome()
}

func (ws *WatchScreen) flip() {
	ws.white = !ws.white
	ws.redraw()
}

// WatchTv follows a TV channel; when its game ends, TV moves on to the next.
func WatchTv(key, name string) {
	ws := Root.watch
	ctx := ws.start("Lichess TV · " + name)
	ws.tv = key
	gen := ws.gen
	go func() {
		for ctx.Err() == nil {
			err := api.StreamTv(ctx, key, func(u api.WatchUpdate) { ws.post(gen, u) })
			if ctx.Err() != nil {
				return
			}
			msg := "[yellow]Connection lost, reconnecting…[-]"
			if err != nil {
				msg = fmt.Sprintf("[yellow]%v; reconnecting…[-]", err)
			}
			Root.App.QueueUpdateDraw(func() {
				if gen == ws.gen {
					ws.setStatus(msg)
				}
			})
			select {
			case <-ctx.Done():
			case <-time.After(3 * time.Second):
			}
		}
	}()
}

// WatchGameByID follows one game (from a link or ID, or a player's current
// game when player is set).
func WatchGameByID(id, player string) {
	ws := Root.watch
	title := "Game " + id
	if player != "" {
		title = player + "'s game"
	}
	ctx := ws.start(title)
	gen := ws.gen
	go func() {
		var g *api.WatchGame
		var err error
		if player != "" {
			g, err = api.GetCurrentGame(player)
		} else {
			g, err = api.GetGame(id)
		}
		if err != nil {
			Root.App.QueueUpdateDraw(func() {
				if gen == ws.gen {
					ws.setStatus(fmt.Sprintf("[red]Couldn't find the game: %v[-]", err))
				}
			})
			return
		}
		ws.post(gen, api.WatchUpdate{Game: g, WhiteClock: -1, BlackClock: -1})
		err = api.StreamGame(ctx, g.ID, func(u api.WatchUpdate) { ws.post(gen, u) })
		if ctx.Err() == nil && err != nil {
			Root.App.QueueUpdateDraw(func() {
				if gen == ws.gen && ws.over == "" {
					ws.setStatus(fmt.Sprintf("[yellow]Stream ended: %v[-]", err))
				}
			})
		}
	}()
}

// post hands an update to the UI goroutine, unless something else is being
// watched by then.
func (ws *WatchScreen) post(gen int, u api.WatchUpdate) {
	Root.App.QueueUpdateDraw(func() {
		if gen == ws.gen {
			ws.apply(u)
		}
	})
}

func (ws *WatchScreen) apply(u api.WatchUpdate) {
	if u.Game != nil {
		ws.newGame(u.Game)
	}
	if u.FEN != "" {
		ws.follow(u.FEN, u.LastMove)
	}
	if u.WhiteClock >= 0 || u.BlackClock >= 0 {
		ws.clockW, ws.clockB, ws.clockAt = u.WhiteClock, u.BlackClock, time.Now()
		ws.clockMax = max(ws.clockMax, u.WhiteClock, u.BlackClock)
	}
	if u.Status != "" {
		ws.over = gameOverText(u.Status, u.Winner)
	}
	ws.updateInfo()
	ws.redraw()
	// the result: from Lichess, or seen on the board (TV doesn't send results)
	switch {
	case u.Status != "" && u.Status != "aborted" && u.Status != "noStart":
		ws.showResult(colorFromName(u.Winner), u.Status)
	case ws.chess != nil && ws.chess.Outcome() != chess.NoOutcome && ws.over == "":
		winner, method := winnerOf(ws.chess.Outcome()), gameMethod(ws.chess)
		ws.over = strings.SplitN(strings.TrimPrefix(resultText(winner, method), "[::b]"), "\n", 2)[1]
		ws.updateInfo()
		ws.showResult(winner, method)
	case u.Game != nil && ws.over == "":
		ws.showResult(chess.NoColor, "") // a new game
	}
}

// showResult puts the result beside the board and on the kings; method ""
// clears it.
func (ws *WatchScreen) showResult(winner chess.Color, method string) {
	if method == "" {
		ws.Result.SetText("")
		ws.Side.ResizeItem(ws.Result, 0, 0)
		return
	}
	ws.Result.SetText(resultText(winner, method))
	ws.Side.ResizeItem(ws.Result, 3, 0)
	ws.Board.SetResult(winner, method)
}

// newGame starts following a game, replaying its moves if they're known.
func (ws *WatchScreen) newGame(g *api.WatchGame) {
	same := ws.game != nil && ws.game.ID == g.ID
	if !same {
		ws.clockMax = 0
	}
	ws.game, ws.over = g, ""
	if !same {
		ws.tc = ""
	}
	if g.ClockInitial > 0 {
		ws.clockMax = g.ClockInitial
	}
	if g.ClockInitial > 0 || g.ClockIncrement > 0 || g.DaysPerTurn > 0 || g.Speed != "" {
		ws.tc = timeControlText(g.ClockInitial, g.ClockIncrement, g.Speed, g.DaysPerTurn)
	}
	if same && ws.chess != nil {
		return
	}
	ws.white = g.Orientation != "black"
	ws.chess = nil
	if g.Moves != "" || g.InitialFEN != "" {
		game, err := gameFromSAN(g.InitialFEN, g.Moves)
		if err == nil {
			ws.chess = game
		}
	} else if ws.tv != "" {
		go ws.fetchEarlierMoves(ws.gen, g.ID)
	}
}

// fetchEarlierMoves fills in a TV game's moves from before we joined. TV
// only sends positions, and Lichess's export of an ongoing game runs a few
// moves behind, so it retries until the export reaches a position we've
// seen live, then joins the two.
func (ws *WatchScreen) fetchEarlierMoves(gen int, id string) {
	for try := 0; try < 18; try++ {
		if try > 0 {
			time.Sleep(10 * time.Second)
		}
		g, err := api.GetGame(id)
		if err != nil {
			continue
		}
		earlier, err := gameFromSAN(g.InitialFEN, g.Moves)
		if err != nil {
			return
		}
		if g.ClockInitial > 0 {
			Root.App.QueueUpdateDraw(func() {
				if gen == ws.gen && ws.game != nil && ws.game.ID == id {
					ws.clockMax = g.ClockInitial // for the clock bars
					ws.tc = timeControlText(g.ClockInitial, g.ClockIncrement, g.Speed, g.DaysPerTurn)
					ws.updateInfo()
				}
			})
		}
		done := make(chan bool, 1)
		Root.App.QueueUpdateDraw(func() {
			if gen != ws.gen || ws.game == nil || ws.game.ID != id {
				done <- true // watching something else now
				return
			}
			if ws.chess == nil {
				done <- false
				return
			}
			if merged := joinGames(earlier, ws.chess); merged != nil {
				ws.chess = merged
				ws.redraw()
				done <- true
				return
			}
			done <- false
		})
		if <-done {
			return
		}
	}
}

// joinGames combines an earlier record of a game (from the start) with the
// game as followed live (from when we joined): it finds the latest live
// position the earlier record reaches and continues from there with the live
// moves. Returns nil if they don't meet.
func joinGames(earlier, live *chess.Game) *chess.Game {
	key := func(p *chess.Position) string {
		f := strings.Fields(p.String())
		return f[0] + " " + f[1]
	}
	at := map[string]int{}
	for i, p := range earlier.Positions() {
		at[key(p)] = i
	}
	livePos, liveMoves := live.Positions(), live.Moves()
	for j := len(livePos) - 1; j >= 0; j-- {
		k, ok := at[key(livePos[j])]
		if !ok {
			continue
		}
		// replay the earlier record up to k, then the live moves after j
		game := chess.NewGame()
		if start := earlier.Positions()[0]; key(start) != key(chess.StartingPosition()) {
			if fen, err := chess.FEN(start.String()); err == nil {
				game = chess.NewGame(fen)
			}
		}
		for _, m := range earlier.Moves()[:k] {
			if game.Move(m) != nil {
				return nil
			}
		}
		for i, m := range liveMoves[j:] {
			uci := chess.UCINotation{}.Encode(livePos[j+i], m)
			next, err := chess.UCINotation{}.Decode(game.Position(), uci)
			if err != nil || game.Move(next) != nil {
				return nil
			}
		}
		return game
	}
	return nil
}

// gameFromSAN replays SAN moves from a start position ("" or "startpos" for
// the usual one).
func gameFromSAN(initialFEN, moves string) (*chess.Game, error) {
	game := chess.NewGame()
	if initialFEN != "" && initialFEN != "startpos" {
		fen, err := chess.FEN(initialFEN)
		if err != nil {
			return nil, err
		}
		game = chess.NewGame(fen)
	}
	for _, san := range strings.Fields(moves) {
		if err := game.MoveStr(san); err != nil {
			return nil, err
		}
	}
	return game, nil
}

// follow moves the game on to a new position: by playing the last move if
// it leads there, otherwise starting over from the position.
func (ws *WatchScreen) follow(fen, lm string) {
	full := completeFEN(fen, lm)
	if ws.chess != nil {
		if samePlacement(ws.chess.Position(), full) {
			return
		}
		if lm != "" {
			if m, err := (chess.UCINotation{}).Decode(ws.chess.Position(), lm); err == nil {
				next := ws.chess.Clone()
				if next.Move(m) == nil && samePlacement(next.Position(), full) {
					ws.chess = next
					return
				}
			}
		}
	}
	f, err := chess.FEN(full)
	if err != nil {
		ws.setStatus("[red]Can't show this position (variant?)[-]")
		return
	}
	ws.chess = chess.NewGame(f)
}

// samePlacement compares the piece placement of a position and a FEN.
func samePlacement(pos *chess.Position, fen string) bool {
	a, b := strings.Fields(pos.String()), strings.Fields(fen)
	return len(a) > 0 && len(b) > 0 && a[0] == b[0]
}

// completeFEN fills in a FEN that has only the board (and maybe the side
// to move): the side to move is the one that didn't just move the piece now
// on the last move's destination square.
func completeFEN(fen, lm string) string {
	f := strings.Fields(fen)
	if len(f) >= 6 {
		return fen
	}
	if len(f) == 2 {
		return f[0] + " " + f[1] + " - - 0 1"
	}
	turn := "w"
	if len(lm) >= 4 {
		if pos, err := chess.FEN(f[0] + " w - - 0 1"); err == nil {
			game := chess.NewGame(pos)
			dest := chess.Square(int(lm[3]-'1')*8 + int(lm[2]-'a'))
			if p := game.Position().Board().Piece(dest); p != chess.NoPiece && p.Color() == chess.White {
				turn = "b"
			}
		}
	}
	return f[0] + " " + turn + " - - 0 1"
}

func gameOverText(status, winner string) string {
	text := map[string]string{
		"mate": "Checkmate", "resign": "Resignation", "stalemate": "Stalemate", "timeout": "Time out",
		"draw": "Draw", "outoftime": "Time out", "aborted": "Aborted", "noStart": "Didn't start",
		"cheat": "Cheat detected", "variantEnd": "Variant ending",
	}[status]
	if text == "" {
		text = status
	}
	switch winner {
	case "white":
		text += " · White wins"
	case "black":
		text += " · Black wins"
	}
	return text
}

func (ws *WatchScreen) setStatus(s string) { ws.Status.SetText(s) }

func (ws *WatchScreen) updateInfo() {
	g := ws.game
	if g == nil {
		return
	}
	if zenMode() { // just the result, if any
		status := ""
		if ws.over != "" {
			status = "[yellow]" + ws.over + "[-]"
		}
		ws.setStatus(status)
		ws.updatePlayers()
		return
	}
	status := "[::b]" + cv.Escape(ws.title) + "[::-]\n"
	if ws.tc != "" && !clockBarsOn() {
		status += ws.tc + " · "
	}
	if g.Speed != "" {
		rated := "casual"
		if g.Rated {
			rated = "rated"
		}
		status += fmt.Sprintf("%s · %s · ", caser.String(g.Speed), rated)
	}
	status += "[#a0a0a0]lichess.org/" + g.ID + "[-]\n"
	if ws.tv == "" {
		status += "[#a0a0a0]Lichess shows ongoing games 3 moves behind.[-]\n"
	}
	if ws.over != "" {
		status += "\n[yellow]" + ws.over + "[-]"
		if ws.tv != "" {
			status += "\n[#a0a0a0]TV will switch to the next game.[-]"
		}
	}
	ws.setStatus(status)
	ws.updatePlayers()
}

func (ws *WatchScreen) updatePlayers() {
	g := ws.game
	if g == nil {
		return
	}
	name := func(p api.WatchPlayer, c string) string {
		if zenMode() {
			if c == "[::b]○[::-] " {
				return c + "White"
			}
			return c + "Black"
		}
		s := cv.Escape(p.Name)
		if p.Title != "" {
			s = "[yellow]" + p.Title + "[-] " + s
		}
		if p.Rating > 0 {
			s += fmt.Sprintf(" [#a0a0a0](%d)[-]", p.Rating)
		}
		return c + s
	}
	white, black := name(g.White, "[::b]○[::-] "), name(g.Black, "[::b]●[::-] ")
	if ws.white {
		ws.Top.SetText(black)
		ws.Bottom.SetText(white)
	} else {
		ws.Top.SetText(white)
		ws.Bottom.SetText(black)
	}
	ws.updateClocks()
}

// updateClocks shows the clocks beside the board (time and bar), counting
// down the side to move.
func (ws *WatchScreen) updateClocks() {
	if ws.clockW < 0 && ws.clockB < 0 {
		ws.Board.HideClockBars()
		return
	}
	w, b := ws.clockW, ws.clockB
	// clocks run once both sides have moved (joining mid-game, they have)
	running := ws.over == "" && ws.chess != nil &&
		(len(ws.chess.Moves()) >= 2 || moveNumber(ws.chess.Positions()[0]) > 1)
	whiteToMove := ws.chess != nil && ws.chess.Position().Turn() == chess.White
	if running {
		gone := int(time.Since(ws.clockAt).Seconds())
		if whiteToMove {
			w = max(0, w-gone)
		} else {
			b = max(0, b-gone)
		}
	}
	frac := func(s int) float64 {
		if ws.clockMax <= 0 {
			return -1
		}
		return float64(s) / float64(ws.clockMax)
	}
	wc := clockBar{label: clockLabel(w), frac: frac(w), active: running && whiteToMove}
	bc := clockBar{label: clockLabel(b), frac: frac(b), active: running && !whiteToMove}
	top, bottom := bc, wc
	if !ws.white {
		top, bottom = wc, bc
	}
	if !clockBarsOn() { // plain text in the sidebar instead
		ws.Board.HideClockBars()
		text := func(c clockBar) string {
			if c.active {
				return "[black:#c8c8c8] " + c.label + " [-:-]"
			}
			return " " + c.label
		}
		ws.TopClock.SetText(text(top))
		ws.BotClock.SetText(text(bottom))
		return
	}
	top.right, bottom.right = ws.tc, ws.tc
	ws.Board.SetClockBars(top, bottom)
}

func (ws *WatchScreen) redraw() {
	if ws.chess == nil {
		return
	}
	ws.Board.SetPosition(ws.chess.Position(), ws.white)
	ws.Board.SetHistory(ws.chess.Positions(), ws.chess.Moves())
	from, to := chess.NoSquare, chess.NoSquare
	if moves := ws.chess.Moves(); len(moves) > 0 {
		last := moves[len(moves)-1]
		from, to = last.S1(), last.S2()
	}
	ws.Board.SetLastMove(from, to)
	ws.History.SetText(moveListText(ws.chess))
	ws.History.ScrollToEnd()
	ws.updatePlayers()
}

// ticker redraws the running clock every second.
func (ws *WatchScreen) ticker() {
	for range time.Tick(time.Second) {
		if !showing("watch") {
			continue
		}
		Root.App.QueueUpdateDraw(func() {
			if name, _ := Root.nav.GetFrontPanel(); name == "watch" {
				ws.updateClocks()
			}
		})
	}
}

func askWatch(mode string) {
	if mode == "player" {
		askText("Whose game?", "Lichess username: ", func(text string) { WatchGameByID("", text) })
	} else {
		askText("Which game?", "Link or ID: ", func(text string) { WatchGameByID(api.GameIDFrom(text), "") })
	}
}

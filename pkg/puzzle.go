package pkg

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// Lichess puzzles: the daily puzzle, or puzzle training (a new puzzle each
// time, at a chosen difficulty), rated or casual like on Lichess.
//
// A puzzle starts after the opponent's move; the solution alternates the
// player's moves (which must be found) with the opponent's replies (played
// automatically). As on Lichess, any move that checkmates also counts, and
// the first wrong move fails the puzzle, though you can keep trying.

type PuzzleScreen struct {
	Board   *BoardView
	Side    *cv.Flex
	Info    *cv.TextView
	Status  *cv.TextView
	History *cv.TextView
	List    *cv.List
	Move    *moveBox

	p        *api.Puzzle
	daily    bool
	game     *chess.Game
	color    chess.Color // the side the player plays
	idx      int         // next solution move
	gen      int         // bumped per puzzle, so stale timers do nothing
	failed   bool
	done     bool // solved, or solution shown
	recorded string
	loading  bool
}

func puzzleRated() bool { return !api.UserInfo.PuzzleCasual }

func puzzleDifficulty() string {
	if d := api.UserInfo.PuzzleDifficulty; d != "" {
		return d
	}
	return "normal"
}

func savePuzzlePrefs(casual bool, difficulty string) {
	_ = api.UpdateConfig(func(c *api.UserConfig) { c.PuzzleCasual, c.PuzzleDifficulty = casual, difficulty })
}

func (ps *PuzzleScreen) Init() cv.Primitive {
	ps.Board = NewBoardView(ps.onSquare)
	ps.Board.EnableSounds()
	ps.Info = textBox(cv.AlignLeft)
	ps.Status = textBox(cv.AlignLeft)
	ps.History = textBox(cv.AlignLeft)
	ps.List = menuList()
	linkMoveList(ps.History, ps.Board)

	side := cv.NewFlex()
	side.SetDirection(cv.FlexRow)
	ps.Side = side
	ps.Move = newMoveBox(side, ps.Board, func(text string) {
		if ps.done || ps.game == nil || ps.game.Position().Turn() != ps.color {
			return
		}
		uci, err := parseMoveInput(ps.game, text)
		if err != nil {
			ps.setStatus(invalidMoveHint(text))
			ps.Move.show()
			return
		}
		ps.tryMove(uci)
	})
	side.AddItem(ps.Info, 5, 0, false)
	side.AddItem(ps.Status, 0, 1, false)
	side.AddItem(ps.History, 0, 1, false)
	side.AddItem(ps.Move.input, 0, 0, false)
	side.AddItem(ps.List, 4, 0, false)

	ps.Board.SetTabFunc(func() { Root.App.SetFocus(ps.List) })
	ps.List.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		if event.Key() == tc.KeyTab || event.Key() == tc.KeyEscape {
			Root.App.SetFocus(ps.Board)
			return nil
		}
		return event
	})
	registerMenus("puzzle", false, ps.List)
	return newGameLayout(ps.Board, side, " chess-cli · puzzles │ Ctrl-K type a move │ click or arrows/hjkl + Enter to move │ Tab menu │ Ctrl-C quit")
}

// startPuzzle opens the puzzle screen and loads a puzzle. Rated puzzles need
// a Lichess login, which is started first if there's no saved one.
func startPuzzle(daily bool) {
	ps := Root.puzzle
	ps.gen++
	ps.p, ps.game, ps.daily, ps.loading = nil, nil, daily, true
	ps.Info.SetText("")
	ps.History.SetText("")
	ps.Board.ClearSelection()
	ps.updateMenu()
	Root.Switch("puzzle")

	if puzzleRated() && !api.HasValidToken() {
		ps.showLogin()
		return
	}
	ps.fetch()
}

func (ps *PuzzleScreen) showLogin() {
	ps.setStatus("[yellow]Rated puzzles need a Lichess login. Logging in…[-]")
	startLogin(func() {
		ps.setStatus("[yellow]Rated puzzles need a Lichess login.[-]\nLog in through your browser; it should open by itself. If it didn't, open this link:\n" + cv.Escape(loginURL))
	}, func(err error) {
		if err != nil {
			ps.setStatus(fmt.Sprintf("[red]Login failed:[-] %v\nChoose Casual mode on the Puzzles menu to play without logging in.", err))
			return
		}
		ps.fetch()
	})
}

func (ps *PuzzleScreen) fetch() {
	what := "puzzle"
	if ps.daily {
		what = "daily puzzle"
	}
	ps.setStatus(fmt.Sprintf("[#a0a0a0]Loading the %s…[-]", what))
	gen, daily, difficulty := ps.gen, ps.daily, puzzleDifficulty()
	go func() {
		var p *api.Puzzle
		var err error
		if daily {
			p, err = api.GetDailyPuzzle()
		} else {
			p, err = api.GetNextPuzzle(difficulty)
		}
		Root.App.QueueUpdateDraw(func() {
			if gen != ps.gen {
				return
			}
			ps.loading = false
			if err == nil {
				err = ps.load(p)
			}
			if err != nil {
				ps.setStatus(fmt.Sprintf("[red]Couldn't load the %s: %v[-]", what, err))
				ps.updateMenu()
			}
		})
	}()
}

// load sets up a fetched puzzle: replays the game up to the puzzle position.
func (ps *PuzzleScreen) load(p *api.Puzzle) error {
	game, err := puzzleGame(p)
	if err != nil {
		return err
	}
	if len(p.Puzzle.Solution) == 0 {
		return errors.New("the puzzle has no solution")
	}
	ps.p, ps.game = p, game
	ps.color = game.Position().Turn()
	ps.idx, ps.failed, ps.done, ps.recorded = 0, false, false, ""
	ps.updateInfo()
	ps.setStatus(fmt.Sprintf("Your turn: find the best move for %s.", colorName(ps.color)))
	ps.updateMenu()
	ps.redraw()
	Root.App.SetFocus(ps.Board)
	return nil
}

// puzzleGame rebuilds the puzzle's game from its PGN (which ends with the
// opponent's move), falling back to the FEN if Lichess sent one.
func puzzleGame(p *api.Puzzle) (*chess.Game, error) {
	game := chess.NewGame()
	ok := p.Game.PGN != ""
	for _, san := range strings.Fields(p.Game.PGN) {
		if err := game.MoveStr(san); err != nil {
			ok = false
			break
		}
	}
	if ok {
		return game, nil
	}
	if p.Puzzle.FEN != "" {
		fen, err := chess.FEN(p.Puzzle.FEN)
		if err != nil {
			return nil, err
		}
		return chess.NewGame(fen), nil
	}
	return nil, errors.New("couldn't read the puzzle's moves")
}

func colorName(c chess.Color) string {
	if c == chess.White {
		return "White"
	}
	return "Black"
}

func (ps *PuzzleScreen) setStatus(s string) { ps.Status.SetText(s) }

func (ps *PuzzleScreen) updateInfo() {
	p := ps.p
	title := "Puzzle"
	if ps.daily {
		title = "Daily puzzle"
	}
	mode := "Casual"
	if puzzleRated() {
		mode = "Rated"
	}
	text := fmt.Sprintf("[::b]%s[::-] [#a0a0a0]#%s[-]\nRating [::b]%d[::-] · played %d times\n%s", title, p.Puzzle.ID, p.Puzzle.Rating, p.Puzzle.Plays, mode)
	if !ps.daily {
		text += " · " + puzzleDifficulty()
	}
	if ps.recorded != "" {
		text += "\n" + ps.recorded
	}
	ps.Info.SetText(text)
}

func (ps *PuzzleScreen) updateMenu() {
	l := ps.List
	l.Clear()
	add := func(text string, key rune, f func()) {
		item := cv.NewListItem(text)
		item.SetShortcut(key)
		item.SetSelectedFunc(f)
		l.AddItem(item)
	}
	if ps.p != nil && !ps.done {
		add("Hint", 'i', ps.hint)
		add("View solution", 's', ps.reveal)
	}
	if ps.p != nil && ps.done || ps.p == nil && !ps.loading {
		add("Next puzzle", 'n', func() { startPuzzle(false) })
	}
	add("Leave puzzles", 'x', gotoWelcome)
	ps.Side.ResizeItem(l, l.GetItemCount(), 0)
}

func (ps *PuzzleScreen) redraw() {
	if ps.game == nil {
		return
	}
	ps.Board.SetPosition(ps.game.Position(), ps.color == chess.White)
	ps.Board.SetHistory(ps.game.Positions(), ps.game.Moves())
	from, to := chess.NoSquare, chess.NoSquare
	if moves := ps.game.Moves(); len(moves) > 0 {
		last := moves[len(moves)-1]
		from, to = last.S1(), last.S2()
	}
	ps.Board.SetLastMove(from, to)
	ps.History.SetText(moveListText(ps.game))
	ps.History.ScrollToEnd()
}

func (ps *PuzzleScreen) onSquare(sq chess.Square) {
	if ps.game == nil {
		return
	}
	canMove := !ps.done && ps.game.Position().Turn() == ps.color
	ps.Board.HandleSquare(ps.game, sq, canMove, ps.tryMove)
}

// play makes a UCI move on the puzzle game.
func (ps *PuzzleScreen) play(uci string) error {
	m, err := chess.UCINotation{}.Decode(ps.game.Position(), uci)
	if err != nil {
		return err
	}
	return ps.game.Move(m)
}

func (ps *PuzzleScreen) tryMove(uci string) {
	if ps.done || ps.game == nil {
		return
	}
	before := ps.game.Clone()
	if err := ps.play(uci); err != nil {
		ps.setStatus(fmt.Sprintf("[red]%v[-]", err))
		return
	}
	correct := uci == ps.p.Puzzle.Solution[ps.idx] || ps.game.Method() == chess.Checkmate
	if !correct {
		ps.game = before
		ps.redraw()
		ps.setStatus("[red]✗ That's not the move.[-] Try again, or view the solution.")
		playSound(soundError)
		if !ps.failed {
			ps.failed = true
			ps.record(false)
		}
		return
	}
	ps.idx++
	ps.redraw()
	if ps.idx >= len(ps.p.Puzzle.Solution) || ps.game.Outcome() != chess.NoOutcome {
		ps.finish()
		return
	}
	ps.setStatus("[green]✓ Best move![-] Keep going.")
	playSound(soundConfirm)
	ps.later(500*time.Millisecond, func() {
		if ps.idx < len(ps.p.Puzzle.Solution) && ps.play(ps.p.Puzzle.Solution[ps.idx]) == nil {
			ps.idx++
			ps.redraw()
			ps.setStatus(fmt.Sprintf("[green]✓ Best move![-] Your turn again: find the best move for %s.", colorName(ps.color)))
		}
	})
}

// later runs f on the UI goroutine after d, unless a new puzzle started.
func (ps *PuzzleScreen) later(d time.Duration, f func()) {
	gen := ps.gen
	time.AfterFunc(d, func() {
		Root.App.QueueUpdateDraw(func() {
			if gen == ps.gen {
				f()
			}
		})
	})
}

func (ps *PuzzleScreen) finish() {
	ps.done = true
	if ps.failed {
		ps.setStatus("Puzzle complete, [yellow]but not solved first time.[-]" + ps.themes())
	} else {
		ps.setStatus("[green]✓ Solved![-]" + ps.themes())
		playSound(soundVictory)
		ps.record(true)
	}
	ps.updateMenu()
}

func (ps *PuzzleScreen) themes() string {
	if len(ps.p.Puzzle.Themes) == 0 {
		return ""
	}
	return "\n\n[#a0a0a0]Themes: " + cv.Escape(strings.Join(ps.p.Puzzle.Themes, ", ")) + "[-]\n[#a0a0a0]From game lichess.org/" + cv.Escape(ps.p.Game.ID) + "[-]"
}

// hint selects the piece to move, as Lichess's hint does.
func (ps *PuzzleScreen) hint() {
	if ps.done || ps.game.Position().Turn() != ps.color {
		return
	}
	m, err := chess.UCINotation{}.Decode(ps.game.Position(), ps.p.Puzzle.Solution[ps.idx])
	if err != nil {
		return
	}
	targets := map[chess.Square]bool{}
	for _, v := range ps.game.ValidMoves() {
		if v.S1() == m.S1() {
			targets[v.S2()] = true
		}
	}
	ps.Board.SetSelection(m.S1(), targets)
	Root.App.SetFocus(ps.Board)
}

// reveal plays the rest of the solution, one move at a time. It counts as
// a failed puzzle.
func (ps *PuzzleScreen) reveal() {
	if ps.done {
		return
	}
	if !ps.failed {
		ps.failed = true
		ps.record(false)
	}
	ps.done = true
	ps.Board.ClearSelection()
	ps.setStatus("Solution:")
	ps.updateMenu()
	var step func()
	step = func() {
		if ps.idx >= len(ps.p.Puzzle.Solution) || ps.play(ps.p.Puzzle.Solution[ps.idx]) != nil {
			ps.setStatus("That was the solution." + ps.themes())
			return
		}
		ps.idx++
		ps.redraw()
		ps.later(700*time.Millisecond, step)
	}
	ps.later(300*time.Millisecond, step)
}

// record sends the result to Lichess (if logged in), once per puzzle.
func (ps *PuzzleScreen) record(win bool) {
	if !api.HasValidToken() {
		ps.recorded = "[#a0a0a0]Not logged in: result not saved[-]"
		ps.updateInfo()
		return
	}
	id, rated, gen := ps.p.Puzzle.ID, puzzleRated(), ps.gen
	go func() {
		res, err := api.SubmitPuzzle(id, win, rated)
		Root.App.QueueUpdateDraw(func() {
			if gen != ps.gen {
				return
			}
			var apiErr *api.APIError
			switch {
			case errors.As(err, &apiErr) && (apiErr.Status == http.StatusUnauthorized || apiErr.Status == http.StatusForbidden):
				ps.recorded = "[yellow]Result not saved: log in again to allow puzzles[-]"
				ps.addRelogin()
			case err != nil:
				ps.recorded = fmt.Sprintf("[red]Result not saved: %v[-]", err)
			case rated && len(res.Rounds) > 0:
				diff := res.Rounds[0].RatingDiff
				col := "green"
				if diff < 0 {
					col = "red"
				}
				ps.recorded = fmt.Sprintf("Your rating: [::b]%.0f[::-] [%s]%+d[-]", res.Glicko.Rating, col, diff)
			default:
				ps.recorded = "[#a0a0a0]Saved to your Lichess history[-]"
			}
			ps.updateInfo()
		})
	}()
}

// addRelogin offers a fresh login, for tokens from before puzzle support
// (they lack the puzzle:write scope).
func (ps *PuzzleScreen) addRelogin() {
	item := cv.NewListItem("Log in again")
	item.SetSelectedFunc(func() {
		_ = api.ClearToken()
		Root.User = nil
		startPuzzle(ps.daily)
	})
	ps.List.InsertItem(0, item)
	ps.Side.ResizeItem(ps.List, ps.List.GetItemCount(), 0)
}

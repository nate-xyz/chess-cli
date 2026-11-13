package pkg

import (
	"strings"
	"sync"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

// BoardView is a chess board primitive that scales to fill whatever space the
// layout gives it. Squares grow with the terminal; once there is room, pieces
// are drawn as half-block pixel art instead of single glyphs.
//
// Mouse: click a piece, then click a destination. Keyboard (when focused):
// arrows / hjkl move a cursor, Enter or Space selects, Esc clears, Tab leaves.
// Promotions open a picker over the promotion square (click a piece or press
// q/r/b/n; Esc or a click elsewhere cancels).
type BoardView struct {
	*cv.Box

	mu       sync.Mutex
	pos      *chess.Position
	white    bool // true = white at the bottom
	selected chess.Square
	targets  map[chess.Square]bool
	lastFrom chess.Square
	lastTo   chess.Square
	cursor   chess.Square
	onSquare func(sq chess.Square)
	onTab    func()
	promo    *promoPick

	// geometry from the last Draw, used to map mouse clicks to squares
	ox, oy, sqW, sqH int
}

// promoPick is an open promotion picker: four squares in a column over the
// promotion square, each offering one piece.
type promoPick struct {
	from, to chess.Square
	color    chess.Color
	squares  map[chess.Square]chess.PieceType
	done     func(uci string)
}

// makePiece builds a Piece from type and colour (the chess library keeps its
// own constructor unexported).
func makePiece(t chess.PieceType, c chess.Color) chess.Piece {
	for p := chess.WhiteKing; p <= chess.BlackPawn; p++ {
		if p.Type() == t && p.Color() == c {
			return p
		}
	}
	return chess.NoPiece
}

// promoOrder is the order pieces appear in the picker, nearest the edge first.
var promoOrder = []chess.PieceType{chess.Queen, chess.Knight, chess.Rook, chess.Bishop}

// NewBoardView makes a board. onSquare is called when the user clicks or
// presses Enter on a square; pass nil for a display-only board.
func NewBoardView(onSquare func(sq chess.Square)) *BoardView {
	return &BoardView{
		Box:      cv.NewBox(),
		pos:      chess.StartingPosition(),
		white:    true,
		selected: chess.NoSquare,
		lastFrom: chess.NoSquare,
		lastTo:   chess.NoSquare,
		cursor:   chess.E2,
		onSquare: onSquare,
	}
}

// SetPosition sets the position to show and which side sits at the bottom.
// If the position changed, any selection or open promotion picker is dropped,
// since it no longer applies (e.g. the opponent just moved).
func (b *BoardView) SetPosition(pos *chess.Position, white bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.pos == nil || pos.String() != b.pos.String() || white != b.white {
		b.selected, b.targets, b.promo = chess.NoSquare, nil, nil
	}
	b.pos = pos
	b.white = white
}

// PromotionOpen reports whether the promotion picker is showing.
func (b *BoardView) PromotionOpen() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.promo != nil
}

// CancelPromotion closes the promotion picker without moving.
func (b *BoardView) CancelPromotion() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.promo = nil
	b.selected, b.targets = chess.NoSquare, nil
}

// SetLastMove highlights the from/to squares of the previous move.
func (b *BoardView) SetLastMove(from, to chess.Square) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastFrom, b.lastTo = from, to
}

// SetSelection highlights a selected square and the squares it can move to.
func (b *BoardView) SetSelection(sq chess.Square, targets map[chess.Square]bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.selected = sq
	b.targets = targets
}

// Selected returns the currently selected square (or chess.NoSquare).
func (b *BoardView) Selected() chess.Square {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.selected
}

// ClearSelection removes the selection and move targets.
func (b *BoardView) ClearSelection() { b.SetSelection(chess.NoSquare, nil) }

// SetTabFunc sets what happens when Tab is pressed while the board has focus.
func (b *BoardView) SetTabFunc(f func()) { b.onTab = f }

// squareAt maps a display row/col (0,0 = top-left) to a board square.
func squareAt(row, col int, white bool) chess.Square {
	if white {
		return chess.NewSquare(chess.File(col), chess.Rank(7-row))
	}
	return chess.NewSquare(chess.File(7-col), chess.Rank(row))
}

// displayOf is the inverse of squareAt.
func displayOf(sq chess.Square, white bool) (row, col int) {
	if white {
		return 7 - int(sq.Rank()), int(sq.File())
	}
	return int(sq.Rank()), 7 - int(sq.File())
}

const (
	labelW = 2 // columns reserved left of the board for rank labels
	labelH = 1 // rows reserved below the board for file labels
)

// boardGeometry picks the largest square size that fits w x h cells. Terminal
// cells are roughly twice as tall as wide, so squares are ~2x wider than tall.
func boardGeometry(w, h int) (sqW, sqH int) {
	w -= labelW
	h -= labelH
	for s := h / 8; s >= 2; s-- {
		if 8*(2*s+2) <= w {
			return 2*s + 2, s
		}
	}
	if h >= 8 && w >= 8*2 {
		return min(4, w/8), 1
	}
	return 0, 0
}

var (
	colLight     = tc.NewRGBColor(145, 130, 109)
	colDark      = tc.NewRGBColor(108, 81, 59)
	colWhiteBody = tc.NewRGBColor(255, 248, 220)
	colWhiteLine = tc.NewRGBColor(60, 48, 36)
	colBlackBody = tc.NewRGBColor(18, 18, 18)
	colBlackLine = tc.NewRGBColor(120, 110, 95)
	colLabel     = tc.NewRGBColor(170, 160, 140)

	tintSelected = tc.NewRGBColor(110, 150, 70)
	tintLastMove = tc.NewRGBColor(205, 190, 80)
	tintCapture  = tc.NewRGBColor(200, 70, 60)
	tintCursor   = tc.NewRGBColor(80, 130, 200)
	colDot       = tc.NewRGBColor(60, 90, 40)
	colPromo     = tc.NewRGBColor(200, 195, 185)
	colPromoDark = tc.NewRGBColor(95, 90, 85)
)

func blend(a, b tc.Color, t float64) tc.Color {
	ar, ag, ab := a.RGB()
	br, bg, bb := b.RGB()
	mix := func(x, y int32) int32 { return int32(float64(x)*(1-t) + float64(y)*t) }
	return tc.NewRGBColor(mix(ar, br), mix(ag, bg), mix(ab, bb))
}

// Draw renders the board, scaled to the primitive's current size.
func (b *BoardView) Draw(screen tc.Screen) {
	if !b.GetVisible() {
		return
	}
	b.Box.Draw(screen)

	b.mu.Lock()
	defer b.mu.Unlock()

	x, y, w, h := b.GetInnerRect()
	sqW, sqH := boardGeometry(w, h)
	b.sqW, b.sqH = sqW, sqH
	if sqW == 0 {
		msg := "Enlarge the terminal to see the board"
		for i, r := range msg {
			if i < w {
				screen.SetContent(x+i, y+h/2, r, nil, tc.StyleDefault)
			}
		}
		return
	}

	totalW := 8*sqW + labelW
	totalH := 8*sqH + labelH
	b.ox = x + (w-totalW)/2 + labelW
	b.oy = y + (h-totalH)/2

	focused := b.HasFocus()
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			sq := squareAt(row, col, b.white)
			b.drawSquare(screen, sq, b.ox+col*sqW, b.oy+row*sqH, focused && sq == b.cursor)
		}
	}

	labelStyle := tc.StyleDefault.Foreground(colLabel).Background(b.GetBackgroundColor())
	for i := 0; i < 8; i++ {
		rank := squareAt(i, 0, b.white).Rank().String()
		screen.SetContent(b.ox-2, b.oy+i*sqH+sqH/2, rune(rank[0]), nil, labelStyle)
		file := squareAt(0, i, b.white).File().String()
		screen.SetContent(b.ox+i*sqW+sqW/2, b.oy+8*sqH, rune(file[0]), nil, labelStyle)
	}
}

func (b *BoardView) drawSquare(screen tc.Screen, sq chess.Square, x, y int, cursor bool) {
	bg := colDark
	if (int(sq.File())+int(sq.Rank()))%2 == 1 {
		bg = colLight
	}
	piece := b.pos.Board().Piece(sq)
	if b.promo != nil {
		if pt, ok := b.promo.squares[sq]; ok {
			piece = makePiece(pt, b.promo.color)
			bg = colPromo
			if piece.Color() == chess.White {
				bg = colPromoDark
			}
			if cursor {
				bg = blend(bg, tintCursor, 0.5)
			}
			b.paintSquare(screen, x, y, bg, piece, false)
			return
		}
		bg = blend(bg, tc.ColorBlack, 0.55)
	}
	switch {
	case sq == b.selected:
		bg = blend(bg, tintSelected, 0.6)
	case b.targets[sq] && piece != chess.NoPiece:
		bg = blend(bg, tintCapture, 0.5)
	case sq == b.lastFrom || sq == b.lastTo:
		bg = blend(bg, tintLastMove, 0.4)
	}
	if cursor {
		bg = blend(bg, tintCursor, 0.5)
	}
	b.paintSquare(screen, x, y, bg, piece, b.targets[sq])
}

// paintSquare draws one square with the given background and piece (or a
// move-target dot if dot is set and the square is empty).
func (b *BoardView) paintSquare(screen tc.Screen, x, y int, bg tc.Color, piece chess.Piece, dot bool) {
	// pixel buffer: 1 column x 2 half-rows per cell
	pw, ph := b.sqW, b.sqH*2
	px := make([]tc.Color, pw*ph)
	for i := range px {
		px[i] = bg
	}

	glyph := rune(0)
	var glyphFg tc.Color
	if piece != chess.NoPiece {
		body, line := colBlackBody, colBlackLine
		if piece.Color() == chess.White {
			body, line = colWhiteBody, colWhiteLine
		}
		if art, k, ok := pickArt(piece.Type(), pw, ph); ok {
			aw, ah := len(art[0])*k, len(art)*k
			margin := 0
			if ph-ah >= 2 {
				margin = 1
			}
			offX := (pw - aw) / 2
			offY := ph - ah - margin // bottom-aligned so pieces share a baseline
			for ay := 0; ay < ah; ay++ {
				for ax := 0; ax < aw; ax++ {
					switch art[ay/k][ax/k] {
					case '#':
						px[(offY+ay)*pw+offX+ax] = body
					case 'o':
						px[(offY+ay)*pw+offX+ax] = line
					}
				}
			}
		} else {
			glyph = []rune(PiecesMap[rune(piece.Type().String()[0])])[0]
			glyphFg = body
		}
	} else if dot {
		if b.sqH >= 3 {
			// a round-ish dot in the middle, ~1/4 of the square
			dw, dh := max(2, pw/4), max(2, ph/4)
			for dy := 0; dy < dh; dy++ {
				for dx := 0; dx < dw; dx++ {
					if (dx == 0 || dx == dw-1) && (dy == 0 || dy == dh-1) && dw > 2 && dh > 2 {
						continue
					}
					px[(ph/2-dh/2+dy)*pw+pw/2-dw/2+dx] = colDot
				}
			}
		} else {
			glyph, glyphFg = '•', colDot
		}
	}

	for cy := 0; cy < b.sqH; cy++ {
		for cx := 0; cx < pw; cx++ {
			top, bot := px[(2*cy)*pw+cx], px[(2*cy+1)*pw+cx]
			if top == bot {
				screen.SetContent(x+cx, y+cy, ' ', nil, tc.StyleDefault.Background(top))
			} else {
				screen.SetContent(x+cx, y+cy, '▀', nil, tc.StyleDefault.Foreground(top).Background(bot))
			}
		}
	}
	if glyph != 0 {
		screen.SetContent(x+pw/2, y+b.sqH/2, glyph, nil, tc.StyleDefault.Foreground(glyphFg).Background(bg))
	}
}

// pickArt returns the largest piece bitmap (with an integer scale factor) that
// fits in pw x ph pixels. Integer scaling keeps the pixel art crisp.
func pickArt(t chess.PieceType, pw, ph int) ([]string, int, bool) {
	if ph < 6 {
		return nil, 0, false
	}
	availW, availH := pw, ph
	if ph >= 8 {
		availW, availH = pw-2, ph-2
	}
	for _, set := range []map[chess.PieceType][]string{artLarge, artSmall} {
		art := set[t]
		k := min(availW/len(art[0]), availH/len(art))
		if k >= 1 {
			return art, k, true
		}
	}
	return nil, 0, false
}

// InputHandler handles keyboard navigation on the board.
func (b *BoardView) InputHandler() func(event *tc.EventKey, setFocus func(p cv.Primitive)) {
	return b.WrapInputHandler(func(event *tc.EventKey, setFocus func(p cv.Primitive)) {
		b.mu.Lock()
		row, col := displayOf(b.cursor, b.white)
		b.mu.Unlock()
		move := func(dr, dc int) {
			row = min(7, max(0, row+dr))
			col = min(7, max(0, col+dc))
			b.mu.Lock()
			b.cursor = squareAt(row, col, b.white)
			b.mu.Unlock()
		}
		switch event.Key() {
		case tc.KeyUp:
			move(-1, 0)
		case tc.KeyDown:
			move(1, 0)
		case tc.KeyLeft:
			move(0, -1)
		case tc.KeyRight:
			move(0, 1)
		case tc.KeyEnter:
			b.activate(b.cursor)
		case tc.KeyEscape:
			b.CancelPromotion()
		case tc.KeyTab, tc.KeyBacktab:
			if b.onTab != nil {
				b.onTab()
			}
		case tc.KeyRune:
			if sq, ok := b.promoSquareFor(event.Str()); ok {
				b.activate(sq)
				return
			}
			switch event.Str() {
			case "k":
				move(-1, 0)
			case "j":
				move(1, 0)
			case "h":
				move(0, -1)
			case "l":
				move(0, 1)
			case " ":
				b.activate(b.cursor)
			}
		}
	})
}

// MouseHandler maps clicks to squares. It deliberately does not take focus,
// so the move input field keeps keyboard focus while you click around.
func (b *BoardView) MouseHandler() func(action cv.MouseAction, event *tc.EventMouse, setFocus func(p cv.Primitive)) (bool, cv.Primitive) {
	return b.WrapMouseHandler(func(action cv.MouseAction, event *tc.EventMouse, setFocus func(p cv.Primitive)) (bool, cv.Primitive) {
		mx, my := event.Position()
		if !b.InRect(mx, my) {
			return false, nil
		}
		if action != cv.MouseLeftDown {
			return true, nil
		}
		b.mu.Lock()
		ox, oy, sqW, sqH, white := b.ox, b.oy, b.sqW, b.sqH, b.white
		b.mu.Unlock()
		if sqW == 0 || mx < ox || my < oy {
			return true, nil
		}
		col, row := (mx-ox)/sqW, (my-oy)/sqH
		if col > 7 || row > 7 {
			return true, nil
		}
		sq := squareAt(row, col, white)
		b.mu.Lock()
		b.cursor = sq
		b.mu.Unlock()
		b.activate(sq)
		return true, nil
	})
}

// promoSquareFor maps q/r/b/n (either case) to the picker square offering that piece.
func (b *BoardView) promoSquareFor(key string) (chess.Square, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.promo == nil {
		return chess.NoSquare, false
	}
	want, ok := map[string]chess.PieceType{"q": chess.Queen, "r": chess.Rook, "b": chess.Bishop, "n": chess.Knight}[strings.ToLower(key)]
	if !ok {
		return chess.NoSquare, false
	}
	for sq, pt := range b.promo.squares {
		if pt == want {
			return sq, true
		}
	}
	return chess.NoSquare, false
}

func (b *BoardView) activate(sq chess.Square) {
	if b.onSquare != nil {
		b.onSquare(sq)
	}
}

// HandleSquare implements click-to-move for a game: select one of the mover's
// pieces, then pick a legal destination. doMove is given a UCI move string.
// Moving a pawn to the last rank opens the promotion picker; the move is made
// once a piece is chosen. Returns true if a move was attempted.
func (b *BoardView) HandleSquare(game *chess.Game, sq chess.Square, canMove bool, doMove func(uci string)) bool {
	b.mu.Lock()
	promo := b.promo
	b.mu.Unlock()
	if promo != nil {
		b.CancelPromotion()
		if pt, ok := promo.squares[sq]; ok {
			promo.done(promo.from.String() + promo.to.String() + pt.String())
			return true
		}
		return false // clicked elsewhere: cancel
	}

	pos := game.Position()
	selected := b.Selected()

	if selected != chess.NoSquare && selected != sq {
		var matches []*chess.Move
		for _, m := range game.ValidMoves() {
			if m.S1() == selected && m.S2() == sq {
				matches = append(matches, m)
			}
		}
		switch {
		case len(matches) == 0:
			// not a legal destination: fall through to (re)selection below
		case !canMove:
			b.ClearSelection()
			return false
		case len(matches) == 1:
			b.ClearSelection()
			doMove(chess.UCINotation{}.Encode(pos, matches[0]))
			return true
		default: // promotion: let the player choose the piece
			b.openPromotion(selected, sq, pos.Turn(), doMove)
			return false
		}
	}

	piece := pos.Board().Piece(sq)
	if sq == selected || piece == chess.NoPiece || piece.Color() != pos.Turn() {
		b.ClearSelection()
		return false
	}
	targets := map[chess.Square]bool{}
	for _, m := range game.ValidMoves() {
		if m.S1() == sq {
			targets[m.S2()] = true
		}
	}
	b.SetSelection(sq, targets)
	return false
}

func (b *BoardView) openPromotion(from, to chess.Square, color chess.Color, doMove func(string)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	row, col := displayOf(to, b.white)
	dir := 1 // promotion square at the top edge: picker extends downwards
	if row == 7 {
		dir = -1
	}
	squares := map[chess.Square]chess.PieceType{}
	for i, pt := range promoOrder {
		squares[squareAt(row+i*dir, col, b.white)] = pt
	}
	b.promo = &promoPick{from: from, to: to, color: color, squares: squares, done: doMove}
	b.targets = nil
	b.cursor = to
}

// Piece bitmaps. '#' = body, 'o' = detail line, '.' = transparent.
var artSmall = map[chess.PieceType][]string{
	chess.Pawn: {
		"...#...",
		"..###..",
		"...#...",
		"..###..",
		".#####.",
	},
	chess.Knight: {
		"..##...",
		".####..",
		"#####..",
		"#..##..",
		"..###..",
		".#####.",
	},
	chess.Bishop: {
		"...#...",
		"..#o#..",
		"..###..",
		"...#...",
		"..###..",
		".#####.",
	},
	chess.Rook: {
		"#.#.#.#",
		"#######",
		".#####.",
		"..###..",
		".#####.",
		"#######",
	},
	chess.Queen: {
		"#..#..#",
		"##.#.##",
		".#####.",
		"..###..",
		".#####.",
		"#######",
	},
	chess.King: {
		"...#...",
		"..###..",
		"#..#..#",
		"#######",
		".#####.",
		"#######",
	},
}

var artLarge = map[chess.PieceType][]string{
	chess.Pawn: {
		"......##......",
		".....####.....",
		"....######....",
		"....######....",
		".....####.....",
		"......##......",
		".....####.....",
		"....oooooo....",
		".....####.....",
		".....####.....",
		"....######....",
		"...########...",
		"..##########..",
		"..##########..",
	},
	chess.Knight: {
		".......#......",
		".....#####....",
		"....#######...",
		"...###o#####..",
		"..##########..",
		".###########..",
		".####..######.",
		"..##...######.",
		"......######..",
		".....######...",
		"....#######...",
		"....#######...",
		"...oooooooo...",
		"..##########..",
		".############.",
		".############.",
	},
	chess.Bishop: {
		"......##......",
		".....####.....",
		"....####o#....",
		"...####o###...",
		"...###o####...",
		"...########...",
		"....######....",
		".....####.....",
		"...oooooooo...",
		"....######....",
		".....####.....",
		".....####.....",
		"...########...",
		"..##########..",
		".############.",
		".############.",
	},
	chess.Rook: {
		"..##.####.##..",
		"..##.####.##..",
		"..##########..",
		"..##########..",
		"...########...",
		"...oooooooo...",
		"....######....",
		"....######....",
		"....######....",
		"....######....",
		"....######....",
		"...oooooooo...",
		"...########...",
		"..##########..",
		".############.",
		".############.",
	},
	chess.Queen: {
		".#....##....#.",
		".#....##....#.",
		".##..####..##.",
		".###.####.###.",
		".############.",
		"..##########..",
		"...########...",
		"....######....",
		"...oooooooo...",
		"....######....",
		"....######....",
		"...########...",
		"...oooooooo...",
		"..##########..",
		".############.",
		".############.",
	},
	chess.King: {
		"......##......",
		"....######....",
		"......##......",
		"..###.##.###..",
		".#####..#####.",
		".############.",
		".############.",
		"..##########..",
		"...oooooooo...",
		"....######....",
		"...########...",
		"....######....",
		"...########...",
		"..oooooooooo..",
		".############.",
		".############.",
	},
}

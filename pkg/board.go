package pkg

import (
	"fmt"
	"image"
	"math"
	"strings"
	"sync"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

// BoardView is a chess board primitive that scales to fill whatever space the
// layout gives it. Squares grow with the terminal; in terminals with kitty's
// text sizing protocol the pieces grow with them (see textsize.go).
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
	right    bool // draw against the right edge instead of centred

	history      []*chess.Position // the game so far, for stepping back (see board_extras.go)
	historyMoves []*chess.Move
	back         int // how many moves back the board is showing (0 = now)
	shapes       []boardShape
	dragFrom     chess.Square // right-button drag start, for arrows
	result       boardResult  // set when the game shown is over (see result.go)
	topBar       clockBar     // clock bars beside the players (see board_extras.go)
	bottomBar    clockBar
	onShown      func(ply int) // tells a linked move list the move shown (linkMoveList)
	shownPly     int
	browse       bool      // browsing a line (studies): no "return to now" label or click
	sounds       bool      // a game board: a new move plays its sound
	glyph        moveGlyph // a move symbol (!, ??...) on glyphSq, see glyphs.go
	glyphSq      chess.Square

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
		Box:       cv.NewBox(),
		pos:       chess.StartingPosition(),
		white:     true,
		selected:  chess.NoSquare,
		lastFrom:  chess.NoSquare,
		lastTo:    chess.NoSquare,
		cursor:    chess.E2,
		dragFrom:  chess.NoSquare,
		topBar:    clockBar{frac: -1},
		bottomBar: clockBar{frac: -1},
		onSquare:  onSquare,
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
		b.shapes = nil
		b.result = boardResult{}
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

// How a board is drawn this frame.
type boardStyle int

const (
	styleGlyph boardStyle = iota // one Unicode symbol per square
	styleSized                   // symbols scaled up with kitty text sizing
	styleImage                   // piece images with kitty graphics
	styleBasic                   // symbols in the 8 standard colours, for any terminal
)

// boardGeometry picks the largest square size that fits w x h cells. Terminal
// cells are about twice as tall as wide, so squares are ~2x wider than tall.
// With big pieces, squares are exactly 2s x s cells so a piece scaled by s
// (the protocol allows 1-7) fills one square. With images, aspect is the
// cell's height/width in pixels, so squares come out as square as possible.
func boardGeometry(w, h int, style boardStyle, aspect float64) (sqW, sqH int) {
	w -= labelW
	h -= labelH
	switch style {
	case styleSized:
		for s := min(7, h/8); s >= 1; s-- {
			if 8*2*s <= w {
				return 2 * s, s
			}
		}
		return 0, 0
	case styleImage:
		for s := h / 8; s >= 1; s-- {
			sw := max(1, int(math.Round(float64(s)*aspect)))
			if 8*sw <= w {
				return sw, s
			}
		}
		return 0, 0
	}
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

// styleFor decides how to draw on this screen: images need the cell size in
// pixels, which some terminals don't report.
func styleFor(screen tc.Screen) (style boardStyle, cellW, cellH int) {
	if PiecesSetting() == PiecesBasic {
		return styleBasic, 0, 0
	}
	if useImages() {
		if cw, ch, ok := cellPixels(screen); ok {
			return styleImage, cw, ch
		}
	}
	if TextSizingMode() {
		return styleSized, 0, 0
	}
	return styleGlyph, 0, 0
}

// WidthFor is how many columns the board needs when given w x h cells.
func (b *BoardView) WidthFor(screen tc.Screen, w, h int) int {
	style, cw, ch := styleFor(screen)
	aspect := 2.0
	if cw > 0 {
		aspect = float64(ch) / float64(cw)
	}
	sqW, _ := boardGeometry(w, h, style, aspect)
	return 8*sqW + labelW
}

var (
	colLight     = tc.NewRGBColor(145, 130, 109)
	colDark      = tc.NewRGBColor(108, 81, 59)
	colWhiteBody = tc.NewRGBColor(255, 248, 220)
	colBlackBody = tc.NewRGBColor(18, 18, 18)
	colLabel     = tc.NewRGBColor(170, 160, 140)

	tintSelected = tc.NewRGBColor(110, 150, 70)
	tintLastMove = tc.NewRGBColor(205, 190, 80)
	tintCapture  = tc.NewRGBColor(200, 70, 60)
	tintCursor   = tc.NewRGBColor(80, 130, 200)
	colDot       = tc.NewRGBColor(60, 90, 40)
	colPromo     = tc.NewRGBColor(200, 195, 185)
	colPromoDark = tc.NewRGBColor(95, 90, 85)

	// Lichess's default ("brown") board
	lichessLastMove = tc.NewRGBColor(155, 199, 0)
	lichessSelected = tc.NewRGBColor(20, 85, 30)
	lichessPromo    = tc.NewRGBColor(0xb0, 0xb0, 0xb0)
)

// Big pieces are drawn at this fraction of the square, centred, leaving a
// little margin.
const pieceScaleN, pieceScaleD = 9, 10

// bigGlyphs are the pieces drawn with text sizing. The filled glyphs are used
// for both colours (coloured by fg); the pawn gets U+FE0E so it renders as
// text, not emoji.
var bigGlyphs = map[chess.PieceType]string{
	chess.King: "♚", chess.Queen: "♛", chess.Rook: "♜",
	chess.Bishop: "♝", chess.Knight: "♞", chess.Pawn: "♟\uFE0E",
}

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
	b.notifyShown()

	style, cellW, cellH := styleFor(screen)
	aspect := 2.0
	if style == styleImage {
		aspect = float64(cellH) / float64(cellW)
	}
	x, y, w, h := b.GetInnerRect()
	if b.right {
		w -= boardMargin
	}
	bars := b.hasClockBars()
	if bars {
		y++ // a row for each bar
		h -= 2
	}
	sqW, sqH := boardGeometry(w, h, style, aspect)
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

	// stepping back through history: draw the earlier position instead
	if b.back > 0 {
		savedPos, savedFrom, savedTo := b.pos, b.lastFrom, b.lastTo
		b.pos, b.lastFrom, b.lastTo = b.shown()
		defer func() { b.pos, b.lastFrom, b.lastTo = savedPos, savedFrom, savedTo }()
	}

	totalW := 8*sqW + labelW
	totalH := 8*sqH + labelH
	b.ox = x + (w-totalW)/2 + labelW
	if b.right {
		b.ox = x + w - totalW + labelW
	}
	b.oy = y + (h-totalH)/2

	focused := b.HasFocus()
	var cells []sizedCell
	var images []imgPlacement
	// images: the largest square that fits the square's pixels, centred
	pieceSize := min(sqW*cellW, sqH*cellH)
	checkSq := b.checkSquare()
	place := func(sx, sy, img, z int) {
		dx, dy := (sqW*cellW-pieceSize)/2, (sqH*cellH-pieceSize)/2
		images = append(images, imgPlacement{x: sx + dx/cellW, y: sy + dy/cellH, px: dx % cellW, py: dy % cellH, img: img, z: z})
	}
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			sq := squareAt(row, col, b.white)
			sx, sy := b.ox+col*sqW, b.oy+row*sqH
			var look squareLook
			if style == styleBasic {
				look = b.basicLookOf(sq, focused && sq == b.cursor)
			} else {
				look = b.lookOf(sq, focused && sq == b.cursor, style == styleImage)
			}
			if sq == checkSq { // a king in check glows red, as on Lichess
				switch style {
				case styleImage:
					place(sx, sy, imgCheck, zMarker)
				case styleBasic:
					look.bg = tc.ColorMaroon
				default:
					look.bg = blend(look.bg, colCheck, 0.7)
				}
			}
			kb := badgeNone
			if look.piece.Type() == chess.King && b.promo == nil {
				kb = b.badgeFor(look.piece.Color())
			}
			if kb != badgeNone && style != styleImage {
				if style == styleBasic {
					look.bg = badgeBasic(kb)
				} else {
					look.bg = blend(look.bg, badgeTc(kb), 0.55)
				}
			}
			for cy := 0; cy < sqH; cy++ {
				for cx := 0; cx < sqW; cx++ {
					screen.SetContent(sx+cx, sy+cy, ' ', nil, tc.StyleDefault.Background(look.bg))
				}
			}
			switch {
			case style == styleImage:
				if look.dot {
					place(sx, sy, imgDot, zMarker)
				}
				if look.capture {
					place(sx, sy, imgRing, zMarker)
				}
				if look.piece != chess.NoPiece {
					place(sx, sy, pieceImageID(look.piece), zPiece)
				}
			case style == styleSized:
				cells = append(cells, look.sized(sx, sy, sqW, sqH))
			case look.piece != chess.NoPiece:
				name := look.piece.Type().String()
				if style == styleBasic && look.piece.Color() == chess.White {
					name = strings.ToUpper(name) // outline symbols for white
				}
				glyph := []rune(PiecesMap[rune(name[0])])
				screen.SetContent(sx+sqW/2, sy+sqH/2, glyph[0], glyph[1:], tc.StyleDefault.Foreground(look.fg).Background(look.bg).Bold(style == styleBasic))
			case look.dot:
				fg := colDot
				if style == styleBasic {
					fg = look.fg
				}
				screen.SetContent(sx+sqW/2, sy+sqH/2, '•', nil, tc.StyleDefault.Foreground(fg).Background(look.bg))
			}
			if b.glyph != 0 && sq == b.glyphSq && b.promo == nil && style != styleSized { // big pieces fill the square
				g := b.glyph
				switch style {
				case styleImage:
					bs := badgeSizeFor(pieceSize)
					dx := sqW*cellW - bs
					images = append(images, imgPlacement{x: sx + dx/cellW, y: sy, px: dx % cellW, img: imgGlyph(g), z: zBadge})
				default:
					bg, fg := g.tc(), tc.ColorWhite
					if style == styleBasic {
						bg, fg = g.basic(), tc.ColorWhite
					}
					text := []rune(g.text())
					for i, r := range text {
						screen.SetContent(sx+sqW-len(text)+i, sy, r, nil, tc.StyleDefault.Foreground(fg).Background(bg).Bold(true))
					}
				}
			}
			if kb != badgeNone {
				switch style {
				case styleImage: // badge in the square's top-right corner
					bs := badgeSizeFor(pieceSize)
					dx := sqW*cellW - bs
					images = append(images, imgPlacement{x: sx + dx/cellW, y: sy, px: dx % cellW, img: imgBadge(kb), z: zBadge})
				case styleGlyph, styleBasic:
					fg := tc.ColorWhite
					screen.SetContent(sx+sqW-1, sy, badgeText[kb], nil, tc.StyleDefault.Foreground(fg).Background(look.bg).Bold(true))
				}
			}
		}
	}
	switch style {
	case styleSized:
		queueSized(screen, rect{b.ox, b.oy, 8 * sqW, 8 * sqH}, cells)
	case styleImage:
		var overlay *image.RGBA
		key := ""
		if len(b.shapes) > 0 {
			key = fmt.Sprint(b.shapes, b.white, sqW*cellW, sqH*cellH)
			overlay = shapesImage(b.shapes, b.white, sqW*cellW, sqH*cellH)
			images = append(images, imgPlacement{x: b.ox, y: b.oy, img: imgOverlay, z: zShapes})
		}
		queueImages(pieceSize, currentPieceSet(), images)
		queueOverlay(overlay, key)
	}
	if b.back > 0 && !b.browse && b.oy > y {
		label := b.historyLabel()
		lx := b.ox + 8*sqW - len([]rune(label))
		for i, r := range label {
			screen.SetContent(lx+i, b.oy-1, r, nil, tc.StyleDefault.Foreground(tc.ColorBlack).Background(tintLastMove))
		}
	}

	if bars {
		bg := b.GetBackgroundColor()
		if b.back == 0 || b.browse { // the history banner uses this row (not in studies)
			drawClockBar(screen, b.ox, b.oy-1, 8*sqW, b.topBar, style == styleBasic, bg)
		}
		drawClockBar(screen, b.ox, b.oy+8*sqH+labelH, 8*sqW, b.bottomBar, style == styleBasic, bg)
	}
	labelStyle := tc.StyleDefault.Foreground(colLabel).Background(b.GetBackgroundColor())
	for i := 0; i < 8; i++ {
		rank := squareAt(i, 0, b.white).Rank().String()
		screen.SetContent(b.ox-2, b.oy+i*sqH+sqH/2, rune(rank[0]), nil, labelStyle)
		file := squareAt(0, i, b.white).File().String()
		screen.SetContent(b.ox+i*sqW+sqW/2, b.oy+8*sqH, rune(file[0]), nil, labelStyle)
	}
}

// squareLook is what one square shows.
type squareLook struct {
	bg, fg  tc.Color
	piece   chess.Piece
	dot     bool // empty square the selected piece can move to
	capture bool // occupied square the selected piece can capture on
}

// sized turns a look into a multicell character filling the square.
//
// Each glyph is followed by a space: kitty centres text by its width in the
// main (monospace) font, which usually has no chess symbols and measures them
// as zero wide, leaving the piece's left edge at the centre. The trailing
// space gives it a width close to the piece's, so the piece lands centred.
func (l squareLook) sized(x, y, sqW, sqH int) sizedCell {
	c := sizedCell{x: x, y: y, s: sqH, w: sqW / sqH, fg: l.fg, bg: l.bg, text: " "}
	switch {
	case l.piece != chess.NoPiece:
		c.text, c.n, c.d = bigGlyphs[l.piece.Type()]+" ", pieceScaleN, pieceScaleD
	case l.dot:
		// ⬤, not ●: kitty draws ● itself and tiles it across the block
		c.text, c.fg, c.n, c.d = "⬤ ", colDot, 2, 5
	}
	return c
}

// lookOf works out a square's colours and contents: base colour, then
// promotion picker, selection, capture targets, last move and cursor.
// lichess picks the Lichess board colours (used with piece images).
func (b *BoardView) lookOf(sq chess.Square, cursor, lichess bool) squareLook {
	light, dark := colLight, colDark
	if lichess {
		t := currentBoard()
		light, dark = t.light, t.dark
	}
	l := squareLook{bg: dark, piece: b.pos.Board().Piece(sq)}
	if (int(sq.File())+int(sq.Rank()))%2 == 1 {
		l.bg = light
	}
	if b.promo != nil {
		if pt, ok := b.promo.squares[sq]; ok {
			l.piece = makePiece(pt, b.promo.color)
			switch {
			case lichess:
				l.bg = lichessPromo
			case b.promo.color == chess.White:
				l.bg = colPromoDark
			default:
				l.bg = colPromo
			}
		} else {
			l.bg = blend(l.bg, tc.ColorBlack, 0.55)
		}
	} else {
		target := b.targets[sq]
		switch {
		case sq == b.selected && lichess:
			l.bg = blend(l.bg, lichessSelected, 0.5)
		case sq == b.selected:
			l.bg = blend(l.bg, tintSelected, 0.6)
		case target && l.piece != chess.NoPiece && !lichess:
			l.bg = blend(l.bg, tintCapture, 0.5)
		case (sq == b.lastFrom || sq == b.lastTo) && lichess:
			l.bg = blend(l.bg, lichessLastMove, 0.41)
		case sq == b.lastFrom || sq == b.lastTo:
			l.bg = blend(l.bg, tintLastMove, 0.4)
		}
		l.dot = target && l.piece == chess.NoPiece
		l.capture = target && l.piece != chess.NoPiece
	}
	if tint, ok := b.shapeTint(sq); ok && !lichess {
		l.bg = blend(l.bg, tint, 0.6)
	}
	if cursor {
		l.bg = blend(l.bg, tintCursor, 0.5)
	}
	l.fg = colBlackBody
	if l.piece != chess.NoPiece && l.piece.Color() == chess.White {
		l.fg = colWhiteBody
	}
	return l
}

// basicLookOf is lookOf for basic rendering: only the 8 standard colours
// (which every colour terminal has, and which follow its colour scheme), and
// both sides in black, white as outline symbols, like a printed diagram.
func (b *BoardView) basicLookOf(sq chess.Square, cursor bool) squareLook {
	l := squareLook{bg: tc.ColorOlive, fg: tc.ColorBlack, piece: b.pos.Board().Piece(sq)}
	if (int(sq.File())+int(sq.Rank()))%2 == 1 {
		l.bg = tc.ColorSilver
	}
	if b.promo != nil {
		if pt, ok := b.promo.squares[sq]; ok {
			l.piece = makePiece(pt, b.promo.color)
			l.bg = tc.ColorSilver
		} else {
			l.bg = tc.ColorBlack
			l.fg = tc.ColorSilver
		}
	} else {
		target := b.targets[sq]
		switch {
		case sq == b.selected:
			l.bg = tc.ColorGreen
		case target && l.piece != chess.NoPiece:
			l.bg = tc.ColorMaroon
		case sq == b.lastFrom || sq == b.lastTo:
			l.bg = tc.ColorTeal
		}
		l.dot = target && l.piece == chess.NoPiece
	}
	if _, ok := b.shapeTint(sq); ok {
		l.bg = tc.ColorGreen
	}
	if cursor {
		l.bg, l.fg = tc.ColorNavy, tc.ColorSilver
	}
	return l
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
		case tc.KeyHome, tc.KeyEnd:
			b.mu.Lock()
			if event.Key() == tc.KeyHome {
				b.step(len(b.history))
			} else {
				b.back = 0
			}
			b.mu.Unlock()
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
			case "[", "]":
				b.mu.Lock()
				if event.Str() == "[" {
					b.step(1)
				} else {
					b.step(-1)
				}
				b.mu.Unlock()
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
		b.mu.Lock()
		ox, oy, sqW, sqH, white := b.ox, b.oy, b.sqW, b.sqH, b.white
		sq := chess.NoSquare
		if sqW > 0 && mx >= ox && my >= oy && (mx-ox)/sqW <= 7 && (my-oy)/sqH <= 7 {
			sq = squareAt((my-oy)/sqH, (mx-ox)/sqW, white)
		}
		switch action {
		case cv.MouseScrollUp:
			b.step(1)
		case cv.MouseScrollDown:
			b.step(-1)
		case cv.MouseRightDown:
			b.dragFrom = sq
		case cv.MouseRightUp:
			if b.dragFrom != chess.NoSquare && sq != chess.NoSquare {
				b.toggleShape(boardShape{from: b.dragFrom, to: sq, color: shapeColorFor(event.Modifiers())})
			}
			b.dragFrom = chess.NoSquare
		}
		if action != cv.MouseLeftDown {
			b.mu.Unlock()
			return true, nil
		}
		b.shapes = nil
		if b.back > 0 && !b.browse { // viewing history: a click comes back to the present
			b.back = 0
			b.mu.Unlock()
			return true, nil
		}
		if sq == chess.NoSquare {
			b.mu.Unlock()
			return true, nil
		}
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

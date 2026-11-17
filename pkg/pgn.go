package pkg

import (
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

// Reading annotated PGN, as Lichess exports studies: several games
// (chapters), each a tree of moves with variations, comments, move and
// position symbols ($1, !?), and the arrows and circles drawn in the study
// ([%cal Ge2e4] and [%csl Rd5] inside comments).

// studyNode is a move in a chapter's tree (or the start, for the root).
type studyNode struct {
	id       int
	move     *chess.Move // nil at the root
	san      string
	pos      *chess.Position // the position after the move
	parent   *studyNode
	children []*studyNode // [0] continues the line; the rest are variations
	comment  string
	nags     []int
	shapes   []boardShape
	ply      int // half-moves from the chapter's start position
}

// studyChapter is one game of the PGN.
type studyChapter struct {
	headers map[string]string
	root    *studyNode
	nodes   []*studyNode // by id
	white   bool         // white at the bottom (the study's orientation)
	problem string       // the first move that couldn't be read, if any
}

// name is the chapter's title: Lichess's ChapterName, or the Event.
func (c *studyChapter) name() string {
	if n := c.headers["ChapterName"]; n != "" {
		return n
	}
	ev := c.headers["Event"]
	if _, after, ok := strings.Cut(ev, ": "); ok && c.headers["StudyName"] != "" {
		return after
	}
	if ev == "" || ev == "?" {
		return "Chapter"
	}
	return ev
}

// studyName is the study's name: Lichess's StudyName, or the Event, which
// Lichess writes as "Study name: Chapter name".
func (c *studyChapter) studyName() string {
	if n := c.headers["StudyName"]; n != "" {
		return n
	}
	ev, ch := c.headers["Event"], c.headers["ChapterName"]
	if ch != "" && strings.HasSuffix(ev, ": "+ch) {
		return strings.TrimSuffix(ev, ": "+ch)
	}
	return ""
}

// chapterID is Lichess's ID for the chapter (from its link), or "".
func (c *studyChapter) chapterID() string {
	u := strings.TrimRight(c.headers["ChapterURL"], "/")
	if i := strings.LastIndex(u, "/"); i >= 0 {
		return u[i+1:]
	}
	return ""
}

// parseStudyPGN reads every game in a PGN text.
func parseStudyPGN(text string) []*studyChapter {
	p := &pgnParser{src: []rune(text)}
	p.run()
	return p.chapters
}

type pgnParser struct {
	src      []rune
	i        int
	chapters []*studyChapter

	c      *studyChapter
	moves  bool       // the current chapter's movetext has started
	cur    *studyNode // the last move read (or the root)
	broken bool       // a move in this line couldn't be read: skip to the line's end
	stack  []pgnFrame
}

type pgnFrame struct {
	cur    *studyNode
	broken bool
}

func (p *pgnParser) run() {
	for p.i < len(p.src) {
		r := p.src[p.i]
		switch {
		case unicode.IsSpace(r):
			p.i++
		case r == '[':
			if p.c == nil || p.moves {
				p.startChapter()
			}
			p.header()
		case r == '{':
			p.need()
			p.moves = true
			end := p.until('}')
			p.addComment(string(p.src[p.i+1 : end]))
			p.i = end + 1
		case r == ';':
			p.need()
			p.moves = true
			end := p.until('\n')
			p.addComment(string(p.src[p.i+1 : end]))
			p.i = end + 1
		case r == '(':
			p.need()
			p.moves = true
			p.i++
			p.stack = append(p.stack, pgnFrame{p.cur, p.broken})
			if p.cur.parent != nil { // the variation replaces the last move
				p.cur = p.cur.parent
			} else {
				p.broken = true // a variation before any move: not valid
			}
		case r == ')':
			p.i++
			if n := len(p.stack); n > 0 {
				p.cur, p.broken = p.stack[n-1].cur, p.stack[n-1].broken
				p.stack = p.stack[:n-1]
			}
		case r == '$':
			p.need()
			p.moves = true
			start := p.i + 1
			p.i = start
			for p.i < len(p.src) && unicode.IsDigit(p.src[p.i]) {
				p.i++
			}
			if n, err := strconv.Atoi(string(p.src[start:p.i])); err == nil && !p.broken {
				p.cur.nags = append(p.cur.nags, n)
			}
		default:
			p.need()
			p.moves = true
			start := p.i
			for p.i < len(p.src) && !unicode.IsSpace(p.src[p.i]) && !strings.ContainsRune("(){}[];$", p.src[p.i]) {
				p.i++
			}
			p.token(string(p.src[start:p.i]))
		}
	}
}

// until finds the next r from the current position (or the end).
func (p *pgnParser) until(r rune) int {
	for j := p.i + 1; j < len(p.src); j++ {
		if p.src[j] == r {
			return j
		}
	}
	return len(p.src)
}

// need starts a chapter if movetext comes without headers.
func (p *pgnParser) need() {
	if p.c == nil {
		p.startChapter()
	}
}

func (p *pgnParser) startChapter() {
	p.c = &studyChapter{headers: map[string]string{}, white: true}
	p.chapters = append(p.chapters, p.c)
	p.moves, p.broken, p.stack = false, false, nil
	p.c.root = p.newNode(nil, "", chess.StartingPosition(), nil)
	p.cur = p.c.root
}

func (p *pgnParser) newNode(m *chess.Move, san string, pos *chess.Position, parent *studyNode) *studyNode {
	n := &studyNode{id: len(p.c.nodes), move: m, san: san, pos: pos, parent: parent}
	if parent != nil {
		n.ply = parent.ply + 1
		parent.children = append(parent.children, n)
	}
	p.c.nodes = append(p.c.nodes, n)
	return n
}

// header reads [Key "Value"].
func (p *pgnParser) header() {
	end := p.i + 1
	inQuote := false
	for ; end < len(p.src); end++ {
		switch c := p.src[end]; {
		case c == '\\' && inQuote:
			end++
		case c == '"':
			inQuote = !inQuote
		case c == ']' && !inQuote:
			goto done
		}
	}
done:
	body := strings.TrimSpace(string(p.src[p.i+1 : min(end, len(p.src))]))
	p.i = end + 1
	key, value, _ := strings.Cut(body, " ")
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(strings.TrimPrefix(value, `"`), `"`)
	value = strings.ReplaceAll(strings.ReplaceAll(value, `\"`, `"`), `\\`, `\`)
	p.c.headers[key] = value
	switch key {
	case "FEN":
		if opt, err := chess.FEN(value); err == nil {
			p.c.root.pos = chess.NewGame(opt).Position()
		} else {
			p.c.problem = "the chapter's start position (FEN) isn't valid"
		}
	case "Orientation":
		p.c.white = value != "black"
	}
}

// moveGlyphs are move symbols written after the move, as NAGs.
var moveGlyphs = map[string]int{"!": 1, "?": 2, "!!": 3, "??": 4, "!?": 5, "?!": 6}

func (p *pgnParser) token(tok string) {
	switch tok {
	case "1-0", "0-1", "1/2-1/2", "½-½", "*":
		return
	}
	// move numbers: "12." "12..." or glued to the move, "12...e5"
	tok = strings.TrimLeft(tok, "0123456789")
	tok = strings.TrimLeft(tok, ".…")
	if tok == "" || p.broken {
		return
	}
	san := strings.TrimRight(tok, "!?")
	glyph := tok[len(san):]
	san = strings.TrimRight(san, "+#")
	san = strings.ReplaceAll(san, "0-0", "O-O")
	if san == "--" || san == "Z0" { // a null move: can't be followed
		p.problem(tok)
		return
	}
	m, err := chess.AlgebraicNotation{}.Decode(p.cur.pos, san)
	if err != nil {
		p.problem(tok)
		return
	}
	// reuse a move already there (a repeated variation)
	for _, ch := range p.cur.children {
		if ch.move.String() == m.String() {
			p.cur = ch
			p.addGlyph(glyph)
			return
		}
	}
	p.cur = p.newNode(m, chess.AlgebraicNotation{}.Encode(p.cur.pos, m), p.cur.pos.Update(m), p.cur)
	p.addGlyph(glyph)
}

func (p *pgnParser) addGlyph(g string) {
	if n, ok := moveGlyphs[g]; ok {
		p.cur.nags = append(p.cur.nags, n)
	}
}

func (p *pgnParser) problem(tok string) {
	p.broken = true
	if p.c.problem == "" {
		p.c.problem = fmt.Sprintf("couldn't follow the move %q (after %s)", tok, describeNode(p.cur))
	}
}

func describeNode(n *studyNode) string {
	if n.move == nil {
		return "the start"
	}
	return moveLabel(n)
}

// moveLabel is a move with its number, e.g. "12... Nf6".
func moveLabel(n *studyNode) string {
	num := moveNumber(n.parent.pos)
	if n.parent.pos.Turn() == chess.White {
		return fmt.Sprintf("%d. %s", num, n.san)
	}
	return fmt.Sprintf("%d... %s", num, n.san)
}

// addComment attaches a comment to the last move (or the chapter start),
// taking out the drawing commands.
func (p *pgnParser) addComment(text string) {
	if p.broken {
		return
	}
	text, shapes := commentShapes(text)
	p.cur.shapes = append(p.cur.shapes, shapes...)
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return
	}
	if p.cur.comment != "" {
		p.cur.comment += " "
	}
	p.cur.comment += text
}

// commentShapes takes the [%...] commands out of a comment, returning the
// text left and the arrows (%cal) and circles (%csl) drawn.
func commentShapes(text string) (string, []boardShape) {
	var shapes []boardShape
	var rest strings.Builder
	for {
		i := strings.Index(text, "[%")
		if i < 0 {
			rest.WriteString(text)
			break
		}
		rest.WriteString(text[:i])
		end := strings.IndexByte(text[i:], ']')
		if end < 0 {
			break
		}
		cmd := strings.Fields(text[i+2 : i+end])
		text = text[i+end+1:]
		if len(cmd) < 2 || (cmd[0] != "cal" && cmd[0] != "csl") {
			continue // %clk, %eval, %emt...
		}
		for _, spec := range strings.Split(cmd[1], ",") {
			if s, ok := parseShape(spec, cmd[0] == "cal"); ok {
				shapes = append(shapes, s)
			}
		}
	}
	return rest.String(), shapes
}

var shapeColorCodes = map[byte]int{'G': 0, 'R': 1, 'B': 2, 'Y': 3}

// parseShape reads "Ge2e4" (arrow) or "Rd5" (circle).
func parseShape(spec string, arrow bool) (boardShape, bool) {
	spec = strings.TrimSpace(spec)
	if len(spec) < 3 {
		return boardShape{}, false
	}
	color, ok := shapeColorCodes[spec[0]]
	if !ok {
		return boardShape{}, false
	}
	from, ok1 := parseSquare(spec[1:3])
	to := from
	ok2 := true
	if arrow {
		if len(spec) < 5 {
			return boardShape{}, false
		}
		to, ok2 = parseSquare(spec[3:5])
	}
	return boardShape{from: from, to: to, color: color}, ok1 && ok2
}

func parseSquare(s string) (chess.Square, bool) {
	if len(s) != 2 || s[0] < 'a' || s[0] > 'h' || s[1] < '1' || s[1] > '8' {
		return chess.NoSquare, false
	}
	return chess.Square(int(s[1]-'1')*8 + int(s[0]-'a')), true
}

// nagText is how a NAG is shown: move symbols right after the move,
// position symbols after a space.
func nagText(n int) (string, bool) {
	switch n {
	case 1:
		return "!", true
	case 2:
		return "?", true
	case 3:
		return "!!", true
	case 4:
		return "??", true
	case 5:
		return "!?", true
	case 6:
		return "?!", true
	case 7:
		return "□", true
	case 22:
		return "⨀", true
	}
	// Lichess's position and observation symbols
	pos := map[int]string{10: "=", 13: "∞", 14: "⩲", 15: "⩱", 16: "±", 17: "∓", 18: "+−", 19: "−+",
		32: "↑↑", 36: "↑", 40: "→", 44: "=∞", 132: "⇆", 138: "⊕", 140: "∆", 146: "N"}[n]
	return pos, false
}

// tagLike matches what cview would read as a colour or region tag.
var tagLike = regexp.MustCompile(`(\[(?:[a-zA-Z]+|#[0-9a-zA-Z]{6}|-)?(?::(?:[a-zA-Z]+|#[0-9a-zA-Z]{6}|-)?(?::(?:[bdilrsu]+|-)?)?)?|\["[a-zA-Z0-9_,;: \-.]*")\]`)

// escapeText escapes text for a cview text view. Unlike cview.Escape it
// leaves other bracketed text alone: "[11...dxe5 12. dxe5]" isn't a tag, and
// escaping it breaks when the line wraps inside it.
func escapeText(s string) string {
	return tagLike.ReplaceAllStringFunc(s, func(m string) string {
		if m == "[]" {
			return "[ ]"
		}
		return m[:len(m)-1] + "[]"
	})
}

// tint styles text lines (not the indent): the current move's comment gets
// a faint background, the side lines that branch off at it are underlined.
// Each line is styled on its own,
// so the colour never runs past the end of a line.
func tint(text string, on bool, style string) string {
	if !on || uiLimited() { // dark tints are all black on a limited terminal
		return text
	}
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		body := strings.TrimLeft(l, " ")
		if body != "" {
			lines[i] = l[:len(l)-len(body)] + style + body + styleReset[style]
		}
	}
	return strings.Join(lines, "\n")
}

// Colours in the move list.
const (
	pgnVariation = "[#a0a0a0]"
	pgnComment   = "[#8fbf8f]"
	tintComment  = "[:#1e2a1e]" // faint green behind the current move's comment
)

// styleReset undoes a tint style.
var styleReset = map[string]string{tintComment: "[:-]"}

// underlineTags underline the side lines branching off at the current move,
// heavier the deeper they're nested: solid, dashed, dotted, double, curly.
// cview only knows a plain underline, so levels past the first are marked
// with blink (never used otherwise) plus strikethrough/italic, which
// underlineScreen turns into the terminal's underline styles.
var underlineTags = []string{"[::u]", "[::ul]", "[::uls]", "[::uli]", "[::ulis]"}

// underlineBranch underlines a side line written at depth (its own lines,
// nested lines one level heavier for each level of nesting; a comment goes
// with the line it belongs to).
func underlineBranch(text string, depth int) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		body := strings.TrimLeft(l, " ")
		if body == "" {
			continue
		}
		level := (len(l)-len(body))/2 - depth
		if strings.HasPrefix(body, pgnComment) { // comments are indented one more
			level--
		}
		tag := underlineTags[max(0, min(level, len(underlineTags)-1))]
		lines[i] = l[:len(l)-len(body)] + tag + body + "[::-]"
	}
	return strings.Join(lines, "\n")
}

// colCurrentMove is behind the move showing, in every move list.
var colCurrentMove = tc.NewHexColor(0x3a3a3a)

// studyMoveText is a chapter's moves for the sidebar, Lichess style: the
// main line flows; a comment or variations break it, each on its own
// indented line. Every move is a region, n<id>, for clicking.
//
// cur is the move showing: its comment and the side lines that branch off
// at it are tinted, so you can see what the board is talking about.
func studyMoveText(c *studyChapter, cur *studyNode) string {
	var b strings.Builder
	if c.root.comment != "" {
		b.WriteString(tint(pgnComment+escapeText(c.root.comment)+"[-]\n", cur == c.root, tintComment))
	}
	if len(c.root.children) > 0 {
		writeLine(&b, c.root.children[0], c.root.children[1:], 0, "", cur)
	}
	if len(c.root.children) == 0 && c.root.comment == "" {
		b.WriteString("[#a0a0a0]No moves in this chapter.[-]")
	}
	if score, how := chapterResult(c); score != "" {
		b.WriteString("\n[::b]" + score + "[::-]\n[#a0a0a0]" + escapeText(how) + "[-]\n")
	}
	return b.String()
}

// chapterResult is how the chapter's game ended, as Lichess shows it under
// the moves: the score ("1–0") and e.g. "Checkmate • White is victorious".
// From the Result and Termination tags and the main line's final position;
// "" if the game isn't over (or the result is unknown).
func chapterResult(c *studyChapter) (score, how string) {
	line := c.root.line()
	end := line[len(line)-1].pos
	result := c.tag("Result")
	reason := ""
	switch end.Status() {
	case chess.Checkmate:
		reason = "Checkmate"
		if result == "" {
			result = map[chess.Color]string{chess.White: "0-1", chess.Black: "1-0"}[end.Turn()]
		}
	case chess.Stalemate:
		reason, result = "Stalemate", "1/2-1/2"
	}
	if reason == "" {
		switch strings.ToLower(c.tag("Termination")) {
		case "time forfeit":
			reason = "Time out"
		case "abandoned":
			reason = "Abandoned"
		case "rules infraction":
			reason = "Rules infraction"
		case "adjudication":
			reason = "Adjudicated"
		case "stalemate":
			reason = "Stalemate"
		case "insufficient material":
			reason = "Insufficient material"
		}
	}
	var winner string
	switch result {
	case "1-0":
		score, winner = "1–0", "White is victorious"
	case "0-1":
		score, winner = "0–1", "Black is victorious"
	case "1/2-1/2", "½-½":
		score, winner = "½–½", "Draw"
	default:
		return "", ""
	}
	if reason != "" {
		return score, reason + " • " + winner
	}
	return score, winner
}

// writeLine writes the line starting with move n (and its continuation),
// where alts are n's alternatives, written as variations after it. open
// starts the line ("(" for a variation).
func writeLine(b *strings.Builder, n *studyNode, alts []*studyNode, depth int, open string, cur *studyNode) {
	color := ""
	if depth > 0 {
		color = pgnVariation
	}
	indent := strings.Repeat("  ", depth)
	needNumber := true // a black move needs "12..." after a break
	b.WriteString(indent + color + open + "[-]")
	for n != nil {
		white := n.parent.pos.Turn() == chess.White
		switch {
		case white:
			fmt.Fprintf(b, "%s%d.[-] ", color, moveNumber(n.parent.pos))
		case needNumber:
			fmt.Fprintf(b, "%s%d...[-] ", color, moveNumber(n.parent.pos))
		}
		needNumber = false
		fmt.Fprintf(b, `%s["n%d"]%s%s[""][-] `, color, n.id, n.san, nagSuffix(n, color))
		var next *studyNode
		var nextAlts []*studyNode
		if len(n.children) > 0 {
			next, nextAlts = n.children[0], n.children[1:]
		}
		if n.comment != "" || len(alts) > 0 {
			b.WriteString("\n")
			if n.comment != "" {
				b.WriteString(tint(indent+"  "+pgnComment+escapeText(n.comment)+"[-]\n", n == cur, tintComment))
			}
			for _, a := range alts {
				var sub strings.Builder
				writeLine(&sub, a, nil, depth+1, "(", cur)
				branch := strings.TrimRight(sub.String(), " \n") + pgnVariation + ")[-]\n"
				if n == cur {
					branch = underlineBranch(branch, depth+1)
				}
				b.WriteString(branch)
			}
			if next != nil {
				b.WriteString(indent)
			}
			needNumber = true
		}
		n, alts = next, nextAlts
	}
	if !strings.HasSuffix(b.String(), "\n") {
		b.WriteString("\n")
	}
}

// nagSuffix is the symbols shown after a move, the move symbols in their
// Lichess colours; lineColor is the line's colour tag to go back to ("" = none).
func nagSuffix(n *studyNode, lineColor string) string {
	var move, pos string
	for _, x := range n.nags {
		t, isMove := nagText(x)
		switch {
		case isMove && glyphColors[moveGlyph(x)] != (color.NRGBA{}):
			c := glyphColors[moveGlyph(x)]
			back := lineColor
			if back == "" {
				back = "[-]"
			}
			move += fmt.Sprintf("[#%02x%02x%02x]%s%s", c.R, c.G, c.B, t, back)
		case isMove:
			move += t
		case t != "":
			pos += " " + t
		}
	}
	return move + pos
}

// line is the path from the root to n, continued along the main line of
// what follows n to its end.
func (n *studyNode) line() []*studyNode {
	var path []*studyNode
	for x := n; x != nil; x = x.parent {
		path = append(path, x)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	for x := n; len(x.children) > 0; {
		x = x.children[0]
		path = append(path, x)
	}
	return path
}

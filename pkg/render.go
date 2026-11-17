package pkg

import (
	"os"
	"strings"
	"sync"

	"github.com/nate-xyz/chess-cli/api"
)

// How pieces are drawn, chosen on the Options screen:
//
//   - PiecesText: Unicode chess symbols, scaled up to fill the squares in
//     terminals with kitty's text sizing protocol (see textsize.go).
//   - PiecesImage: the Lichess board colours and cburnett piece images, drawn
//     with kitty's graphics protocol (see images.go) in terminals that
//     support it. Elsewhere the board falls back to text.
//   - PiecesBasic: plain symbols in the 8 standard colours, like the
//     original chess-cli; for older or limited terminals.
const (
	PiecesText  = "text"
	PiecesImage = "image"
	PiecesBasic = "basic"
)

var rendering struct {
	sync.Mutex
	graphics bool   // terminal supports the kitty graphics protocol
	pieces   string // user's choice; "" = automatic
}

// SetupRendering asks the terminal what it can draw and loads the saved
// choice. Call it before the UI takes over the terminal, after the config is
// loaded. CHESS_CLI_BIG_PIECES and CHESS_CLI_IMAGES (0 or 1) override the
// detection.
func SetupRendering() {
	sized, graphics := probeTerminal()
	if v, ok := envBool("CHESS_CLI_BIG_PIECES"); ok {
		sized = v
	}
	if v, ok := envBool("CHESS_CLI_IMAGES"); ok {
		graphics = v
	}
	textSizing.Lock()
	textSizing.enabled = sized
	textSizing.Unlock()

	rendering.Lock()
	rendering.graphics = graphics
	rendering.pieces = api.UserInfo.Pieces
	rendering.Unlock()
}

func envBool(name string) (value, set bool) {
	switch strings.ToLower(os.Getenv(name)) {
	case "0", "off", "false", "no":
		return false, true
	case "1", "on", "true", "yes":
		return true, true
	}
	return false, false
}

// GraphicsSupported reports whether piece images can be shown.
func GraphicsSupported() bool {
	rendering.Lock()
	defer rendering.Unlock()
	return rendering.graphics
}

// PiecesSetting is the chosen piece style: the saved choice, or images when
// the terminal supports them.
func PiecesSetting() string {
	rendering.Lock()
	defer rendering.Unlock()
	switch rendering.pieces {
	case PiecesText, PiecesImage, PiecesBasic:
		return rendering.pieces
	}
	if rendering.graphics {
		return PiecesImage
	}
	return PiecesText
}

// previewPieces switches the piece style without saving it.
func previewPieces(style string) {
	rendering.Lock()
	rendering.pieces = style
	rendering.Unlock()
}

// RecommendedPieces is the best style this terminal can show.
func RecommendedPieces() string {
	switch {
	case GraphicsSupported():
		return PiecesImage
	case trueColor():
		return PiecesText
	}
	return PiecesBasic
}

// trueColor guesses whether the terminal shows 24-bit colour.
func trueColor() bool {
	ct := strings.ToLower(os.Getenv("COLORTERM"))
	return ct == "truecolor" || ct == "24bit" || TextSizingMode() ||
		strings.Contains(os.Getenv("TERM"), "kitty") || strings.Contains(os.Getenv("TERM"), "direct")
}

// terminalName describes the terminal, as best the environment tells.
func terminalName() string {
	name := os.Getenv("TERM_PROGRAM")
	switch {
	case name == "iTerm.app":
		name = "iTerm2"
	case name == "Apple_Terminal":
		name = "Terminal.app"
	case name == "vscode":
		name = "the VS Code terminal"
	case os.Getenv("KITTY_WINDOW_ID") != "" || strings.HasSuffix(os.Getenv("TERM"), "kitty"):
		name = "kitty"
	case os.Getenv("WT_SESSION") != "":
		name = "Windows Terminal"
	case name == "" && os.Getenv("TERM") != "":
		name = os.Getenv("TERM")
	case name == "":
		name = "an unknown terminal"
	}
	if v := os.Getenv("TERM_PROGRAM_VERSION"); v != "" && os.Getenv("TERM_PROGRAM") != "" {
		name += " " + v
	}
	if os.Getenv("TMUX") != "" {
		name += ", inside tmux"
	} else if strings.HasPrefix(os.Getenv("TERM"), "screen") {
		name += ", inside screen"
	}
	return name
}

// SetPiecesSetting changes the piece style and saves it.
func SetPiecesSetting(style string) error {
	rendering.Lock()
	rendering.pieces = style
	rendering.Unlock()
	return api.UpdateConfig(func(c *api.UserConfig) { c.Pieces = style })
}

// useImages reports whether boards should be drawn with images now.
func useImages() bool {
	return GraphicsSupported() && PiecesSetting() == PiecesImage
}

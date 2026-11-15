package pkg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// A local chess engine (Stockfish), spoken to over UCI: the standard text
// protocol on its stdin/stdout.

// stockfishLevels are Lichess-style levels 1-8: Stockfish's skill level
// (0-20), how long it may think and how deep it may search.
var stockfishLevels = []struct {
	skill, movetimeMS, depth int
}{
	{0, 50, 1}, {2, 100, 2}, {5, 150, 4}, {8, 200, 6},
	{11, 300, 8}, {14, 400, 12}, {17, 700, 16}, {20, 1000, 22},
}

// FindStockfish returns the Stockfish to use: $CHESS_CLI_STOCKFISH, then
// "stockfish" on the PATH, then common install locations. "" if none.
func FindStockfish() string {
	if p := os.Getenv("CHESS_CLI_STOCKFISH"); p != "" {
		return p
	}
	if p, err := exec.LookPath("stockfish"); err == nil {
		return p
	}
	for _, p := range []string{"/usr/games/stockfish", "/usr/local/bin/stockfish", "/opt/homebrew/bin/stockfish",
		"/snap/bin/stockfish", `C:\Program Files\Stockfish\stockfish.exe`} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

// Engine is a running UCI engine.
type Engine struct {
	cmd   *exec.Cmd
	in    io.WriteCloser
	lines chan string // the engine's output, line by line
	level int

	mu     sync.Mutex // one search at a time
	closed atomic.Bool
}

// StartEngine runs the engine at path and sets it up for a level (1-8).
func StartEngine(path string, level int) (*Engine, error) {
	cmd := exec.Command(path)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	e := &Engine{cmd: cmd, in: in, lines: make(chan string, 64), level: max(1, min(8, level))}
	go func() {
		sc := bufio.NewScanner(out)
		sc.Buffer(make([]byte, 64<<10), 1<<20)
		for sc.Scan() {
			e.lines <- sc.Text()
		}
		close(e.lines)
	}()
	if err := e.handshake(); err != nil {
		e.Close()
		return nil, err
	}
	return e, nil
}

func (e *Engine) send(cmd string) error {
	_, err := io.WriteString(e.in, cmd+"\n")
	return err
}

// waitFor reads lines until one starts with prefix, returning it.
func (e *Engine) waitFor(ctx context.Context, prefix string) (string, error) {
	for {
		select {
		case line, ok := <-e.lines:
			if !ok {
				return "", errors.New("the engine stopped")
			}
			if strings.HasPrefix(line, prefix) {
				return line, nil
			}
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func (e *Engine) handshake() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.send("uci"); err != nil {
		return err
	}
	if _, err := e.waitFor(ctx, "uciok"); err != nil {
		return fmt.Errorf("not a UCI chess engine? %w", err)
	}
	lv := stockfishLevels[e.level-1]
	_ = e.send(fmt.Sprintf("setoption name Skill Level value %d", lv.skill))
	_ = e.send("ucinewgame")
	_ = e.send("isready")
	_, err := e.waitFor(ctx, "readyok")
	return err
}

// BestMove asks for a move (UCI) in a position given as a start FEN plus the
// UCI moves played since, so the engine knows about repetitions.
func (e *Engine) BestMove(ctx context.Context, startFEN string, moves []string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed.Load() {
		return "", errors.New("the engine was closed")
	}
	pos := "position fen " + startFEN
	if len(moves) > 0 {
		pos += " moves " + strings.Join(moves, " ")
	}
	lv := stockfishLevels[e.level-1]
	if err := e.send(pos); err != nil {
		return "", err
	}
	if err := e.send(fmt.Sprintf("go movetime %d depth %d", lv.movetimeMS, lv.depth)); err != nil {
		return "", err
	}
	line, err := e.waitFor(ctx, "bestmove")
	if err != nil {
		_ = e.send("stop")
		return "", err
	}
	f := strings.Fields(line)
	if len(f) < 2 || f[1] == "(none)" {
		return "", errors.New("the engine has no move")
	}
	return f[1], nil
}

// Close quits the engine (killing it if it doesn't go quietly).
func (e *Engine) Close() {
	if e == nil {
		return
	}
	if e.closed.Swap(true) {
		return
	}
	_ = e.send("quit") // ends any search too; BestMove sees the output close
	_ = e.in.Close()
	done := make(chan struct{})
	go func() { _ = e.cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		_ = e.cmd.Process.Kill()
		<-done
	}
}

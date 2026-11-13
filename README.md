<!-- # chess-cli ♟️ -->
<h1 align='center'>chess-cli ♟️</h1>
<p align="center">
  <b>A chess tui app written in Go.</b><br/>
  <sub>Play locally with a friend or online with Lichess!</sub>
</p>
<br />

<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>

chess-cli is a tui (terminal user interface) that allows you to play chess, both offline with a friend or online as a Lichess api client.

![home](https://user-images.githubusercontent.com/57846867/182972093-c9917d32-81c3-4ba5-a6ec-a783bf254f33.png)
![game](https://user-images.githubusercontent.com/42559123/183780181-a4cb62c0-8515-4371-8406-932e11e84f77.png)

<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>

# Install

**Download a binary** for Linux, macOS or Windows from the
[releases page](https://github.com/nate-xyz/chess-cli/releases), unpack it and
run `chess-cli`.

**Or install with Go** (1.24+):

```sh
go install github.com/nate-xyz/chess-cli@latest
```

**Or build from source:**

```sh
git clone https://github.com/nate-xyz/chess-cli
cd chess-cli
go run .            # or: go build -o chess-cli . && ./chess-cli
```

A terminal with true colour and a font with the Unicode chess symbols gives the
best result. The board scales with your terminal, so make it big.

# Playing

| | |
|---|---|
| **Mouse** | Click a piece, then click where it should go. Legal moves are dotted, captures tinted. |
| **Typing** | Enter moves in UCI (`e2e4`) in the move box. Promotions need the piece: `e7e8q`. |
| **Keyboard on the board** | `Tab` moves focus to the board; arrows or `hjkl` move the cursor, `Enter`/`Space` selects, `Esc` cancels. |
| **Promotion** | A picker opens over the promotion square: click a piece, or press `q` `r` `b` `n`. |
| **Online chat** | `Tab` cycles move box → board → chat. |

Local games are two players on one keyboard; the board flips each move. Games
can be saved and resumed.

# Lichess

Choose *Lichess* on the home screen and log in through your browser (if it
doesn't open, the link is shown in the app). chess-cli uses the
[Lichess Board API](https://lichess.org/api#tag/board):

- Challenge a friend, seek a random opponent, or play Lichess's Stockfish (levels 1–8)
- All variants, rated or casual, real-time, correspondence or unlimited
- Resume ongoing games; accept, decline (`d`) and cancel challenges
- Chat, draw offers, takebacks, resign/abort
- Claim victory or a draw when your opponent leaves
- Rematch from the game-over screen
- `Esc` on the waiting screen withdraws your challenge or seek

**Time controls:** Lichess only allows Rapid, Classical and Correspondence
through the Board API, plus Blitz for direct challenges and games against the AI.
Bullet isn't available to third-party clients.

Your Lichess token and saved games are kept in `~/.config/chess-cli`
(`chess-cli --help` prints the exact folder on your system).

<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>

# Development

```sh
go test -race ./...
```

The tests include an end-to-end online game: the real UI running on tcell's
mock terminal against a fake Lichess server.

**Releasing:** push a version tag and GitHub Actions builds and publishes the
binaries with [GoReleaser](https://goreleaser.com):

```sh
git tag v0.2.0 && git push origin v0.2.0
```

### Dependency documentation

- [Lichess API](https://lichess.org/api)
- [notnil/chess](https://pkg.go.dev/github.com/notnil/chess): rules and move generation
- [cview](https://pkg.go.dev/codeberg.org/tslocum/cview): TUI widgets
- [tcell](https://pkg.go.dev/github.com/gdamore/tcell/v3): terminal handling

<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>

### Todo

#### general

- Choose the input notation (SAN like `Nf3` as well as UCI)

#### local client

- Time controls
- Play a local engine (Stockfish, if installed)

#### online

- Challenge construction
  - Set default positions for sliders
  - Custom starting position
- Watch games (Lichess TV, follow a player)
- Puzzles, opening explorer, game history/PGN export
- Variants: the board follows standard chess rules only. Games in variants
  like crazyhouse or atomic can be created but the app can't follow them yet;
  Chess960 is untested.

<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>

# chess-cli ♟️

run with: go run .

if libraries are missing: go mod tidy


TODO: 

    - fix prompt window breaking in alacritty x
    - get mouse input working in board window x
    - get lichess requests working? wip
      - get profile info
      - get challenges / make challenges
      - list of ongoing games
      - game screen w/ working request loop and translation of move sequence into FEN string
    - get oauth working ✓
    - get stockfish working through lichess? wip
    - screen handler at different depth? maybe a generic handler with functions passed as parameters


♞♞♞♞♞♞♞♞♞♞♞

DEV DOCS:
https://lichess.org/api for lichess

https://pkg.go.dev/github.com/notnil/chess?utm_source=godoc for go chess

https://pkg.go.dev/github.com/rthornton128/goncurses for goncurses

https://pkg.go.dev/net/http for http
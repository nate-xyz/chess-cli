# chess-cli ♟️

## tips

run with: go run .

if libraries are missing: go mod tidy

♞♞♞♞♞♞♞♞♞♞♞

## TODO:


### all
- [x] fix prompt window breaking in alacritty
  - [ ] fixed with mouse input active
- [ ] non unicode support for tty
- [ ] windows support with pancurses

### local
- [x] fix legal move list on info window ✓
- [ ] choose notation to start?
  - [ ] maybe settings screen for notation / color to start / unicode on or off, accessible through escape
- [ ] choose color to start
- [ ] get mouse input working in board window
- [x] local game screen handler

### lichess
- [ ] get lichess requests working? wip
  - [x] get profile info ✓
  - [x] get challenges
  - [ ] make challenges
    - [x] create a challenge (specific user)
    - [ ] create an open-ended challenge
    - [ ] create a seek (random user)
    - [ ] challenge the ai
  - [ ] list of ongoing games
    - [x] request
    - [ ] screen
  - [ ] (after challenge or join ongoing game) lichess game screen w/ working request loop and translation of move sequence into FEN string
    - [ ] tranlate move sequence into FEN w local chess lib
    - [ ] update FEN on board event from stream
    - [ ] make move
 - [x] stream incoming events
   - [x] start as goroutine initially in main and implement notifications for every screen
- [x] get oauth working ✓
- [x] screen handler at different depth? maybe a generic handler with functions passed as parameters
  - [x] implemented lichess screen handler


### ai
- [ ] get stockfish working through lichess? wip



♞♞♞♞♞♞♞♞♞♞♞

## DEV DOCS:

https://lichess.org/api for lichess

https://pkg.go.dev/github.com/notnil/chess?utm_source=godoc for go chess

https://pkg.go.dev/github.com/rthornton128/goncurses for goncurses

https://pkg.go.dev/net/http for http

♞♞♞♞♞♞♞♞♞♞♞

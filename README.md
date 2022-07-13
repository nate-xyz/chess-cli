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
- [ ] fix board flipping visual errors when black to move

### local
- [x] fix legal move list on info window ✓
- [ ] choose notation to start?
  - [ ] maybe settings screen for notation / color to start / unicode on or off, accessible through escape
- [ ] choose color to start
- [ ] get mouse input working in board window
- [x] local game screen handler

### lichess

- [ ] oauth
  - [x] get oauth working ✓
  - [ ] check if token is expired and get new one if so
- [ ] api
  - [ ] get lichess requests working? wip
    - [x] get profile info ✓
    - [x] get challenges
    - [ ] make challenges
      - [x] create a challenge (specific user)
      - [ ] create an open-ended challenge
      - [ ] create a seek (random user)
      - [ ] challenge the ai
        - [ ] generate CreateAiChallengeType type fields through windows in CreateLichessGame
    - [ ] list of ongoing games
      - [x] request
      - [ ] screen
   - [x] stream incoming events
     - [x] start as goroutine initially in main and implement notifications for every screen
- [ ] game client
  - [x] screen handler at different depth? maybe a generic handler with functions passed as parameters
    - [x] implemented lichess screen handler
  - [ ] (after challenge or join ongoing game) lichess game screen w/ working request loop and translation of move sequence into FEN string
    - [x] translate move sequence into FEN w local chess lib
    - [x] update FEN on board event from stream
    - [x] make move
    - [ ] handles lichess game abort
    - [ ] handle game end
    - [ ] handle custom starting positions

### ai
- [ ] get stockfish working through lichess? wip



♞♞♞♞♞♞♞♞♞♞♞

## DEV DOCS:

https://lichess.org/api for lichess

https://pkg.go.dev/github.com/notnil/chess?utm_source=godoc for go chess

https://pkg.go.dev/github.com/rthornton128/goncurses for goncurses

https://pkg.go.dev/net/http for http

https://mholt.github.io/json-to-go/

♞♞♞♞♞♞♞♞♞♞♞

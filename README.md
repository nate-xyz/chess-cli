<h1 align='center'>chess-cli ♟️</h1>
<p align="center">
  <b>A chess tui app written in Go.</b><br/>
  <sub>Play locally with a friend or online with Lichess!</a></sub>
</p>
<br />


<!-- # chess-cli ♟️ -->

## tips

run with: `go run .`

if libraries are missing: `go mod tidy`

♞♞♞♞♞♞♞♞♞♞♞



## DEV DOCS:

https://lichess.org/api for lichess

https://pkg.go.dev/github.com/notnil/chess?utm_source=godoc for go chess

https://docs.rocketnine.space/code.rocketnine.space/tslocum/cview/ CVIEW

https://code.rocketnine.space/tslocum/cview/src/branch/master/demos CVIEW DEMOS

https://pkg.go.dev/net/http for http

https://mholt.github.io/json-to-go/ JSON

♞♞♞♞♞♞♞♞♞♞♞


## TODO:


### local
- [ ] choose notation to start?
  - [ ] maybe settings screen for notation / color to start / unicode on or off, accessible through escape
- [ ] get mouse input working in board window
- [ ] captured pieces window

### lichess

- [ ] oauth
  - [ ] check if token is expired and get new one if so
  - [ ] handle if oauth process rejected

- [ ] api
  - [ ] make challenges
    - [ ] set default interval position
    - [ ] create an open-ended challenge
    - [ ] list of ongoing games
      - [ ] request
      - [ ] screen

- [ ] game client
  - [ ] exit option on challenge waiting screen
  - [ ] add timers and update them from stream
  - [x] handle game end
    - [ ] post game options
      - [ ] rematch
      - [ ] new game
      - [ ] back to lichess home
      - [ ] back to chess-cli home
    - [ ] handle custom starting positions
    - [ ] rematches

### ai
- [ ] get stockfish working through lichess? wip

♞♞♞♞♞♞♞♞♞♞♞


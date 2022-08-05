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



# DEV DOCS:

https://lichess.org/api for lichess

https://pkg.go.dev/github.com/notnil/chess?utm_source=godoc for go chess

https://docs.rocketnine.space/code.rocketnine.space/tslocum/cview/ CVIEW

https://code.rocketnine.space/tslocum/cview/src/branch/master/demos CVIEW DEMOS

https://pkg.go.dev/net/http for http

https://mholt.github.io/json-to-go/ JSON

♞♞♞♞♞♞♞♞♞♞♞

# TODO:

### general 

- Handle all notation type or choose notation to start
- exit option on all loaders 
  
### local client 

- save local games
- add time options?
  

### oauth

  - handle if oauth process rejected

### api

  - BULLET CHALLENGES BUG: when minutes are < 3 minutes for a realtime challenge, api provides an empty json for the GameFull event.

### online 

  - challenge construction 
    - set default interval position
    - custom starting position option for challenges
    - see variants on board
  
### online game client

  - rating info next to name
  - chat
   

♞♞♞♞♞♞♞♞♞♞♞


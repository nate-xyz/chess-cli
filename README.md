<h1 align='center'>chess-cli ♟️</h1>
<p align="center">
  <b>A chess tui app written in Go.</b><br/>
  <sub>Play locally with a friend or online with Lichess!</a></sub>
</p>
<br />

![home](https://user-images.githubusercontent.com/57846867/182972093-c9917d32-81c3-4ba5-a6ec-a783bf254f33.png)



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
- go back to online after going back to start page from online page bug
- Handle all notation type or choose notation to start
- exit games w/o quit both local and lichess (save, resign)
- exit option on all loaders 
  
### local client 
- save local games
- add time options?
  

### oauth
  ✔️ check if token is expired and get new one if so
  - handle if oauth process rejected

### api

  - BULLET CHALLENGES BUG: when minutes are < 3 minutes for a realtime challenge, api provides an empty json for the GameFull event.

### online 
  - see and join ongoing games
    - api request
    - screen
  - see incoming and outgoing challenges, join game from accept
  - challenge construction 
    - set default interval position
    - custom starting position option for challenges
    - see variants on board
  
### online game client
  - post game options (rematch, new game, etc)
  - rating info next to name
  - chat
   
    

♞♞♞♞♞♞♞♞♞♞♞


<!-- # chess-cli ♟️ -->
<h1 align='center'>chess-cli ♟️</h1>
<p align="center">
  <b>A chess tui app written in Go.</b><br/>
  <sub>Play locally with a friend or online with Lichess!</a></sub>
</p>
<br />

<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>

chess-cli is a tui (terminal user interface) that allows you to play chess, both offline with a friend or online as a Lichess api client.



![home](https://user-images.githubusercontent.com/57846867/182972093-c9917d32-81c3-4ba5-a6ec-a783bf254f33.png)
![game](https://user-images.githubusercontent.com/42559123/183780181-a4cb62c0-8515-4371-8406-932e11e84f77.png)


<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>

# Features 

  
- Mouse input support 
- Displays move history
- UCI chess notation input
- Saving and restoring local games 
- Lichess features:
  - Lichess login w/ oauth
  - Lichess challenge request for real time, unlimited, or correspondence games, with all variants and options supported
  - Challenge a friend or a random player
  - Play against the stockfish AI
  - Join ongoing games
  - Join from your incoming challenges





<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>



### dependency documentation:

https://lichess.org/api for lichess

https://pkg.go.dev/github.com/notnil/chess  chess lib

https://docs.rocketnine.space/code.rocketnine.space/tslocum/cview/ cview

https://code.rocketnine.space/tslocum/cview/src/branch/master/demos cview demos

https://pkg.go.dev/net/http 

https://mholt.github.io/json-to-go/ 

<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>

### todo:

#### general 

- Handle all notation types or choose notation to start
  
#### local client 

- add time options
  
#### oauth

  - handle if oauth rejected mid process 

#### online 

- exit option on loader 
- challenge construction 
  - set default interval position for sliders
  - custom starting position option for challenges
- client
  - chat
  - rematch option post game
   
<p align="center">♞♞♞♞♞♞♞♞♞♞♞</p>


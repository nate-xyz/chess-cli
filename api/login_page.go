package api

import (
	"errors"
	"html/template"
	"io"
)

// The page the browser lands on after logging in to Lichess. Everything is
// inline (no fonts, scripts or images from elsewhere), so it works offline
// and tells nobody anything.

type loginPage struct {
	OK       bool
	Title    string
	Message  string
	Username string
	Glyph    string
}

func writeLoginPage(w io.Writer, username string, err error) error {
	p := loginPage{OK: err == nil, Username: username, Glyph: "♞"}
	switch {
	case err == nil:
		p.Title = "You're logged in"
		p.Message = "You can close this tab and go back to chess-cli in your terminal."
	case errors.Is(err, ErrLoginCancelled):
		p.Title, p.Glyph = "Login cancelled", "♚"
		p.Message = "chess-cli wasn't given access. Close this tab, and try again from chess-cli whenever you like."
	default:
		p.Title, p.Glyph = "Login failed", "♚"
		p.Message = err.Error() + ". Close this tab and try again from chess-cli."
	}
	return loginTemplate.Execute(w, p)
}

var loginTemplate = template.Must(template.New("login").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light dark">
<title>chess-cli · {{.Title}}</title>
<style>
:root {
  --bg: #f4efe6; --card: #fffdf8; --ink: #2b2620; --muted: #6f655a;
  --light: #f0d9b5; --dark: #b58863; --ok: #5a8f3c; --bad: #b5452f;
  --line: #e4dac9;
}
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #161512; --card: #262421; --ink: #e8e3da; --muted: #a39b8f;
    --ok: #7fb35b; --bad: #e0705a; --line: #3a3733;
  }
}
* { box-sizing: border-box; }
html, body { height: 100%; margin: 0; }
body {
  background: var(--bg); color: var(--ink);
  font: 16px/1.55 system-ui, -apple-system, "Segoe UI", Roboto, Ubuntu, sans-serif;
  display: grid; place-items: center; padding: 16px;
}
.card {
  width: 100%; max-width: 440px; background: var(--card);
  border: 1px solid var(--line); border-radius: 14px; overflow: hidden;
  box-shadow: 0 10px 30px rgba(0,0,0,.12);
}
.board {
  display: grid; grid-template-columns: repeat(8, 1fr); height: 28px;
}
.board span:nth-child(odd) { background: var(--light); }
.board span:nth-child(even) { background: var(--dark); }
.body { padding: 28px 28px 24px; text-align: center; }
.piece {
  width: 84px; height: 84px; margin: 0 auto 14px; border-radius: 50%;
  display: grid; place-items: center; font-size: 66px; line-height: 1;
  background: var(--light); color: #2b2620; position: relative;
}
.badge {
  position: absolute; right: -2px; bottom: -2px; width: 30px; height: 30px;
  border-radius: 50%; display: grid; place-items: center;
  color: #fff; font-size: 17px; font-weight: 700;
  border: 3px solid var(--card);
}
.ok .badge { background: var(--ok); }
.bad .badge { background: var(--bad); }
h1 { font-size: 22px; margin: 0 0 6px; }
.user { margin: 0 0 10px; color: var(--muted); }
.user b { color: var(--ink); }
p.msg { margin: 0 auto; max-width: 34ch; color: var(--muted); }
.term {
  margin: 22px auto 0; display: inline-block; text-align: left;
  font: 13px/1.5 ui-monospace, "SF Mono", Menlo, Consolas, monospace;
  background: #161512; color: #d8d2c6; border-radius: 8px; padding: 10px 14px;
}
.term .p { color: #7fb35b; }
footer { border-top: 1px solid var(--line); padding: 10px; text-align: center;
  font-size: 13px; color: var(--muted); }
</style>
</head>
<body>
<main class="card {{if .OK}}ok{{else}}bad{{end}}">
  <div class="board" aria-hidden="true"><span></span><span></span><span></span><span></span><span></span><span></span><span></span><span></span></div>
  <div class="body">
    <div class="piece" aria-hidden="true">{{.Glyph}}<span class="badge">{{if .OK}}✓{{else}}✕{{end}}</span></div>
    <h1>{{.Title}}</h1>
    {{if .Username}}<p class="user">Logged in to Lichess as <b>{{.Username}}</b></p>{{end}}
    <p class="msg">{{.Message}}</p>
    {{if .OK}}<div class="term"><span class="p">$</span> chess-cli <span style="opacity:.6">— ready to play</span></div>{{end}}
  </div>
  <footer>chess-cli · Lichess login</footer>
</main>
</body>
</html>
`))

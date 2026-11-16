package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeLichess points the API at a test server for the duration of a test.
func fakeLichess(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	old, oldUser := HostUrl, UserInfo
	HostUrl = srv.URL
	UserInfo = UserConfig{ApiToken: "test-token"}
	t.Cleanup(func() {
		srv.Close()
		HostUrl, UserInfo = old, oldUser
	})
	return srv
}

func TestResponseErrorMessages(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   string
	}{
		{400, `{"error":"Not your turn, or game already over"}`, "lichess: Not your turn, or game already over"},
		{400, `{"error":{"clock.limit":["Invalid value"],"days":["Required"]}}`, "lichess: clock.limit: Invalid value; days: Required"},
		{429, ``, "lichess: rate limited by Lichess, wait a minute and try again"},
		{404, `not json`, "lichess: 404 Not Found"},
	} {
		fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			fmt.Fprint(w, tc.body)
		}))
		err := MakeMove("abcd1234", "e2e4")
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Status != tc.status || err.Error() != tc.want {
			t.Errorf("status %d: got %v, want %q", tc.status, err, tc.want)
		}
	}
}

func TestRequestsSendAuthAndForm(t *testing.T) {
	var got *http.Request
	var form url.Values
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		r.ParseForm()
		form = r.PostForm
		fmt.Fprint(w, `{"ok":true}`)
	}))
	if err := WriteChat("game1234", "player", "good game"); err != nil {
		t.Fatal(err)
	}
	if got.URL.Path != "/api/board/game/game1234/chat" || got.Header.Get("Authorization") != "Bearer test-token" {
		t.Errorf("bad request: %s %v", got.URL.Path, got.Header)
	}
	if form.Get("room") != "player" || form.Get("text") != "good game" {
		t.Errorf("bad form: %v", form)
	}
}

func TestCreateChallengeKeepsAliveUntilDeclined(t *testing.T) {
	release := make(chan struct{})
	var form url.Values
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		form = r.PostForm
		fmt.Fprintln(w, `{"challenge":{"id":"chal1234","status":"created"}}`)
		w.(http.Flusher).Flush()
		<-release
		fmt.Fprintln(w, `{"done":"declined"}`)
	}))
	p, err := CreateChallenge(CreateChallengeType{
		DestUser: "friend", Variant: "standard", Rated: "false", Color: "white",
		TimeOption: TimeRealTime, ClockLimit: "600", ClockIncrement: "5",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "chal1234" {
		t.Fatalf("id %q", p.ID)
	}
	if form.Get("keepAliveStream") != "true" || form.Get("clock.limit") != "600" || form.Get("clock.increment") != "5" {
		t.Errorf("form %v", form)
	}
	select {
	case r := <-p.Done():
		t.Fatalf("finished early: %q", r)
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	if r := <-p.Done(); r != "declined" {
		t.Fatalf("done = %q", r)
	}
}

func TestCancelChallengeClosesConnection(t *testing.T) {
	closed := make(chan struct{})
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"challenge":{"id":"chal1234"}}`)
		w.(http.Flusher).Flush()
		<-r.Context().Done() // Lichess cancels the challenge when we hang up
		close(closed)
	}))
	p, err := CreateChallenge(CreateChallengeType{DestUser: "friend", TimeOption: TimeUnlimited})
	if err != nil {
		t.Fatal(err)
	}
	p.Cancel()
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("connection not closed by Cancel")
	}
}

func TestSeekUsesMinutes(t *testing.T) {
	var form url.Values
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		form = r.PostForm
		// real-time seeks stream keep-alive newlines until paired, then close
		fmt.Fprint(w, "\n")
	}))
	p, err := CreateSeek(CreateChallengeType{Variant: "standard", Rated: "true", Color: "random",
		TimeOption: TimeRealTime, ClockLimit: "900", MinTurn: 15, ClockIncrement: "10"})
	if err != nil {
		t.Fatal(err)
	}
	<-p.Done()
	if form.Get("time") != "15" || form.Get("increment") != "10" || form.Has("clock.limit") || form.Has("color") {
		t.Errorf("seek form %v", form)
	}
}

func TestStreamBoardStateEvents(t *testing.T) {
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, line := range []string{
			`{"type":"gameFull","id":"g1","white":{"id":"harry","name":"Harry"},"black":{"aiLevel":3},"initialFen":"startpos","state":{"type":"gameState","moves":"e2e4","wtime":60000,"btime":60000,"status":"started"}}`,
			``,
			`{"type":"chatLine","username":"lichess","text":"Good luck","room":"player"}`,
			`{"type":"somethingNew"}`,
			`{"type":"opponentGone","gone":true,"claimWinInSeconds":10}`,
			`{"type":"gameState","moves":"e2e4 e7e5","status":"started","bdraw":true}`,
		} {
			fmt.Fprintln(w, line)
		}
	}))
	events := make(chan BoardEvent, 10)
	errs := make(chan error, 1)
	StreamBoardState(context.Background(), events, errs, "g1")
	close(events)
	var got []BoardEvent
	for ev := range events {
		got = append(got, ev)
	}
	if len(errs) != 0 {
		t.Fatal(<-errs)
	}
	if len(got) != 5 {
		t.Fatalf("got %d events: %+v", len(got), got)
	}
	if got[0].Type != GameFull || got[0].Full.Black.AiLevel != 3 || got[0].Full.State.Moves != "e2e4" {
		t.Errorf("gameFull: %+v", got[0].Full)
	}
	if got[1].Type != ChatLine || got[1].Chat.Text != "Good luck" {
		t.Errorf("chat: %+v", got[1])
	}
	if got[2].Type != OpponentGone || !got[2].Gone.Gone || got[2].Gone.ClaimWinInSeconds != 10 {
		t.Errorf("gone: %+v", got[2])
	}
	if got[3].Type != GameState || !got[3].State.Bdraw {
		t.Errorf("state: %+v", got[3])
	}
	if got[4].Type != EOF {
		t.Errorf("last event should be EOF: %+v", got[4])
	}
}

func TestStreamBoardStateStopsOnCancel(t *testing.T) {
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		StreamBoardState(ctx, make(chan BoardEvent), make(chan error), "g1")
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not stop on cancel")
	}
}

// browserLogin plays the browser's part of the OAuth flow: it follows the
// login URL's redirect_uri with the given query (code/state/error).
func browserLogin(t *testing.T, authURL string, query func(state string) url.Values) {
	u, err := url.Parse(authURL)
	if err != nil {
		t.Error(err)
		return
	}
	q := u.Query()
	if q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" || q.Get("client_id") != ClientID {
		t.Errorf("bad auth url: %s", authURL)
	}
	go func() {
		res, err := http.Get(q.Get("redirect_uri") + "?" + query(q.Get("state")).Encode())
		if err == nil {
			res.Body.Close()
		}
	}()
}

func useTempConfig(t *testing.T) string {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir) // os.UserConfigDir on macOS
	t.Setenv("AppData", dir)
	return dir
}

func TestOAuthLoginSavesTokenPrivately(t *testing.T) {
	dir := useTempConfig(t)
	var tokenForm url.Values
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/token" {
			r.ParseForm()
			tokenForm = r.PostForm
			fmt.Fprint(w, `{"token_type":"Bearer","access_token":"lio_secret","expires_in":31536000}`)
		}
	}))
	UserInfo = UserConfig{}
	token, err := AuthUser(context.Background(), func(authURL string) {
		browserLogin(t, authURL, func(state string) url.Values {
			return url.Values{"code": {"the-code"}, "state": {state}}
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if token != "lio_secret" || tokenForm.Get("code") != "the-code" || tokenForm.Get("code_verifier") == "" {
		t.Fatalf("token %q form %v", token, tokenForm)
	}
	path := filepath.Join(dir, "chess-cli", "config.json")
	if runtimeConfig, err := ConfigDir(); err == nil {
		path = filepath.Join(runtimeConfig, "config.json")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 { // Windows has no Unix permissions
		t.Errorf("config perms %v, want 0600", info.Mode().Perm())
	}
	if time.Until(UserInfo.TokenExpirationDate) < 364*24*time.Hour {
		t.Errorf("expiry not taken from expires_in: %v", UserInfo.TokenExpirationDate)
	}
}

// Logging in (e.g. again, for rated puzzles) must keep the other settings.
func TestOAuthLoginKeepsSettings(t *testing.T) {
	useTempConfig(t)
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"new-token","expires_in":3600}`)
	}))
	if err := UpdateConfig(func(c *UserConfig) { c.Pieces, c.PieceSet, c.Zen = "image", "merida", true }); err != nil {
		t.Fatal(err)
	}
	if _, err := AuthUser(context.Background(), func(authURL string) {
		browserLogin(t, authURL, func(state string) url.Values {
			return url.Values{"code": {"c"}, "state": {state}}
		})
	}); err != nil {
		t.Fatal(err)
	}
	UserInfo = UserConfig{}
	if err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if UserInfo.ApiToken != "new-token" || UserInfo.Pieces != "image" || UserInfo.PieceSet != "merida" || !UserInfo.Zen {
		t.Errorf("settings after login: %+v", UserInfo)
	}
}

func TestOAuthLoginCancelled(t *testing.T) {
	useTempConfig(t)
	fakeLichess(t, http.NotFoundHandler())
	_, err := AuthUser(context.Background(), func(authURL string) {
		browserLogin(t, authURL, func(state string) url.Values {
			return url.Values{"error": {"access_denied"}, "state": {state}}
		})
	})
	if !errors.Is(err, ErrLoginCancelled) {
		t.Fatalf("got %v, want ErrLoginCancelled", err)
	}
}

func TestOAuthIgnoresWrongState(t *testing.T) {
	useTempConfig(t)
	fakeLichess(t, http.NotFoundHandler())
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	_, err := AuthUser(ctx, func(authURL string) {
		browserLogin(t, authURL, func(string) url.Values {
			return url.Values{"code": {"x"}, "state": {"forged"}}
		})
	})
	if err == nil || !strings.Contains(err.Error(), "timed out") && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("forged callback should be ignored until timeout, got %v", err)
	}
}

func TestLegacyConfigMigrated(t *testing.T) {
	useTempConfig(t)
	work := t.TempDir()
	t.Chdir(work)
	os.WriteFile(legacyConfigFile, []byte(`{"ApiToken":"old-token"}`), 0o644)
	UserInfo = UserConfig{}
	if err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if UserInfo.ApiToken != "old-token" {
		t.Fatalf("token %q", UserInfo.ApiToken)
	}
	if _, err := os.Stat(legacyConfigFile); !os.IsNotExist(err) {
		t.Errorf("world-readable legacy token file should be removed")
	}
}

// After Lichess rejects the token the event stream stops; logging in again
// must be able to start it again.
func TestEventStreamRestartsAfterRejection(t *testing.T) {
	useTempConfig(t)
	var calls atomic.Int32
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/stream/event" {
			if calls.Add(1) == 1 {
				http.Error(w, `{"error":"No such token"}`, http.StatusUnauthorized)
				return
			}
			fmt.Fprintln(w, `{"type":"gameStart","game":{"id":"abcd1234"}}`)
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		}
	}))
	events := make(chan StreamEventType, 4)
	StartEventStream(events)
	deadline := time.Now().Add(3 * time.Second)
	for streamRunning.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if streamRunning.Load() {
		t.Fatal("stream kept running after a 401")
	}
	StartEventStream(events) // e.g. after logging in again
	defer StopEventStream()
	select {
	case e := <-events:
		if e.GameID != "abcd1234" {
			t.Errorf("event %+v", e)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("stream didn't restart")
	}
}

// The login landing page: success names the account, failures say what
// went wrong (escaped), and nothing is loaded from elsewhere.
func TestLoginPage(t *testing.T) {
	var ok, failed, cancelled strings.Builder
	if err := writeLoginPage(&ok, "hfwitte", nil); err != nil {
		t.Fatal(err)
	}
	_ = writeLoginPage(&failed, "", errors.New("lichess: <bad> grant"))
	_ = writeLoginPage(&cancelled, "", ErrLoginCancelled)
	if !strings.Contains(ok.String(), "re logged in</h1>") || !strings.Contains(ok.String(), "<b>hfwitte</b>") {
		t.Error("success page missing title or username")
	}
	if !strings.Contains(failed.String(), "Login failed") || !strings.Contains(failed.String(), "&lt;bad&gt;") {
		t.Error("failure page missing title or not escaping the error")
	}
	if !strings.Contains(cancelled.String(), "Login cancelled") {
		t.Error("cancelled page missing title")
	}
	for _, page := range []string{ok.String(), failed.String()} {
		if strings.Contains(page, "src=") || strings.Contains(page, "http") {
			t.Error("login page loads something from elsewhere")
		}
	}
}

package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	pkce "github.com/jimlambrt/go-oauth-pkce-code-verifier"
	"github.com/skratchdot/open-golang/open"
)

// LoginTimeout is how long to wait for the browser login before giving up.
const LoginTimeout = 5 * time.Minute

// ErrLoginCancelled is returned when the user denies access on Lichess.
var ErrLoginCancelled = errors.New("login was cancelled on Lichess")

// ConfigDir is where chess-cli keeps its files (~/.config/chess-cli on Linux).
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "chess-cli"), nil
}

func configPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// configMu guards UserInfo against the login (which runs in the background)
// and saving. Settings are read freely on the UI goroutine; changes go
// through UpdateConfig/SetConfig so they're locked against a save in
// progress.
var configMu sync.Mutex

// UpdateConfig changes settings and saves them.
func UpdateConfig(change func(c *UserConfig)) error {
	configMu.Lock()
	defer configMu.Unlock()
	change(&UserInfo)
	return saveLocked()
}

// SetConfig changes settings without saving (e.g. previewing a choice).
func SetConfig(change func(c *UserConfig)) {
	configMu.Lock()
	defer configMu.Unlock()
	change(&UserInfo)
}

// ResetPreferences puts every setting back to its default and saves. The
// Lichess login is kept (log out separately); saved games aren't settings
// and aren't touched.
func ResetPreferences() error {
	configMu.Lock()
	defer configMu.Unlock()
	UserInfo = UserConfig{
		ApiToken:            UserInfo.ApiToken,
		TokenCreationDate:   UserInfo.TokenCreationDate,
		TokenExpirationDate: UserInfo.TokenExpirationDate,
	}
	return saveLocked()
}

// token is the saved Lichess token, if any.
func token() string {
	configMu.Lock()
	defer configMu.Unlock()
	return UserInfo.ApiToken
}

// setToken records a new token (keeping all other settings) and saves it.
func setToken(tok string, expiresIn int64) error {
	configMu.Lock()
	defer configMu.Unlock()
	UserInfo.ApiToken, UserInfo.TokenCreationDate, UserInfo.TokenExpirationDate = tok, time.Now(), time.Time{}
	if expiresIn > 0 {
		UserInfo.TokenExpirationDate = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
	return saveLocked()
}

// LoadConfig reads the saved settings and token, migrating the old
// ./user_config.json (which older versions wrote world-readable into the
// working directory).
func LoadConfig() error {
	configMu.Lock()
	defer configMu.Unlock()
	path, err := configPath()
	if err != nil {
		return err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if old, oldErr := os.ReadFile(legacyConfigFile); oldErr == nil {
			if json.Unmarshal(old, &UserInfo) == nil && UserInfo.ApiToken != "" {
				if err := saveLocked(); err != nil {
					return err
				}
				_ = os.Remove(legacyConfigFile)
			}
		}
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &UserInfo)
}

// SaveConfig writes the settings and token, readable only by the current user.
func SaveConfig() error {
	configMu.Lock()
	defer configMu.Unlock()
	return saveLocked()
}

func saveLocked() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(&UserInfo, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ClearToken forgets the saved token (e.g. after Lichess rejects it).
func ClearToken() error {
	configMu.Lock()
	defer configMu.Unlock()
	// forget the login, keep the preferences
	UserInfo.ApiToken, UserInfo.TokenCreationDate, UserInfo.TokenExpirationDate = "", time.Time{}, time.Time{}
	return saveLocked()
}

// HasValidToken reports whether a saved token exists and hasn't expired.
func HasValidToken() bool {
	configMu.Lock()
	defer configMu.Unlock()
	return UserInfo.ApiToken != "" &&
		(UserInfo.TokenExpirationDate.IsZero() || time.Now().Before(UserInfo.TokenExpirationDate))
}

// PerformOAuth returns a usable token: the saved one if still valid, otherwise
// it runs the browser login. showURL is called with the login URL so the UI
// can display it in case the browser doesn't open by itself.
// The settings were loaded at startup (LoadConfig).
func PerformOAuth(ctx context.Context, showURL func(string)) (string, error) {
	if HasValidToken() {
		return token(), nil
	}
	return AuthUser(ctx, showURL)
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// quickUsername is the account's name for the login page, if Lichess
// answers quickly; the page doesn't wait long for it.
func quickUsername() string {
	ch := make(chan string, 1)
	go func() {
		name, _ := GetUsername()
		ch <- name
	}()
	select {
	case name := <-ch:
		return name
	case <-time.After(2 * time.Second):
		return ""
	}
}

// AuthUser runs the OAuth2 PKCE flow: it opens the browser at Lichess, waits on
// a localhost redirect for the authorization code, and trades it for a token.
func AuthUser(ctx context.Context, showURL func(string)) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, LoginTimeout)
	defer cancel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("could not start local login server: %w", err)
	}
	redirectURL := fmt.Sprintf("http://%s/", ln.Addr().String())

	verifier, err := pkce.CreateCodeVerifier()
	if err != nil {
		ln.Close()
		return "", err
	}
	state, err := randomString(16)
	if err != nil {
		ln.Close()
		return "", err
	}
	authURL := HostUrl + "/oauth?" + url.Values{
		"response_type":         {"code"},
		"client_id":             {ClientID},
		"redirect_uri":          {redirectURL},
		"code_challenge_method": {"S256"},
		"code_challenge":        {verifier.CodeChallengeS256()},
		"scope":                 {strings.Join(Scopes, " ")},
		"state":                 {state},
	}.Encode()

	type result struct {
		token string
		err   error
	}
	results := make(chan result, 1)
	finish := func(w http.ResponseWriter, r result) {
		name := ""
		if r.err == nil {
			name = quickUsername()
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = writeLoginPage(w, name, r.err)
		select {
		case results <- r:
		default:
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case q.Get("state") != state:
			http.Error(w, "unexpected login callback", http.StatusBadRequest)
		case q.Get("error") != "":
			if q.Get("error") == "access_denied" {
				finish(w, result{err: ErrLoginCancelled})
			} else {
				finish(w, result{err: fmt.Errorf("lichess: %s", q.Get("error"))})
			}
		case q.Get("code") == "":
			finish(w, result{err: errors.New("no authorization code in callback")})
		default:
			tok, expiresIn, err := getAccessToken(ctx, q.Get("code"), verifier.String(), redirectURL)
			if err == nil {
				err = setToken(tok, expiresIn) // keeps the other settings
			}
			finish(w, result{token: tok, err: err})
		}
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go srv.Serve(ln)
	defer func() {
		shutdownCtx, c := context.WithTimeout(context.Background(), 2*time.Second)
		defer c()
		srv.Shutdown(shutdownCtx) // lets the success page finish sending
	}()

	if showURL != nil {
		showURL(authURL)
	}
	_ = open.Start(authURL) // if this fails the user can still use the shown URL

	select {
	case r := <-results:
		return r.token, r.err
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", errors.New("login timed out, try again")
		}
		return "", ctx.Err()
	}
}

// getAccessToken trades the authorization code for an access token.
func getAccessToken(ctx context.Context, code, codeVerifier, redirectURL string) (string, int64, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {codeVerifier},
		"redirect_uri":  {redirectURL},
		"client_id":     {ClientID},
	}
	req, err := http.NewRequestWithContext(ctx, "POST", HostUrl+"/api/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := apiClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", 0, responseError(res)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tok); err != nil {
		return "", 0, err
	}
	if tok.AccessToken == "" {
		return "", 0, errors.New("lichess returned no access token")
	}
	return tok.AccessToken, tok.ExpiresIn, nil
}

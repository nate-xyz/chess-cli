package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCleanString(t *testing.T) {
	for in, want := range map[string]string{
		"plain text":             "plain text",
		"gg\x1b]0;pwned\x07 wp":  "gg]0;pwned wp",
		"line one\nline two\tok": "line one\nline two\tok",
		"csi\u009b31m":           "csi31m",
		"abc\u202edcba":          "abcdcba",
		"ünïcødé ♞ 象棋":           "ünïcødé ♞ 象棋",
		"\x00\x01\x7f":           "",
	} {
		if got := cleanString(in); got != want {
			t.Errorf("cleanString(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeNested(t *testing.T) {
	type inner struct{ Name string }
	v := struct {
		Title   string
		List    []inner
		Ptr     *inner
		Map     map[string]string
		Structs map[string]inner
		Any     any
		private string
	}{
		Title:   "a\x1bb",
		List:    []inner{{"c\x1bd"}},
		Ptr:     &inner{"e\x1bf"},
		Map:     map[string]string{"k": "g\x1bh"},
		Structs: map[string]inner{"k": {"i\x1bj"}},
		Any:     map[string]any{"x": "k\x1bl"},
		private: "m\x1bn",
	}
	sanitize(&v)
	all := fmt.Sprint(v.Title, v.List, *v.Ptr, v.Map, v.Structs, v.Any)
	if strings.ContainsRune(all, 0x1b) {
		t.Errorf("control character left: %q", all)
	}
}

// End to end: a Lichess response with an escape sequence in a player's name
// arrives clean.
func TestResponsesAreSanitized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"username":"bad\u001b[2Jguy"}`)
	}))
	defer srv.Close()
	old := HostUrl
	HostUrl = srv.URL
	defer func() { HostUrl = old }()
	var out struct{ Username string }
	if err := call("GET", "/api/account", nil, &out); err != nil {
		t.Fatal(err)
	}
	if out.Username != "bad[2Jguy" {
		t.Errorf("username = %q", out.Username)
	}
}

func TestValidation(t *testing.T) {
	for _, ok := range []string{"DrNykterstein", "bob", "a_b-c9"} {
		if !ValidUsername(ok) {
			t.Errorf("%q rejected", ok)
		}
	}
	for _, bad := range []string{"", "x", "../account", "a/b", "bob?x=1", "has space", strings.Repeat("a", 31)} {
		if ValidUsername(bad) {
			t.Errorf("%q accepted as a username", bad)
		}
	}
	if !ValidGameID("q7ZvsdUF") || ValidGameID("q7ZvsdUF/") || ValidGameID("..") || ValidGameID("short") {
		t.Error("game ID validation")
	}
}

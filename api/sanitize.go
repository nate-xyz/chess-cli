package api

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"
)

// Everything decoded from the network goes through sanitize before the UI
// sees it. Text like chat, player names, blog titles and error messages ends
// up on the terminal, and the terminal library passes control characters
// straight through, so a string containing ESC could inject escape sequences
// (move the cursor, change the title, worse with some terminals). JSON can
// carry those as \u001b escapes, so they're removed after decoding.

// cleanString removes control characters (keeping newline and tab), C1
// controls, and Unicode bidirectional overrides (which can make text display
// differently from what it says).
func cleanString(s string) string {
	clean := true
	for _, r := range s {
		if unsafeRune(r) {
			clean = false
			break
		}
	}
	if clean {
		return s
	}
	return strings.Map(func(r rune) rune {
		if unsafeRune(r) {
			return -1
		}
		return r
	}, s)
}

func unsafeRune(r rune) bool {
	switch {
	case r == '\n' || r == '\t':
		return false
	case r < 0x20, r >= 0x7f && r <= 0x9f:
		return true
	case r >= 0x202a && r <= 0x202e, r >= 0x2066 && r <= 0x2069, r == 0x200e, r == 0x200f, r == 0x061c:
		return true
	}
	return false
}

// sanitize cleans every string reachable from v (a pointer), in place.
func sanitize(v any) {
	sanitizeValue(reflect.ValueOf(v), 0)
}

func sanitizeValue(v reflect.Value, depth int) {
	if depth > 32 || !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			if v.Kind() == reflect.Interface {
				// values inside interfaces aren't addressable: rebuild them
				inner := v.Elem()
				if inner.Kind() == reflect.String && v.CanSet() {
					v.Set(reflect.ValueOf(cleanString(inner.String())))
					return
				}
				if inner.Kind() == reflect.Map || inner.Kind() == reflect.Slice || inner.Kind() == reflect.Pointer {
					sanitizeValue(inner, depth+1)
				}
				return
			}
			sanitizeValue(v.Elem(), depth+1)
		}
	case reflect.String:
		if v.CanSet() {
			v.SetString(cleanString(v.String()))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				sanitizeValue(v.Field(i), depth+1)
			}
		}
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
			return // []byte / json.RawMessage: cleaned when decoded further
		}
		for i := 0; i < v.Len(); i++ {
			sanitizeValue(v.Index(i), depth+1)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			val := v.MapIndex(k)
			if val.Kind() == reflect.String {
				v.SetMapIndex(k, reflect.ValueOf(cleanString(val.String())).Convert(val.Type()))
				continue
			}
			if val.Kind() == reflect.Interface && val.Elem().Kind() == reflect.String {
				v.SetMapIndex(k, reflect.ValueOf(cleanString(val.Elem().String())))
				continue
			}
			if val.Kind() == reflect.Pointer || val.Kind() == reflect.Map || val.Kind() == reflect.Interface {
				sanitizeValue(val, depth+1)
			} else if val.Kind() == reflect.Slice {
				sanitizeValue(val, depth+1)
			} else if val.Kind() == reflect.Struct {
				// struct values in maps aren't addressable: copy, clean, store
				cp := reflect.New(val.Type()).Elem()
				cp.Set(val)
				sanitizeValue(cp, depth+1)
				v.SetMapIndex(k, cp)
			}
		}
	}
}

// unmarshal is json.Unmarshal followed by sanitize.
func unmarshal(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	sanitize(v)
	return nil
}

// decode reads the next JSON value from d into v and sanitizes it.
func decode(d *json.Decoder, v any) error {
	if err := d.Decode(v); err != nil {
		return err
	}
	sanitize(v)
	return nil
}

var (
	usernameRe = regexp.MustCompile(`^[A-Za-z0-9_-]{2,30}$`)
	gameIDRe   = regexp.MustCompile(`^[A-Za-z0-9]{8}$`)
)

// ErrBadUsername / ErrBadGameID: what was typed can't be a Lichess
// username or game ID (checked before it goes into a request's URL).
var (
	ErrBadUsername = errors.New("that isn't a valid Lichess username")
	ErrBadGameID   = errors.New("that isn't a valid Lichess game ID")
)

// ValidUsername reports whether s could be a Lichess username.
func ValidUsername(s string) bool { return usernameRe.MatchString(s) }

// ValidGameID reports whether s could be a Lichess game ID.
func ValidGameID(s string) bool { return gameIDRe.MatchString(s) }

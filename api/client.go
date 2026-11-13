package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var (
	// apiClient is for ordinary request/response calls.
	apiClient = &http.Client{Timeout: 30 * time.Second}
	// streamClient has no overall timeout: streams stay open for a whole game.
	streamClient = &http.Client{}
)

// APIError is a non-2xx response from Lichess, with its error message if the
// body had one.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("lichess: %d %s", e.Status, http.StatusText(e.Status))
	}
	return fmt.Sprintf("lichess: %s", e.Message)
}

// newRequest builds an authenticated request. form, if non-nil, is sent as an
// application/x-www-form-urlencoded body.
func newRequest(ctx context.Context, method, path string, form url.Values) (*http.Request, error) {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, HostUrl+path, body)
	if err != nil {
		return nil, err
	}
	if UserInfo.ApiToken != "" {
		req.Header.Set("Authorization", "Bearer "+UserInfo.ApiToken)
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req, nil
}

// call performs a request and, if out is non-nil, decodes the JSON response
// into it. Non-2xx responses become *APIError.
func call(method, path string, form url.Values, out any) error {
	req, err := newRequest(context.Background(), method, path, form)
	if err != nil {
		return err
	}
	res, err := apiClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return responseError(res)
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, res.Body)
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// openStream starts a streaming (ndjson) request. The caller must close the
// returned body; cancelling ctx also closes it.
func openStream(ctx context.Context, method, path string, form url.Values) (io.ReadCloser, error) {
	req, err := newRequest(ctx, method, path, form)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/x-ndjson")
	res, err := streamClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		defer res.Body.Close()
		return nil, responseError(res)
	}
	return res.Body, nil
}

// responseError turns an error response into an *APIError. Lichess sends
// {"error": "message"} or {"error": {"field": ["message", ...]}}.
func responseError(res *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	e := &APIError{Status: res.StatusCode}
	var parsed struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(body, &parsed) == nil && len(parsed.Error) > 0 {
		var msg string
		var fields map[string][]string
		switch {
		case json.Unmarshal(parsed.Error, &msg) == nil:
			e.Message = msg
		case json.Unmarshal(parsed.Error, &fields) == nil:
			var parts []string
			for field, msgs := range fields {
				parts = append(parts, field+": "+strings.Join(msgs, ", "))
			}
			sort.Strings(parts)
			e.Message = strings.Join(parts, "; ")
		}
	}
	if e.Message == "" && res.StatusCode == http.StatusTooManyRequests {
		e.Message = "rate limited by Lichess, wait a minute and try again"
	}
	return e
}

package api

import (
	"context"
	"errors"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Lichess studies: a user's list of studies, and a study's chapters as PGN
// (with variations, comments and the arrows and circles drawn in it).
// Private studies need the study:read scope, which the login asks for.

// StudyInfo is one study in a user's list.
type StudyInfo struct {
	ID      string
	Name    string
	Updated time.Time
}

// ErrBadStudyID: what was typed isn't a Lichess study link or ID.
var ErrBadStudyID = errors.New("that isn't a Lichess study link or ID")

var studyIDRe = regexp.MustCompile(`^[a-zA-Z0-9]{8}$`)

// ValidStudyID reports whether s looks like a study (or chapter) ID.
func ValidStudyID(s string) bool { return studyIDRe.MatchString(s) }

// StudyIDFrom takes a study link (lichess.org/study/abcd1234[/chapter]) or a
// bare ID and returns the study and chapter IDs (chapter may be "").
func StudyIDFrom(s string) (study, chapter string) {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "?#"); i >= 0 {
		s = s[:i]
	}
	if i := strings.Index(s, "/study/"); i >= 0 {
		s = s[i+len("/study/"):]
	}
	parts := strings.Split(strings.Trim(s, "/"), "/")
	study = parts[0]
	if len(parts) > 1 && ValidStudyID(parts[1]) {
		chapter = parts[1]
	}
	return study, chapter
}

// ListStudies lists a user's studies, most recently changed first. With a
// token for that user it includes their private studies.
func ListStudies(username string) ([]StudyInfo, error) {
	if !ValidUsername(username) {
		return nil, ErrBadUsername
	}
	req, err := newRequest(context.Background(), "GET", "/api/study/by/"+url.PathEscape(username), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/x-ndjson")
	res, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, responseError(res)
	}
	var out []StudyInfo
	err = eachLine(res.Body, func(line []byte) {
		var s struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			UpdatedAt int64  `json:"updatedAt"`
		}
		if unmarshal(line, &s) == nil && ValidStudyID(s.ID) {
			out = append(out, StudyInfo{ID: s.ID, Name: strings.TrimSpace(s.Name), Updated: time.UnixMilli(s.UpdatedAt)})
		}
	})
	sort.SliceStable(out, func(i, j int) bool { return out[i].Updated.After(out[j].Updated) })
	return out, err
}

// maxStudyPGN caps a study download (Lichess allows 64 chapters).
const maxStudyPGN = 16 << 20

// GetStudyPGN downloads all of a study's chapters as PGN.
func GetStudyPGN(id string) (string, error) {
	if !ValidStudyID(id) {
		return "", ErrBadStudyID
	}
	q := url.Values{"comments": {"true"}, "variations": {"true"}, "clocks": {"false"}, "orientation": {"true"}}
	req, err := newRequest(context.Background(), "GET", "/api/study/"+id+".pgn?"+q.Encode(), nil)
	if err != nil {
		return "", err
	}
	res, err := apiClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", responseError(res)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxStudyPGN))
	if err != nil {
		return "", err
	}
	return cleanString(string(body)), nil // no terminal escapes from comments
}

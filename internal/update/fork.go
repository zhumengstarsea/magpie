package update

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yetone/magpie/internal/source"
)

// ErrNoRelease means the fork has not published an installable release yet.
// It is not an invitation to fetch an upstream build over this fork.
var ErrNoRelease = errors.New("no releases published for this magpie fork")

func forkRequest(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "magpie")
	return source.Do(client, req)
}

func latestFork(ctx context.Context, lang string) (*Release, error) {
	if updateSite == "" {
		return nil, errors.New("update service not configured in this build")
	}
	res, err := forkRequest(ctx, withChannel(withLang(Feed(), lang)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNoRelease
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fork update feed: %s", res.Status)
	}
	var published Release
	if err := json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&published); err != nil {
		return nil, fmt.Errorf("fork update feed: %w", err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(published.Version), "v")
	if parse(version) == nil {
		return nil, fmt.Errorf("fork update feed: no published release version")
	}
	rel := &Release{Version: version, Notes: InLang(published.Notes, lang), URL: ReleasePage(version), Assets: map[string]Asset{}}
	for name, a := range published.Assets {
		sum := a.SHA256
		decoded, err := hex.DecodeString(sum)
		if err != nil || len(decoded) != 32 || !strings.HasPrefix(a.URL, Site+"/releases/download/v"+version+"/") {
			continue
		}
		rel.Assets[name] = Asset{URL: a.URL, Size: a.Size, SHA256: strings.ToLower(sum)}
	}
	return rel, nil
}

func forkNotes(ctx context.Context, after, upto, lang string) ([]Note, error) {
	if updateSite == "" {
		return nil, errors.New("update service not configured in this build")
	}
	q := url.Values{"upto": {upto}}
	if after != "" {
		q.Set("after", after)
	}
	if lang != "" {
		q.Set("lang", lang)
	}
	res, err := forkRequest(ctx, NotesFeed()+"?"+q.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fork release notes: %s", res.Status)
	}
	var releases struct {
		Releases []Note `json:"releases"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(&releases); err != nil {
		return nil, fmt.Errorf("fork release notes: %w", err)
	}
	var notes []Note
	for _, r := range releases.Releases {
		if parse(r.Version) == nil {
			continue
		}
		version := strings.TrimPrefix(strings.TrimSpace(r.Version), "v")
		notes = append(notes, Note{Version: version, Notes: InLang(r.Notes, lang), URL: ReleasePage(version)})
	}
	return Between(notes, after, upto), nil
}

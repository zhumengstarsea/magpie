package update

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/yetone/magpie/internal/source"
)

// ErrNoRelease means the fork has not published an installable release yet.
// It is not an invitation to fetch an upstream build over this fork.
var ErrNoRelease = errors.New("no releases published for this magpie fork")

type forkRelease struct {
	Tag        string `json:"tag_name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name   string `json:"name"`
		URL    string `json:"browser_download_url"`
		Size   int64  `json:"size"`
		Digest string `json:"digest"`
	} `json:"assets"`
}

func forkRequest(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	req.Header.Set("User-Agent", "magpie-native-pool")
	return source.Do(client, req)
}

func latestFork(ctx context.Context, lang string) (*Release, error) {
	res, err := forkRequest(ctx, Feed())
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
	var published forkRelease
	if err := json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&published); err != nil {
		return nil, fmt.Errorf("fork update feed: %w", err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(published.Tag), "v")
	if parse(version) == nil || published.Draft || published.Prerelease {
		return nil, fmt.Errorf("fork update feed: no published release version")
	}
	rel := &Release{Version: version, Notes: InLang(published.Body, lang), URL: ReleasePage(version), Assets: map[string]Asset{}}
	for _, a := range published.Assets {
		sum, ok := strings.CutPrefix(a.Digest, "sha256:")
		decoded, err := hex.DecodeString(sum)
		if !ok || err != nil || len(decoded) != 32 || !strings.HasPrefix(a.URL, Site+"/releases/download/") {
			continue
		}
		rel.Assets[a.Name] = Asset{URL: a.URL, Size: a.Size, SHA256: strings.ToLower(sum)}
	}
	return rel, nil
}

func forkNotes(ctx context.Context, after, upto, lang string) ([]Note, error) {
	res, err := forkRequest(ctx, NotesFeed()+"?per_page=100")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fork release notes: %s", res.Status)
	}
	var releases []forkRelease
	if err := json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(&releases); err != nil {
		return nil, fmt.Errorf("fork release notes: %w", err)
	}
	var notes []Note
	for _, r := range releases {
		if r.Draft || r.Prerelease || parse(r.Tag) == nil {
			continue
		}
		version := strings.TrimPrefix(strings.TrimSpace(r.Tag), "v")
		notes = append(notes, Note{Version: version, Notes: InLang(r.Body, lang), URL: ReleasePage(version)})
	}
	return Between(notes, after, upto), nil
}

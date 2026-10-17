package update

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type forkTransport func(*http.Request) (*http.Response, error)

func (f forkTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func forkFixture(t *testing.T, status int, body string, seen *[]string) {
	t.Helper()
	t.Setenv("MAGPIE_UPDATE_FEED", "")
	t.Setenv("MAGPIE_NOTES_FEED", "")
	t.Setenv("MAGPIE_MIRRORS", "")
	old := client
	client = &http.Client{Transport: forkTransport(func(r *http.Request) (*http.Response, error) {
		*seen = append(*seen, r.URL.String())
		if !strings.HasPrefix(r.URL.String(), releasesAPI) {
			t.Errorf("unexpected update source %s", r.URL)
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("not a GitHub release request: %v", r.Header)
		}
		return &http.Response{StatusCode: status, Status: http.StatusText(status), Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})}
	t.Cleanup(func() { client = old })
}

func TestForkReleaseFeed(t *testing.T) {
	sum := strings.Repeat("ab", 32)
	var seen []string
	body := `{"tag_name":"v0.1.1276-native.2","body":"English notes\n<!-- lang:zh -->\n中文说明","assets":[{"name":"magpie-windows-amd64.exe","browser_download_url":"https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1276-native.2/magpie-windows-amd64.exe","size":42,"digest":"sha256:` + sum + `"},{"name":"unsigned.exe","browser_download_url":"https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1276-native.2/unsigned.exe"},{"name":"upstream.exe","browser_download_url":"https://github.com/yetone/magpie-releases/releases/download/v9.9.9/magpie.exe","digest":"sha256:` + sum + `"},{"name":"invalid.exe","browser_download_url":"https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1276-native.2/invalid.exe","digest":"sha256:nope"}]}`
	forkFixture(t, 200, body, &seen)
	rel, err := LatestIn(context.Background(), "zh")
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 || seen[0] != releasesAPI+"/latest" {
		t.Fatalf("requests %v", seen)
	}
	if rel.Version != "0.1.1276-native.2" || rel.Notes != "中文说明" || rel.URL != Site+"/releases/tag/v0.1.1276-native.2" {
		t.Fatalf("release %+v", rel)
	}
	if len(rel.Assets) != 1 || rel.Assets["magpie-windows-amd64.exe"].SHA256 != sum || rel.Assets["magpie-windows-amd64.exe"].Size != 42 {
		t.Fatalf("unsafe or missing assets %+v", rel.Assets)
	}
	if Feed() != releasesAPI+"/latest" || NotesFeed() != releasesAPI {
		t.Fatalf("feed=%s notes=%s", Feed(), NotesFeed())
	}
}

func TestForkWithoutRelease(t *testing.T) {
	var seen []string
	forkFixture(t, 404, `{"message":"Not Found"}`, &seen)
	rel, err := Latest(context.Background())
	if rel != nil || !errors.Is(err, ErrNoRelease) || len(seen) != 1 {
		t.Fatalf("release=%v error=%v requests=%v", rel, err, seen)
	}
}

func TestForkReleaseNotes(t *testing.T) {
	var seen []string
	forkFixture(t, 200, `[{"tag_name":"v0.1.1276-native.2","body":"- English\n<!-- lang:zh -->\n- 中文"},{"tag_name":"v0.1.1276-native.3","body":"- draft","draft":true},{"tag_name":"v0.1.1276-native.1","body":"- earlier"}]`, &seen)
	notes, err := NotesBetween(context.Background(), "0.1.1276-native.1", "0.1.1276-native.2", "zh")
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0].Notes != "- 中文" || notes[0].URL != ReleasePage("0.1.1276-native.2") {
		t.Fatalf("notes %+v", notes)
	}
	if len(seen) != 1 || seen[0] != releasesAPI+"?per_page=100" {
		t.Fatalf("requests %v", seen)
	}
}

func TestForkRejectsInvalidRelease(t *testing.T) {
	for _, body := range []string{`{`, `{"tag_name":"not-a-version"}`, `{"tag_name":"v1.0.0","draft":true}`, `{"tag_name":"v1.0.0","prerelease":true}`} {
		t.Run(body, func(t *testing.T) {
			var seen []string
			forkFixture(t, 200, body, &seen)
			if _, err := Latest(context.Background()); err == nil {
				t.Fatal("invalid release accepted")
			}
		})
	}
}

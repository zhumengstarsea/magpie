package update

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/settings"
)

type forkTransport func(*http.Request) (*http.Response, error)

func (f forkTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func forkFixture(t *testing.T, status int, body string, seen *[]string) {
	t.Helper()
	before := updateSite
	updateSite = "https://updates.example.invalid"
	t.Cleanup(func() { updateSite = before })
	t.Setenv("MAGPIE_UPDATE_FEED", "")
	t.Setenv("MAGPIE_NOTES_FEED", "")
	t.Setenv("MAGPIE_MIRRORS", "")
	old := client
	client = &http.Client{Transport: forkTransport(func(r *http.Request) (*http.Response, error) {
		*seen = append(*seen, r.URL.String())
		if !strings.HasPrefix(r.URL.String(), updateSite+"/api/") {
			t.Errorf("unexpected update source %s", r.URL)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("not an update service request: %v", r.Header)
		}
		return &http.Response{StatusCode: status, Status: http.StatusText(status), Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})}
	t.Cleanup(func() { client = old })
}

func TestForkReleaseFeed(t *testing.T) {
	sum := strings.Repeat("ab", 32)
	var seen []string
	body := `{"version":"0.1.1661","notes":"English notes\n<!-- lang:zh -->\n中文说明","assets":{"magpie-windows-amd64.exe":{"url":"https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1661/magpie-windows-amd64.exe","size":42,"sha256":"` + sum + `"},"unsigned.exe":{"url":"https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1661/unsigned.exe"},"upstream.exe":{"url":"https://github.com/yetone/magpie-releases/releases/download/v9.9.9/magpie.exe","sha256":"` + sum + `"},"wrong-version.exe":{"url":"https://github.com/zhumengstarsea/magpie/releases/download/v9.9.9/magpie.exe","sha256":"` + sum + `"},"invalid.exe":{"url":"https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1661/invalid.exe","sha256":"nope"}}}`
	stableFor(t, "")
	forkFixture(t, 200, body, &seen)
	rel, err := LatestIn(context.Background(), "zh")
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 || seen[0] != updateSite+"/api/latest?lang=zh" {
		t.Fatalf("requests %v", seen)
	}
	if rel.Version != "0.1.1661" || rel.Notes != "中文说明" || rel.URL != Site+"/releases/tag/v0.1.1661" {
		t.Fatalf("release %+v", rel)
	}
	if len(rel.Assets) != 1 || rel.Assets["magpie-windows-amd64.exe"].SHA256 != sum || rel.Assets["magpie-windows-amd64.exe"].Size != 42 {
		t.Fatalf("unsafe or missing assets %+v", rel.Assets)
	}
	if Feed() != updateSite+"/api/latest" || NotesFeed() != updateSite+"/api/notes" {
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
	forkFixture(t, 200, `{"releases":[{"version":"0.1.1276-native.2","notes":"- English\n<!-- lang:zh -->\n- 中文"},{"version":"invalid","notes":"- invalid"},{"version":"0.1.1276-native.1","notes":"- earlier"}]}`, &seen)
	notes, err := NotesBetween(context.Background(), "0.1.1276-native.1", "0.1.1276-native.2", "zh")
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0].Notes != "- 中文" || notes[0].URL != ReleasePage("0.1.1276-native.2") {
		t.Fatalf("notes %+v", notes)
	}
	if len(seen) != 1 || seen[0] != updateSite+"/api/notes?after=0.1.1276-native.1&lang=zh&upto=0.1.1276-native.2" {
		t.Fatalf("requests %v", seen)
	}
}

func TestForkRejectsInvalidRelease(t *testing.T) {
	for _, body := range []string{`{`, `{"version":"not-a-version"}`, `{"tag_name":"v1.0.0"}`} {
		t.Run(body, func(t *testing.T) {
			var seen []string
			forkFixture(t, 200, body, &seen)
			if _, err := Latest(context.Background()); err == nil {
				t.Fatal("invalid release accepted")
			}
		})
	}
}

// Unlike a custom test feed, the default source used to bypass the channel.
func TestForkStableChannelUsesService(t *testing.T) {
	var seen []string
	stableFor(t, settings.UpdateStable)
	forkFixture(t, 200, `{"version":"0.1.1326","notes":"稳定版"}`, &seen)
	rel, err := LatestIn(context.Background(), "zh")
	if err != nil {
		t.Fatal(err)
	}
	if rel.Version != "0.1.1326" || len(seen) != 1 || seen[0] != updateSite+"/api/latest?channel=stable&lang=zh" {
		t.Fatalf("release=%+v requests=%v", rel, seen)
	}
}

func TestForkWithoutConfiguredService(t *testing.T) {
	before := updateSite
	updateSite = ""
	t.Cleanup(func() { updateSite = before })
	var seen []string
	old := client
	client = &http.Client{Transport: forkTransport(func(r *http.Request) (*http.Response, error) {
		seen = append(seen, r.URL.String())
		return nil, errors.New("unexpected network request")
	})}
	t.Cleanup(func() { client = old })
	if _, err := latestFork(context.Background(), ""); err == nil || len(seen) != 0 {
		t.Fatalf("unconfigured update: error=%v requests=%v", err, seen)
	}
	if _, err := forkNotes(context.Background(), "", "0.1.1661", ""); err == nil || len(seen) != 0 {
		t.Fatalf("unconfigured notes: error=%v requests=%v", err, seen)
	}
}

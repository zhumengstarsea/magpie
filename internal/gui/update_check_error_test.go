package gui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A failed check must give the reason without exposing the configured feed
// in the version row or its tooltip (both use the API's error field).
func TestUpdateCheckHidesFeedAddress(t *testing.T) {
	sandboxHome(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Errorf("hijack: %v", err)
			return
		}
		conn.Close()
	}))
	defer srv.Close()
	feed := srv.URL + "/private/feed?key=private-value"
	t.Setenv("MAGPIE_UPDATE_FEED", feed)
	t.Setenv("MAGPIE_MIRRORS", "")
	u := &updater{}
	u.check()
	j := u.json()
	if j.State != "error" || !strings.Contains(j.Error, "EOF") {
		t.Fatalf("expected the connection error, got %+v", j)
	}
	if strings.Contains(j.Error, srv.URL) || strings.Contains(j.Error, "private-value") || strings.Contains(j.Error, "/private/feed") {
		t.Fatalf("update error reveals the feed address: %q", j.Error)
	}
}

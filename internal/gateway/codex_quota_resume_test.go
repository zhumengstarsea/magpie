package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/access"
	"github.com/yetone/magpie/internal/provider"
)

func quotaResumeTest(t *testing.T) {
	t.Helper()
	codexSignedIn(t, "spare@example.com")
	oldMax, oldPause := codexQuotaResumeMax, retryPause
	codexQuotaResumeMax, retryPause = 3*time.Second, time.Millisecond
	t.Cleanup(func() { codexQuotaResumeMax, retryPause = oldMax, oldPause })
}

func quotaReply(w http.ResponseWriter) {
	w.WriteHeader(http.StatusTooManyRequests)
	io.WriteString(w, `{"error":{"type":"usage_limit_reached","message":"The usage limit has been reached","resets_in_seconds":1}}`)
}

func quotaPong(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	io.WriteString(w, sse(
		`data: {"type":"response.created","response":{"id":"r1","model":"gpt-5.5"}}`,
		`data: {"type":"response.output_text.delta","delta":"pong"}`,
		`data: {"type":"response.completed","response":{"id":"r1","usage":{"input_tokens":7,"output_tokens":1}}}`))
}

func quotaBackend(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	up := httptest.NewServer(handler)
	t.Cleanup(up.Close)
	old := provider.CodexBase
	provider.CodexBase = up.URL + "/backend-api/codex"
	t.Cleanup(func() { provider.CodexBase = old })
}

// A pool already resting is reduced to one probe. Its one remaining candidate
// must still be able to wait and resume the original pool.
func TestCodexQuotaResumeAfterSingleProbe(t *testing.T) {
	quotaResumeTest(t)
	var asks atomic.Int32
	quotaBackend(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		if asks.Add(1) == 1 {
			quotaReply(w)
			return
		}
		quotaPong(w)
	})
	p, err := provider.Find("codex")
	if err != nil {
		t.Fatal(err)
	}
	s := New()
	cs, _ := s.plan(*p, "gpt-5.5", provider.Responses)
	if len(cs) != 2 {
		t.Fatalf("pool: %d accounts", len(cs))
	}
	for _, c := range cs {
		s.restAfter(c, 429, nil, []byte(`{"error":{"type":"usage_limit_reached","resets_in_seconds":1}}`))
	}
	code, body := codexPost(t, `{"model":"gpt-5.5","stream":true,"input":"ping"}`)
	if code != 200 || !strings.Contains(body, "pong") || asks.Load() != 2 {
		t.Fatalf("single probe did not resume: status=%d asks=%d body=%s", code, asks.Load(), body)
	}
}

// A transport failure can rest a nearly spent account until its allowance
// window resets. Its "other/window" rest must not veto a pool's quota wait.
func TestCodexQuotaResumeWithWindowRest(t *testing.T) {
	quotaResumeTest(t)
	reset := time.Now().Add(time.Second)
	old := allowances
	t.Cleanup(func() { allowances = old })
	allowances = func(string) map[string]provider.Allowance {
		used := 98.0
		if !time.Now().Before(reset) {
			used = 0
		}
		return map[string]provider.Allowance{
			"me@example.com":    {{Used: 100, Resets: reset, Span: 5 * time.Hour}},
			"spare@example.com": {{Used: used, Resets: reset.Add(time.Second), Span: 5 * time.Hour}},
		}
	}
	var asks atomic.Int32
	quotaBackend(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		asks.Add(1)
		if time.Now().Before(reset) {
			quotaReply(w)
			return
		}
		quotaPong(w)
	})
	p, err := provider.Find("codex")
	if err != nil {
		t.Fatal(err)
	}
	s := New()
	cs, _ := s.plan(*p, "gpt-5.5", provider.Responses)
	if len(cs) != 2 {
		t.Fatalf("pool: %d accounts", len(cs))
	}
	for _, c := range cs {
		if c.p.Account.User == "spare@example.com" {
			rest := s.restAfter(c, 502, nil, []byte(`upstream EOF`))
			if rest.Why != failOther || rest.By != "window" {
				t.Fatalf("wanted the reporter's other/window rest, got %+v", rest)
			}
		} else {
			s.restAfter(c, 429, nil, []byte(`{"error":{"type":"usage_limit_reached","resets_in_seconds":1}}`))
		}
	}
	code, body := codexPost(t, `{"model":"gpt-5.5","stream":true,"input":"ping"}`)
	if code != 200 || !strings.Contains(body, "pong") || asks.Load() < 2 {
		t.Fatalf("mixed window rest ended the request: status=%d asks=%d body=%s", code, asks.Load(), body)
	}
}

func TestCodexQuotaWaitDoesNotHideOtherFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		used   float64
	}{
		{"invalid login", http.StatusUnauthorized, 98},
		{"forbidden", http.StatusForbidden, 98},
		{"transport failure with allowance left", http.StatusBadGateway, 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			quotaResumeTest(t)
			capUsage(t, map[string]float64{"me@example.com": 100, "spare@example.com": tc.used})
			p, err := provider.Find("codex")
			if err != nil {
				t.Fatal(err)
			}
			s := New()
			cs, _ := s.plan(*p, "gpt-5.5", provider.Responses)
			if len(cs) != 2 {
				t.Fatalf("pool: %d accounts", len(cs))
			}
			for _, c := range cs {
				if c.p.Account.User == "spare@example.com" {
					s.restAfter(c, tc.status, nil, []byte(`upstream failure`))
				} else {
					s.restAfter(c, 429, nil, []byte(`{"error":{"type":"usage_limit_reached","resets_in_seconds":1}}`))
				}
			}
			if at, wait := codexQuotaBack(cs); wait {
				t.Fatalf("unrelated failure entered quota wait until %s", at)
			}
		})
	}
}

func mixedQuotaRequest(t *testing.T, s *Server) (int, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(
		`{"model":"group/quota-resume","stream":true,"input":"ping"}`)).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer magpie-codex")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func quotaRelay(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	up := httptest.NewServer(handler)
	t.Cleanup(up.Close)
	if err := provider.Save(provider.Provider{
		ID: "relay", Name: "Relay", Key: "test", Models: []string{"m"}, Responses: up.URL + "/v1",
	}); err != nil {
		t.Fatal(err)
	}
	if err := provider.SaveGroup(provider.Group{
		ID: "quota-resume", Name: "Quota resume",
		Members: []string{"codex/gpt-5.5", "relay/m"}, Routing: provider.Ordered, Patience: -1,
	}); err != nil {
		t.Fatal(err)
	}
}

// Quota recovery still applies after the final candidate is a third-party
// relay: the Codex accounts remain the same request's first choice when back.
func TestCodexQuotaResumeWithFailedRelay(t *testing.T) {
	quotaResumeTest(t)
	var asks atomic.Int32
	quotaBackend(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		if asks.Add(1) <= 2 {
			quotaReply(w)
			return
		}
		quotaPong(w)
	})
	quotaRelay(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusBadGateway)
		io.WriteString(w, `{"error":{"message":"upstream EOF"}}`)
	})
	code, body := mixedQuotaRequest(t, New())
	if code != 200 || !strings.Contains(body, "pong") {
		t.Fatalf("mixed route did not resume: status=%d asks=%d body=%s", code, asks.Load(), body)
	}
}

// A relay with no response must not monopolize the request past a known
// official quota reset. Its attempt is canceled, not the client's request.
func TestCodexQuotaResumePreemptsSilentRelay(t *testing.T) {
	quotaResumeTest(t)
	var asks atomic.Int32
	quotaBackend(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		if asks.Add(1) <= 2 {
			quotaReply(w)
			return
		}
		quotaPong(w)
	})
	relayCanceled := make(chan struct{})
	quotaRelay(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		<-r.Context().Done()
		close(relayCanceled)
	})
	code, body := mixedQuotaRequest(t, New())
	if code != 200 || !strings.Contains(body, "pong") {
		t.Fatalf("silent relay blocked recovery: status=%d asks=%d body=%s", code, asks.Load(), body)
	}
	select {
	case <-relayCanceled:
	case <-time.After(5 * time.Second):
		t.Fatal("relay request was not canceled after switching back")
	}
}

// A real SSE event, unlike a comment, resets Codex's stream idle timeout.
// It carries no assistant text, response id, tool call or completion.
func TestCodexQuotaHeartbeatIsAnEvent(t *testing.T) {
	rec := httptest.NewRecorder()
	h := newHoldWriter(rec, true)
	h.alive = &keptAlive{proto: provider.Responses, streams: true, events: true}
	h.keepAlive()
	body := rec.Body.String()
	if !strings.Contains(body, "data: {\"type\":\"magpie.keepalive\"}\n\n") {
		t.Fatalf("Codex receives only a comment, not an event: %q", body)
	}
	for _, forbidden := range []string{"output_text", "response.created", "response.completed", "tool_call"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("heartbeat invents response content: %q", body)
		}
	}
}

func TestCodexQuotaResumeHonorsAccountLimit(t *testing.T) {
	quotaResumeTest(t)
	var asks atomic.Int32
	quotaBackend(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		if got := r.Header.Get("chatgpt-account-id"); got != "acct-2" {
			t.Errorf("request escaped key's account limit: %q", got)
		}
		if asks.Add(1) == 1 {
			quotaReply(w)
			return
		}
		quotaPong(w)
	})
	keys, secrets := newCaller(t, "Only spare")
	if _, err := access.Update("accounts-key", access.Change{
		Key: keys[0].ID, Accounts: []string{"codex/spare@example.com"},
	}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(
		`{"model":"codex/gpt-5.5","stream":true,"input":"ping"}`))
	req.Header.Set("Authorization", "Bearer "+secrets[0])
	rec := httptest.NewRecorder()
	New().Handler().ServeHTTP(rec, req)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "pong") || asks.Load() != 2 {
		t.Fatalf("restricted account did not resume: %d asks=%d %s", rec.Code, asks.Load(), rec.Body.String())
	}
}

func TestCodexQuotaResumeDoesNotReplaceReplyInProgress(t *testing.T) {
	quotaResumeTest(t)
	var asks atomic.Int32
	quotaBackend(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		asks.Add(1)
		quotaReply(w)
	})
	quotaRelay(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, sse(
			`data: {"type":"response.created","response":{"id":"relay"}}`,
			`data: {"type":"response.output_text.delta","delta":"relay answer"}`))
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
			return
		case <-time.After(1500 * time.Millisecond):
		}
		io.WriteString(w, sse(`data: {"type":"response.completed","response":{"id":"relay","status":"completed"}}`))
	})
	code, body := mixedQuotaRequest(t, New())
	if code != 200 || !strings.Contains(body, "relay answer") || asks.Load() != 2 ||
		strings.Count(body, `"type":"response.created"`) != 1 {
		t.Fatalf("reply replaced after content: %d asks=%d %s", code, asks.Load(), body)
	}
}

func TestCodexQuotaResumeStopsWhenClientLeaves(t *testing.T) {
	quotaResumeTest(t)
	var asks atomic.Int32
	quotaBackend(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		asks.Add(1)
		quotaReply(w)
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	quotaRelay(t, func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		cancel()
	})
	req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(
		`{"model":"group/quota-resume","stream":true,"input":"ping"}`)).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer magpie-codex")
	s := New()
	s.Handler().ServeHTTP(httptest.NewRecorder(), req)
	if got := lastRoute(s); got.Status != 499 || asks.Load() != 2 {
		t.Fatalf("canceled request resumed: status=%d asks=%d", got.Status, asks.Load())
	}
}

func TestCodexQuotaHeartbeatDoesNotSplitPartialUpstreamFrame(t *testing.T) {
	rec := httptest.NewRecorder()
	h := newHoldWriter(rec, true)
	h.alive = &keptAlive{proto: provider.Responses, streams: true, events: true}
	h.passing = true
	h.pulse()
	if got := rec.Body.String(); got != ": keepalive\n" {
		t.Fatalf("data event inserted into an in-progress upstream frame: %q", got)
	}
}

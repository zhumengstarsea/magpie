package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	"github.com/yetone/magpie/internal/provider"
)

// ClinePass's DeepSeek models are served by DeepSeek's own API alone when
// the provider pins them (White Immortal on Discord, as ClinePass +
// ProviderPin does): each request for one carries
// providerOptions.gateway.only ["deepseek"], on Chat and when an Anthropic
// or Responses request is built into one. Its GLM models are pinned to
// Z.ai's the same way, ["zai"] (sdytman on Discord). Its other makers'
// models, Cline's free ones, a provider that doesn't pin, and a body
// naming providerOptions itself go as they were.
func TestClinePinsDeepSeekAndGLM(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	cline, err := provider.FromPreset("clinepass")
	if err != nil {
		t.Fatal(err)
	}
	cline.ID, cline.Key, cline.PinUpstream = "cline", "clp_key", true
	cline.Models = []string{"cline-pass/deepseek-v4-pro", "cline-pass/glm-5.3", "cline-pass/glm-5.3-flash", "cline-pass/kimi-k3", "cline-pass/qwen3.8-max", "cline-free/deepseek-v4.1-flash", "cline-free/glm-5.3-flash"}
	plain := cline
	plain.ID, plain.Name, plain.PinUpstream = "cline2", "ClinePass 2", false
	for _, p := range []provider.Provider{cline, plain} {
		if err := provider.Save(p); err != nil {
			t.Fatal(err)
		}
	}
	var sent []byte
	s := New()
	s.client = &http.Client{Transport: countTransport(func(r *http.Request) (*http.Response, error) {
		sent, _ = io.ReadAll(r.Body)
		if gjson.GetBytes(sent, "stream").Bool() {
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}},
				Body: io.NopCloser(strings.NewReader("data: {\"id\":\"c\",\"object\":\"chat.completion.chunk\",\"model\":\"m\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"pong\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"))}, nil
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{"id":"c","object":"chat.completion","model":"m","choices":[{"index":0,"message":{"role":"assistant","content":"pong"},"finish_reason":"stop"}]}`))}, nil
	})}
	ask := func(path, body string) string {
		t.Helper()
		sent = nil
		req := httptest.NewRequest("POST", path, strings.NewReader(body))
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Fatalf("%s %s: %d %s", path, body, rec.Code, rec.Body)
		}
		return gjson.GetBytes(sent, "providerOptions").Raw
	}
	chat := func(model, extra string) string {
		return ask("/v1/chat/completions", `{"model":"`+model+`","messages":[{"role":"user","content":"Say pong"}]`+extra+`}`)
	}
	const pinned = `{"gateway":{"only":["deepseek"]}}`
	if got := chat("cline/cline-pass/deepseek-v4-pro", ""); got != pinned {
		t.Errorf("chat: providerOptions %s, want %s (sent %s)", got, pinned, sent)
	}
	if got := ask("/v1/messages", `{"model":"cline/cline-pass/deepseek-v4-pro","max_tokens":64,"messages":[{"role":"user","content":"Say pong"}]}`); got != pinned {
		t.Errorf("messages: providerOptions %s, want %s", got, pinned)
	}
	if got := ask("/v1/responses", `{"model":"cline/cline-pass/deepseek-v4-pro","input":"Say pong","stream":false}`); got != pinned {
		t.Errorf("responses: providerOptions %s, want %s", got, pinned)
	}
	const zai = `{"gateway":{"only":["zai"]}}`
	for _, m := range []string{"cline/cline-pass/glm-5.3", "cline/cline-pass/glm-5.3-flash"} {
		if got := chat(m, ""); got != zai {
			t.Errorf("%s: providerOptions %s, want %s (sent %s)", m, got, zai, sent)
		}
	}
	if got := ask("/v1/messages", `{"model":"cline/cline-pass/glm-5.3","max_tokens":64,"messages":[{"role":"user","content":"Say pong"}]}`); got != zai {
		t.Errorf("messages, GLM: providerOptions %s, want %s", got, zai)
	}
	for _, m := range []string{"cline/cline-pass/kimi-k3", "cline/cline-pass/qwen3.8-max", "cline/cline-free/deepseek-v4.1-flash", "cline/cline-free/glm-5.3-flash", "cline2/cline-pass/deepseek-v4-pro", "cline2/cline-pass/glm-5.3"} {
		if got := chat(m, ""); got != "" {
			t.Errorf("%s: providerOptions %s, want none", m, got)
		}
	}
	for _, m := range []string{"cline/cline-pass/deepseek-v4-pro", "cline/cline-pass/glm-5.3"} {
		// a second providerOptions after the client's would be the one a
		// JSON reader keeps, so the body must hold the client's alone
		if got := chat(m, `,"providerOptions":{"gateway":{"order":["novita"]}}`); got != `{"gateway":{"order":["novita"]}}` || strings.Count(string(sent), `"providerOptions"`) != 1 {
			t.Errorf("%s: the client's own providerOptions became %s (sent %s)", m, got, sent)
		}
	}
}

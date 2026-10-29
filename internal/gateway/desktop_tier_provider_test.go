package gateway

import (
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// cybercaius on Discord: Claude Desktop's chat was on AgentRouter's
// gpt-6.1-sol, and Claude Code's haiku calls in it (its small, fast
// model) went to Leee's claude-haiku-5-5, a provider the user never picked
// in Desktop. /v1/models tagged Leee's haiku with the haiku tier unasked,
// so Claude Code asked for it by name. Untagged, Claude Code asks for its
// own claude-haiku id, which stays on the chat's provider: its model of
// that tier when it has one, else the chat's model. A tier the user picked
// keeps its pick.
func TestClaudeDesktopBackgroundStaysOnChatProvider(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	desktopPicked.Lock()
	desktopPicked.model, desktopPicked.from = "", ""
	desktopPicked.Unlock()
	for _, p := range []provider.Provider{
		{ID: "agentrouter", Name: "AgentRouter", Key: "k", Anthropic: "http://127.0.0.1:1", Models: []string{"gpt-6.1-sol"}},
		{ID: "leee", Name: "Leee", Key: "k", Anthropic: "http://127.0.0.1:1", Models: []string{"claude-haiku-5-5"}},
		{ID: "claudeish", Name: "Claudeish", Key: "k", Anthropic: "http://127.0.0.1:1", Models: []string{"claude-opus-5-5", "claude-haiku-4-5", "claude-haiku-5-5"}},
	} {
		if err := provider.Save(p); err != nil {
			t.Fatal(err)
		}
	}
	for _, m := range desktopModels(provider.Catalog(), true) {
		if tier, _ := m["anthropic_family_tier"].(string); tier != "" {
			t.Fatalf("%v tagged %q with nothing picked", m["id"], tier)
		}
	}

	turn := []byte(`{"max_tokens":32000,"tools":[{"name":"Bash"}],"messages":[{"role":"user","content":"hi"}]}`)
	small := []byte(`{"max_tokens":3000,"system":[{"type":"text","text":"Summarize."}],"messages":[{"role":"user","content":"…"}]}`)
	sub := []byte(`{"max_tokens":32000,"tools":[{"name":"Read"}],"messages":[{"role":"user","content":"look"}]}`)

	// the chat on AgentRouter's GPT: Claude Code's haiku calls go there,
	// though Leee serves claude-haiku-5-5 by that very name
	if got := desktopTurn("agentrouter/gpt-6.1-sol", turn, true, false); got != "agentrouter/gpt-6.1-sol" {
		t.Fatalf("turn sent %q", got)
	}
	for _, c := range []struct {
		asked string
		body  []byte
		sub   bool
	}{
		{"claude-haiku-5-5", small, false},
		{"claude-haiku-4-5-20251001", small, false},
		{"claude-haiku-5-5", sub, true},
	} {
		if got := desktopTurn(c.asked, c.body, true, c.sub); got != "agentrouter/gpt-6.1-sol" {
			t.Errorf("chat on AgentRouter, %s (subagent %v) sent to %q", c.asked, c.sub, got)
		}
	}

	// the chat on a provider with haikus of its own: its newest haiku,
	// or the very one asked for when it has it
	if got := desktopTurn("claudeish/claude-opus-5-5", turn, true, false); got != "claudeish/claude-opus-5-5" {
		t.Fatalf("turn sent %q", got)
	}
	for asked, want := range map[string]string{
		"claude-haiku-5-5":          "claudeish/claude-haiku-5-5",
		"claude-haiku-4-5":          "claudeish/claude-haiku-4-5",
		"claude-haiku-4-5-20251001": "claudeish/claude-haiku-5-5",
		"claude-haiku-9[1m]":        "claudeish/claude-haiku-5-5[1m]",
	} {
		if got := desktopTurn(asked, small, true, false); got != want {
			t.Errorf("chat on Claudeish, %s sent to %q, want %q", asked, got, want)
		}
	}

	// a haiku the user picked for the tier is tagged and kept
	if err := SetDesktopTier("haiku", "leee/claude-haiku-5-5"); err != nil {
		t.Fatal(err)
	}
	tagged := false
	for _, m := range desktopModels(provider.Catalog(), true) {
		if DesktopCatalogID(m["id"].(string)) == "leee/claude-haiku-5-5" {
			tagged = m["anthropic_family_tier"] == "haiku" && m["is_family_default"] == true
		}
	}
	if !tagged {
		t.Fatal("the picked haiku isn't tagged")
	}
	if got := desktopTurn("leee/claude-haiku-5-5", small, true, false); got != "leee/claude-haiku-5-5" {
		t.Fatalf("the picked haiku sent to %q", got)
	}
}

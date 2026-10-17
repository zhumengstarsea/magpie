package agent

import (
	"strings"
	"testing"
)

// A quota result never changes Codex's sign-in mode. An official ChatGPT
// sign-in remains beside magpie, while magpie API is an explicit choice.
func TestCodexRunsOutAfterPick(t *testing.T) {
	home, read := codexHome(t, `{"tokens":{"access_token":"x","id_token":"x.e30.x"}}`, "")
	cx := codex(home)
	if err := cx.Fields[0].Set("fake/m1"); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); strings.Contains(cfg, "model_provider =") || !strings.Contains(cfg, "openai_base_url") {
		t.Fatalf("beside the sign-in:\n%s", cfg)
	}
	codexUsedUp = func() bool { return true }
	if err := cx.Sync(); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); strings.Contains(cfg, `model_provider = "magpie"`) || strings.Contains(cfg, "model_catalog_json") ||
		!strings.Contains(cfg, "openai_base_url") || !strings.Contains(cfg, `model = "fake/m1"`) {
		t.Fatalf("quota changed the sign-in mode:\n%s", cfg)
	}
	if err := cx.Sync(); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); strings.Contains(cfg, `model_provider = "magpie"`) {
		t.Fatalf("quota changed the sign-in mode on the next sync:\n%s", cfg)
	}
	codexUsedUp = func() bool { return false }
	if err := cx.Sync(); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); strings.Contains(cfg, "model_provider =") || strings.Contains(cfg, "model_catalog_json") ||
		!strings.Contains(cfg, "openai_base_url") || !strings.Contains(cfg, `model = "fake/m1"`) {
		t.Fatalf("allowance recovery changed the sign-in mode:\n%s", cfg)
	}

	// magpie as the provider because the user asked for it stays so
	if err := cx.Field("login").Set("api"); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); !strings.Contains(cfg, `model_provider = "magpie"`) {
		t.Fatalf("api:\n%s", cfg)
	}
	if err := cx.Sync(); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); !strings.Contains(cfg, `model_provider = "magpie"`) {
		t.Fatalf("api, allowance there, synced:\n%s", cfg)
	}
}

// The explicit Always ChatGPT choice remains stable across Sync as well.
func TestCodexKeptOnChatGPT(t *testing.T) {
	home, read := codexHome(t, `{"tokens":{"access_token":"x","id_token":"x.e30.x"}}`, "")
	cx := codex(home)
	if err := cx.Fields[0].Set("fake/m1"); err != nil {
		t.Fatal(err)
	}
	codexUsedUp = func() bool { return true }
	t.Cleanup(func() { codexUsedUp = func() bool { return false } })
	if err := cx.Sync(); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); strings.Contains(cfg, `model_provider = "magpie"`) {
		t.Fatalf("ChatGPT, used up: magpie became its provider:\n%s", cfg)
	}

	beside := func(when string) {
		t.Helper()
		if cfg := read(); strings.Contains(cfg, "model_provider =") || strings.Contains(cfg, "model_catalog_json") ||
			!strings.Contains(cfg, "openai_base_url") || !strings.Contains(cfg, `model = "fake/m1"`) {
			t.Fatalf("%s:\n%s", when, cfg)
		}
	}
	if err := cx.Field("login").Set("chatgpt"); err != nil {
		t.Fatal(err)
	}
	beside("kept on ChatGPT, used up")
	if err := cx.Sync(); err != nil {
		t.Fatal(err)
	}
	beside("kept on ChatGPT, used up, synced")
	if err := cx.Fields[0].Set("fake/m1"); err != nil {
		t.Fatal(err)
	}
	beside("kept on ChatGPT, used up, picked")
	if got := cx.Field("login").Get(); got != "chatgpt" {
		t.Fatalf("login %q", got)
	}

	// Switching back to the default ChatGPT mode keeps it beside magpie too.
	if err := cx.Field("login").Set(""); err != nil {
		t.Fatal(err)
	}
	beside("ChatGPT again, used up")
}

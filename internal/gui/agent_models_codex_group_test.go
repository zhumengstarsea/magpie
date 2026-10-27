package gui

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
)

// Codex's own models, a ChatGPT account's, stay in its model picker with a
// routing group switched on and its own models switched off (shiqi_40904 on
// Discord: "开启路由组后 … 模型列表里不会显示 codex 模型"). magpie's list shows
// them Codex as its Codex section, and its pickers offer them as Codex's own
// (codex/… would only add a hop), so the switch for the agent's own models
// never takes them out. It did once the group gave the picker one of
// magpie's models, which is what the switch waits for: the picker was the
// group alone, and with the group off every model was back.
func TestCodexOwnModelsStayWithGroupOn(t *testing.T) {
	home := sandboxHome(t)
	codexHome := filepath.Join(home, ".codex")
	t.Setenv("CODEX_HOME", codexHome)
	os.MkdirAll(codexHome, 0o700)
	jwt := func(c any) string {
		b, _ := json.Marshal(c)
		return "e30." + base64.RawURLEncoding.EncodeToString(b) + ".sig"
	}
	auth, _ := json.Marshal(map[string]any{"auth_mode": "chatgpt", "tokens": map[string]any{
		"id_token":     jwt(map[string]any{"email": "a@x.y", "https://api.openai.com/auth": map[string]any{"chatgpt_plan_type": "plus", "chatgpt_account_id": "acc"}}),
		"access_token": jwt(map[string]any{"exp": time.Now().Add(24 * time.Hour).Unix()}), "refresh_token": "r", "account_id": "acc"}})
	os.WriteFile(filepath.Join(codexHome, "auth.json"), auth, 0o600)
	os.WriteFile(filepath.Join(codexHome, "config.toml"), []byte("model = \"gpt-5.5\"\n"), 0o600)
	ids := []string{"gpt-5.5", "gpt-5.4", "gpt-5.4-mini", "gpt-5.3-codex", "gpt-5.2"}
	var cache []map[string]any
	var live []catalog.Model
	for i, id := range ids {
		cache = append(cache, map[string]any{"slug": id, "display_name": id, "visibility": "list", "priority": i, "context_window": 400000})
		live = append(live, catalog.Model{ID: id, Context: 400000})
	}
	b, _ := json.Marshal(map[string]any{"etag": "x", "models": cache})
	os.WriteFile(filepath.Join(codexHome, "models_cache.json"), b, 0o600)
	provider.ForgetAccounts()
	t.Cleanup(provider.ForgetAccounts)
	catalog.Reset()
	t.Cleanup(catalog.Reset)
	if err := catalog.SaveLive("codex", provider.CodexBase, live); err != nil {
		t.Fatal(err)
	}
	if err := provider.SetOwnModelsHidden("codex", true); err != nil {
		t.Fatal(err)
	}
	codex, err := agent.Find("codex")
	if err != nil {
		t.Fatal(err)
	}
	picks := func() (out []string) {
		t.Helper()
		for _, f := range agentFields(codex, codex.Values()) {
			if f.Key == "model" {
				for _, o := range f.Options {
					out = append(out, o.Value)
				}
			}
		}
		return out
	}
	for _, id := range ids {
		if got := picks(); !slices.Contains(got, id) {
			t.Fatalf("group off: %s not in Codex's picker %v", id, got)
		}
	}
	if err := provider.SaveGroup(provider.Group{ID: "mygpt", Name: "MyGPT", Members: []string{"codex/gpt-5.5", "codex/gpt-5.4"}}); err != nil {
		t.Fatal(err)
	}
	got := picks()
	for _, id := range append(slices.Clone(ids), "group/mygpt") {
		if !slices.Contains(got, id) {
			t.Fatalf("group on: %s not in Codex's picker %v", id, got)
		}
	}
	// nothing of Codex's own is left for the switch to take out, so the
	// list has none to show
	if codex.OwnModels(codex.Values()) {
		t.Fatal("Codex's list offers a switch for its own models, every one of them magpie's too")
	}
	// one the list takes out is the switch's again: magpie no longer shows
	// it Codex
	if err := provider.SetHiddenModels("codex", []string{"codex/gpt-5.4"}); err != nil {
		t.Fatal(err)
	}
	if got := picks(); slices.Contains(got, "gpt-5.4") || !slices.Contains(got, "gpt-5.4-mini") {
		t.Fatalf("gpt-5.4 taken out of the list, own models off: picker %v", got)
	}
}

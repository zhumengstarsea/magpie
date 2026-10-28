package usageapi

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/testenv"
	"github.com/yetone/magpie/internal/usage"
)

func TestMain(m *testing.M) { testenv.Main(m) }

// A provider the usage log names but that isn't set up now is still named
// as magpie knows it (@ByronVon1 on X): the Claude Code subscription,
// signed out of since, showed as "claude" in the Requests page's Provider
// filter, between providers named by their names. A preset's removed
// since is named by the preset; a custom one magpie can't know stays its
// id. Every place the page names a provider from the log says the same:
// the filter, each row, the split-by legends, and the summary's models
// and accounts.
func TestProvidersNotSetUpAreNamed(t *testing.T) {
	for _, p := range provider.All() {
		if p.ID == "claude" || p.ID == "deepseek" {
			t.Fatalf("the test home has %s set up", p.ID)
		}
	}
	now := time.Now()
	for i, pid := range []string{"claude", "deepseek", "gone-custom"} {
		usage.Append(usage.Record{Time: now.Add(-time.Duration(i+1) * time.Minute), Agent: "claude", Provider: pid, Model: "m-" + pid, Input: 10, Output: 2, ProviderAccount: "me@example.com"})
	}
	want := map[string]string{"claude": "Claude Code", "deepseek": "DeepSeek", "gone-custom": "gone-custom"}

	l := LedgerPage(usage.All, usage.Filter{}, 0, 100)
	if len(l.Providers) != len(want) {
		t.Fatalf("providers: %+v", l.Providers)
	}
	for _, p := range l.Providers {
		if p.Name != want[p.ID] {
			t.Errorf("the Provider filter names %s %q, want %q", p.ID, p.Name, want[p.ID])
		}
	}
	for _, r := range l.Rows {
		if r.ProviderName != want[r.Provider] {
			t.Errorf("a row names %s %q, want %q", r.Provider, r.ProviderName, want[r.Provider])
		}
	}
	for _, s := range l.By["provider"] {
		if s.Name != want[s.ID] {
			t.Errorf("the split by provider names %s %q, want %q", s.ID, s.Name, want[s.ID])
		}
	}
	for _, s := range l.By["modelAt"] {
		if s.Name == "m-claude · claude" {
			t.Errorf("the split by model at provider names %q", s.Name)
		}
	}
	if len(l.Accounts) != 1 || len(l.Accounts[0].Providers) != 3 {
		t.Errorf("accounts: %+v", l.Accounts)
	}
	for _, a := range l.Accounts {
		for _, n := range a.Providers {
			if n == "claude" || n == "deepseek" {
				t.Errorf("the Account filter names a provider %q", n)
			}
		}
	}

	s := State(usage.All)
	for _, g := range append(s.Models, s.Accounts...) {
		if g.Sub == "claude" || g.Sub == "deepseek" {
			t.Errorf("the summary names %q's provider %q", g.Name, g.Sub)
		}
	}
	if len(s.Models) != 3 {
		t.Errorf("models: %+v", s.Models)
	}
}

// cybercaius on Discord: the Usage page's Provider filter lists a
// provider deleted and the one added under its id since apart, the
// deleted one by its own name with when it was deleted.
func TestDeletedProviderListedApartAndMarked(t *testing.T) {
	home := t.TempDir()
	testenv.SetHome(t, home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	if err := provider.Save(provider.Provider{ID: "relay", Name: "Old relay", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"gpt-9"}}); err != nil {
		t.Fatal(err)
	}
	usage.Append(usage.Record{Time: time.Now().Add(-time.Hour), Agent: "codex", Provider: "relay", Model: "gpt-9", Input: 10, Output: 2})
	if err := provider.Delete("relay"); err != nil {
		t.Fatal(err)
	}
	// while no other has its id, its calls stay under it, named as it was
	if ps := LedgerPage(usage.All, usage.Filter{}, 0, 100).Providers; len(ps) != 1 || ps[0].ID != "relay" || ps[0].Name != "Old relay" || ps[0].Deleted == nil {
		t.Fatalf("deleted, its id not taken: %+v", ps)
	}
	if err := provider.Save(provider.Provider{ID: "relay", Name: "New relay", Key: "k2", Chat: "http://127.0.0.1:2/v1", Models: []string{"gpt-9"}}); err != nil {
		t.Fatal(err)
	}
	usage.Append(usage.Record{Time: time.Now().Add(time.Second), Agent: "codex", Provider: "relay", Model: "gpt-9", Input: 10, Output: 2})

	got := map[string]Named{}
	for _, p := range LedgerPage(usage.All, usage.Filter{}, 0, 100).Providers {
		got[p.Name] = p
	}
	if n := got["New relay"]; n.ID != "relay" || n.Deleted != nil {
		t.Errorf("the new one: %+v", n)
	}
	if n := got["Old relay"]; !strings.HasPrefix(n.ID, "relay~") || n.Deleted == nil {
		t.Errorf("the deleted one: %+v (all %+v)", n, got)
	}
}

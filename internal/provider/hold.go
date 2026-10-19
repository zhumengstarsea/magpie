package provider

import (
	"hash/maphash"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/filememo"
	"github.com/yetone/magpie/internal/plugin"
	"github.com/yetone/magpie/internal/settings"
)

// The catalog is rebuilt from every provider's files on each Resolve (tens
// of milliseconds with a few hundred models), and a page of the GUI
// resolves a model for each row: /api/state and /api/groups took 4s and
// 7s with 24 providers. A request that only reads holds the catalog it
// built first for all its look-ups.
var held struct {
	sync.Mutex
	holds int
	gen   uint64
	// what was built, and when, by what it is
	built map[string]heldValue
	// env is the environment it was built in (environSum)
	env uint64
	// fgen is filememo's generation it was built in: a write that only
	// tells filememo (settings.Save, an account's file) has it built again
	fgen uint64
	// roots is the providers' and the settings' files as last looked at
	roots [2]rootStamp
}

// rootStamp is a file's time and size, or that it wasn't there.
type rootStamp struct {
	mod  time.Time
	size int64
	ok   bool
}

func stampOf(path string) rootStamp {
	fi, err := os.Stat(path)
	if err != nil {
		return rootStamp{}
	}
	return rootStamp{fi.ModTime(), fi.Size(), true}
}

// heldFreshLocked drops what is built when a file was written since, by
// this process (filememo.Gen) or, for the providers' and the settings'
// files, by another (the CLI's "magpie group add", seen by the next
// request, not a second later). Called with held held.
func heldFreshLocked(look bool) {
	if g := filememo.Gen(); g != held.fgen {
		held.built, held.fgen = nil, g
	}
	if !look {
		return
	}
	for i, path := range [2]string{Path(), settings.Path()} {
		if st := stampOf(path); st != held.roots[i] {
			held.roots[i] = st
			held.built = nil
			filememo.Look(path)
		}
	}
}

var environSeed = maphash.MakeSeed()

// environSum is a hash of this process's environment: the folders every
// file and account is read from are named by it.
func environSum() uint64 {
	var h maphash.Hash
	h.SetSeed(environSeed)
	for _, kv := range os.Environ() {
		h.WriteString(kv)
		h.WriteByte(0)
	}
	return h.Sum64()
}

type heldValue struct {
	v  any
	at time.Time
}

func init() {
	catalog.Forget = Changed
	// a plugin's sign-in saved or renewed, a plugin added, and the models
	// a plugin lists once the host has answered: the catalog may be
	// another, so the model lists agents keep in their own files follow,
	// as for a provider saved (catalog.Changed). Dropping only what is
	// held left Pi's models.json without a plugin's models signed in to
	// after it was written (#2373).
	plugin.OnChange(catalog.Touched)
	plugin.OnListed(catalog.Touched)
	// and what a request holds goes at once: a SyncCatalog running holds
	// the plugins as they were before a sign-in, and a look at the
	// accounts right after it was given those for a second (#2449)
	plugin.OnChangeNow(dropHeld)
}

// heldFor is how long a held catalog is used before it is built again,
// while one request after another keeps it held.
var heldFor = time.Second

// Hold keeps the catalog built once until release is called, for a
// request that only reads; a write meanwhile (Changed) drops it.
func Hold() (release func()) {
	held.Lock()
	if held.holds == 0 {
		// what was built under another environment (a HOME or config
		// folder of its own: a test's) isn't this one's
		if env := environSum(); env != held.env {
			held.built, held.env = nil, env
		}
		heldFreshLocked(true)
	}
	held.holds++
	held.Unlock()
	return sync.OnceFunc(func() {
		held.Lock()
		if held.holds--; held.holds == 0 {
			// what is built stays for the next request within heldFor, as
			// it stays for one that overlaps this one: one request after
			// another, each built the catalog again from every provider's
			// files and accounts (a fifth of a request's CPU with 30
			// providers). Only what is past heldFor goes, for memory.
			for k, h := range held.built {
				if time.Since(h.at) >= heldFor {
					delete(held.built, k)
				}
			}
		}
		held.Unlock()
	})
}

// Changed drops what is held: a provider, an account, a plugin or a
// setting was written.
func Changed() {
	filememo.Forget()
	dropHeld()
}

// dropHeld is Changed for what is built on the plugins alone, which
// read their files afresh each time.
func dropHeld() {
	held.Lock()
	held.gen++
	held.built = nil
	held.Unlock()
}

// heldOf is what build gives, built once while a request holds it.
func heldOf[T any](key string, build func() T) T {
	held.Lock()
	if held.holds == 0 {
		held.Unlock()
		return build()
	}
	heldFreshLocked(false)
	if h, ok := held.built[key]; ok && time.Since(h.at) < heldFor {
		held.Unlock()
		return h.v.(T)
	}
	gen := held.gen
	held.Unlock()
	v := build()
	held.Lock()
	if held.holds > 0 && held.gen == gen {
		if held.built == nil {
			held.built = map[string]heldValue{}
		}
		held.built[key] = heldValue{v, time.Now()}
	}
	held.Unlock()
	return v
}

func heldEntries(build func() []Entry) []Entry {
	// clipped: a caller's append copies
	return heldOf("entries", func() []Entry { return slices.Clip(build()) })
}

// heldPlugins is plugin.Cached, which reads the plugins' list and
// sign-ins each time, and PluginOf asks for every model.
func heldPlugins() []plugin.Provider {
	return heldOf("plugins", func() []plugin.Provider { return slices.Clip(plugin.Cached()) })
}

// fromEntries is what was built from a catalog, and which catalog: its
// first entry and its length.
type fromEntries[T any] struct {
	at *Entry
	n  int
	v  T
}

// heldFrom is heldOf for what is built from entries: held while a request
// holds the catalog, and built again for another catalog than it was built
// from (providerEntries built afresh after heldFor, or one a caller made).
// A request asks for the groups' entries several times (Catalog, Served),
// and each built them from every member of every group again (the CPU a
// request to a group took at once, with many providers).
func heldFrom[T any](key string, entries []Entry, build func([]Entry) T) T {
	var at *Entry
	if len(entries) > 0 {
		at = &entries[0]
	}
	held.Lock()
	if held.holds == 0 {
		held.Unlock()
		return build(entries)
	}
	heldFreshLocked(false)
	if h, ok := held.built[key]; ok && time.Since(h.at) < heldFor {
		if f := h.v.(fromEntries[T]); f.at == at && f.n == len(entries) {
			held.Unlock()
			return f.v
		}
	}
	gen := held.gen
	held.Unlock()
	v := build(entries)
	held.Lock()
	if held.holds > 0 && held.gen == gen {
		if held.built == nil {
			held.built = map[string]heldValue{}
		}
		held.built[key] = heldValue{fromEntries[T]{at, len(entries), v}, time.Now()}
	}
	held.Unlock()
	return v
}

// heldSettings is the settings, read once while a request holds the
// catalog: a look at the agents asks them for every model of every agent.
// Only for reading — what it gives is shared by the request's look-ups.
func heldSettings() settings.Settings { return heldOf("settings", settings.Load) }

// HeldSettings is the settings as heldSettings reads them, for a look
// outside the package made many times in one request.
func HeldSettings() settings.Settings { return heldSettings() }

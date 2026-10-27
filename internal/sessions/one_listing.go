package sessions

import (
	"slices"
	"sync"
	"time"
)

// held is the session folders' listings kept while a request that reads
// them more than once is answered. A Sessions poll lists every Claude
// Code and Codex file for its list (allFiles), then again for the Codex
// threads' effort (the Requests state's CallSources): ~21k files, read
// twice a poll.
var held struct {
	sync.Mutex
	n     int
	since time.Time // when the latest hold began
	lists map[string]heldList
}

type heldList struct {
	at    time.Time
	files []file
}

// OneListing has the session folders listed once until done is called:
// the reads between share each folder's listing, taken after the latest
// hold began, so none sees the files as they were before its own request
// came. done lets the listings go once no hold is left.
func OneListing() (done func()) {
	held.Lock()
	held.n++
	held.since = time.Now()
	held.Unlock()
	return sync.OnceFunc(func() {
		held.Lock()
		if held.n--; held.n == 0 {
			held.lists = nil
		}
		held.Unlock()
	})
}

// heldFiles is list(), or, under a hold, the listing of key taken since
// the latest hold began. The folders are listed outside the lock.
func heldFiles(key string, list func() []file) []file {
	held.Lock()
	if held.n == 0 {
		held.Unlock()
		return list()
	}
	if l, ok := held.lists[key]; ok && !l.at.Before(held.since) {
		held.Unlock()
		return slices.Clone(l.files)
	}
	held.Unlock()
	at := time.Now()
	files := list()
	held.Lock()
	if held.n > 0 && !at.Before(held.since) {
		if held.lists == nil {
			held.lists = map[string]heldList{}
		}
		held.lists[key] = heldList{at, slices.Clone(files)}
	}
	held.Unlock()
	return files
}

package sessions

import "testing"

// Under a hold a folder is listed once; a hold begun after a listing has
// it listed again, so no request reads files from before it came; with
// no hold left every read lists.
func TestOneListing(t *testing.T) {
	n := 0
	list := func() []file { n++; return []file{{path: "a"}} }
	read := func() {
		if fs := heldFiles("k", list); len(fs) != 1 || fs[0].path != "a" {
			t.Fatalf("heldFiles = %v", fs)
		}
	}
	read()
	read()
	if n != 2 {
		t.Fatalf("listed %d times with no hold, want 2", n)
	}
	n = 0
	done := OneListing()
	read()
	read()
	if n != 1 {
		t.Fatalf("listed %d times under one hold, want 1", n)
	}
	other := OneListing()
	read()
	if n != 2 {
		t.Fatalf("a hold begun after the listing read it: listed %d times, want 2", n)
	}
	read()
	done()
	read()
	if n != 2 {
		t.Fatalf("listed %d times while a hold was left, want 2", n)
	}
	other()
	other()
	read()
	if n != 3 {
		t.Fatalf("listed %d times after the holds ended, want 3", n)
	}
	if held.n != 0 || held.lists != nil {
		t.Fatalf("holds left: %d, lists %v", held.n, held.lists)
	}
}

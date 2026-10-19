package sessions

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
)

// A scan of a small session file takes a buffer the size of the file, not
// 1 MiB: a start reads thousands of them, and their buffers were most of
// what it allocated (#2776).
func TestScanOfASmallFileTakesASmallBuffer(t *testing.T) {
	p := filepath.Join(t.TempDir(), "s.jsonl")
	if err := os.WriteFile(p, []byte(`{"type":"user","n":1}`+"\n"+`{"type":"assistant","n":2}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	const scans = 200
	lines := 0
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	for range scans {
		if _, err := scanAt(p, 0, nil, func(b []byte, _, _ int64) bool { lines++; return true }); err != nil {
			t.Fatal(err)
		}
	}
	runtime.ReadMemStats(&after)
	if lines != 2*scans {
		t.Fatalf("read %d lines, want %d", lines, 2*scans)
	}
	// 16 KiB each, with room for the file handles and anything else
	if per := (after.TotalAlloc - before.TotalAlloc) / scans; per > 64<<10 {
		t.Fatalf("a scan of a two-line file allocated %d bytes", per)
	}
}

// A line longer than the buffer (the file grew after its size was read) is
// still handed over whole.
func TestScanReadsALineLongerThanItsBuffer(t *testing.T) {
	p := filepath.Join(t.TempDir(), "s.jsonl")
	if err := os.WriteFile(p, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	size := scanBuffer(f, 0)
	f.Close()
	long := `{"text":"` + strings.Repeat("x", 3*size) + `"}`
	if err := os.WriteFile(p, []byte(long+"\n{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var got []int
	if _, err := scanAt(p, 0, nil, func(b []byte, _, _ int64) bool { got = append(got, len(b)); return true }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != len(long) || got[1] != 2 {
		t.Fatalf("lines %v, want [%d 2]", got, len(long))
	}
}

// Every other reader of a session file sizes its buffer as scanAt does:
// the calls readers of MiniMax Code, agy and WorkBuddy, dsh's packed
// frames and a search's first look each took 1 MiB per file read (#2776).
func TestReadersOfASmallFileTakeASmallBuffer(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		t.Helper()
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	line := `{"type":"user","n":1}` + "\n"
	plain := write("s.jsonl", line+line)
	enc, err := zstd.NewWriter(nil)
	if err != nil {
		t.Fatal(err)
	}
	framed := write("s.v3.jsonl.zstd", string(enc.EncodeAll([]byte(line+line), nil)))
	enc.Close()
	for _, c := range []struct {
		name string
		read func()
	}{
		{"minimax", func() { readMiniMaxCalls(file{path: plain, key: "minimax-code:s"}) }},
		{"agy", func() { readAgyCalls(file{path: plain, key: "agy:s"}) }},
		{"workbuddy", func() { readWorkBuddyCalls(file{agent: "workbuddy", path: plain, key: "workbuddy:s"}) }},
		{"frames", func() { scanFrames(framed, 0, 0, func([]byte) bool { return true }) }},
		{"search", func() { mayHold(context.Background(), plain, "absent") }},
	} {
		t.Run(c.name, func(t *testing.T) {
			const reads = 50
			c.read() // anything a first read sets up once
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			for range reads {
				c.read()
			}
			runtime.ReadMemStats(&after)
			if per := (after.TotalAlloc - before.TotalAlloc) / reads; per > 256<<10 {
				t.Fatalf("a read of a two-line file allocated %d bytes", per)
			}
		})
	}
}

package usage

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
)

// Queries reading at once hold one priced copy of a gateway block between
// them, kept until the last is done: the GUI's requests at a cold start
// each held the whole history's, and the app doubled (#2278). A query of
// other prices or names prices its own.
func TestPricedBlocksSharedWhileHeld(t *testing.T) {
	var raw rowChunk
	at := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	for i := 0; i < logBlockRows; i++ {
		raw.add(Row{Record: Record{Time: at, Provider: "openai", Model: "gpt-5", Input: 100 + i, Output: 10}}, "", int64(i), false)
	}
	packed := raw.pack()
	if packed == &raw {
		t.Fatal("the block was not packed")
	}
	pricer := func(meta string) *blockPricer {
		b := newBlockPricer(provider.Renames{}, nil, func(Record) *catalog.Price { return nil })
		b.share(sha256.Sum256([]byte(meta)))
		return b
	}
	held := func() int {
		sharedPriced.Lock()
		defer sharedPriced.Unlock()
		n := 0
		for k := range sharedPriced.blocks {
			if k.block == packed {
				n++
			}
		}
		return n
	}
	a, b, other := pricer("m"), pricer("m"), pricer("n")
	ca, cb, co := a.block(packed), b.block(packed), other.block(packed)
	if ca == nil || ca != cb {
		t.Fatalf("two queries of one meta priced %p and %p", ca, cb)
	}
	if co == ca {
		t.Fatal("a query of other prices read another's block")
	}
	if ca.row(7).Input != 107 || co.row(7).Input != 107 {
		t.Fatalf("rows read back %d, %d", ca.row(7).Input, co.row(7).Input)
	}
	a.release()
	if held() != 2 || b.block(packed) != cb {
		t.Fatalf("a block one query still holds was let go: %d held", held())
	}
	b.release()
	b.release() // a second release lets nothing go twice
	other.release()
	if n := held(); n != 0 {
		t.Fatalf("%d blocks held after every query was done", n)
	}
	if c := pricer("m").block(packed); c == ca {
		t.Fatal("a block let go was handed out again")
	}
}

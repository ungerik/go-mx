package twmerge

import (
	"strconv"
	"sync"
	"testing"
)

func TestLRUCache(t *testing.T) {
	t.Run("get returns what set stored", func(t *testing.T) {
		c := newLRUCache(2)
		c.set("a", "1")
		if v, ok := c.get("a"); !ok || v != "1" {
			t.Errorf("get(a) = %q, %v; want 1, true", v, ok)
		}
		if _, ok := c.get("b"); ok {
			t.Error("get(b) found a key never set")
		}
	})

	t.Run("set overwrites in place", func(t *testing.T) {
		c := newLRUCache(2)
		c.set("a", "1")
		c.set("a", "2")
		if v, _ := c.get("a"); v != "2" {
			t.Errorf("get(a) = %q, want 2", v)
		}
		if c.size != 1 {
			t.Errorf("size = %d after overwriting one key, want 1", c.size)
		}
	})

	t.Run("an overfull generation becomes the previous one", func(t *testing.T) {
		// maxSize 2: the third entry starts a new generation, as upstream's
		// cacheSize > maxCacheSize check does.
		c := newLRUCache(2)
		c.set("a", "1")
		c.set("b", "2")
		c.set("c", "3")
		if len(c.current) != 0 || len(c.previous) != 3 {
			t.Fatalf("current %v, previous %v; want empty, a b c", c.current, c.previous)
		}
		// Another full generation drops the previous one as a whole.
		c.set("d", "4")
		c.set("e", "5")
		c.set("f", "6")
		if _, ok := c.get("a"); ok {
			t.Error("a survived two generation switches")
		}
		if v, ok := c.get("d"); !ok || v != "4" {
			t.Errorf("get(d) = %q, %v; want 4, true", v, ok)
		}
	})

	t.Run("a hit in the previous generation is promoted", func(t *testing.T) {
		c := newLRUCache(2)
		c.set("a", "1")
		c.set("b", "2")
		c.set("c", "3") // a, b, c are now the previous generation
		if v, ok := c.get("a"); !ok || v != "1" {
			t.Fatalf("get(a) = %q, %v; want 1, true", v, ok)
		}
		if _, ok := c.current["a"]; !ok {
			t.Error("a was not promoted into the current generation")
		}
	})

	t.Run("maxSize below 1 disables the cache", func(t *testing.T) {
		c := newLRUCache(0)
		c.set("a", "1")
		if _, ok := c.get("a"); ok {
			t.Error("a disabled cache returned a value")
		}
	})

	t.Run("concurrent use", func(t *testing.T) {
		// Run with -race: Merge is called from concurrent renders.
		c := newLRUCache(8)
		var wg sync.WaitGroup
		for g := range 8 {
			wg.Go(func() {
				for i := range 100 {
					key := strconv.Itoa((g + i) % 20)
					c.set(key, key)
					if v, ok := c.get(key); ok && v != key {
						t.Errorf("get(%s) = %q", key, v)
					}
				}
			})
		}
		wg.Wait()
	})
}

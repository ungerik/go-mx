package twmerge

// The result cache, ported from tailwind-merge v3.7.0 (src/lib/lru-cache.ts,
// used by src/lib/create-tailwind-merge.ts).

import "sync"

// cacheSize is the default-config cacheSize of tailwind-merge v3.7.0: how
// many merge results the current generation of [lruCache] holds.
const cacheSize = 500

// cache holds the results of [Merge] by their input class list.
var cache = newLRUCache(cacheSize)

// lruCache is upstream's approximate LRU cache: two maps, the current
// generation and the previous one. A new entry goes into the current map;
// when that holds more than maxSize entries it becomes the previous map and
// a new current map starts, so the previous generation is dropped as a
// whole. A hit in the previous map is promoted into the current one, which
// keeps the class lists still in use.
//
// Upstream's cache lives in single-threaded JavaScript; here [Merge] runs
// concurrently, so a read-write mutex guards the maps. A hit in the current
// generation, the common case once the class lists in use are cached, only
// reads and takes the read lock, so concurrent renders do not wait on each
// other; promoting from the previous generation and storing take the write
// lock.
type lruCache struct {
	mu       sync.RWMutex
	maxSize  int
	size     int
	current  map[string]string
	previous map[string]string
}

// newLRUCache returns a cache whose current generation holds up to maxSize
// entries. A maxSize below 1 disables caching, as upstream's does.
func newLRUCache(maxSize int) *lruCache {
	return &lruCache{
		maxSize:  maxSize,
		current:  make(map[string]string),
		previous: make(map[string]string),
	}
}

// get returns the cached value of key, promoting it from the previous
// generation into the current one.
func (c *lruCache) get(key string) (string, bool) {
	if c.maxSize < 1 {
		return "", false
	}
	c.mu.RLock()
	value, ok := c.current[key]
	c.mu.RUnlock()
	if ok {
		return value, true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	// Another goroutine may have promoted or stored key, or started a new
	// generation, since the read lock was released.
	if value, ok := c.current[key]; ok {
		return value, true
	}
	if value, ok := c.previous[key]; ok {
		c.update(key, value)
		return value, true
	}
	return "", false
}

// set stores value under key.
func (c *lruCache) set(key, value string) {
	if c.maxSize < 1 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.current[key]; ok {
		c.current[key] = value
		return
	}
	c.update(key, value)
}

// update adds key to the current generation and starts a new generation
// once the current one is over maxSize. The caller holds c.mu.
func (c *lruCache) update(key, value string) {
	c.current[key] = value
	c.size++
	if c.size > c.maxSize {
		c.size = 0
		c.previous = c.current
		c.current = make(map[string]string, c.maxSize)
	}
}

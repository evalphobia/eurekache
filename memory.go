package eurekache

import (
	"bytes"
	"encoding/gob"
	"sync"
	"time"
)

// MemoryCacheTTL is a cache source for on-memory cache
// When item size reaches maxSize, the item is selected by FIFO or TTL and erased.
type MemoryCacheTTL struct {
	itemsMu    sync.RWMutex
	items      map[string]*Item
	maxSize    int
	defaultTTL int64
}

// NewMemoryCacheTTL returns initialized MemoryCacheTTL
// max value limits maximum saved item size.
func NewMemoryCacheTTL(max int) *MemoryCacheTTL {
	if max == 0 {
		return nil
	}

	return &MemoryCacheTTL{
		items:   make(map[string]*Item),
		maxSize: max,
	}
}

// SetTTL sets default TTL (milliseconds)
func (c *MemoryCacheTTL) SetTTL(ttl int64) {
	c.defaultTTL = ttl
}

// Get searches cache on memory by given key and returns flag of cache is existed or not.
// when cache hit, data is assigned.
func (c *MemoryCacheTTL) Get(key string, data interface{}) bool {
	c.itemsMu.RLock()
	defer c.itemsMu.RUnlock()

	item, ok := c.items[key]
	switch {
	case !ok:
		return false
	case !c.isValidItem(item):
		return false
	default:
		return copyValue(data, item.Value)
	}
}

// GetInterface searches cache on memory by given key and returns interface value.
func (c *MemoryCacheTTL) GetInterface(key string) (interface{}, bool) {
	c.itemsMu.RLock()
	defer c.itemsMu.RUnlock()

	if item, ok := c.items[key]; ok {
		if c.isValidItem(item) {
			return item.Value, true
		}
	}

	return nil, false
}

// GetGobBytes searches cache on memory by given key and returns gob-encoded value.
func (c *MemoryCacheTTL) GetGobBytes(key string) ([]byte, bool) {
	c.itemsMu.RLock()
	defer c.itemsMu.RUnlock()

	if item, ok := c.items[key]; ok {
		if c.isValidItem(item) {
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(item.Value)
			if err == nil {
				return buf.Bytes(), true
			}
		}
	}

	return []byte{}, false
}

// Set sets data.
func (c *MemoryCacheTTL) Set(key string, data interface{}) error {
	return c.SetExpire(key, data, c.defaultTTL)
}

// SetExpire sets data with TTL.
func (c *MemoryCacheTTL) SetExpire(key string, data interface{}, ttl int64) error {
	if key == "" {
		return nil
	}

	c.itemsMu.Lock()
	defer c.itemsMu.Unlock()

	replaceKey, item := c.getNextReplacement()
	if replaceKey != "" {
		delete(c.items, replaceKey)
	}

	item.init()
	item.SetExpire(ttl)
	item.Value = data
	c.items[key] = item
	return nil
}

// isValidItem checks if the item is expired or not
func (c *MemoryCacheTTL) isValidItem(item *Item) bool {
	return item.ExpiredAt > time.Now().UnixNano()
}

// getNextReplacement returns new item.
// when it reaches maximum item size, any expired item or oldest item is returned.
func (c *MemoryCacheTTL) getNextReplacement() (string, *Item) {
	now := time.Now().UnixNano()

	var replaceItem *Item
	var replaceKey string
	var oldestTime int64

	for key, item := range c.items {
		// return expired item
		if now > item.ExpiredAt {
			return key, item
		}

		// save older item
		timeDelta := now - item.CreatedAt
		if timeDelta > oldestTime {
			oldestTime = timeDelta
			replaceKey = key
			replaceItem = item
		}
	}

	// return oldest item when the cache reaches maximum size
	if len(c.items) >= c.maxSize {
		return replaceKey, replaceItem
	}

	return "", &Item{}
}

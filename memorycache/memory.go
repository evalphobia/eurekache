package memorycache

import (
	"bytes"
	"encoding/gob"
	"sync"
	"time"

	"github.com/evalphobia/eurekache"
)

// CacheTTL is a cache source for on-memory cache
// When item size reaches maxSize, the item is selected by FIFO or TTL and erased.
type CacheTTL struct {
	itemsMu     sync.RWMutex
	items       map[string]*eurekache.Item
	deleteQueue []string
	maxSize     int
	defaultTTL  int64
}

// NewCacheTTL returns initialized CacheTTL
// max value limits maximum saved item size.
func NewCacheTTL(max int) *CacheTTL {
	if max == 0 {
		return nil
	}

	return &CacheTTL{
		items:       make(map[string]*eurekache.Item),
		maxSize:     max,
		deleteQueue: make([]string, 0, max),
	}
}

// SetTTL sets default TTL (milliseconds)
func (c *CacheTTL) SetTTL(ttl int64) {
	c.defaultTTL = ttl
}

// Get searches cache on memory by given key and returns flag of cache is existed or not.
// when cache hit, data is assigned.
func (c *CacheTTL) Get(key string, data interface{}) bool {
	c.itemsMu.RLock()
	defer c.itemsMu.RUnlock()

	item, ok := c.items[key]
	switch {
	case !ok:
		return false
	case !c.isValidItem(item):
		return false
	default:
		return eurekache.CopyValue(data, item.Value)
	}
}

// GetInterface searches cache on memory by given key and returns interface value.
func (c *CacheTTL) GetInterface(key string) (interface{}, bool) {
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
func (c *CacheTTL) GetGobBytes(key string) ([]byte, bool) {
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
func (c *CacheTTL) Set(key string, data interface{}) error {
	return c.SetExpire(key, data, c.defaultTTL)
}

// SetExpire sets data with TTL.
func (c *CacheTTL) SetExpire(key string, data interface{}, ttl int64) error {
	if key == "" {
		return nil
	}

	c.itemsMu.Lock()
	defer c.itemsMu.Unlock()

	if data == nil {
		delete(c.items, key)
		return nil
	}

	// return oldest item when the cache reaches maximum size
	if len(c.deleteQueue) >= c.maxSize {
		c.deleteOldest()
	}

	item := eurekache.NewItem()
	item.SetExpire(ttl)
	item.Value = data
	c.items[key] = item
	c.deleteQueue = append(c.deleteQueue, key)
	return nil
}

func (c *CacheTTL) deleteOldest() {
	if len(c.deleteQueue) == 0 {
		return
	}

	oldestKey := c.deleteQueue[0]
	c.deleteQueue = c.deleteQueue[1:]

	// retry delete when missing
	if _, ok := c.items[oldestKey]; !ok {
		c.deleteOldest()
		return
	}

	delete(c.items, oldestKey)
}

// isValidItem checks if the item is expired or not
func (c *CacheTTL) isValidItem(item *eurekache.Item) bool {
	if item.Value == nil {
		return false
	}
	return item.ExpiredAt > time.Now().UnixNano()
}

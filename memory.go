package eurekache

import (
	"bytes"
	"encoding/gob"
	"sync"
	"time"
)

type MemoryCacheTTL struct {
	itemsMu    sync.RWMutex
	items      map[string]*Item
	maxSize    int
	defaultTTL int64
}

func NewMemoryCacheTTL(max int) *MemoryCacheTTL {
	if max == 0 {
		return nil
	}

	return &MemoryCacheTTL{
		items:   make(map[string]*Item),
		maxSize: max,
	}
}

func (c *MemoryCacheTTL) SetTTL(ttl int64) {
	c.defaultTTL = ttl
}

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

func (c *MemoryCacheTTL) Set(key string, data interface{}) error {
	return c.SetExpire(key, data, c.defaultTTL)
}

// ttl=milli second
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

func (c *MemoryCacheTTL) isValidItem(item *Item) bool {
	return item.ExpiredAt > time.Now().UnixNano()
}

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

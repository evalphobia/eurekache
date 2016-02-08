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

func (m *MemoryCacheTTL) Get(key string, data interface{}) bool {
	m.itemsMu.RLock()
	defer m.itemsMu.RUnlock()

	item, ok := m.items[key]
	switch {
	case !ok:
		return false
	case !m.isValidItem(item):
		return false
	default:
		return copyValue(data, item.value)
	}
}

func (m *MemoryCacheTTL) GetInterface(key string) (interface{}, bool) {
	m.itemsMu.RLock()
	defer m.itemsMu.RUnlock()

	if item, ok := m.items[key]; ok {
		if m.isValidItem(item) {
			return item.value, true
		}
	}

	return nil, false
}

func (m *MemoryCacheTTL) GetGobBytes(key string) ([]byte, bool) {
	m.itemsMu.RLock()
	defer m.itemsMu.RUnlock()

	if item, ok := m.items[key]; ok {
		if m.isValidItem(item) {
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(item.value)
			if err == nil {
				return buf.Bytes(), true
			}
		}
	}

	return []byte{}, false
}

func (m *MemoryCacheTTL) Set(key string, data interface{}) {
	m.SetExpire(key, data, 0)
}

// ttl=milli second
func (m *MemoryCacheTTL) SetExpire(key string, data interface{}, ttl int64) {
	if key == "" {
		return
	}

	m.itemsMu.Lock()
	defer m.itemsMu.Unlock()

	replaceKey, item := m.getNextReplacement()
	if replaceKey != "" {
		delete(m.items, replaceKey)
	}

	item.init()
	item.SetExpire(ttl)
	item.value = data
	m.items[key] = item
}

func (m *MemoryCacheTTL) isValidItem(item *Item) bool {
	return item.ExpiredAt > time.Now().UnixNano()
}

func (m *MemoryCacheTTL) getNextReplacement() (string, *Item) {
	now := time.Now().UnixNano()

	var replaceItem *Item
	var replaceKey string
	var oldestTime int64 = 0

	for key, item := range m.items {
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
	if len(m.items) >= m.maxSize {
		return replaceKey, replaceItem
	}

	return "", &Item{}
}

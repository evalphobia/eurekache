package eurekache

import (
	"math"
	"time"
)

// Item contains actual value and meta data for cache.
type Item struct {
	// unix nanosec of creation time
	CreatedAt int64

	// unix nanosec of expires time
	ExpiredAt int64

	// The actual value stored in this item.
	Value interface{}
}

// Init initializes Item
func (i *Item) Init() {
	i.CreatedAt = time.Now().UnixNano()
	i.ExpiredAt = math.MaxInt64
	i.Value = nil
}

// SetExpire updates ExpiredAt from given ttl millisec
func (i *Item) SetExpire(ttl int64) {
	if ttl == 0 {
		return
	}
	i.ExpiredAt = i.CreatedAt + ttl*int64(time.Millisecond)
}

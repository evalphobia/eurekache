package eurekache

import (
	"math"
	"time"
)

type Item struct {
	// creation time
	CreatedAt int64
	ExpiredAt int64

	// The actual item stored in this item.
	value interface{}
}

func (i *Item) init() {
	i.CreatedAt = time.Now().UnixNano()
	i.ExpiredAt = math.MaxInt64
	i.value = nil
}

func (i *Item) SetExpire(ttl int64) {
	if ttl == 0 {
		return
	}
	i.ExpiredAt = i.CreatedAt + ttl*int64(time.Millisecond)
}

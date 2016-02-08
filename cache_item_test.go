package eurekache

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestItemInit(t *testing.T) {
	assert := assert.New(t)
	start := time.Now().UnixNano()

	item := &Item{}
	item.value = "the value"
	item.init()

	end := time.Now().UnixNano()

	assert.True(start < item.CreatedAt)
	assert.True(item.CreatedAt < end)
	assert.EqualValues(math.MaxInt64, item.ExpiredAt)
	assert.Nil(item.value)
}

func TestItemSetExpire(t *testing.T) {
	assert := assert.New(t)

	item := &Item{}
	item.SetExpire(0)
	assert.EqualValues(0, item.ExpiredAt)
	assert.EqualValues(0, item.CreatedAt)
	assert.Nil(item.value)

	item.SetExpire(100)
	assert.EqualValues(100*int64(time.Millisecond), item.ExpiredAt)
	assert.EqualValues(0, item.CreatedAt)
	assert.Nil(item.value)

	item.CreatedAt = time.Now().UnixNano()
	item.SetExpire(100)
	assert.EqualValues(item.CreatedAt+100*int64(time.Millisecond), item.ExpiredAt)
}

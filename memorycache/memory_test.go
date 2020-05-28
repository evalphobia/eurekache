package memorycache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/evalphobia/eurekache"
)

func TestNewCacheTTL(t *testing.T) {
	assert := assert.New(t)
	m := NewCacheTTL(1)

	assert.NotNil(m)
	assert.Len(m.items, 0)
	assert.Equal(1, m.maxSize)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	m := NewCacheTTL(1)
	strValue := "the value"

	item := &eurekache.Item{
		CreatedAt: 1,
		ExpiredAt: math.MaxInt64,
		Value:     strValue,
	}
	m.items["key"] = item

	var result string
	var ok bool

	// miss cache
	ok = m.Get("nokey", &result)
	assert.False(ok)

	// hit cache
	ok = m.Get("key", &result)
	assert.True(ok)
	assert.Equal(strValue, result)

	// nil value
	item.Value = nil
	ok = m.Get("key", &result)
	assert.False(ok)
}

func TestGetInterface(t *testing.T) {
	assert := assert.New(t)
	m := NewCacheTTL(1)
	strValue := "the value"

	item := &eurekache.Item{
		CreatedAt: 1,
		ExpiredAt: math.MaxInt64,
		Value:     strValue,
	}
	m.items["key"] = item

	var result interface{}
	var ok bool

	// miss cache
	result, ok = m.GetInterface("nokey")
	assert.False(ok)
	assert.Nil(result)

	// hit cache
	result, ok = m.GetInterface("key")
	assert.True(ok)
	assert.Equal(strValue, result.(string))

	// nil value
	item.Value = nil
	result, ok = m.GetInterface("key")
	assert.False(ok)
	assert.Nil(result)
}

func TestGetGobBytes(t *testing.T) {
	assert := assert.New(t)
	m := NewCacheTTL(1)
	strValue := "the value"

	item := &eurekache.Item{
		CreatedAt: 1,
		ExpiredAt: math.MaxInt64,
		Value:     strValue,
	}
	m.items["key"] = item

	var b []byte
	var ok bool

	// miss cache
	b, ok = m.GetGobBytes("nokey")
	assert.False(ok)
	assert.Empty(b)

	// hit cache
	b, ok = m.GetGobBytes("key")
	assert.True(ok)

	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	var str string
	dec.Decode(&str)
	assert.Equal(strValue, str)

	// nil value
	item.Value = nil
	b, ok = m.GetGobBytes("key")
	assert.False(ok)
	assert.Empty(b)
}

func TestSet(t *testing.T) {
	assert := assert.New(t)
	m := NewCacheTTL(1)
	strValue := "the value"

	m.Set("key", strValue)

	var item *eurekache.Item
	var ok bool

	// miss cache
	item, ok = m.items["nokey"]
	assert.False(ok)
	assert.Nil(item)

	// hit cache
	item, ok = m.items["key"]
	assert.True(ok)
	assert.Equal(strValue, item.Value)
	assert.EqualValues(math.MaxInt64, item.ExpiredAt)
}

func TestSetExpire(t *testing.T) {
	assert := assert.New(t)
	m := NewCacheTTL(1)
	strValue := "the value"

	m.SetExpire("key", strValue, 2000)

	var item *eurekache.Item
	var ok bool

	// miss cache
	item, ok = m.items["nokey"]
	assert.False(ok)
	assert.Nil(item)

	// hit cache
	item, ok = m.items["key"]
	assert.True(ok)
	assert.Equal(strValue, item.Value)

	expected := item.CreatedAt + 2000*int64(time.Millisecond)
	assert.EqualValues(expected, item.ExpiredAt)
}

func TestDeleteOldest(t *testing.T) {
	assert := assert.New(t)
	m := NewCacheTTL(4)

	m.Set("key1", "value1")
	m.Set("key2", "value2")
	m.Set("key3", "value3")
	m.Set("key4", "value4")
	assert.Len(m.items, 4)
	assert.Len(m.deleteQueue, 4)
	assert.Equal("key1", m.deleteQueue[0])

	m.deleteOldest()
	assert.Len(m.items, 3)
	assert.Len(m.deleteQueue, 3)
	assert.Equal("key2", m.deleteQueue[0])

	// item is deleted, but not deleted in queue
	m.Set("key2", nil)
	assert.Len(m.items, 2)
	assert.Len(m.deleteQueue, 3)
	assert.Equal("key2", m.deleteQueue[0])

	// delete key2 and key3
	m.deleteOldest()
	assert.Len(m.items, 1)
	assert.Len(m.deleteQueue, 1)
	assert.Equal("key4", m.deleteQueue[0])

	m.deleteOldest()
	assert.Len(m.items, 0)
	assert.Len(m.deleteQueue, 0)
}

func TestClear(t *testing.T) {
	assert := assert.New(t)
	m := NewCacheTTL(4)

	err := m.Clear()
	assert.NoError(err)

	// set data
	m.Set("key1", "value1")
	m.Set("key2", "value2")
	m.Set("key3", "value3")
	m.Set("key4", "value4")
	assert.Len(m.items, 4)
	assert.Len(m.deleteQueue, 4)
	assert.Equal("key1", m.deleteQueue[0])

	// clear
	err = m.Clear()
	assert.NoError(err)
	assert.Len(m.items, 0)
	assert.Len(m.deleteQueue, 0)

	// set again
	m.Set("key1", "value1")
	m.Set("key2", "value2")
	assert.Len(m.items, 2)
	assert.Len(m.deleteQueue, 2)
	assert.Equal("key1", m.deleteQueue[0])

	// clear again
	err = m.Clear()
	assert.NoError(err)
	assert.Len(m.items, 0)
	assert.Len(m.deleteQueue, 0)
}

func TestIsValidItem(t *testing.T) {
	assert := assert.New(t)
	m := NewCacheTTL(1)
	strValue := "the value"

	m.SetExpire("key", strValue, 100)

	var item *eurekache.Item
	var ok bool
	item, ok = m.items["key"]
	assert.True(ok)

	ok = m.isValidItem(item)
	assert.True(ok)

	time.Sleep(10 * time.Millisecond)
	ok = m.isValidItem(item)
	assert.True(ok)

	time.Sleep(100 * time.Millisecond)
	ok = m.isValidItem(item)
	assert.False(ok)

	// nil value
	m.SetExpire("key", nil, 10000)
	ok = m.isValidItem(item)
	assert.False(ok)
}

func TestItems(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		max   int
		items int
	}{
		{1, 1},
		{1, 2},
		{3, 4},
		{100, 10},
		{100, 99},
		{100, 100},
		{100, 101},
		{100, 1000},
	}

	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)

		m := NewCacheTTL(tt.max)

		// set and check data
		for i, max := 0, tt.max; i < max; i++ {
			key := strconv.Itoa(i)

			var result string
			var ok bool

			ok = m.Get(key, &result)
			assert.False(ok, target)
			assert.Empty(result, target)

			m.Set(key, key)

			ok = m.Get(key, &result)
			assert.Equal(true, ok, target)
			assert.Equal(key, result, target)
		}

		// above maximum
		for i, max := tt.max, tt.items; i < max; i++ {
			delIndex := i - tt.max
			delKey := strconv.Itoa(delIndex)

			// check before set
			assert.Equal(delKey, m.deleteQueue[0], target)
			assert.Len(m.items, tt.max, target)
			assert.Len(m.deleteQueue, tt.max, target)

			// set data
			key := strconv.Itoa(i)
			m.Set(key, key)

			var result string
			ok := m.Get(key, &result)
			assert.Equal(true, ok, target)
			assert.Equal(key, result, target)

			// check after set
			assert.NotEqual(delKey, m.deleteQueue[0], target)
			assert.Len(m.deleteQueue, tt.max, target)
		}

		// check delete
		m = NewCacheTTL(tt.max)

		// set data
		for i, max := 0, tt.max; i < max; i++ {
			key := strconv.Itoa(i)
			m.Set(key, key)

			if i%3 == 0 {
				m.Set(key, nil)
			}
		}

		// check data with deleted key
		for i, max := 0, tt.items; i < max; i++ {
			delKey := "@" + strconv.Itoa(i)

			// check before set
			assert.True(len(m.deleteQueue) <= tt.max, target)

			// set data
			key := strconv.Itoa(i)
			m.Set(key, key)

			var result string
			ok := m.Get(key, &result)
			assert.Equal(true, ok, target)
			assert.Equal(key, result, target)

			// check after set
			assert.NotEqual(delKey, m.deleteQueue[0], target)
			assert.True(len(m.deleteQueue) <= tt.max, target)
		}
	}
}

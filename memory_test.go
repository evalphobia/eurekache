package eurekache

import (
	"bytes"
	"encoding/gob"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryCacheTTL(t *testing.T) {
	assert := assert.New(t)
	m := NewMemoryCacheTTL(1)

	assert.NotNil(m)
	assert.Len(m.items, 0)
	assert.Equal(1, m.maxSize)
}

func TestMemoryCacheTTLGet(t *testing.T) {
	assert := assert.New(t)
	m := NewMemoryCacheTTL(1)
	strValue := "the value"

	item := &Item{
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
}

func TestMemoryCacheTTLGetInterface(t *testing.T) {
	assert := assert.New(t)
	m := NewMemoryCacheTTL(1)
	strValue := "the value"

	item := &Item{
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
}

func TestMemoryCacheTTLGetGobBytes(t *testing.T) {
	assert := assert.New(t)
	m := NewMemoryCacheTTL(1)
	strValue := "the value"

	item := &Item{
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
}

func TestMemoryCacheTTLSet(t *testing.T) {
	assert := assert.New(t)
	m := NewMemoryCacheTTL(1)
	strValue := "the value"

	m.Set("key", strValue)

	var item *Item
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

func TestMemoryCacheTTLSetExpire(t *testing.T) {
	assert := assert.New(t)
	m := NewMemoryCacheTTL(1)
	strValue := "the value"

	m.SetExpire("key", strValue, 2000)

	var item *Item
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

func TestMemoryCacheTTLIsValidItem(t *testing.T) {
	assert := assert.New(t)
	m := NewMemoryCacheTTL(1)
	strValue := "the value"

	m.SetExpire("key", strValue, 100)

	var item *Item
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
}

func TestMemoryCacheTTLGetNextReplacement(t *testing.T) {
	assert := assert.New(t)
	k1 := "key1"
	v1 := "value1"
	k2 := "key2"
	v2 := "value2"
	k3 := "key3"
	v3 := "value3"
	var key string
	var item *Item

	// maximum slot = 1
	m := NewMemoryCacheTTL(1)
	m.Set(k1, v1)
	key, item = m.getNextReplacement()
	assert.Equal(k1, key, "it shuld be first item")
	assert.Equal(v1, item.Value, "it shuld be first item")

	// maximum slot = 2
	m = NewMemoryCacheTTL(2)
	m.Set(k1, v1)
	key, item = m.getNextReplacement()
	assert.Equal("", key, "it shuld be new empty item")
	assert.Nil(item.Value, "it shuld be new empty item")

	m.Set(k2, v2)
	key, item = m.getNextReplacement()
	assert.Equal(k1, key, "it shuld be first item")
	assert.Equal(v1, item.Value, "it shuld be first item")

	m.Set(k3, v3)
	key, item = m.getNextReplacement()
	assert.Equal(k2, key, "it shuld be second item")
	assert.Equal(v2, item.Value, "it shuld be second item")
}

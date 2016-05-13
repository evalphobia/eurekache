package eurekache_test

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/evalphobia/eurekache"
	"github.com/evalphobia/eurekache/memory"
)

func TestExtraEurekacheGet(t *testing.T) {
	assert := assert.New(t)

	e := New()
	m := memory.NewMemoryCacheTTL(1)
	e.SetCacheSources([]Cache{m})

	m.Set("key", "value")

	var ok bool
	var valInt int
	var valStr string

	// int value
	ok = e.Get("key", &valInt)
	assert.False(ok)
	assert.Empty(valInt)

	// string value
	ok = e.Get("key", &valStr)
	assert.True(ok)
	assert.Equal("value", valStr)

	// struct value
	var valItem1, valItem2 Item
	var valNonItem *Eurekache
	valItem1 = Item{
		Value: "val",
	}
	m.Set("item", valItem1)

	ok = e.Get("item", valNonItem)
	assert.False(ok)
	assert.Nil(valNonItem)

	ok = e.Get("item", &valItem2)
	assert.True(ok)
	assert.Equal(valItem1, valItem2)

	// pointer value
	var valPtr1, valPtr2 *Item
	m.Set("item_ptr", &valPtr1)

	ok = e.Get("item_ptr", &valPtr2)
	assert.True(ok)
	assert.Equal(valPtr1, valPtr2)

}

func TestExtraEurekacheGetInterface(t *testing.T) {
	assert := assert.New(t)
	val := "value"

	e := New()
	m := memory.NewMemoryCacheTTL(1)
	e.SetCacheSources([]Cache{m})
	m.Set("key", val)

	var result interface{}
	var ok bool

	result, ok = e.GetInterface("key")
	assert.True(ok)
	assert.Equal(val, result)

	result, ok = e.GetInterface("nokey")
	assert.False(ok)
	assert.Nil(result)
}

func TestExtraEurekacheGetGobBytes(t *testing.T) {
	assert := assert.New(t)
	val := "value"

	// encode value
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(val)
	assert.Nil(err)

	e := New()
	m := memory.NewMemoryCacheTTL(1)
	e.SetCacheSources([]Cache{m})
	m.Set("key", val)

	var b []byte
	var ok bool

	// check gob encoded result
	b, ok = e.GetGobBytes("key")
	assert.True(ok)
	assert.Equal(buf.Bytes(), b)

	// check gob encoded result
	b, ok = e.GetGobBytes("nokey")
	assert.False(ok)
	assert.Empty(b)
}

func TestExtraEurekacheSet(t *testing.T) {
	assert := assert.New(t)
	val := "value"

	e := New()
	m := memory.NewMemoryCacheTTL(1)
	e.SetCacheSources([]Cache{m})
	e.Set("key", val)

	var item *Item
	var ok bool

	var result string
	ok = m.Get("nokey", &result)
	assert.False(ok)
	assert.Nil(item)

	ok = m.Get("key", &result)
	assert.True(ok)
	assert.Equal(val, result)

}

func TestExtraEurekacheSetExpire(t *testing.T) {
	assert := assert.New(t)
	val := "value"

	e := New()
	m := memory.NewMemoryCacheTTL(1)
	e.SetCacheSources([]Cache{m})
	e.SetExpire("key", val, 100)

	var result string
	var ok bool

	ok = m.Get("nokey", &result)
	assert.False(ok)
	assert.Empty(result)

	result = ""
	ok = e.Get("key", &result)
	assert.True(ok)
	assert.Equal(val, result)

	result = ""
	time.Sleep(100 * time.Millisecond)
	ok = e.Get("key", &result)
	assert.False(ok)
	assert.Empty(result)
}

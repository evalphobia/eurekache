package eurekache

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	e := New()
	assert.NotNil(e)
	assert.Empty(e.caches)
}

func TestEurekacheSetCacheAlgolithms(t *testing.T) {
	assert := assert.New(t)

	e := New()
	e.SetCacheAlgolithms(nil)
	assert.Empty(e.caches)

	m := NewMemoryCacheTTL(1)
	e.SetCacheAlgolithms([]cache{m})
	assert.Len(e.caches, 1)
	assert.Equal(m, e.caches[0])
}

func TestEurekacheGet(t *testing.T) {
	assert := assert.New(t)

	e := New()
	m := NewMemoryCacheTTL(1)
	e.SetCacheAlgolithms([]cache{m})

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
		value: "val",
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

func TestEurekacheGetInterface(t *testing.T) {
	assert := assert.New(t)
	val := "value"

	e := New()
	m := NewMemoryCacheTTL(1)
	e.SetCacheAlgolithms([]cache{m})
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

func TestEurekacheGetGobBytes(t *testing.T) {
	assert := assert.New(t)
	val := "value"

	// encode value
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(val)
	assert.Nil(err)

	e := New()
	m := NewMemoryCacheTTL(1)
	e.SetCacheAlgolithms([]cache{m})
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

func TestEurekacheSet(t *testing.T) {
	assert := assert.New(t)
	val := "value"

	e := New()
	m := NewMemoryCacheTTL(1)
	e.SetCacheAlgolithms([]cache{m})
	e.Set("key", val)

	var item *Item
	var ok bool

	item, ok = m.items["key"]
	assert.True(ok)
	assert.Equal(val, item.value)

	item, ok = m.items["nokey"]
	assert.False(ok)
	assert.Nil(item)
}

func TestEurekacheSetExpire(t *testing.T) {
	assert := assert.New(t)
	val := "value"

	e := New()
	m := NewMemoryCacheTTL(1)
	e.SetCacheAlgolithms([]cache{m})
	e.SetExpire("key", val, 100)

	var item *Item
	var ok bool

	item, ok = m.items["key"]
	assert.True(ok)
	assert.Equal(val, item.value)
	expected := item.CreatedAt + 100*int64(time.Millisecond)
	assert.EqualValues(expected, item.ExpiredAt)

	item, ok = m.items["nokey"]
	assert.False(ok)
	assert.Nil(item)

	t.Skip("Eurekache.Get() must be implemented")
}

func TestCopyValue(t *testing.T) {
	assert := assert.New(t)

	var ok bool
	var valStr1, valStr2, valStr3 string
	var valInt1, valInt2, valInt3 int
	var valSlice1, valSlice2, valSlice3 []string
	var valMap1, valMap2, valMap3 map[interface{}]interface{}
	var valStruct1, valStruct2, valStruct3 Item

	valStr1 = "val"
	valInt1 = 99
	valSlice1 = []string{"val1", "val2"}
	valMap1 = map[interface{}]interface{}{
		"key1": "value1",
		"key2": 100,
	}
	valStruct1 = Item{
		value: "value",
	}

	// string
	ok = copyValue(&valStr2, valStr1)
	assert.True(ok)
	assert.Equal(valStr1, valStr2)
	ok = copyValue(&valStr3, &valStr1)
	assert.Equal(valStr1, valStr3)

	// int
	ok = copyValue(&valInt2, valInt1)
	assert.True(ok)
	assert.Equal(valInt1, valInt2)
	ok = copyValue(&valInt3, &valInt1)
	assert.Equal(valInt1, valInt3)

	// slice
	ok = copyValue(&valSlice2, valSlice1)
	assert.True(ok)
	assert.Equal(valSlice1, valSlice2)
	ok = copyValue(&valSlice3, &valSlice1)
	assert.Equal(valSlice1, valSlice3)

	// map
	ok = copyValue(&valMap2, valMap1)
	assert.True(ok)
	assert.Equal(valMap1, valMap2)
	ok = copyValue(&valMap3, &valMap1)
	assert.Equal(valMap1, valMap3)

	// struct
	ok = copyValue(&valStruct2, valStruct1)
	assert.True(ok)
	assert.Equal(valStruct1, valStruct2)
	ok = copyValue(&valStruct3, &valStruct1)
	assert.Equal(valStruct1, valStruct3)
}

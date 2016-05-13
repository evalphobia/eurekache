package eurekache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	e := New()
	assert.NotNil(e)
	assert.Empty(e.caches)
}

func TestSetCacheSources(t *testing.T) {
	assert := assert.New(t)

	e := New()
	e.SetCacheSources(nil)
	assert.Empty(e.caches)

	m := newDummyCache()
	e.SetCacheSources([]Cache{m})
	assert.Len(e.caches, 1)
	assert.Equal(m, e.caches[0])
}

func TestAddCacheSource(t *testing.T) {
	assert := assert.New(t)

	e := New()
	e.AddCacheSource(nil)
	assert.Empty(e.caches)

	m1 := newDummyCache()
	e.AddCacheSource(m1)
	assert.Len(e.caches, 1)

	m2 := newDummyCache()
	e.AddCacheSource(m2)
	assert.Len(e.caches, 2)

	assert.Equal(m1, e.caches[0])
	assert.Equal(m2, e.caches[1])
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	e := New()
	m := newDummyCache()
	e.SetCacheSources([]Cache{m})

	// dummy
	var result string
	ok := e.Get("key", &result)
	assert.False(ok)
	assert.Empty(result)
}

func TestGetInterface(t *testing.T) {
	assert := assert.New(t)

	e := New()
	m := newDummyCache()
	e.SetCacheSources([]Cache{m})

	result, ok := e.GetInterface("key")
	assert.False(ok)
	assert.Empty(result)
}

func TestGetGobBytes(t *testing.T) {
	assert := assert.New(t)

	e := New()
	m := newDummyCache()
	e.SetCacheSources([]Cache{m})

	b, ok := e.GetGobBytes("key")
	assert.False(ok)
	assert.Empty(b)
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
		Value: "value",
	}

	// string
	ok = CopyValue(&valStr2, valStr1)
	assert.True(ok)
	assert.Equal(valStr1, valStr2)
	ok = CopyValue(&valStr3, &valStr1)
	assert.Equal(valStr1, valStr3)

	// int
	ok = CopyValue(&valInt2, valInt1)
	assert.True(ok)
	assert.Equal(valInt1, valInt2)
	ok = CopyValue(&valInt3, &valInt1)
	assert.Equal(valInt1, valInt3)

	// slice
	ok = CopyValue(&valSlice2, valSlice1)
	assert.True(ok)
	assert.Equal(valSlice1, valSlice2)
	ok = CopyValue(&valSlice3, &valSlice1)
	assert.Equal(valSlice1, valSlice3)

	// map
	ok = CopyValue(&valMap2, valMap1)
	assert.True(ok)
	assert.Equal(valMap1, valMap2)
	ok = CopyValue(&valMap3, &valMap1)
	assert.Equal(valMap1, valMap3)

	// struct
	ok = CopyValue(&valStruct2, valStruct1)
	assert.True(ok)
	assert.Equal(valStruct1, valStruct2)
	ok = CopyValue(&valStruct3, &valStruct1)
	assert.Equal(valStruct1, valStruct3)
}

type dummyCache struct{}

func (d *dummyCache) Get(k string, v interface{}) bool {
	return false
}

func (d *dummyCache) GetInterface(k string) (interface{}, bool) {
	return nil, false
}

func (d *dummyCache) GetGobBytes(k string) ([]byte, bool) {
	return nil, false
}

func (d *dummyCache) Set(k string, v interface{}) error {
	return nil
}

func (d *dummyCache) SetExpire(k string, v interface{}, i int64) error {
	return nil
}

func newDummyCache() *dummyCache {
	return &dummyCache{}
}

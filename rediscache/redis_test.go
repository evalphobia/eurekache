package rediscache

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"

	"github.com/evalphobia/eurekache"
	"github.com/evalphobia/eurekache/test/helper"
)

var testRedisPrefix = "eurekache:"

func TestNewRedisCache(t *testing.T) {
	assert := assert.New(t)

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)

	assert.NotNil(c)
	assert.Equal(pool, c.pool)
	assert.Equal(c.dbno, "0")
	assert.EqualValues(c.defaultTTL, 0)
}

func TesSetPrefix(t *testing.T) {
	assert := assert.New(t)

	c := NewRedisCache(nil)
	assert.Equal(c.prefix, "")

	c.SetPrefix(testRedisPrefix)
	assert.Equal(c.prefix, testRedisPrefix)
}

func TesSelect(t *testing.T) {
	assert := assert.New(t)

	c := NewRedisCache(nil)
	assert.Equal(c.dbno, "0")

	c.Select(1)
	assert.Equal(c.dbno, "1")
}

func TesGet(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	// set data
	b := helper.TestGobItem("valueTesGet")
	_, err := pool.Get().Do("SETEX", testRedisPrefix+key, 300, b)
	assert.Nil(err)

	// get data
	var result string
	ok := c.Get(key, &result)
	assert.True(ok)
	assert.Equal("valueTesGet", result)

	// nil value
	b = helper.TestGobItem(nil)
	_, err = pool.Get().Do("SETEX", testRedisPrefix+key, 300, b)
	assert.Nil(err)
	var result2 string
	ok = c.Get(key, &result2)
	assert.False(ok)
	assert.Empty(result2)
}

func TesGetInterface(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	// set data
	b := helper.TestGobItem("valueTesGetInterface")
	_, err := pool.Get().Do("SETEX", testRedisPrefix+key, 300, b)
	assert.Nil(err)

	// get data
	v, ok := c.GetInterface(key)
	assert.True(ok)
	assert.Equal("valueTesGetInterface", v)
}

func TesGetGobBytes(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	// set data
	b := helper.TestGobItem("valueTesGetGobBytes")
	_, err := pool.Get().Do("SETEX", testRedisPrefix+key, 300, b)
	assert.Nil(err)

	// get data
	b, ok := c.GetGobBytes(key)
	assert.True(ok)

	var result string
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&result)
	assert.Nil(err)
	assert.Equal("valueTesGetGobBytes", result)
}

func TesSet(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	err := c.Set(key, "valueTesSet")
	assert.Nil(err)

	// get data
	b, err := pool.Get().Do("GET", testRedisPrefix+key)
	assert.Nil(err)
	b, err = redis.Bytes(b, err)
	assert.Nil(err)

	buf := bytes.NewBuffer(b.([]byte))
	dec := gob.NewDecoder(buf)

	item := &eurekache.Item{}
	err = dec.Decode(&item)
	assert.Nil(err)
	assert.Equal("valueTesSet", item.Value)
}

func TesSetExpire(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	err := c.SetExpire(key, "valueTesSetExpire", 1000)
	assert.Nil(err)

	// get data
	var v string
	var ok bool

	ok = c.Get(key, &v)
	assert.True(ok)

	time.Sleep(200 * time.Millisecond)
	ok = c.Get(key, &v)
	assert.True(ok)

	time.Sleep(1 * time.Second)
	ok = c.Get(key, &v)
	assert.False(ok)
}

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

func TestSetTTL(t *testing.T) {
	assert := assert.New(t)

	c := NewRedisCache(nil)
	assert.EqualValues(c.defaultTTL, 0)

	c.SetTTL(100)
	assert.EqualValues(c.defaultTTL, 100)
}

func TestSetPrefix(t *testing.T) {
	assert := assert.New(t)

	c := NewRedisCache(nil)
	assert.Equal(c.prefix, "")

	c.SetPrefix(testRedisPrefix)
	assert.Equal(c.prefix, testRedisPrefix)
}

func TestSelect(t *testing.T) {
	assert := assert.New(t)

	c := NewRedisCache(nil)
	assert.Equal(c.dbno, "0")

	c.Select(1)
	assert.Equal(c.dbno, "1")
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	// set data
	b := helper.TestGobItem("valueTestGet")
	_, err := pool.Get().Do("SETEX", testRedisPrefix+key, 300, b)
	assert.Nil(err)

	// get data
	var result string
	ok := c.Get(key, &result)
	assert.True(ok)
	assert.Equal("valueTestGet", result)

	// nil value
	b = helper.TestGobItem(nil)
	_, err = pool.Get().Do("SETEX", testRedisPrefix+key, 300, b)
	assert.Nil(err)
	var result2 string
	ok = c.Get(key, &result2)
	assert.False(ok)
	assert.Empty(result2)
}

func TestGetInterface(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	// set data
	b := helper.TestGobItem("valueTestGetInterface")
	_, err := pool.Get().Do("SETEX", testRedisPrefix+key, 300, b)
	assert.Nil(err)

	// get data
	v, ok := c.GetInterface(key)
	assert.True(ok)
	assert.Equal("valueTestGetInterface", v)
}

func TestGetGobBytes(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	// set data
	b := helper.TestGobItem("valueTestGetGobBytes")
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
	assert.Equal("valueTestGetGobBytes", result)
}

func TestSet(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	err := c.Set(key, "valueTestSet")
	assert.NoError(err)

	// get data
	b, err := pool.Get().Do("GET", testRedisPrefix+key)
	assert.NoError(err)
	b, err = redis.Bytes(b, err)
	assert.NoError(err)

	buf := bytes.NewBuffer(b.([]byte))
	dec := gob.NewDecoder(buf)

	item := &eurekache.Item{}
	err = dec.Decode(&item)
	assert.NoError(err)
	assert.Equal("valueTestSet", item.Value)

	// delete data
	err = c.Set(key, nil)
	assert.NoError(err)

	b, err = pool.Get().Do("GET", testRedisPrefix+key)
	assert.NoError(err)
	assert.Nil(b)

}

func TestSetExpire(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := helper.TestGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	err := c.SetExpire(key, "valueTestSetExpire", 1000)
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

func TestConn(t *testing.T) {
	assert := assert.New(t)

	c := NewRedisCache(helper.TestGetPool())
	conn, err := c.conn()
	assert.NoError(err)
	assert.NotNil(conn)

	c.pool = nil
	conn, err = c.conn()
	assert.Equal(errNilPool, err)
	assert.Nil(conn)
}

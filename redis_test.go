package eurekache

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

var testRedisHost string = "127.0.0.1:6379"
var testRedisPrefix = "eurekache_"

func TestNewRedisCache(t *testing.T) {
	assert := assert.New(t)

	pool := testGetPool()
	c := NewRedisCache(pool)

	assert.NotNil(c)
	assert.Equal(pool, c.pool)
	assert.Equal(c.dbno, "0")
	assert.EqualValues(c.defaultTTL, 0)
}

func TestRedisCacheSetPrefix(t *testing.T) {
	assert := assert.New(t)

	c := NewRedisCache(nil)
	assert.Equal(c.prefix, "")

	c.SetPrefix(testRedisPrefix)
	assert.Equal(c.prefix, testRedisPrefix)
}

func TestRedisCacheSelect(t *testing.T) {
	assert := assert.New(t)

	c := NewRedisCache(nil)
	assert.Equal(c.dbno, "0")

	c.Select(1)
	assert.Equal(c.dbno, "1")
}

func TestRedisCacheGet(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := testGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	// set data
	b := testGobItem("valueTestRedisCacheGet")
	_, err := pool.Get().Do("SETEX", testRedisPrefix+key, 300, b)
	assert.Nil(err)

	// get data
	var result string
	ok := c.Get(key, &result)
	assert.True(ok)
	assert.Equal("valueTestRedisCacheGet", result)
}

func TestRedisCacheGetInterface(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := testGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	// set data
	b := testGobItem("valueTestRedisCacheGetInterface")
	_, err := pool.Get().Do("SETEX", testRedisPrefix+key, 300, b)
	assert.Nil(err)

	// get data
	v, ok := c.GetInterface(key)
	assert.True(ok)
	assert.Equal("valueTestRedisCacheGetInterface", v)
}

func TestRedisCacheGetGobBytes(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := testGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	// set data
	b := testGobItem("valueTestRedisCacheGetGobBytes")
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
	assert.Equal("valueTestRedisCacheGetGobBytes", result)
}

func TestRedisCacheSet(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := testGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	err := c.Set(key, "valueTestRedisCacheSet")
	assert.Nil(err)

	// get data
	b, err := pool.Get().Do("GET", testRedisPrefix+key)
	assert.Nil(err)
	b, err = redis.Bytes(b, err)
	assert.Nil(err)

	buf := bytes.NewBuffer(b.([]byte))
	dec := gob.NewDecoder(buf)

	item := &Item{}
	err = dec.Decode(&item)
	assert.Nil(err)
	assert.Equal("valueTestRedisCacheSet", item.Value)
}

func TestRedisCacheSetExpire(t *testing.T) {
	assert := assert.New(t)
	key := "key"

	pool := testGetPool()
	c := NewRedisCache(pool)
	c.SetPrefix(testRedisPrefix)

	err := c.SetExpire(key, "valueTestRedisCacheSetExpire", 1000)
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

func testGetPool() *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", testRedisHost)
		},
	}
}
func testGobItem(v interface{}) []byte {
	item := &Item{}
	item.init()
	item.Value = v

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(item)
	return buf.Bytes()
}

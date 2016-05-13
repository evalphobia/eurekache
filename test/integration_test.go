package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/evalphobia/eurekache"
	"github.com/evalphobia/eurekache/memorycache"
	"github.com/evalphobia/eurekache/rediscache"
	"github.com/evalphobia/eurekache/test/helper"
)

var testRedisPrefix = "eurekache:integration:"

func TestIntegrationGet(t *testing.T) {
	assert := assert.New(t)
	key := "testintegrationget"
	val := "TestIntegrationGet"

	mc := memorycache.NewCacheTTL(3)
	mc.SetTTL(200)
	rc := rediscache.NewRedisCache(helper.TestGetPool())
	rc.SetTTL(1000)
	rc.SetPrefix(testRedisPrefix)

	e := eurekache.New()
	e.SetCacheSources([]eurekache.Cache{mc, rc})
	e.Set(key, val)

	var ok bool
	var result string

	// miss cache
	ok = e.Get("key", &result)
	assert.False(ok)
	assert.Empty(result)

	// hit memory
	ok = e.Get(key, &result)
	assert.True(ok)
	assert.Equal(val, result)

	// hit redis
	result = ""
	time.Sleep(300 * time.Millisecond)
	ok = e.Get(key, &result)
	assert.True(ok)
	assert.Equal(val, result)

	// cache expired
	result = ""
	time.Sleep(1000 * time.Millisecond)
	ok = e.Get(key, &result)
	assert.False(ok)
	assert.Equal("", result)
}

func TestIntegrationGetTimeout(t *testing.T) {
	assert := assert.New(t)
	key := "testintegrationgettimeout"
	val := "TestIntegrationGetTimeout"

	dc := newDummySleepCache(100 * time.Millisecond)
	mc := memorycache.NewCacheTTL(3)

	e := eurekache.New()
	e.SetCacheSources([]eurekache.Cache{dc, mc})
	e.Set(key, val)

	var ok bool
	var result string

	// default timeout(hour)
	ok = e.Get(key, &result)
	assert.True(ok)
	assert.Equal(val, result)

	// timeout within 20ms
	result = ""
	e.SetTimeout(20 * time.Millisecond)
	ok = e.Get(key, &result)
	assert.False(ok)
	assert.Equal("", result)
}

type dummySleepCache struct {
	sleep time.Duration
}

func (d *dummySleepCache) Get(k string, v interface{}) bool {
	time.Sleep(d.sleep)
	return false
}

func (d *dummySleepCache) GetInterface(k string) (interface{}, bool) {
	time.Sleep(d.sleep)
	return nil, false
}

func (d *dummySleepCache) GetGobBytes(k string) ([]byte, bool) {
	time.Sleep(d.sleep)
	return nil, false
}

func (d *dummySleepCache) Set(k string, v interface{}) error {
	time.Sleep(d.sleep)
	return nil
}

func (d *dummySleepCache) SetExpire(k string, v interface{}, i int64) error {
	time.Sleep(d.sleep)
	return nil
}

func newDummySleepCache(sleep time.Duration) *dummySleepCache {
	return &dummySleepCache{sleep}
}

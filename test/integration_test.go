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

	mc := memorycache.NewCacheTTL(3)
	mc.SetTTL(200)
	rc := rediscache.NewRedisCache(helper.TestGetPool())
	rc.SetTTL(1000)
	rc.SetPrefix(testRedisPrefix)

	e := eurekache.New()
	e.SetCacheSources([]eurekache.Cache{mc, rc})
	e.Set("integration", "TestIntegrationGet")

	var ok bool
	var result string

	// miss cache
	ok = e.Get("key", &result)
	assert.False(ok)
	assert.Empty(result)

	// hit memory
	ok = e.Get("integration", &result)
	assert.True(ok)
	assert.Equal("TestIntegrationGet", result)

	// hit redis
	result = ""
	time.Sleep(300 * time.Millisecond)
	ok = e.Get("integration", &result)
	assert.True(ok)
	assert.Equal("TestIntegrationGet", result)

	// cache expired
	result = ""
	time.Sleep(1000 * time.Millisecond)
	ok = e.Get("integration", &result)
	assert.False(ok)
	assert.Equal("", result)
}

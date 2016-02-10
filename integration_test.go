package eurekache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationGet(t *testing.T) {
	assert := assert.New(t)

	mc := NewMemoryCacheTTL(3)
	mc.SetTTL(200)
	rc := NewRedisCache(testGetPool())
	rc.SetTTL(1000)
	rc.SetPrefix(testRedisPrefix)

	e := New()
	e.SetCacheSources([]cache{mc, rc})
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

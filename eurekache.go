// Package eurekache provides fallback cache system with multiple cache source
package eurekache

import (
	"reflect"
	"time"
)

// Cache is interface for storing data
type Cache interface {
	Get(string, interface{}) bool
	GetInterface(string) (interface{}, bool)
	GetGobBytes(string) ([]byte, bool)
	Set(string, interface{}) error
	SetExpire(string, interface{}, int64) error
}

// Eurekache will contains multiple cache source
type Eurekache struct {
	caches       []Cache
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// New returns empty new Eurekache
func New() *Eurekache {
	return &Eurekache{
		readTimeout:  time.Hour,
		writeTimeout: time.Hour,
	}
}

// SetCacheSources sets cache sources
func (e *Eurekache) SetCacheSources(caches []Cache) {
	e.caches = caches
}

// AddCacheSource adds cache source
func (e *Eurekache) AddCacheSource(cache Cache) {
	if cache == nil {
		return
	}

	e.caches = append(e.caches, cache)
}

// SetTimeout sets r/w timeout
func (e *Eurekache) SetTimeout(d time.Duration) {
	e.readTimeout = d
	e.writeTimeout = d
}

// SetReadTimeout sets read timeout
func (e *Eurekache) SetReadTimeout(d time.Duration) {
	e.readTimeout = d
}

// SetWriteTimeout sets write timeout
func (e *Eurekache) SetWriteTimeout(d time.Duration) {
	e.writeTimeout = d
}

// Get searches cache by given key and returns flag of cache is existed or not.
// when cache hit, data is assigned.
func (e *Eurekache) Get(key string, data interface{}) (ok bool) {
	ch := make(chan bool, 1)
	// get cache
	go func() {
		for _, c := range e.caches {
			ok = c.Get(key, data)
			if ok {
				ch <- true
				return
			}
		}
		ch <- false
	}()

	// get cache or timeout
	select {
	case <-ch:
		return
	case <-time.After(e.readTimeout):
		return false
	}
}

// GetInterface searches cache by given key and returns interface value.
func (e *Eurekache) GetInterface(key string) (v interface{}, ok bool) {
	ch := make(chan bool, 1)
	// get cache
	go func() {
		for _, c := range e.caches {
			v, ok = c.GetInterface(key)
			if ok {
				ch <- true
				return
			}
		}
		ch <- false
	}()

	// get cache or timeout
	select {
	case <-ch:
		return
	case <-time.After(e.readTimeout):
		return nil, false
	}
}

// GetGobBytes searches cache by given key and returns gob-encoded value.
func (e *Eurekache) GetGobBytes(key string) (b []byte, ok bool) {
	ch := make(chan bool, 1)
	// get cache
	go func() {
		for _, c := range e.caches {
			b, ok = c.GetGobBytes(key)
			if ok {
				ch <- true
				return
			}
		}
		ch <- false
	}()

	// get cache or timeout
	select {
	case <-ch:
		return
	case <-time.After(e.readTimeout):
		return nil, false
	}
}

// Set sets data into all of cache sources.
func (e *Eurekache) Set(key string, data interface{}) {
	ch := make(chan bool, 1)
	// set cache
	go func() {
		for _, c := range e.caches {
			c.Set(key, data)
		}
		ch <- true
	}()

	// set cache or timeout
	select {
	case <-ch:
		return
	case <-time.After(e.writeTimeout):
		return
	}
}

// SetExpire sets data with TTL.
func (e *Eurekache) SetExpire(key string, data interface{}, ttl int64) {
	ch := make(chan bool, 1)
	// set cache
	go func() {
		for _, c := range e.caches {
			c.SetExpire(key, data, ttl)
		}
		ch <- true
	}()

	// set cache or timeout
	select {
	case <-ch:
		return
	case <-time.After(e.writeTimeout):
		return
	}
}

// CopyValue copies srv value into dst.
func CopyValue(dst, src interface{}) bool {
	vvDst := reflect.ValueOf(dst)
	switch {
	case vvDst.Kind() != reflect.Ptr:
		// cannot assign value for non-pointer
		return false
	case vvDst.IsNil():
		// cannot assign value for nil
		return false
	}

	vvDst = vvDst.Elem()
	if !vvDst.CanSet() {
		return false
	}

	vvSrc := reflect.ValueOf(src)
	if vvSrc.Kind() == reflect.Ptr {
		vvSrc = vvSrc.Elem()
	}

	// type check
	switch {
	case vvSrc.Kind() != vvDst.Kind():
		return false
	case vvDst.Type() != vvSrc.Type():
		return false
	}

	vvDst.Set(vvSrc)
	return true
}

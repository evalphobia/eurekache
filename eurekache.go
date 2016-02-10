// Package eurekache provides fallback cache system with multiple cache source
package eurekache

import "reflect"

type cache interface {
	Get(string, interface{}) bool
	GetInterface(string) (interface{}, bool)
	GetGobBytes(string) ([]byte, bool)
	Set(string, interface{}) error
	SetExpire(string, interface{}, int64) error
}

// Eurekache will contains multiple cache source
type Eurekache struct {
	caches []cache
}

// New returns empty new Eurekache
func New() *Eurekache {
	return &Eurekache{}
}

// SetCacheSources sets cache sources
func (e *Eurekache) SetCacheSources(caches []cache) {
	e.caches = caches
}

// Get searches cache by given key and returns flag of cache is existed or not.
// when cache hit, data is assigned.
func (e *Eurekache) Get(key string, data interface{}) (ok bool) {
	for _, c := range e.caches {
		ok = c.Get(key, data)
		if ok {
			return
		}
	}
	return
}

// GetInterface searches cache by given key and returns interface value.
func (e *Eurekache) GetInterface(key string) (v interface{}, ok bool) {
	for _, c := range e.caches {
		v, ok = c.GetInterface(key)
		if ok {
			return
		}
	}
	return
}

// GetGobBytes searches cache by given key and returns gob-encoded value.
func (e *Eurekache) GetGobBytes(key string) (b []byte, ok bool) {
	for _, c := range e.caches {
		b, ok = c.GetGobBytes(key)
		if ok {
			return
		}
	}
	return
}

// Set sets data into all of cache sources.
func (e *Eurekache) Set(key string, data interface{}) {
	for _, c := range e.caches {
		c.Set(key, data)
	}
}

// SetExpire sets data with TTL.
func (e *Eurekache) SetExpire(key string, data interface{}, ttl int64) {
	for _, c := range e.caches {
		c.SetExpire(key, data, ttl)
	}
}

// copyValue copies srv value into dst.
func copyValue(dst, src interface{}) bool {
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

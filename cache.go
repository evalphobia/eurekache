package eurekache

import "reflect"

type cache interface {
	Get(string, interface{}) bool
	GetInterface(string) (interface{}, bool)
	GetGobBytes(string) ([]byte, bool)
	Set(string, interface{}) error
	SetExpire(string, interface{}, int64) error
}

type Eurekache struct {
	caches []cache
}

func New() *Eurekache {
	return &Eurekache{}
}

func (e *Eurekache) SetCacheAlgolithms(caches []cache) {
	e.caches = caches
}

func (e *Eurekache) Get(key string, data interface{}) (ok bool) {
	for _, c := range e.caches {
		ok = c.Get(key, data)
		if ok {
			return
		}
	}
	return
}

func (e *Eurekache) GetInterface(key string) (v interface{}, ok bool) {
	for _, c := range e.caches {
		v, ok = c.GetInterface(key)
		if ok {
			return
		}
	}
	return
}

func (e *Eurekache) GetGobBytes(key string) (b []byte, ok bool) {
	for _, c := range e.caches {
		b, ok = c.GetGobBytes(key)
		if ok {
			return
		}
	}
	return
}

func (e *Eurekache) Set(key string, data interface{}) {
	for _, c := range e.caches {
		c.Set(key, data)
	}
}

func (e *Eurekache) SetExpire(key string, data interface{}, ttl int64) {
	for _, c := range e.caches {
		c.SetExpire(key, data, ttl)
	}
}

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

package benches

import (
	"github.com/VictoriaMetrics/fastcache"
	"github.com/coocood/freecache"
	"github.com/qianbin/directcache"
)

const capacity = 32 * 1024 * 1024

type cache interface {
	get(key []byte) ([]byte, bool)
	set(key, val []byte)
	capacity() int
}

type getFunc func(key []byte) ([]byte, bool)
type setFunc func(key, val []byte)
type capacityFunc func() int
type nowFunc func() uint32

func (f getFunc) get(key []byte) ([]byte, bool) { return f(key) }
func (f setFunc) set(key, val []byte)           { f(key, val) }
func (f capacityFunc) capacity() int            { return f() }
func (f nowFunc) Now() uint32                   { return f() }

func newDirectCache() cache {
	c := directcache.New(capacity)
	return &struct {
		getFunc
		setFunc
		capacityFunc
	}{
		func(key []byte) ([]byte, bool) { return c.Get(key) },
		func(key, val []byte) { c.Set(key, val) },
		func() int { return c.Capacity() },
	}
}

func newDirectCacheWithPolicy(shouldEvict func(entry directcache.Entry) bool) cache {
	c := directcache.New(capacity)
	c.SetEvictionPolicy(shouldEvict)
	return &struct {
		getFunc
		setFunc
		capacityFunc
	}{
		func(key []byte) ([]byte, bool) { return c.Get(key) },
		func(key, val []byte) { c.Set(key, val) },
		func() int { return c.Capacity() },
	}
}

func newFreeCache() cache {
	t := uint32(0)
	c := freecache.NewCacheCustomTimer(capacity, nowFunc(func() uint32 {
		t++
		return t
	}))
	return &struct {
		getFunc
		setFunc
		capacityFunc
	}{
		func(key []byte) ([]byte, bool) {
			val, err := c.Get(key)
			return val, err == nil
		},
		func(key, val []byte) { c.Set(key, val, 0) },
		func() int { return capacity },
	}
}

func newFastCache() cache {
	c := fastcache.New(capacity)
	return &struct {
		getFunc
		setFunc
		capacityFunc
	}{
		func(key []byte) ([]byte, bool) { return c.HasGet(nil, key) },
		func(key, val []byte) { c.Set(key, val) },
		func() int { return capacity },
	}
}

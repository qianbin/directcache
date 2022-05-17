package benches

import (
	"encoding/binary"
	"math/rand"
	"testing"

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

func newDirectCache(cap int) cache {
	c := directcache.New(cap)
	return &struct {
		getFunc
		setFunc
		capacityFunc
	}{
		func(key []byte) ([]byte, bool) { return c.Get(key) },
		func(key, val []byte) { c.Set(key, val) },
		func() int { return cap },
	}
}

func newFreeCache(cap int) cache {
	t := uint32(0)
	c := freecache.NewCacheCustomTimer(cap, nowFunc(func() uint32 {
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
		func() int { return cap },
	}
}

func newFastCache(cap int) cache {
	c := fastcache.New(cap)
	return &struct {
		getFunc
		setFunc
		capacityFunc
	}{
		func(key []byte) ([]byte, bool) { return c.HasGet(nil, key) },
		func(key, val []byte) { c.Set(key, val) },
		func() int { return cap },
	}
}

func benchmarkGet(b *testing.B, c cache) {
	var (
		key     = make([]byte, 8)
		val     = make([]byte, 16)
		entries = c.capacity() / (len(key) + len(val)) / 2
	)

	// fill the cache
	for i := 0; i < entries; i++ {
		binary.BigEndian.PutUint64(key[:], uint64(i))
		c.set(key, val)
	}

	b.SetBytes(int64(len(key) + len(val)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint64(key[:], uint64(i%entries))
		c.get(key)
	}
}

func benchmarkSet(b *testing.B, c cache) {
	var (
		key = make([]byte, 8)
		val = make([]byte, 16)
	)

	b.SetBytes(int64(len(key) + len(val)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint64(key, uint64(i))
		c.set(key, val)
	}
}

func BenchmarkGet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkGet(b, newDirectCache(capacity)) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkGet(b, newFreeCache(capacity)) })
	b.Run("FastCache", func(b *testing.B) { benchmarkGet(b, newFastCache(capacity)) })
}

func BenchmarkSet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkSet(b, newDirectCache(capacity)) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkSet(b, newFreeCache(capacity)) })
	b.Run("FastCache", func(b *testing.B) { benchmarkSet(b, newFastCache(capacity)) })
}

func testHitrate(t *testing.T, c cache, entries int) {
	var (
		key    = make([]byte, 8)
		val    = make([]byte, c.capacity()/entries-len(key))
		z      = rand.NewZipf(rand.New(rand.NewSource(1)), 1.0001, 1, uint64(entries*10))
		hits   = 0
		misses = 0
	)

	for i := 0; i < entries*100; i++ {
		binary.BigEndian.PutUint64(key, z.Uint64())
		if _, has := c.get(key); has {
			hits++
		} else {
			c.set(key, val)
			misses++
		}
	}
	hitrate := float64(hits) / float64(hits+misses)
	t.Logf("hits: %d\tmisses: %d\thitrate: %.2f%%", hits, misses, hitrate*100)
}

func TestHitrate(t *testing.T) {
	const entries = 10000

	t.Run("DirectCache", func(t *testing.T) { testHitrate(t, newDirectCache(capacity), entries) })
	t.Run("FreeCache", func(t *testing.T) { testHitrate(t, newFreeCache(capacity), entries) })
	t.Run("FastCache", func(t *testing.T) { testHitrate(t, newFastCache(capacity), entries) })
}

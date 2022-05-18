package benches

import (
	"encoding/binary"
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/qianbin/directcache"
)

const capacity = 32 * 1024 * 1024

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

func benchmarkParallelGet(b *testing.B, c cache) {
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

	b.RunParallel(func(p *testing.PB) {
		i := 0
		for p.Next() {
			i++
			binary.BigEndian.PutUint64(key[:], uint64(i%entries))
			c.get(key)
		}
	})
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

func benchmarkParallelSet(b *testing.B, c cache) {
	var (
		key     = make([]byte, 8)
		val     = make([]byte, 16)
		counter uint64
	)

	b.SetBytes(int64(len(key) + len(val)))
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			binary.BigEndian.PutUint64(key, uint64(i))
			c.set(key, val)
		}
	})
}

func benchmarkParallelSetGet(b *testing.B, c cache) {
	var (
		key     = make([]byte, 8)
		val     = make([]byte, 16)
		counter uint64
	)
	const frac = 8

	b.SetBytes(int64(len(key) + len(val)))
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			binary.BigEndian.PutUint64(key, i/frac)
			if i%frac == 0 {
				c.set(key, val)
			}
			c.get(key)
		}
	})
}

func BenchmarkGet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkGet(b, newDirectCache(capacity)) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkGet(b, newFreeCache(capacity)) })
	b.Run("FastCache", func(b *testing.B) { benchmarkGet(b, newFastCache(capacity)) })
}

func BenchmarkParallelGet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkParallelGet(b, newDirectCache(capacity)) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkParallelGet(b, newFreeCache(capacity)) })
	b.Run("FastCache", func(b *testing.B) { benchmarkParallelGet(b, newFastCache(capacity)) })
}

func BenchmarkSet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkSet(b, newDirectCache(capacity)) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkSet(b, newFreeCache(capacity)) })
	b.Run("FastCache", func(b *testing.B) { benchmarkSet(b, newFastCache(capacity)) })
}

func BenchmarkParallelSet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkParallelSet(b, newDirectCache(capacity)) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkParallelSet(b, newFreeCache(capacity)) })
	b.Run("FastCache", func(b *testing.B) { benchmarkParallelSet(b, newFastCache(capacity)) })
}

func BenchmarkParallelSetGet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkParallelSetGet(b, newDirectCache(capacity)) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkParallelSetGet(b, newFreeCache(capacity)) })
	b.Run("FastCache", func(b *testing.B) { benchmarkParallelSetGet(b, newFastCache(capacity)) })
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
	t.Run("DirectCache(custom policy)", func(t *testing.T) {
		testHitrate(t, newDirectCacheWithPolicy(capacity, func(entry directcache.Entry) bool {
			return binary.BigEndian.Uint64(entry.Key()) > entries
		}), entries)
	})
	t.Run("FreeCache", func(t *testing.T) { testHitrate(t, newFreeCache(capacity), entries) })
	t.Run("FastCache", func(t *testing.T) { testHitrate(t, newFastCache(capacity), entries) })
}

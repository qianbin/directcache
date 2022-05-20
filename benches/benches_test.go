package benches

import (
	"encoding/binary"
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/qianbin/directcache"
)

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
		key := make([]byte, len(key))
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
	const (
		keyLen = 8
		valLen = 16
	)

	var counter uint64

	b.SetBytes(int64(keyLen + valLen))
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		key := make([]byte, keyLen)
		val := make([]byte, valLen)
		for p.Next() {
			i := atomic.AddUint64(&counter, 1) - 1
			binary.BigEndian.PutUint64(key, uint64(i))
			c.set(key, val)
		}
	})
}

func benchmarkParallelSetGet(b *testing.B, c cache) {
	const (
		keyLen = 8
		valLen = 16
	)

	var counter uint64

	const frac = 8

	b.SetBytes(keyLen + valLen)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		key := make([]byte, keyLen)
		val := make([]byte, valLen)
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
	b.Run("DirectCache", func(b *testing.B) { benchmarkGet(b, newDirectCache()) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkGet(b, newFreeCache()) })
	b.Run("FastCache", func(b *testing.B) { benchmarkGet(b, newFastCache()) })
}

func BenchmarkParallelGet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkParallelGet(b, newDirectCache()) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkParallelGet(b, newFreeCache()) })
	b.Run("FastCache", func(b *testing.B) { benchmarkParallelGet(b, newFastCache()) })
}

func BenchmarkSet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkSet(b, newDirectCache()) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkSet(b, newFreeCache()) })
	b.Run("FastCache", func(b *testing.B) { benchmarkSet(b, newFastCache()) })
}

func BenchmarkParallelSet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkParallelSet(b, newDirectCache()) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkParallelSet(b, newFreeCache()) })
	b.Run("FastCache", func(b *testing.B) { benchmarkParallelSet(b, newFastCache()) })
}

func BenchmarkParallelSetGet(b *testing.B) {
	b.Run("DirectCache", func(b *testing.B) { benchmarkParallelSetGet(b, newDirectCache()) })
	b.Run("FreeCache", func(b *testing.B) { benchmarkParallelSetGet(b, newFreeCache()) })
	b.Run("FastCache", func(b *testing.B) { benchmarkParallelSetGet(b, newFastCache()) })
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

	t.Run("DirectCache", func(t *testing.T) { testHitrate(t, newDirectCache(), entries) })
	t.Run("DirectCache(custom policy)", func(t *testing.T) {
		testHitrate(t, newDirectCacheWithPolicy(func(entry directcache.Entry) bool {
			return binary.BigEndian.Uint64(entry.Key()) > entries
		}), entries)
	})
	t.Run("FreeCache", func(t *testing.T) { testHitrate(t, newFreeCache(), entries) })
	t.Run("FastCache", func(t *testing.T) { testHitrate(t, newFastCache(), entries) })
}

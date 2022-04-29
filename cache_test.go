package directcache_test

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	"github.com/cespare/xxhash/v2"
	"github.com/qianbin/directcache"
	"github.com/stretchr/testify/assert"
)

func ExampleCache() {
	// capacity will be set to directcache.MinCapacity
	c := directcache.New(0)

	key := []byte("DirectCache")
	val := []byte("is awesome,")

	c.Set(key, val)
	got, ok := c.Get(key)

	fmt.Println(string(key), string(got), ok)
	// Output: DirectCache is awesome, true
}

func TestCache(t *testing.T) {
	c := directcache.New(0)
	assert.Equal(t, directcache.MinCapacity, c.Capacity(), "cap should be at least MinCapacity")

	c.Reset(directcache.MinCapacity * 2)
	assert.Equal(t, directcache.MinCapacity*2, c.Capacity())

	k := "key"
	v := "val"
	// set
	assert.True(t, c.Set([]byte(k), []byte(v)))
	// has get
	assert.True(t, c.Has([]byte(k)))
	got, ok := c.Get([]byte(k))
	assert.True(t, ok)
	assert.Equal(t, v, string(got))

	// del has
	assert.True(t, c.Del([]byte(k)))
	assert.False(t, c.Del([]byte(k)))
	assert.False(t, c.Has([]byte(k)))

	// advget
	c.Set([]byte(k), []byte(v))
	got = got[:0]
	assert.True(t, c.AdvGet([]byte(k), func(val []byte) {
		got = append(got, val...)
	}, false))
	assert.Equal(t, v, string(got))
}

func TestCacheHitrate(t *testing.T) {
	k := make([]byte, 8)
	v := make([]byte, 16)

	n := 1000000

	totalEntrySize := (len(k) + len(v) + 4) * n
	c := directcache.New(totalEntrySize / 10)

	hit, miss := 0, 0
	for i := 0; i < n; i++ {
		binary.BigEndian.PutUint64(k, uint64(i))
		binary.BigEndian.PutUint64(v, uint64(i))
		c.Set(k, v)

		if i < n/100 {
			continue
		}
		// access 1% previously inserted entries
		p := rand.Intn(n/100 + 1)
		binary.BigEndian.PutUint64(k, uint64(p))
		if val, ok := c.Get(k); ok {
			binary.BigEndian.PutUint64(v, uint64(p))
			assert.Equal(t, v, val)
			hit++
		} else {
			miss++
		}
	}
	hitrate := float64(hit) / float64(hit+miss)
	t.Logf("hits: %d misses: %d hitrate: %.2f%%", hit, miss, hitrate*100)
}

func BenchmarkCacheGetSet(b *testing.B) {
	c := directcache.New(1024 * 1024 * 100)

	k := make([]byte, 32)
	v := make([]byte, 64)

	b.Run("set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			binary.BigEndian.PutUint64(v, uint64(i))
			binary.BigEndian.PutUint64(k, uint64(i))
			c.Set(k, v)
		}
	})
	b.Run("get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			binary.BigEndian.PutUint64(v, uint64(i))
			binary.BigEndian.PutUint64(k, uint64(i))
			c.Get(k)
		}
	})
}

func BenchmarkHash(b *testing.B) {
	b.Run("xxhash.Sum64", func(b *testing.B) {
		data := make([]byte, 32)
		for i := 0; i < b.N; i++ {
			xxhash.Sum64(data)
		}
	})
}

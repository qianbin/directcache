package directcache_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/cespare/xxhash/v2"
	"github.com/qianbin/directcache"
	"github.com/stretchr/testify/require"
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
	require.Equal(t, directcache.MinCapacity, c.Capacity(), "cap should be at least MinCapacity")

	c.Reset(directcache.MinCapacity * 2)
	require.Equal(t, directcache.MinCapacity*2, c.Capacity())

	k := "key"
	v := "val"
	// set
	require.True(t, c.Set([]byte(k), []byte(v)))
	// has get
	require.True(t, c.Has([]byte(k)))
	got, ok := c.Get([]byte(k))
	require.True(t, ok)
	require.Equal(t, v, string(got))

	// del has
	require.True(t, c.Del([]byte(k)))
	require.False(t, c.Del([]byte(k)))
	require.False(t, c.Has([]byte(k)))

	// advget
	c.Set([]byte(k), []byte(v))
	got = got[:0]
	require.True(t, c.AdvGet([]byte(k), func(val []byte) {
		got = append(got, val...)
	}, false))
	require.Equal(t, v, string(got))
}

func BenchmarkCacheSetGet(b *testing.B) {
	const nEntries = 1000000
	b.Run("directcache", func(b *testing.B) {
		k := make([]byte, 8)
		v := make([]byte, 8)

		entrySize := len(k) + len(v) + 4
		nBytes := entrySize * nEntries

		c := directcache.New(nBytes * 2)

		b.Run("set", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				binary.BigEndian.PutUint64(k, uint64(i%nEntries))
				c.Set(k, v)
			}
		})
		b.Run("get", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				binary.BigEndian.PutUint64(k, uint64(i%nEntries))
				c.AdvGet(k, func(val []byte) {}, false)
			}
		})
	})
	b.Run("map", func(b *testing.B) {
		m := map[uint64]int{}
		b.Run("map.set", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m[uint64(i%nEntries)] = i
			}
		})
		b.Run("map.get", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = m[uint64(i%nEntries)]
			}
		})
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

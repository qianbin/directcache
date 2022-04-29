package directcache

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/cespare/xxhash/v2"
)

func ExampleCache() {
	c := New(0)
	k := "foo"
	v := "bar"

	c.Set([]byte(k), []byte(v))

	val, ok := c.Get([]byte(k))
	fmt.Println(string(val), ok)
	// Output: bar true
}

func BenchmarkCacheGetSet(b *testing.B) {
	c := New(1024 * 1024 * 100)

	k := make([]byte, 32)
	v := make([]byte, 64)

	b.Run("set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			binary.PutVarint(v, int64(i))
			binary.PutVarint(k, int64(i))
			c.Set(k, v)
		}
	})
	b.Run("get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			binary.PutVarint(v, int64(i))
			binary.PutVarint(k, int64(i))
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

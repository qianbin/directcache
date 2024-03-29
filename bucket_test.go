package directcache

import (
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/cespare/xxhash/v2"
	"github.com/stretchr/testify/require"
)

func Test_bucket(t *testing.T) {
	var bkt bucket
	bkt.Reset(100)

	k := "key"
	v := "val"
	kh := xxhash.Sum64String(k)

	// set
	ok := bkt.Set([]byte(k), kh, len(v), func(val []byte) { copy(val, v) })
	require.True(t, ok)

	// get
	var gv []byte
	ok = bkt.Get([]byte(k), kh, func(val []byte) {
		gv = append(gv[:0], val...)
	}, false)
	require.True(t, ok)
	require.Equal(t, v, string(gv))

	// del
	ok = bkt.Del([]byte(k), kh)
	require.True(t, ok)
	ok = bkt.Get([]byte(k), kh, func(val []byte) {
		require.Fail(t, "deleted, should not callback")
	}, false)
	require.False(t, ok, "deleted, should get nothing")
	require.False(t, bkt.Del([]byte(k), kh), "deleted, re-delete should fail")

	// in-place overwrite
	bkt.Set([]byte(k), kh, len(v), func(val []byte) { copy(val, v) })
	require.True(t, bkt.Set([]byte(k), kh, len(v), func(val []byte) { copy(val, v) }))

	// non-in-place overwrite
	require.True(t, bkt.Set([]byte(k), kh, len(v)*2, func(val []byte) { copy(val, v+v) }))

	// entry too large
	require.False(t, bkt.Set([]byte(k), kh, bkt.q.Cap()+1, func(val []byte) {}), "entry too large, should fail")

	// buffer overflow
	bkt.Reset(100)
	for i := 0; i < 100; i++ {
		n := rand.Intn(bkt.q.Cap() / 2)
		k := make([]byte, n)
		rand.Read(k)
		require.True(t, bkt.Set(k, xxhash.Sum64(k), 0, func(val []byte) {}))
		if i%2 == 0 {
			bkt.Get(k, xxhash.Sum64(k), nil, false) // add recently-used flag
		} else {
			bkt.Del(k, xxhash.Sum64(k)) // add deleted flag
		}
	}
}

func Test_bucketDump(t *testing.T) {
	var bkt bucket
	bkt.Reset(40)
	var ser []byte
	// overfill, the first inserted kv should be evicted
	for i := 0; i < 6; i++ {
		k := []byte{'k', byte(i)}
		v := []byte{'v', byte(i)}
		bkt.Set(k, xxhash.Sum64(k), len(v), func(val []byte) { copy(val, v) })
		ser = append(ser, k...)
		ser = append(ser, v...)
	}

	var dumps []byte
	require.True(t, bkt.Dump(func(e Entry) bool {
		dumps = append(dumps, e.Key()...)
		dumps = append(dumps, e.Value()...)
		return true
	}))
	// skip the first inserted kv
	require.Equal(t, ser[4:], dumps)

	calls := 0
	require.False(t, bkt.Dump(func(e Entry) bool {
		calls++
		return false
	}))
	require.Equal(t, 1, calls)
}

func Test_bucketHitrate(t *testing.T) {
	const maxEntries = 100
	k := make([]byte, 8)

	var bkt bucket
	bkt.Reset(entrySize(len(k), 0, 0) * maxEntries)
	bkt.SetEvictionPolicy(func(entry Entry) bool {
		if entry.RecentlyUsed() {
			return false
		}
		if binary.BigEndian.Uint64(entry.Key()) <= maxEntries {
			return false
		}
		return true
	})

	zipf := rand.NewZipf(rand.New(rand.NewSource(1)), 1.0001, 1, uint64(maxEntries*10))

	hit, miss := 0, 0
	for i := 0; i < maxEntries*100; i++ {
		binary.BigEndian.PutUint64(k, zipf.Uint64())
		hash := xxhash.Sum64(k)
		if bkt.Get(k, hash, nil, false) {
			hit++
		} else {
			miss++
			bkt.Set(k, hash, 0, func(val []byte) {})
		}
	}
	hitrate := float64(hit) / float64(hit+miss)
	// v0.9.0  hits: 5920 misses: 4080 hitrate: 59.20%
	// v0.9.1  hits: 6662 misses: 3338 hitrate: 66.62% (custom eviction policy)
	t.Logf("hits: %d misses: %d hitrate: %.2f%%", hit, miss, hitrate*100)
}

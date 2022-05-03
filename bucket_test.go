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
	ok := bkt.Set([]byte(k), kh, []byte(v))
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
	bkt.Set([]byte(k), kh, []byte(v))
	require.True(t, bkt.Set([]byte(k), kh, []byte(v)))

	// non-in-place overwrite
	require.True(t, bkt.Set([]byte(k), kh, []byte(v+v)))

	// entry too large
	require.False(t, bkt.Set([]byte(k), kh, make([]byte, bkt.q.Cap()+1)), "entry too large, should fail")

	// buffer overflow
	bkt.Reset(100)
	for i := 0; i < 100; i++ {
		n := rand.Intn(bkt.q.Cap() / 2)
		k := make([]byte, n)
		rand.Read(k)
		require.True(t, bkt.Set(k, xxhash.Sum64(k), nil))
		if i%2 == 0 {
			bkt.Get(k, xxhash.Sum64(k), nil, false) // add recently-used flag
		} else {
			bkt.Del(k, xxhash.Sum64(k)) // add deleted flag
		}
	}
}

func Test_bucketHitrate(t *testing.T) {
	maxEntries := 100
	k := make([]byte, 8)

	var bkt bucket
	bkt.Reset(entrySize(len(k), 0, 0) * maxEntries)

	zipf := rand.NewZipf(rand.New(rand.NewSource(1)), 1.0001, 1, uint64(maxEntries*10))

	hit, miss := 0, 0
	for i := 0; i < maxEntries*100; i++ {
		binary.BigEndian.PutUint64(k, zipf.Uint64())
		hash := xxhash.Sum64(k)
		if bkt.Get(k, hash, nil, false) {
			hit++
		} else {
			miss++
			bkt.Set(k, hash, nil)
		}
	}
	hitrate := float64(hit) / float64(hit+miss)
	// v0.9.0: 59.20%
	t.Logf("hits: %d misses: %d hitrate: %.2f%%", hit, miss, hitrate*100)
}

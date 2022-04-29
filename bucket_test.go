package directcache

import (
	"strconv"
	"testing"

	"github.com/cespare/xxhash/v2"
	"github.com/stretchr/testify/assert"
)

func TestBucket(t *testing.T) {
	var bkt bucket
	bkt.Reset(100)

	k := "key"
	v := "val"
	kh := xxhash.Sum64String(k)

	// set
	ok := bkt.Set([]byte(k), kh, []byte(v))
	assert.True(t, ok)

	// get
	var gv []byte
	ok = bkt.Get([]byte(k), kh, func(val []byte) {
		gv = append(gv[:0], val...)
	}, false)
	assert.True(t, ok)
	assert.Equal(t, v, string(gv))

	// del
	ok = bkt.Del([]byte(k), kh)
	assert.True(t, ok)
	ok = bkt.Get([]byte(k), kh, func(val []byte) {
		assert.Fail(t, "deleted, should not callback")
	}, false)
	assert.False(t, ok, "deleted, should get nothing")
	assert.False(t, bkt.Del([]byte(k), kh), "deleted, re-delete should fail")

	// in-place overwrite
	bkt.Set([]byte(k), kh, []byte(v))
	assert.True(t, bkt.Set([]byte(k), kh, []byte(v)))

	// non-in-place overwrite
	assert.True(t, bkt.Set([]byte(k), kh, []byte(v+v)))

	// entry too large
	assert.False(t, bkt.Set([]byte(k), kh, make([]byte, bkt.q.Cap()+1)), "entry too large, should fail")

	// buffer overflow
	bkt.Reset(100)
	for i := 0; i < 100; i++ {
		k := k + strconv.Itoa(i)
		v := v + strconv.Itoa(i)
		assert.True(t, bkt.Set([]byte(k), xxhash.Sum64String(k), []byte(v)))
		bkt.Get([]byte(k), xxhash.Sum64String(k), nil, false)
	}
}

package directcache

import "github.com/cespare/xxhash/v2"

const (
	bucketCount  = 256
	minBucketCap = 64 * 1024

	MinCapacity = bucketCount * minBucketCap
)

type Cache struct {
	cap     int
	buckets [bucketCount]bucket
}

func New(capacity int) *Cache {
	c := &Cache{}
	c.Reset(capacity)
	return c
}

func (c *Cache) Capacity() int { return c.cap }

func (c *Cache) Reset(capacity int) {
	bktCap := capacity / bucketCount
	if bktCap < minBucketCap {
		bktCap = minBucketCap
	}

	c.cap = bktCap * bucketCount
	for i := 0; i < bucketCount; i++ {
		c.buckets[i].Reset(bktCap)
	}
}

func (c *Cache) Set(key, val []byte) bool {
	keyHash := xxhash.Sum64(key)
	index := keyHash % bucketCount
	return c.buckets[index].Set(key, keyHash, val)
}

func (c *Cache) Del(key []byte) bool {
	keyHash := xxhash.Sum64(key)
	index := keyHash % bucketCount
	return c.buckets[index].Del(key, keyHash)
}

func (c *Cache) Get(key []byte) (val []byte, ok bool) {
	ok = c.GetEx(key, func(_val []byte) {
		val = append(val, _val...)
	}, false)
	return
}
func (c *Cache) Has(key []byte) bool {
	return c.GetEx(key, nil, false)
}

func (c *Cache) GetEx(key []byte, fn func(val []byte), peek bool) bool {
	keyHash := xxhash.Sum64(key)
	index := keyHash % bucketCount
	return c.buckets[index].Get(key, keyHash, fn, peek)
}

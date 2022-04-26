package directcache

import "github.com/cespare/xxhash/v2"

const bucketCount = 256
const minBucketCap = 64 * 1024

type Cache struct {
	cap     int
	buckets [bucketCount]bucket
}

func New(capacity int) *Cache {
	minCap := minBucketCap * bucketCount

	if capacity < minCap {
		capacity = minCap
	}

	var c Cache

	bktCap := capacity / bucketCount
	c.cap = bktCap * bucketCount
	for i := 0; i < bucketCount; i++ {
		c.buckets[i].Reset(bktCap)
	}
	return &c
}

func (c *Cache) Set(key, val []byte) error {
	keyHash := xxhash.Sum64(key)
	index := keyHash % bucketCount
	return c.buckets[index].Set(key, keyHash, val)
}

func (c *Cache) Get(key []byte) ([]byte, bool) {
	keyHash := xxhash.Sum64(key)
	index := keyHash % bucketCount
	return c.buckets[index].Get(key, keyHash)
}

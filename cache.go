// Package directcache is a high performance GC-free cache library.
package directcache

import "github.com/cespare/xxhash/v2"

const (
	// BucketCount is the count of buckets in a Cache instance.
	BucketCount = 256
	// MinCapacity is the minimum capacity in bytes of the Cache.
	MinCapacity = BucketCount * 256
)

// Cache caches key-value entries of type []byte.
type Cache struct {
	buckets [BucketCount]bucket
	cap     int
}

// New creates a new Cache instance with the given capacity in bytes.
// The instance capacity will be set to MinCapacity at minimum.
func New(capacity int) *Cache {
	c := &Cache{}
	c.Reset(capacity)
	return c
}

// Capacity returns the cache capacity.
func (c *Cache) Capacity() int { return c.cap }

// Reset resets the cache with new capacity and drops all cached entries.
func (c *Cache) Reset(capacity int) {
	if capacity < MinCapacity {
		capacity = MinCapacity
	}
	bktCap := capacity / BucketCount
	for i := 0; i < BucketCount; i++ {
		c.buckets[i].Reset(bktCap)
	}
	c.cap = capacity
}

// SetEvictionPolicy customizes the cache eviction policy.
// shouldEvict is called when no space to insert the new entry and have to evict an old entry.
// If shouldEvict returns true the old entry will evict immediately, and if false the old entry
// will likely be kept. The provided entry is read-only and never modify its key or value.
func (c *Cache) SetEvictionPolicy(shouldEvict func(entry Entry) bool) {
	for i := 0; i < BucketCount; i++ {
		c.buckets[i].SetEvictionPolicy(shouldEvict)
	}
}

// Set stores the (key, val) entry in the cache, and returns false on failure.
// It always succeeds unless the size of the entry exceeds 1/BucketCount of the cache capacity.
//
// It's safe to modify contents of key and val after Set returns.
func (c *Cache) Set(key, val []byte) bool {
	keyHash := xxhash.Sum64(key)
	return c.buckets[keyHash%BucketCount].Set(key, keyHash, val)
}

// Del deletes the entry matching the given key from the cache.
// false is returned if no entry matched.
//
// It's safe to modify contents of key after Del returns.
func (c *Cache) Del(key []byte) bool {
	keyHash := xxhash.Sum64(key)
	return c.buckets[keyHash%BucketCount].Del(key, keyHash)
}

// Get returns the value of the entry matching the given key.
// It returns false if no matched entry.
//
// It's safe to modify contents of key after Get returns.
func (c *Cache) Get(key []byte) (val []byte, ok bool) {
	keyHash := xxhash.Sum64(key)
	ok = c.buckets[keyHash%BucketCount].Get(key, keyHash, func(_val []byte) {
		val = append(val, _val...)
	}, false)
	return
}

// Has returns false if no entry matching the given key.
//
// It's safe to modify contents of key after Has returns.
func (c *Cache) Has(key []byte) bool {
	keyHash := xxhash.Sum64(key)
	return c.buckets[keyHash%BucketCount].Get(key, keyHash, nil, false)
}

// AdvGet is the advanced version of Get. val is (zero-copy) accessed via fn callback.
// It returns false if no entry matching the given key, and fn will not be called then.
// If peek is true, the entry's recently-used flag is not updated.
//
// val is only valid inside fn and should never be modified.
// It's safe to modify contents of key after AdvGet returns.
func (c *Cache) AdvGet(key []byte, fn func(val []byte), peek bool) bool {
	keyHash := xxhash.Sum64(key)
	return c.buckets[keyHash%BucketCount].Get(key, keyHash, fn, peek)
}

// Dump dumps all saved entires bucket by bucket in the order of insertion.
// It's interrupted if f returns false。
// The provided entry is read-only and never modify its key or value.
func (c *Cache) Dump(f func(Entry) bool) {
	for i := 0; i < BucketCount; i++ {
		if !c.buckets[i].Dump(f) {
			break
		}
	}
}

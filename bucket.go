package directcache

import (
	"bytes"
	"errors"
	"sync"

	"github.com/cespare/xxhash/v2"
)

// bucket indexes and holds entries.
type bucket struct {
	m           map[uint64]int         // maps key hash to offset
	q           fifo                   // the queue buffer stores entries
	shouldEvict func(entry Entry) bool // the custom evention policy
	lock        sync.RWMutex
}

// Reset resets the bucket with new capacity and new eviction method.
// If shouldEvict is nil, the default LRU policy is used.
// It drops all entries.
func (b *bucket) Reset(capacity int) {
	b.lock.Lock()
	b.m = make(map[uint64]int)
	b.q.Reset(capacity)
	b.lock.Unlock()
}

// SetEvictionPolicy customizes the cache eviction policy.
func (b *bucket) SetEvictionPolicy(shouldEvict func(entry Entry) bool) {
	b.lock.Lock()
	b.shouldEvict = shouldEvict
	b.lock.Unlock()
}

// Set set val for key.
// false returned and nonting changed if the new entry size exceeds the capacity of this bucket.
func (b *bucket) Set(key []byte, keyHash uint64, val []byte) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	if offset, found := b.m[keyHash]; found {
		ent := b.entryAt(offset)
		if spare := ent.BodySize() - len(key) - len(val); spare >= 0 { // in-place update
			ent.Init(key, val, spare)
			ent.AddFlag(recentlyUsedFlag) // avoid evicted too early
			return true
		}
		// key not matched or in-place update failed
		ent.AddFlag(deletedFlag)
	}
	// insert new entry
	if offset, ok := b.insertEntry(key, val, 0); ok {
		b.m[keyHash] = offset
		return true
	}
	return false
}

// Del deletes the key.
// false is returned if key does not exist.
func (b *bucket) Del(key []byte, keyHash uint64) bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	if offset, found := b.m[keyHash]; found {
		if ent := b.entryAt(offset); bytes.Equal(ent.Key(), key) {
			delete(b.m, keyHash)
			ent.AddFlag(deletedFlag)
			return true
		}
	}
	return false
}

// Get get the value for key.
// false is returned if the key not found.
// If peek is true, the entry will not be marked as recently-used.
func (b *bucket) Get(key []byte, keyHash uint64, fn func(val []byte), peek bool) bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if offset, found := b.m[keyHash]; found {
		if ent := b.entryAt(offset); bytes.Equal(ent.Key(), key) {
			if !peek {
				ent.AddFlag(recentlyUsedFlag)
			}
			if fn != nil {
				fn(ent.Value())
			}
			return true
		}
	}
	return false
}

// entryAt creates an entry object at the offset of the queue buffer.
func (b *bucket) entryAt(offset int) entry {
	return b.q.Slice(offset)
}

// insertEntry insert a new entry and returns its offset.
// Old entries are evicted like LRU strategy if no enough space.
func (b *bucket) insertEntry(key, val []byte, spare int) (int, bool) {
	entrySize := entrySize(len(key), len(val), spare)
	if entrySize > b.q.Cap() {
		return 0, false
	}

	pushLimit := 8
	for {
		// have a try
		if offset, ok := b.q.Push(nil, entrySize); ok {
			entry(b.q.Slice(offset)).Init(key, val, spare)
			return offset, true
		}

		// no space
		// pop an entry at the front of the queue buffer
		ent := b.entryAt(b.q.Front())
		ent = ent[:ent.Size()]
		if _, ok := b.q.Pop(len(ent)); !ok {
			// will never go here if entry is correctly implemented
			panic(errors.New("bucket.allocEntry: pop entry failed"))
		}

		// good, deleted entry
		if ent.HasFlag(deletedFlag) {
			continue
		}

		keyHash := xxhash.Sum64(ent.Key())
		// pushLimit exceeded
		if pushLimit < 1 {
			delete(b.m, keyHash)
			continue
		}

		if b.shouldEvict == nil {
			// the default LRU policy
			if !ent.HasFlag(recentlyUsedFlag) {
				delete(b.m, keyHash)
				continue
			}
		} else {
			// the custom eviction policy
			if b.shouldEvict(ent) {
				delete(b.m, keyHash)
				continue
			}
		}

		pushLimit--
		ent.RemoveFlag(recentlyUsedFlag)
		//  and push back to the queue
		if offset, ok := b.q.Push(ent, 0); ok {
			// update the offset
			b.m[keyHash] = offset
		} else {
			panic("bucket.allocEntry: push entry failed")
		}
	}
}

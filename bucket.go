package directcache

import (
	"bytes"
	"errors"
	"sync"

	"github.com/cespare/xxhash/v2"
)

// bucket indexes and holds entries.
type bucket struct {
	m    map[uint64]int // maps key hash to offset
	q    fifo           // the queue buffer stores entries
	lock sync.RWMutex
}

// Reset resets the bucket with new capacity.
// It drops all entries.
func (b *bucket) Reset(capacity int) {
	b.lock.Lock()
	b.m = make(map[uint64]int)
	b.q.Reset(capacity)
	b.lock.Unlock()
}

// Set set val for key.
// false returned and nonting changed if the new entry size exceeds the capacity of this bucket.
func (b *bucket) Set(key []byte, keyHash uint64, val []byte) (ok bool) {
	entrySize := entrySize(len(key), len(val), 0)

	b.lock.Lock()
	if entrySize <= b.q.Cap() {
		if offset, found := b.m[keyHash]; found {
			ent := b.entryAt(offset)
			if bytes.Equal(key, ent.Key()) && ent.UpdateValue(val) {
				ent.AddFlag(recentlyUsedFlag)
				ok = true
			} else {
				ent.AddFlag(deletedFlag)
			}
		}
		if !ok {
			newEnt, offset := b.allocEntry(entrySize)
			newEnt.Init(key, val, 0)
			b.m[keyHash] = offset
			ok = true
		}
	}
	b.lock.Unlock()
	return
}

// Del deletes the key.
// false is returned if key does not exist.
func (b *bucket) Del(key []byte, keyHash uint64) (ok bool) {
	b.lock.Lock()
	if offset, found := b.m[keyHash]; found {
		if ent := b.entryAt(offset); bytes.Equal(ent.Key(), key) {
			delete(b.m, keyHash)
			ent.AddFlag(deletedFlag)
			ok = true
		}
	}
	b.lock.Unlock()
	return
}

// Get get the value for key.
// false is returned if the key not found.
// If peek is true, the entry will not be marked as recently-used.
func (b *bucket) Get(key []byte, keyHash uint64, fn func(val []byte), peek bool) (ok bool) {
	b.lock.RLock()
	if offset, found := b.m[keyHash]; found {
		if ent := b.entryAt(offset); bytes.Equal(ent.Key(), key) {
			if !peek {
				ent.AddFlag(recentlyUsedFlag)
			}
			if fn != nil {
				fn(ent.Value())
			}
			ok = true
		}
	}
	b.lock.RUnlock()
	return
}

// entryAt creates an entry object at the offset of the queue buffer.
func (b *bucket) entryAt(offset int) entry {
	return b.q.Slice(offset)
}

// allocEntry allocs space on the queue buffer for the new entry.
// Entries are evicted like LRU strategy if no enough space.
func (b *bucket) allocEntry(size int) (entry, int) {
	pushCount := 0
	for {
		// have a try
		if offset, ok := b.q.Push(nil, size); ok {
			return b.q.Slice(offset)[:size], offset
		}

		// no space
		// pop an entry at the front of the queue buffer
		ent := b.entryAt(b.q.Front())
		popped, ok := b.q.Pop(ent.Size())
		if !ok {
			// will never go here if entry is correctly implemented
			panic(errors.New("bucket.allocEntry: pop entry failed"))
		}

		// good, deleted entry
		if ent.HasFlag(deletedFlag) {
			continue
		}

		keyHash := xxhash.Sum64(ent.Key())
		// pushed too many times or entry not recently used, delete it
		if pushCount > 4 || !ent.HasFlag(recentlyUsedFlag) {
			delete(b.m, keyHash)
			continue
		}

		ent.RemoveFlag(recentlyUsedFlag)
		//  and push back to the queue
		if offset, ok := b.q.Push(popped, 0); ok {
			pushCount++
			// update the offset
			b.m[keyHash] = offset
		} else {
			panic("bucket.allocEntry: push entry failed")
		}
	}
}

package directcache

import (
	"errors"
	"sync"
)

var (
	ErrEntryTooLarge = errors.New("entry too large")
)

type bucket struct {
	m    map[uint64]int
	fifo fifo
	sync.Mutex
}

func (b *bucket) Reset(capacity int) {
	b.Lock()
	defer b.Unlock()

	b.m = make(map[uint64]int)
	b.fifo.Reset(capacity)
}

func (b *bucket) Set(key []byte, keyHash uint64, val []byte) error {
	entrySize := calcEntrySize(len(key), len(val), 0)
	if entrySize > b.fifo.Cap() {
		return ErrEntryTooLarge
	}

	b.Lock()
	defer b.Unlock()

	if offset, found := b.m[keyHash]; found {
		ent := b.entryAt(offset)
		if ent.Match(key) && ent.UpdateValue(val) {
			return nil
		}
		delete(b.m, keyHash)
		ent.AddFlag(deletedFlag)
	}

	newEnt, offset := b.allocEntry(entrySize)
	newEnt.Init(key, keyHash, val, 0)
	b.m[keyHash] = offset
	return nil
}

func (b *bucket) Del(key []byte, keyHash uint64) bool {
	b.Lock()
	defer b.Unlock()

	if offset, found := b.m[keyHash]; found {
		if ent := b.entryAt(offset); ent.Match(key) {
			delete(b.m, keyHash)
			ent.AddFlag(deletedFlag)
			return true
		}
	}
	return false
}

func (b *bucket) Get(key []byte, keyHash uint64, fn func(val []byte), peek bool) bool {
	b.Lock()
	defer b.Unlock()

	if offset, found := b.m[keyHash]; found {
		if ent := b.entryAt(offset); ent.Match(key) {
			if !peek {
				ent.AddFlag(activeFlag)
			}
			if fn != nil {
				fn(ent.Value())
			}
			return true
		}
	}
	return false
}

func (b *bucket) entryAt(offset int) entry {
	ent := entry(b.fifo.Slice(offset))
	return ent[:ent.Size()]
}

func (b *bucket) allocEntry(size int) (entry, int) {
	windCount := 0
	for {
		if offset, ok := b.fifo.Push(nil, size); ok {
			return b.fifo.Slice(offset)[:size], offset
		}

		ent := b.entryAt(b.fifo.Front())
		popped, ok := b.fifo.Pop(len(ent))
		if !ok {
			panic(errors.New("bucket.allocEntry: pop entry failed"))
		}

		if ent.HasFlag(deletedFlag) {
			continue
		}

		keyHash := ent.KeyHash()
		if windCount > 4 || !ent.HasFlag(activeFlag) {
			delete(b.m, keyHash)
			continue
		}

		windCount++
		ent.RemoveFlag(activeFlag)
		if offset, ok := b.fifo.Push(popped, 0); ok {
			b.m[keyHash] = offset
		} else {
			delete(b.m, keyHash)
		}
	}
}

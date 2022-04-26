package directcache

import (
	"sync"
	"unsafe"
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
	b.Lock()
	defer b.Unlock()

	if ent, found := b.lookup(keyHash); found {
		if ent.Match(key) && ent.UpdateValue(val) {
			return nil
		}
		ent.deleted = true
	}

	size := headerSize + len(key) + len(val)
	alignedSize := alignEntrySize(size)

	newEnt, offset := b.allocEntry(alignedSize)
	newEnt.keyHash = keyHash
	newEnt.Assign(key, val, uint32(alignedSize-size))
	b.m[keyHash] = offset
	return nil
}

func (b *bucket) Get(key []byte, keyHash uint64) ([]byte, bool) {
	b.Lock()
	defer b.Unlock()

	ent, found := b.lookup(keyHash)
	if !found {
		return nil, false
	}

	if !ent.Match(key) {
		return nil, false
	}

	return append([]byte(nil), ent.Value()...), true
}

func (b *bucket) lookup(keyHash uint64) (ent entry, found bool) {
	offset, found := b.m[keyHash]
	if found {
		ent = b.entryAt(offset)
	}
	return
}

func (b *bucket) entryAt(offset int) entry {
	v := b.fifo.View(offset)
	return entry{
		(*header)(unsafe.Pointer(&v[0])),
		v[headerSize:],
	}
}

func (b *bucket) allocEntry(size int) (entry, int) {
	windCount := 0
	for {
		if offset, ok := b.fifo.Push(nil, size); ok {
			return b.entryAt(offset), offset
		}

		ent := b.entryAt(b.fifo.Front())

		popped, ok := b.fifo.Pop(ent.Size())
		if !ok {
			panic("")
		}

		if !ent.deleted {
			keyHash := ent.keyHash
			if windCount < 5 {
				windCount++
				if offset, ok := b.fifo.Push(popped, 0); ok {
					b.m[keyHash] = offset
				} else {
					delete(b.m, keyHash)
				}
			} else {
				delete(b.m, keyHash)
			}
		}
	}
}

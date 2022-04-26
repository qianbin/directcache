package directcache

import (
	"bytes"
	"unsafe"
)

const headerSize = int(unsafe.Sizeof(header{}))

func alignEntrySize(size int) int {
	const align = int(unsafe.Alignof(header{}))
	return (size + align - 1) / align * align
}

type header struct {
	keyHash uint64
	valLen  uint32
	spare   uint32
	access  uint32
	keyLen  uint16
	deleted bool
}

type entry struct {
	*header
	body []byte
}

func (e *entry) Size() int {
	return headerSize + int(e.keyLen) + int(e.valLen) + int(e.spare)
}

func (e *entry) Match(key []byte) bool {
	return len(key) == int(e.keyLen) && bytes.Equal(key, e.body[:e.keyLen])
}

func (e *entry) Value() []byte {
	return e.body[e.keyLen : uint32(e.keyLen)+e.valLen]
}

func (e *entry) UpdateValue(val []byte) bool {
	n := uint32(len(val))
	if cap := e.valLen + e.spare; cap >= n {
		copy(e.body[e.keyLen:], val)
		e.valLen = n
		e.spare = cap - n
		return true
	}
	return false
}

func (e *entry) Assign(key, val []byte, spare uint32) {
	copy(e.body, key)
	copy(e.body[len(key):], val)
	e.keyLen = uint16(len(key))
	e.valLen = uint32(len(val))
	e.spare = spare
}

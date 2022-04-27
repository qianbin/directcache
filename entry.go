package directcache

import (
	"bytes"
	"unsafe"
)

const (
	headerSize = int(unsafe.Sizeof(header{}))

	deletedFlag = 1
	activeFlag  = 2

	keyHashMask = 0xff
)

type header struct {
	keyHash0 uint32
	keyHash1 uint16
	keyHash2 uint8
	flags    uint8
	keyLen   uint32
	valLen   uint32
	spare    uint32
}

func (h *header) HasFlag(flag uint8) bool { return (h.flags & flag) != 0 }
func (h *header) AddFlag(flag uint8)      { h.flags |= flag }
func (h *header) RemoveFlag(flag uint8)   { h.flags &^= flag }
func (h *header) EntrySize() int {
	return headerSize + int(h.keyLen) + int(h.valLen) + int(h.spare)
}
func (h *header) KeyHash() uint64 {
	return (uint64(h.keyHash0) << 32) | (uint64(h.keyHash1) << 16) | (uint64(h.keyHash2) << 8)
}

type entry []byte

func (e entry) Header() *header {
	return (*header)(unsafe.Pointer(&e[0]))
}

func (e entry) Init(key []byte, keyHash uint64, val []byte, spare uint32) {
	hdr := e.Header()
	hdr.keyHash0 = uint32(keyHash >> 32)
	hdr.keyHash1 = uint16(keyHash >> 16)
	hdr.keyHash2 = uint8(keyHash >> 8)
	hdr.flags = 0

	hdr.keyLen = uint32(len(key))
	hdr.valLen = uint32(len(val))
	hdr.spare = spare

	copy(e[headerSize:], key)
	copy(e[headerSize+len(key):], val)
}

func (e entry) Match(key []byte) bool {
	hdr := e.Header()
	return len(key) == int(hdr.keyLen) &&
		bytes.Equal(key, e[headerSize:][:hdr.keyLen])
}

func (e entry) Value() []byte {
	hdr := e.Header()
	return e[headerSize:][hdr.keyLen:][:hdr.valLen]
}

func (e entry) UpdateValue(val []byte) bool {
	hdr := e.Header()
	n := uint32(len(val))
	if cap := hdr.valLen + hdr.spare; cap >= n {
		copy(e[headerSize:][hdr.keyLen:], val)
		hdr.valLen = n
		hdr.spare = cap - n
		return true
	}
	return false
}

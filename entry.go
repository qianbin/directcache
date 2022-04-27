package directcache

import (
	"bytes"
	"math"
	"unsafe"
)

const (
	headerSize      = int(unsafe.Sizeof(header{}))
	subHeader8Size  = int(unsafe.Sizeof(subHeader8{}))
	subHeader32Size = int(unsafe.Sizeof(subHeader32{}))

	deletedFlag = 1
	activeFlag  = 2
	compactFlag = 4

	keyHashMask = 0xff
)

type header struct {
	keyHash0 uint32
	keyHash1 uint16
	keyHash2 uint8
	flags    uint8
}

func (h *header) HasFlag(flag uint8) bool { return (h.flags & flag) != 0 }
func (h *header) AddFlag(flag uint8)      { h.flags |= flag }
func (h *header) RemoveFlag(flag uint8)   { h.flags &^= flag }
func (h *header) KeyHash() uint64 {
	return (uint64(h.keyHash0) << 32) | (uint64(h.keyHash1) << 16) | (uint64(h.keyHash2) << 8)
}

type subHeader8 struct {
	keyLen, valLen, spare uint8
}

type subHeader32 struct {
	keyLen, valLen, spare uint32
}

type entry []byte

func (e entry) Header() *header {
	return (*header)(unsafe.Pointer(&e[0]))
}

func (e entry) layout() (bodyOffset, keyLen, valLen, spare int) {
	hdr := e.Header()
	subHdrPtr := unsafe.Pointer(&e[headerSize])
	if hdr.HasFlag(compactFlag) {
		sub := (*subHeader8)(subHdrPtr)
		return headerSize + subHeader8Size, int(sub.keyLen), int(sub.valLen), int(sub.spare)
	}
	sub := (*subHeader32)(subHdrPtr)
	return headerSize + subHeader32Size, int(sub.keyLen), int(sub.valLen), int(sub.spare)
}

func (e entry) setLayout(keyLen, valLen, spare int) (bodyOffset int) {
	subHdrPtr := unsafe.Pointer(&e[headerSize])
	if keyLen <= math.MaxUint8 && (valLen+spare) <= math.MaxUint8 {
		e.Header().AddFlag(compactFlag)
		sub := (*subHeader8)(subHdrPtr)
		sub.keyLen, sub.valLen, sub.spare = uint8(keyLen), uint8(valLen), uint8(spare)
		return headerSize + subHeader8Size
	}
	sub := (*subHeader32)(subHdrPtr)
	sub.keyLen, sub.valLen, sub.spare = uint32(keyLen), uint32(valLen), uint32(spare)
	return headerSize + subHeader32Size
}

func (e entry) Size() int {
	bodyOffset, keyLen, valLen, spare := e.layout()
	return bodyOffset + keyLen + valLen + spare
}

func (e entry) Init(key []byte, keyHash uint64, val []byte, spare int) {
	hdr := e.Header()
	hdr.keyHash0 = uint32(keyHash >> 32)
	hdr.keyHash1 = uint16(keyHash >> 16)
	hdr.keyHash2 = uint8(keyHash >> 8)
	hdr.flags = 0

	bodyOffset := e.setLayout(len(key), len(val), spare)
	copy(e[bodyOffset:], key)
	copy(e[bodyOffset+len(key):], val)
}

func (e entry) Match(key []byte) bool {
	bodyOffset, keyLen, _, _ := e.layout()
	return len(key) == keyLen &&
		bytes.Equal(key, e[bodyOffset:][:keyLen])
}

func (e entry) Value() []byte {
	bodyOffset, keyLen, valLen, _ := e.layout()
	return e[bodyOffset:][keyLen:][:valLen]
}

func (e entry) UpdateValue(val []byte) bool {
	bodyOffset, keyLen, valLen, spare := e.layout()
	newValLen := len(val)
	if cap := valLen + spare; cap >= newValLen {
		copy(e[bodyOffset:][keyLen:], val)
		e.setLayout(keyLen, newValLen, cap-newValLen)
		return true
	}
	return false
}

func calcEntrySize(keyLen, valLen, spare int) int {
	subHdrSize := subHeader32Size
	if keyLen <= math.MaxUint8 && (valLen+spare) <= math.MaxUint8 {
		subHdrSize = subHeader8Size
	}
	return headerSize + subHdrSize + keyLen + valLen + spare
}

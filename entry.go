package directcache

import (
	"bytes"
	"math"
	"unsafe"

	"github.com/cespare/xxhash/v2"
)

const (
	deletedFlag      = 1 // the entry was deleted
	recentlyUsedFlag = 2 // the entry is recently accessed
)

// entry consists of header and body.
type entry []byte

// flag ops. flags are stored in the first 4-bit of the entry.
func (e entry) HasFlag(flag uint8) bool { return e[0]&(flag<<4) != 0 }
func (e entry) AddFlag(flag uint8)      { e[0] |= (flag << 4) }
func (e entry) RemoveFlag(flag uint8)   { e[0] &^= (flag << 4) }

// kvw extracts the width of key/val length.
// kw and vw are stored in 5-8th bits of the entry.
func (e entry) kvw() (kw int, vw int) {
	b := e[0]
	return 1 << ((b >> 2) & byte(3)), 1 << (b & byte(3))
}

func (e entry) hdrSize() int {
	kw, vw := e.kvw()
	return 1 + kw + vw*2
}

func (e entry) keyLen() int {
	kw, _ := e.kvw()
	return e.intAt(1, kw)
}

func (e entry) valLen() (int, int) {
	kw, vw := e.kvw()
	return e.intAt(1+kw, vw), e.intAt(1+kw+vw, vw)
}

// Size returns the entry size.
func (e entry) Size() int {
	valLen, spare := e.valLen()
	return e.hdrSize() + e.keyLen() + valLen + spare
}

// Init initializes the entry with key, val and spare.
//
// The entry must be pre-alloced.
func (e entry) Init(key []byte, val []byte, spare int) {
	e[0] = 0 // reset flags

	keyLen, valLen := len(key), len(val)

	kw, km := width(keyLen)
	vw, vm := width(valLen + spare)

	// pack the width of key length
	e[0] &^= (uint8(3) << 2)
	e[0] |= (km << 2)

	// pack the width of value length
	e[0] &^= uint8(3)
	e[0] |= vm

	e.setIntAt(1, kw, keyLen)
	e.setIntAt(1+kw, vw, valLen)
	e.setIntAt(1+kw+vw, vw, spare)

	hdrSize := 1 + kw + vw*2

	copy(e[hdrSize:], key)
	copy(e[hdrSize:][keyLen:], val)
}

// Match tests if the key matched.
func (e entry) Match(key []byte) bool {
	if keyLen := len(key); keyLen == e.keyLen() {
		return bytes.Equal(key, e[e.hdrSize():][:keyLen])
	}
	return false
}

// KeyHash calculates the hash of the key.
func (e entry) KeyHash() uint64 {
	return xxhash.Sum64(e[e.hdrSize():][:e.keyLen()])
}

// Value returns the value stored in the entry.
func (e entry) Value() []byte {
	valLen, _ := e.valLen()
	return e[e.hdrSize():][e.keyLen():][:valLen]
}

// UpdateValue updates the value.
//
// It fails if no enough space.
func (e entry) UpdateValue(val []byte) bool {
	valLen, spare := e.valLen()
	newValLen := len(val)
	if cap := valLen + spare; cap >= newValLen {
		keyLen := e.keyLen()
		copy(e[e.hdrSize():][keyLen:], val)

		kw, vw := e.kvw()
		e.setIntAt(1+kw, vw, newValLen)
		e.setIntAt(1+kw+vw, vw, cap-newValLen)
		return true
	}
	return false
}

func (e entry) intAt(i int, width int) int {
	switch width {
	case 1:
		return int(e[i])
	case 2:
		return int(*(*uint16)(unsafe.Pointer(&e[i])))
	case 4:
		return int(*(*uint32)(unsafe.Pointer(&e[i])))
	default:
		return int(*(*uint64)(unsafe.Pointer(&e[i])))
	}
}

func (e entry) setIntAt(i int, width int, n int) {
	switch width {
	case 1:
		e[i] = byte(n)
	case 2:
		*(*uint16)(unsafe.Pointer(&e[i])) = uint16(n)
	case 4:
		*(*uint32)(unsafe.Pointer(&e[i])) = uint32(n)
	default:
		*(*uint64)(unsafe.Pointer(&e[i])) = uint64(n)
	}
}

// entrySize returns the size of an entry for given kv lengths.
func entrySize(keyLen, valLen, spare int) int {
	kw, _ := width(keyLen)
	vw, _ := width(valLen + spare)

	hdrSize := 1 + kw + vw*2
	return hdrSize + keyLen + valLen + spare
}

// width returns how many bytes needed to store len.
//
// mask is the bit mask for needed number of bytes.
func width(len int) (width int, mask byte) {
	switch {
	case len <= math.MaxUint8:
		return 1, 0
	case len <= math.MaxUint16:
		return 2, 1
	case int64(len) <= math.MaxUint32: // cast to int64 to avoid compile error on 32-bit env
		return 4, 2
	default:
		return 8, 3
	}
}

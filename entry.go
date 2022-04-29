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

// flag ops
func (e entry) HasFlag(flag uint8) bool { return e[0]&(flag<<4) != 0 }
func (e entry) AddFlag(flag uint8)      { e[0] |= (flag << 4) }
func (e entry) RemoveFlag(flag uint8)   { e[0] &^= (flag << 4) }

// sizes returns the size of each part.
func (e entry) sizes() (hdrSize, keyLen, valLen, spare int) {
	kw := 1 << ((e[0] >> 2) & byte(3)) // extract the width of key length
	vw := 1 << (e[0] & byte(3))        // extract the width of value length

	hdrSize = 1 + kw + vw*2

	keyLen = e.intAt(1, kw)
	valLen = e.intAt(1+kw, vw)
	spare = e.intAt(1+kw+vw, vw)
	return
}

// setSizes sets sizes of key, value and spare, returns the header size.
func (e entry) setSizes(keyLen, valLen, spare int) (hdrSize int) {
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
	return 1 + kw + vw*2
}

// Size returns the entry size.
func (e entry) Size() int {
	hdrSize, keyLen, valLen, spare := e.sizes()
	return hdrSize + keyLen + valLen + spare
}

// Init initializes the entry with key, val and spare.
//
// The entry must be pre-alloced.
func (e entry) Init(key []byte, val []byte, spare int) {
	e[0] = 0 // reset flags

	keyLen, valLen := len(key), len(val)

	hdrSize := e.setSizes(keyLen, valLen, spare)
	copy(e[hdrSize:], key)
	copy(e[hdrSize:][keyLen:], val)
}

// Match tests if the key matched.
func (e entry) Match(key []byte) bool {
	hdrSize, keyLen, _, _ := e.sizes()
	return len(key) == keyLen &&
		bytes.Equal(key, e[hdrSize:][:keyLen])
}

// KeyHash calculates the hash of the key.
func (e entry) KeyHash() uint64 {
	hdrSize, keyLen, _, _ := e.sizes()
	return xxhash.Sum64(e[hdrSize:][:keyLen])
}

// Value returns the value stored in the entry.
func (e entry) Value() []byte {
	hdrSize, keyLen, valLen, _ := e.sizes()
	return e[hdrSize:][keyLen:][:valLen]
}

// UpdateValue updates the value.
//
// It fails if no enough space.
func (e entry) UpdateValue(val []byte) bool {
	hdrSize, keyLen, valLen, spare := e.sizes()
	newValLen := len(val)
	if cap := valLen + spare; cap >= newValLen {
		copy(e[hdrSize:][keyLen:], val)
		e.setSizes(keyLen, newValLen, cap-newValLen)
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

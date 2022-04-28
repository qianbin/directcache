package directcache

import (
	"bytes"
	"math"
	"unsafe"

	"github.com/cespare/xxhash/v2"
)

const (
	deletedFlag = 1
	activeFlag  = 2
)

type entry []byte

func (e entry) HasFlag(flag uint8) bool { return e[0]&flag != 0 }
func (e entry) AddFlag(flag uint8)      { e[0] |= flag }
func (e entry) RemoveFlag(flag uint8)   { e[0] &^= flag }

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

func (e entry) hdr() (hdrSize, keyLen, valLen, spare int) {
	kw := 1 << (e[0] >> 6)
	vw := 1 << ((e[0] >> 4) & byte(3))

	hdrSize = 1 + kw + vw*2

	keyLen = e.intAt(1, kw)
	valLen = e.intAt(1+kw, vw)
	spare = e.intAt(1+kw+vw, vw)
	return
}

func (e entry) setHdr(keyLen, valLen, spare int) (hdrSize int) {
	kw, km := width(keyLen)
	vw, vm := width(valLen + spare)

	hdrSize = 1 + kw + vw*2

	e[0] &^= (uint8(3) << 6)
	e[0] |= (km << 6)

	e[0] &^= (uint8(3) << 4)
	e[0] |= (vm << 4)

	e.setIntAt(1, kw, keyLen)
	e.setIntAt(1+kw, vw, valLen)
	e.setIntAt(1+kw+vw, vw, spare)
	return
}

func (e entry) Size() int {
	hdrSize, keyLen, valLen, spare := e.hdr()
	return hdrSize + keyLen + valLen + spare
}

func (e entry) Init(key []byte, val []byte, spare int) {
	e[0] = 0 // reset flags

	keyLen, valLen := len(key), len(val)

	hdrSize := e.setHdr(keyLen, valLen, spare)
	copy(e[hdrSize:], key)
	copy(e[hdrSize:][keyLen:], val)
}

func (e entry) Match(key []byte) bool {
	hdrSize, keyLen, _, _ := e.hdr()
	return len(key) == keyLen &&
		bytes.Equal(key, e[hdrSize:][:keyLen])
}

func (e entry) KeyHash() uint64 {
	hdrSize, keyLen, _, _ := e.hdr()
	return xxhash.Sum64(e[hdrSize:][:keyLen])
}

func (e entry) Value() []byte {
	hdrSize, keyLen, valLen, _ := e.hdr()
	return e[hdrSize:][keyLen:][:valLen]
}

func (e entry) UpdateValue(val []byte) bool {
	hdrSize, keyLen, valLen, spare := e.hdr()
	newValLen := len(val)
	if cap := valLen + spare; cap >= newValLen {
		copy(e[hdrSize:][keyLen:], val)
		e.setHdr(keyLen, newValLen, cap-newValLen)
		return true
	}
	return false
}

func calcEntrySize(keyLen, valLen, spare int) int {
	kw, _ := width(keyLen)
	vw, _ := width(valLen + spare)

	hdrSize := 1 + kw + vw*2
	return hdrSize + keyLen + valLen + spare
}

func width(len int) (width int, mask byte) {
	switch {
	case len <= math.MaxUint8:
		return 1, 0
	case len <= math.MaxUint16:
		return 2, 1
	case int64(len) <= math.MaxUint32:
		return 4, 2
	default:
		return 8, 3
	}
}

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
	compactFlag = 4
)

type header8 struct {
	keyLen, valLen, spare uint8
}

type header32 struct {
	keyLen, valLen, spare uint32
}

type entry []byte

func (e entry) HasFlag(flag uint8) bool { return e[0]&flag != 0 }
func (e entry) AddFlag(flag uint8)      { e[0] |= flag }
func (e entry) RemoveFlag(flag uint8)   { e[0] &^= flag }

func (e entry) hdr() (hdrSize, keyLen, valLen, spare int) {
	hdrPtr := unsafe.Pointer(&e[1])
	if e.HasFlag(compactFlag) {
		hdr := (*header8)(hdrPtr)
		return 1 + int(unsafe.Sizeof(*hdr)), int(hdr.keyLen), int(hdr.valLen), int(hdr.spare)
	}
	hdr := (*header32)(hdrPtr)
	return 1 + int(unsafe.Sizeof(*hdr)), int(hdr.keyLen), int(hdr.valLen), int(hdr.spare)
}

func (e entry) setHdr(keyLen, valLen, spare int) (hdrSize int) {
	hdrPtr := unsafe.Pointer(&e[1])
	if keyLen <= math.MaxUint8 && (valLen+spare) <= math.MaxUint8 {
		e.AddFlag(compactFlag)
		hdr := (*header8)(hdrPtr)
		hdr.keyLen, hdr.valLen, hdr.spare = uint8(keyLen), uint8(valLen), uint8(spare)
		return 1 + int(unsafe.Sizeof(*hdr))
	}
	hdr := (*header32)(hdrPtr)
	hdr.keyLen, hdr.valLen, hdr.spare = uint32(keyLen), uint32(valLen), uint32(spare)
	return 1 + int(unsafe.Sizeof(*hdr))
}

func (e entry) Size() int {
	hdrSize, keyLen, valLen, spare := e.hdr()
	return hdrSize + keyLen + valLen + spare
}

func (e entry) Init(key []byte, keyHash uint64, val []byte, spare int) {
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
	var hdrSize int
	if keyLen <= math.MaxUint8 && (valLen+spare) <= math.MaxUint8 {
		hdrSize = 1 + int(unsafe.Sizeof(header8{}))
	} else {
		hdrSize = 1 + int(unsafe.Sizeof(header32{}))
	}
	return hdrSize + keyLen + valLen + spare
}

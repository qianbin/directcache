package directcache

import (
	"encoding/binary"
	"math"
)

const (
	deletedFlag      = 1 // the entry was deleted
	recentlyUsedFlag = 2 // the entry is recently accessed
)

// Entry presents the entry of a key-value pair.
type Entry interface {
	Key() []byte
	Value() []byte
	RecentlyUsed() bool
}

// entry consists of header and body.
type entry []byte

// flag ops. flags are stored in the first 4 bits of e[0].
func (e entry) HasFlag(flag uint8) bool { return e[0]&(flag<<4) != 0 }
func (e entry) AddFlag(flag uint8)      { e[0] |= (flag << 4) }
func (e entry) RemoveFlag(flag uint8)   { e[0] &^= (flag << 4) }

// RecentlyUsed complies Entry interface.
func (e entry) RecentlyUsed() bool { return e.HasFlag(recentlyUsedFlag) }

// lw extracts the number of bytes to present key/val length.
// It's stored in the last 2 bits of e[0].
func (e entry) lw() int { return 1 << (e[0] & 3) }

func (e entry) hdrSize() int { return 1 + e.lw()*3 }
func (e entry) keyLen() int  { return e.intAt(1, e.lw()) }
func (e entry) valLen() int  { return e.intAt(1+e.lw(), e.lw()) }
func (e entry) spare() int   { return e.intAt(1+e.lw()*2, e.lw()) }

// Size returns the entry size.
func (e entry) Size() int { return e.hdrSize() + e.keyLen() + e.valLen() + e.spare() }

// BodySize returns the sum of key, val length and spare.
func (e entry) BodySize() int { return e.keyLen() + e.valLen() + e.spare() }

// Key returns the key of the entry.
func (e entry) Key() []byte { return e[e.hdrSize():][:e.keyLen()] }

// Value returns the value of the entry.
func (e entry) Value() []byte { return e[e.hdrSize():][e.keyLen():][:e.valLen()] }

// Init initializes the entry with key, val and spare.
//
// The entry must be pre-alloced.
func (e entry) Init(key []byte, val []byte, spare int) {
	keyLen, valLen := len(key), len(val)
	lb := bitw(keyLen + valLen + spare)

	// init header
	e[0] = lb
	lw := 1 << lb
	e.setIntAt(1, lw, keyLen)
	e.setIntAt(1+lw, lw, valLen)
	e.setIntAt(1+lw*2, lw, spare)

	// init key and value
	hdrSize := 1 + lw*3
	copy(e[hdrSize:], key)
	copy(e[hdrSize:][keyLen:], val)
}

func (e entry) intAt(i int, w int) int {
	switch w {
	case 1:
		return int(e[i])
	case 2:
		return int(binary.BigEndian.Uint16(e[i:]))
	default:
		return int(binary.BigEndian.Uint32(e[i:]))
	}
}

func (e entry) setIntAt(i int, w int, n int) {
	switch w {
	case 1:
		e[i] = byte(n)
	case 2:
		binary.BigEndian.PutUint16(e[i:], uint16(n))
	default:
		binary.BigEndian.PutUint32(e[i:], uint32(n))
	}
}

// entrySize returns the size of an entry for given kv lengths.
func entrySize(keyLen, valLen, spare int) int {
	return 1 + (3 << bitw(keyLen+valLen+spare)) + // hdr
		keyLen + valLen + spare //body
}

// bitw where 1<<bitw is how many bytes needed to present n.
func bitw(n int) byte {
	switch {
	case n <= math.MaxUint8:
		return 0
	case n <= math.MaxUint16:
		return 1
	default:
		return 2
	}
}

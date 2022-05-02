package directcache

import (
	"encoding/binary"
	"math"
)

const (
	deletedFlag      = 1 // the entry was deleted
	recentlyUsedFlag = 2 // the entry is recently accessed
)

// entry consists of header and body.
type entry []byte

// flag ops. flags are stored in the first 4 bits of e[0].
func (e entry) HasFlag(flag uint8) bool { return e[0]&(flag<<4) != 0 }
func (e entry) AddFlag(flag uint8)      { e[0] |= (flag << 4) }
func (e entry) RemoveFlag(flag uint8)   { e[0] &^= (flag << 4) }

// kv vw extracts the width of key/val length.
// kw and vw are stored in the last 4 bits of e[0].
func (e entry) kw() int { return 1 << ((e[0] >> 2) & 3) }
func (e entry) vw() int { return 1 << (e[0] & 3) }

func (e entry) hdrSize() int { return 1 + e.kw() + e.vw()*2 }
func (e entry) keyLen() int  { return e.intAt(1, e.kw()) }
func (e entry) valLen() int  { return e.intAt(1+e.kw(), e.vw()) }
func (e entry) spare() int   { return e.intAt(1+e.kw()+e.vw(), e.vw()) }

// Size returns the entry size.
func (e entry) Size() int { return e.hdrSize() + e.keyLen() + e.valLen() + e.spare() }

// Key returns the key of the entry.
func (e entry) Key() []byte { return e[e.hdrSize():][:e.keyLen()] }

// Value returns the value of the entry.
func (e entry) Value() []byte { return e[e.hdrSize():][e.keyLen():][:e.valLen()] }

// Init initializes the entry with key, val and spare.
//
// The entry must be pre-alloced.
func (e entry) Init(key []byte, val []byte, spare int) {
	keyLen, valLen := len(key), len(val)
	kb, vb := bitw(keyLen), bitw(valLen+spare)

	// init header
	e[0] = (kb << 2) | vb
	kw, vw := 1<<kb, 1<<vb
	e.setIntAt(1, kw, keyLen)
	e.setIntAt(1+kw, vw, valLen)
	e.setIntAt(1+kw+vw, vw, spare)

	// init key and value
	hdrSize := 1 + kw + vw*2
	copy(e[hdrSize:], key)
	copy(e[hdrSize:][keyLen:], val)
}

// UpdateValue updates the value.
//
// It fails if no enough space.
func (e entry) UpdateValue(val []byte) bool {
	cap := e.valLen() + e.spare()
	if nvl := len(val); nvl <= cap {
		e.setIntAt(1+e.kw(), e.vw(), nvl)            // new value len
		e.setIntAt(1+e.kw()+e.vw(), e.vw(), cap-nvl) // new spare
		copy(e[e.hdrSize():][e.keyLen():], val)
		return true
	}
	return false
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
	return 1 + (1 << bitw(keyLen)) + (2 << bitw(valLen+spare)) + // hdr
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

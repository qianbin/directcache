package directcache

type map16 map[uint32]uint16
type map24 map[uint32]uint32
type map32 map[uint64]uint32
type map64 map[uint64]uint64

// vmap is the warpper of a map for mapping key hash to entry offset.
// The purpose of vmap is to reduce bucket RAM overhead.
type vmap struct {
	m interface{}
}

func (v *vmap) Reset(maxv int) {
	switch {
	case int64(maxv) <= 1<<16-1:
		v.m = make(map16)
	case int64(maxv) <= 1<<24-1:
		v.m = make(map24)
	case int64(maxv) <= 1<<32-1:
		v.m = make(map32)
	default:
		v.m = make(map64)
	}
}

func (v *vmap) Get(key uint64) (int, bool) {
	switch m := v.m.(type) {
	case map16:
		v, ok := m[uint32(key>>32)]
		return int(v), ok
	case map24:
		v, ok := m[uint32(key>>32)]
		return int(v), ok
	case map32:
		v, ok := m[key]
		return int(v), ok
	case map64:
		v, ok := m[key]
		return int(v), ok
	}
	return 0, false
}

func (v *vmap) Set(key uint64, val int) {
	switch m := v.m.(type) {
	case map16:
		m[uint32(key>>32)] = uint16(val)
	case map24:
		m[uint32(key>>32)] = uint32(val)
	case map32:
		m[key] = uint32(val)
	case map64:
		m[key] = uint64(val)
	}
}

func (v *vmap) Del(key uint64) {
	switch m := v.m.(type) {
	case map16:
		delete(m, uint32(key>>32))
	case map24:
		delete(m, uint32(key>>32))
	case map32:
		delete(m, key)
	case map64:
		delete(m, key)
	}
}

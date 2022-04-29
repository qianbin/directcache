package directcache

import (
	"errors"
)

// fifo is the no-split ring buffer based fifo queue.
// It ensures each pushed/popped data block within contiguous memory.
type fifo struct {
	buf         []byte
	front, back int
}

// Reset resets the capacity.
func (f *fifo) Reset(capacity int) {
	f.buf = make([]byte, 0, capacity)
	f.front, f.back = 0, 0
}

// Cap returns the capacity.
func (f *fifo) Cap() int { return cap(f.buf) }

// Size returns bytes of data in the queue.
func (f *fifo) Size() int {
	if n := len(f.buf); n == 0 {
		return 0
	} else if f.back > f.front {
		return f.back - f.front
	} else {
		return n - (f.front - f.back)
	}
}

// Front returns the offset of the front end.
func (f *fifo) Front() int { return f.front }

// Back returns the offset of the back end.
func (f *fifo) Back() int { return f.back }

// Slice returns the slice of the inner buffer after the given offset.
func (f *fifo) Slice(offset int) []byte {
	return f.buf[offset:]
}

// Push appends b after the back end, and returns the offset of the pushed data block.
//
// It fails if no enough contiguous space.
func (f *fifo) Push(b []byte, padding int) (offset int, ok bool) {
	if padding < 0 {
		panic(errors.New("fifo.Push: negative padding value"))
	}
	n := len(b) + padding
	if n == 0 {
		return f.back, true
	}

	// 'back' should be after 'front'
	if f.back == len(f.buf) {
		// extends buf
		if f.back+n <= cap(f.buf) {
			f.buf = f.buf[:f.back+n]
			copy(f.buf[f.back:], b)
			offset = f.back
			f.back += n
			return offset, true
		}

		// wrap 'back' to 0 offset
		if n <= f.front {
			copy(f.buf[:], b)
			f.back = n
			return 0, true
		}
		return 0, false
	}

	if f.back+n <= f.front {
		copy(f.buf[f.back:], b)
		offset = f.back
		f.back += n
		return offset, true
	}
	return 0, false
}

// Pop pops n bytes data from the front of the queue.
//
// It fails if no enough contiguous bytes after the front end.
func (f *fifo) Pop(n int) ([]byte, bool) {
	if n < 0 {
		panic(errors.New("fifo.Pop: negative n value"))
	}
	if n == 0 {
		return nil, true
	}

	if f.front+n <= f.back {
		s := f.buf[f.front:][:n]
		f.front += n
		// shrink buf to 0
		if f.front == f.back {
			f.front, f.back = 0, 0
			f.buf = f.buf[:0]
		}
		return s, true
	}

	if f.front+n <= len(f.buf) {
		s := f.buf[f.front:][:n]
		f.front += n
		// wrap 'front' to 0 offset
		if f.front == len(f.buf) {
			f.front = 0
			f.buf = f.buf[:f.back]
		}
		return s, true
	}
	return nil, false
}

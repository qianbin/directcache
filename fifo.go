package directcache

import "errors"

// fifo is the no-split ring buffer based fifo queue.
// It ensures each pushed/popped data block within contiguous memory.
type fifo struct {
	data        []byte
	front, back int
}

// Reset initializes with the capacity.
func (f *fifo) Reset(capacity int) {
	f.data = make([]byte, 0, capacity)
	f.front, f.back = 0, 0
}

// Front returns the offset of the front.
func (f *fifo) Front() int { return f.front }

// Back returns the offset of the back.
func (f *fifo) Back() int { return f.back }

// Cap returns the capacity.
func (f *fifo) Cap() int { return cap(f.data) }

// Size returns length of data in the queue.
func (f *fifo) Size() int {
	if n := len(f.data); n == 0 {
		return 0
	} else if f.back > f.front {
		return f.back - f.front
	} else {
		return n - (f.front - f.back)
	}
}

// View returns the reference of the inner data buffer at the given offset.
func (f *fifo) View(offset int) []byte {
	if l := len(f.data); l > 0 {
		return f.data[offset%l:]
	}
	return nil
}

// Push pushes the data block to the back of the queue.
func (f *fifo) Push(b []byte, padding int) (offset int, ok bool) {
	if padding < 0 {
		panic(errors.New("fifo.Push: negative padding value"))
	}
	n := len(b) + padding
	if n == 0 {
		return f.back, true
	}

	if f.back == len(f.data) {
		if f.back+n <= cap(f.data) {
			f.data = f.data[:f.back+n]
			copy(f.data[f.back:], b)
			offset = f.back
			f.back += n
			return offset, true
		}

		if n <= f.front {
			copy(f.data[:], b)
			f.back = n
			return 0, true
		}
		return 0, false
	}

	if f.back+n <= f.front {
		copy(f.data[f.back:], b)
		offset = f.back
		f.back += n
		return offset, true
	}
	return 0, false
}

// Pop pops n length data from the front of the queue.
func (f *fifo) Pop(n int) ([]byte, bool) {
	if n < 0 {
		panic(errors.New("fifo.Pop: negative n value"))
	}
	if n == 0 {
		return nil, true
	}

	if f.front+n <= f.back {
		s := f.data[f.front : f.front+n]
		f.front += n
		if f.front == f.back {
			f.front, f.back = 0, 0
			f.data = f.data[:0]
		}
		return s, true
	}

	if f.front+n <= len(f.data) {
		s := f.data[f.front : f.front+n]
		f.front += n
		if f.front == len(f.data) {
			f.front = 0
			f.data = f.data[:f.back]
		}
		return s, true
	}
	return nil, false
}

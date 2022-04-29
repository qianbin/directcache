package directcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fifo(t *testing.T) {
	var q fifo
	q.Reset(10)

	assert.Equal(t, 10, q.Cap())
	assert.Zero(t, q.Size())
	assert.Zero(t, q.Front())
	assert.Zero(t, q.Back())
	assert.Empty(t, q.Slice(0))

	q.Reset(20)
	assert.Equal(t, 20, q.Cap())
}

func Test_fifo_PushPop(t *testing.T) {
	var q fifo
	q.Reset(10)

	foo := "foo"
	bar := "bar"

	offset, _ := q.Push([]byte(foo), 0)
	assert.Equal(t, 0, offset)
	offset, _ = q.Push([]byte(bar), 0)
	assert.Equal(t, len(foo), offset)
	offset, _ = q.Push([]byte(foo), 0)
	assert.Equal(t, len(foo)+len(bar), offset)

	assert.Equal(t, []byte(foo+bar+foo), q.Slice(q.Front()))
	assert.Equal(t, len(foo)+len(bar)+len(foo), q.Size())

	popped, _ := q.Pop(len(foo))
	assert.Equal(t, foo, string(popped))
	popped, _ = q.Pop(len(bar))
	assert.Equal(t, bar, string(popped))
	popped, _ = q.Pop(len(foo))
	assert.Equal(t, foo, string(popped))

	assert.Zero(t, q.Size())
	assert.Zero(t, q.Front())
	assert.Zero(t, q.Back())
}

func Test_fifo_PushPop_wrap(t *testing.T) {
	var q fifo
	q.Reset(10)

	foo := "foo"
	bar := "bar"

	q.Push([]byte(foo), 0)
	q.Push([]byte(bar), 0)
	q.Push([]byte(foo), 0)

	_, ok := q.Push([]byte(bar), 0)
	assert.False(t, ok, "no space, should fail")

	q.Pop(len(foo))
	offset, ok := q.Push([]byte(bar), 0)
	assert.True(t, ok, "space returned by pop, should ok")
	assert.Zero(t, offset, "offset should wrap")
	assert.Equal(t, len(bar)+len(foo)+len(foo), q.Size())

	_, ok = q.Pop(q.Size())
	assert.False(t, ok, "data warpped, should fail")

	q.Pop(len(bar))
	// back < front
	q.Push([]byte(foo), 0)

	q.Pop(len(foo))
	assert.Zero(t, q.Front(), "front should wrap")
}

package directcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_fifo(t *testing.T) {
	var q fifo
	q.Reset(10)

	require.Equal(t, 10, q.Cap())
	require.Zero(t, q.Size())
	require.Zero(t, q.Front())
	require.Zero(t, q.Back())
	require.Empty(t, q.Slice(0))

	q.Reset(20)
	require.Equal(t, 20, q.Cap())
}

func Test_fifo_PushPop(t *testing.T) {
	var q fifo
	q.Reset(10)

	foo := "foo"
	bar := "bar"

	offset, _ := q.Push([]byte(foo), 0)
	require.Equal(t, 0, offset)
	offset, _ = q.Push([]byte(bar), 0)
	require.Equal(t, len(foo), offset)
	offset, _ = q.Push([]byte(foo), 0)
	require.Equal(t, len(foo)+len(bar), offset)

	require.Equal(t, []byte(foo+bar+foo), q.Slice(q.Front()))
	require.Equal(t, len(foo)+len(bar)+len(foo), q.Size())

	popped, _ := q.Pop(len(foo))
	require.Equal(t, foo, string(popped))
	popped, _ = q.Pop(len(bar))
	require.Equal(t, bar, string(popped))
	popped, _ = q.Pop(len(foo))
	require.Equal(t, foo, string(popped))

	require.Zero(t, q.Size())
	require.Zero(t, q.Front())
	require.Zero(t, q.Back())
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
	require.False(t, ok, "no space, should fail")

	q.Pop(len(foo))
	offset, ok := q.Push([]byte(bar), 0)
	require.True(t, ok, "space returned by pop, should ok")
	require.Zero(t, offset, "offset should wrap")
	require.Equal(t, len(bar)+len(foo)+len(foo), q.Size())

	_, ok = q.Pop(q.Size())
	require.False(t, ok, "data warpped, should fail")

	q.Pop(len(bar))
	// back < front
	q.Push([]byte(foo), 0)

	q.Pop(len(foo))
	require.Zero(t, q.Front(), "front should wrap")
}

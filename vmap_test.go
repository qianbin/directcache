package directcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_vmap(t *testing.T) {
	var m vmap

	m.Reset(1<<16 - 1)
	m.Set(1, 2)
	got, ok := m.Get(1)
	require.True(t, ok)
	require.Equal(t, 2, got)
	m.Del(1)
	_, ok = m.Get(1)
	require.False(t, ok)

	m.Reset(1<<24 - 1)
	m.Set(1, 2)
	got, ok = m.Get(1)
	require.True(t, ok)
	require.Equal(t, 2, got)
	m.Del(1)
	_, ok = m.Get(1)
	require.False(t, ok)

	m.Reset(1<<32 - 1)
	m.Set(1, 2)
	got, ok = m.Get(1)
	require.True(t, ok)
	require.Equal(t, 2, got)
	m.Del(1)
	_, ok = m.Get(1)
	require.False(t, ok)

	m.Reset(1 << 32)
	m.Set(1, 2)
	got, ok = m.Get(1)
	require.True(t, ok)
	require.Equal(t, 2, got)
	m.Del(1)
	_, ok = m.Get(1)
	require.False(t, ok)
}

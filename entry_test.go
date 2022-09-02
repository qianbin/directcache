package directcache

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_entry(t *testing.T) {
	t.Run("kv", func(t *testing.T) {
		key := "foo"
		val := "bar"

		ent := make(entry, entrySize(len(key), len(val), len(val)))
		copy(ent.Init([]byte(key), len(val), len(val)), val)

		require.Equal(t, entrySize(len(key), len(val), len(val)), ent.Size())
		require.Equal(t, key, string(ent.Key()))
		require.Equal(t, val, string(ent.Value()))
	})

	t.Run("k8v16", func(t *testing.T) {
		key := strings.Repeat("s", 10)
		val := strings.Repeat("v", 1000)

		ent := make(entry, entrySize(len(key), len(val), 0))
		copy(ent.Init([]byte(key), len(val), 0), val)

		require.Equal(t, key, string(ent.Key()))
		require.Equal(t, val, string(ent.Value()))
	})

	t.Run("k8v32", func(t *testing.T) {
		key := strings.Repeat("s", 10)
		val := strings.Repeat("v", 70000)

		ent := make(entry, entrySize(len(key), len(val), 0))
		copy(ent.Init([]byte(key), len(val), 0), val)

		require.Equal(t, key, string(ent.Key()))
		require.Equal(t, val, string(ent.Value()))
	})

	t.Run("k16v8", func(t *testing.T) {
		key := strings.Repeat("s", 1000)
		val := strings.Repeat("v", 10)

		ent := make(entry, entrySize(len(key), len(val), 0))
		copy(ent.Init([]byte(key), len(val), 0), val)

		require.Equal(t, key, string(ent.Key()))
		require.Equal(t, val, string(ent.Value()))
	})

	t.Run("k16v16", func(t *testing.T) {
		key := strings.Repeat("s", 1000)
		val := strings.Repeat("v", 1000)

		ent := make(entry, entrySize(len(key), len(val), 0))
		copy(ent.Init([]byte(key), len(val), 0), val)

		require.Equal(t, key, string(ent.Key()))
		require.Equal(t, val, string(ent.Value()))
	})

	t.Run("k16v32", func(t *testing.T) {
		key := strings.Repeat("s", 1000)
		val := strings.Repeat("v", 70000)

		ent := make(entry, entrySize(len(key), len(val), 0))
		copy(ent.Init([]byte(key), len(val), 0), val)

		require.Equal(t, key, string(ent.Key()))
		require.Equal(t, val, string(ent.Value()))
	})

	t.Run("k32v8", func(t *testing.T) {
		key := strings.Repeat("s", 70000)
		val := strings.Repeat("v", 10)

		ent := make(entry, entrySize(len(key), len(val), 0))
		copy(ent.Init([]byte(key), len(val), 0), val)

		require.Equal(t, key, string(ent.Key()))
		require.Equal(t, val, string(ent.Value()))
	})

	t.Run("k32v16", func(t *testing.T) {
		key := strings.Repeat("s", 70000)
		val := strings.Repeat("v", 1000)

		ent := make(entry, entrySize(len(key), len(val), 0))
		copy(ent.Init([]byte(key), len(val), 0), val)

		require.Equal(t, key, string(ent.Key()))
		require.Equal(t, val, string(ent.Value()))
	})

	t.Run("k32v32", func(t *testing.T) {
		key := strings.Repeat("s", 70000)
		val := strings.Repeat("v", 70000)

		ent := make(entry, entrySize(len(key), len(val), 0))
		copy(ent.Init([]byte(key), len(val), 0), val)

		require.Equal(t, key, string(ent.Key()))
		require.Equal(t, val, string(ent.Value()))
	})

	t.Run("flags", func(t *testing.T) {
		key := "foo"
		val := "bar"

		ent := make(entry, entrySize(len(key), len(val), len(val)))
		copy(ent.Init([]byte(key), len(val), len(val)), val)

		require.False(t, ent.HasFlag(deletedFlag))
		ent.AddFlag(deletedFlag)
		require.True(t, ent.HasFlag(deletedFlag))
		ent.RemoveFlag(deletedFlag)
		require.False(t, ent.HasFlag(deletedFlag))
	})
}

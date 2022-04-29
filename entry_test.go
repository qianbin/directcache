package directcache

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntry(t *testing.T) {
	t.Run("kv", func(t *testing.T) {
		key := "foo"
		val := "bar"

		ent := make(entry, entrySize(len(key), len(val), len(val)))
		ent.Init([]byte(key), []byte(val), len(val))

		assert.True(t, ent.Match([]byte(key)))
		assert.Equal(t, val, string(ent.Value()))

		assert.True(t, ent.UpdateValue([]byte(val+val)), "has spare space, should ok")
		assert.Equal(t, val+val, string(ent.Value()), "should be the updated value")

		assert.False(t, ent.UpdateValue([]byte(val+val+val)), "no space, should fail")
	})

	t.Run("bigkv", func(t *testing.T) {
		key := strings.Repeat("s", 300)
		val := strings.Repeat("v", 65536)

		ent := make(entry, entrySize(len(key), len(val), 0))
		ent.Init([]byte(key), []byte(val), 0)

		assert.True(t, ent.Match([]byte(key)))
		assert.Equal(t, val, string(ent.Value()))
	})

	t.Run("flags", func(t *testing.T) {
		key := "foo"
		val := "bar"

		ent := make(entry, entrySize(len(key), len(val), len(val)))
		ent.Init([]byte(key), []byte(val), len(val))

		assert.False(t, ent.HasFlag(deletedFlag))
		ent.AddFlag(deletedFlag)
		assert.True(t, ent.HasFlag(deletedFlag))
		ent.RemoveFlag(deletedFlag)
		assert.False(t, ent.HasFlag(deletedFlag))
	})
}

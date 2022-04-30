package seen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeen(t *testing.T) {
	seen := NewSeenList()

	ok := seen.Exist("a")
	require.False(t, ok, "should not exist")
	seen.Store("a", "foo")
	ok = seen.Exist("a")
	require.True(t, ok, "should exist")
	for i := 0; i < 1000; i++ {
		seen.Store("a", "foo")
	}
	require.True(t, ok, "should exist")

	seen.Delete("a")
	ok = seen.Exist("a")
	require.False(t, ok, "should not exist")

	for i := 0; i < 1000000; i++ {
		seen.Store(i, i)
	}

	m := seen.Map()
	require.Len(t, m, 1000000, "length is not correct")
}

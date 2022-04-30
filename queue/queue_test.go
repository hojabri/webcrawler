package queue

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	q := NewQueue()
	require.Equal(t, 0, q.Len(), "new Queue should have 0 length")
	q.Add(10)
	require.Equal(t, 1, q.Len(), "length of Queue should be 1 after adding one item")
	require.Equal(t, 10, q.Next().(int), "could not get correct value from queue")
	require.Equal(t, 0, q.Len(), "Queue should have 0 item after getting the single item")
	q.Add(11)
	require.Equal(t, 1, q.Len(), "length of Queue should be 1 after adding one item")
	require.Equal(t, 11, q.Next().(int), "could not get correct value from queue")
	require.Equal(t, 0, q.Len(), "Queue should have 0 item after getting the single item")
}

func TestAddLargeItems_Ints(t *testing.T) {
	q := NewQueue()

	for i := 0; i < 1000000; i++ {
		q.Add(i)
	}
	require.Equal(t, 1000000, q.Len(), "length of Queue should be 1 million")
	for i := 0; i < 1000000; i++ {
		item := q.Next()
		require.Equal(t, i, item.(int), "item is not correct")
	}
	require.Nilf(t, q.Next(), "should be nil")
}

func TestAddLargeItems_Structs(t *testing.T) {
	q := NewQueue()
	item := struct {
		Num int
		Str string
		Mp  map[string]int
	}{
		Num: 0,
		Str: "0",
		Mp:  map[string]int{"1": 1, "2": 2},
	}
	for i := 0; i < 1000000; i++ {
		item.Num = i
		item.Str = fmt.Sprint(i)
		q.Add(item)
	}
	require.Equal(t, 1000000, q.Len(), "length of Queue should be 1 million")
	for i := 0; i < 1000000; i++ {
		q.Next()
	}
	require.Equal(t, 0, q.Len(), "length of Queue should be 0")
}

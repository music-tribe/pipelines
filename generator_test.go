package pipelines

import (
	"context"
	"testing"
)

func TestGenerateFromSlice(t *testing.T) {
	t.Run("when we provide an empty slice, we should get an empty, closed stream back", func(t *testing.T) {
		ttStream := GenerateFromSlice(context.Background(), []string{})
		expectStreamLengthToBe(0, ttStream, t)
		expectClosedChannel(true, ttStream, t)
	})

	t.Run("when we provide a slice with a length of 1, we should get a stream with a length of 1 back", func(t *testing.T) {
		list := []string{"hello"}
		ttStream := GenerateFromSlice(context.Background(), list)
		expectStreamLengthToBe(len(list), ttStream, t)
		expectClosedChannel(true, ttStream, t)
	})

	t.Run("when we provide a slice of many items, we should get a stream of equal length back", func(t *testing.T) {
		list := []string{"hello", "world", "it's", "cold", "out"}
		ttStream := GenerateFromSlice(context.Background(), list)
		expectStreamLengthToBe(len(list), ttStream, t)
		expectClosedChannel(true, ttStream, t)
	})

	t.Run("when we values via the slice, we should receive those exact same values back from the stream", func(t *testing.T) {
		list := []string{"hello", "world", "it's", "cold", "out"}
		ttStream := GenerateFromSlice(context.Background(), list)
		expectOrderedResultsList(list, ttStream, t)
		expectClosedChannel(true, ttStream, t)
	})

	t.Run("when we cancel the Generator, we should expect to receive a truncated stream", func(t *testing.T) {
		// due to the random nature of the select statement, we could receive either 1 or 0 results via the channel
		list := []string{"hello", "world"}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ttStream := GenerateFromSlice(ctx, list)
		expectStreamLengthToBeLessThan(len(list), ttStream, t)
		expectClosedChannel(true, ttStream, t)
	})
}

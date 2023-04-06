package pipelines

import (
	"context"
	"reflect"
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

func TestGenerateHashFromStream(t *testing.T) {
	t.Run("when the inStream has a nil value, we should force a panic", func(t *testing.T) {
		defer func() {
			if perr := recover(); perr == nil {
				t.Errorf("expected FanIn to panic but got %v", perr)
			}
		}()
		_ = GenerateHashFromStream[int, string](context.Background(), nil)
	})

	t.Run("when we pass an item into the stream, we should receive that item back", func(t *testing.T) {
		list := []struct {
			Key int
			Val string
		}{
			{0, "thing0"},
		}
		ctx := context.Background()
		stream := GenerateFromSlice(ctx, list)
		got := GenerateHashFromStream(ctx, stream)
		res := map[int]string{0: "thing0"}
		if !reflect.DeepEqual(res, got) {
			t.Errorf("expected \n%+v\n  but got \n%+v\n", res, got)
		}
	})
}

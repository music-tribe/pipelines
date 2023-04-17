package pipelines

import (
	"context"
	"sync"
	"testing"
)

func TestOrDone(t *testing.T) {
	t.Run("when the inStream has a nil value, then the method should panic", func(t *testing.T) {
		defer func() {
			if perr := recover(); perr == nil {
				t.Error("expected the method to panic")
			}
		}()
		_ = OrDone[string](context.Background(), nil)
	})

	t.Run("when we cancel the context before the OrDone pipeline, we should receive a closed channel with less than one item in there", func(t *testing.T) {
		list := []string{"something", "nother", "another"}
		ctx, cancel := context.WithCancel(context.Background())
		inStream := GenerateFromSlice(ctx, list)
		cancel()
		outStream := OrDone(ctx, inStream)
		expectStreamLengthToBeLessThan(2, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when we pass a single item into the stream, we should receive a single item back", func(t *testing.T) {
		list := []string{"something"}
		ctx := context.Background()
		outStream := OrDone(ctx, GenerateFromSlice(ctx, list))
		expectStreamLengthToBe(len(list), outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when we pass multiple values into the stream, we should receive the same number of values back", func(t *testing.T) {
		list := []string{"something", "nother", "another"}
		ctx := context.Background()
		outStream := OrDone(ctx, GenerateFromSlice(ctx, list))
		expectStreamLengthToBe(len(list), outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when we pass multiple values into the stream, we should receive those same values back", func(t *testing.T) {
		list := []string{"something", "nother", "another"}
		ctx := context.Background()
		outStream := OrDone(ctx, GenerateFromSlice(ctx, list))
		expectOrderedResultsList(list, outStream, t)
		expectClosedChannel(true, outStream, t)
	})
}

func TestTeeSplitter(t *testing.T) {
	t.Run("when the inStream has a nil value, then the method should panic", func(t *testing.T) {
		defer func() {
			if perr := recover(); perr == nil {
				t.Error("expected the method to panic")
			}
		}()
		_, _ = TeeSplitter[string](context.Background(), nil)
	})

	t.Run("when we cancel the context before the TeeSplitter pipeline, we should receive 2 closed channel with containing less than one item", func(t *testing.T) {
		list := []string{"something", "nother", "another"}
		ctx, cancel := context.WithCancel(context.Background())
		inStream := GenerateFromSlice(ctx, list)
		cancel()
		o1, o2 := TeeSplitter(ctx, inStream)
		cnts := countAllStreamLengths(o1, o2)
		for _, cnt := range cnts {
			if cnt > 2 {
				t.Errorf("expected less than one item but got %d", cnt)

			}
		}
	})

	t.Run("when we provide an empty stream, we should recieve 2 empty closed streams", func(t *testing.T) {
		list := []string{}
		ctx := context.Background()
		o1, o2 := TeeSplitter(ctx, GenerateFromSlice(ctx, list))
		expectStreamLengthToBe(len(list), o1, t)
		expectStreamLengthToBe(len(list), o2, t)
		expectClosedChannel(true, o1, t)
		expectClosedChannel(true, o2, t)
	})

	t.Run("when we pass a stream that contains a single item, we should recieve 2 closed streams containing a single item", func(t *testing.T) {
		list := []string{"hello"}
		ctx := context.Background()
		o1, o2 := TeeSplitter(ctx, GenerateFromSlice(ctx, list))
		cnts := countAllStreamLengths(o1, o2)

		if cnts[0] != len(list) || cnts[0] != cnts[1] {
			t.Errorf("expected both stream length to be %d but stream 1 length was %d and stream 2 length was %d", len(list), cnts[0], cnts[1])
		}
		expectClosedChannel(true, o1, t)
		expectClosedChannel(true, o2, t)
	})

	t.Run("when we pass a single value into the stream, we should recieve 2 identical, closed streams containing that same single value", func(t *testing.T) {
		list := []string{"hello"}
		ctx := context.Background()
		o1, o2 := TeeSplitter(ctx, GenerateFromSlice(ctx, list))

		var idx int
		for val1 := range o1 {
			if val2 := <-o2; val1 != list[idx] || val1 != val2 {
				t.Errorf("expected value to be %s but stream1 = %s and stream2 = %s", list[idx], val1, val2)
			}
			idx++
		}

		expectClosedChannel(true, o1, t)
		expectClosedChannel(true, o2, t)
	})

	t.Run("when we pass a multiple values into the stream, we should recieve 2 identical, closed streams containing those same values", func(t *testing.T) {
		list := []string{"hello", "hi", "bonjour", "salut"}
		ctx := context.Background()
		o1, o2 := TeeSplitter(ctx, GenerateFromSlice(ctx, list))

		var idx int
		for val1 := range o1 {
			if val2 := <-o2; val1 != list[idx] || val1 != val2 {
				t.Errorf("expected value to be %s but stream1 = %s and stream2 = %s", list[idx], val1, val2)
			}
			idx++
		}

		expectClosedChannel(true, o1, t)
		expectClosedChannel(true, o2, t)
	})
}

func countAllStreamLengths[T any](streams ...(<-chan T)) []int {
	wg := sync.WaitGroup{}
	wg.Add(len(streams))
	counters := make([]int, len(streams))

	for idx, stream := range streams {
		go func(idx int, stream <-chan T) {
			defer wg.Done()
			for range stream {
				counters[idx]++
			}
		}(idx, stream)
	}

	wg.Wait()
	return counters
}

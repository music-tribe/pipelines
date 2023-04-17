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

func TestCombine(t *testing.T) {

	t.Run("when a single stream with a nil value has been passed, we should return an empty closed channel", func(t *testing.T) {
		ctx := context.Background()
		outStream := Combine[string](ctx, nil)
		expectStreamLengthToBe(0, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when no streams are passed, we should return an empty closed channel", func(t *testing.T) {
		ctx := context.Background()
		outStream := Combine[string](ctx)
		expectStreamLengthToBe(0, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when a single empty stream has been passed, we should return an empty closed channel", func(t *testing.T) {
		ctx := context.Background()
		outStream := Combine(ctx, GenerateFromSlice(ctx, []string{}))
		expectStreamLengthToBe(0, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when the context is cancelled before the pipeline, we should return a closed channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		a := []string{"hello"}
		b := []string{"bonjour"}
		streamA := GenerateFromSlice(ctx, a)
		streamB := GenerateFromSlice(ctx, b)
		cancel()
		outStream := Combine(ctx, streamA, streamB)
		expectStreamLengthToBeLessThan(1, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when a one of the streams has a nil value, it is ignored and only values from the other streams are returned", func(t *testing.T) {
		list := []string{"hello"}
		ctx := context.Background()
		stream1 := GenerateFromSlice(ctx, list)
		var stream2 chan string
		outStream := Combine(ctx, stream1, stream2)
		expectStreamLengthToBe(len(list), outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when we provide many streams with a single value, the resluting channel length should be the sum of all these streams content", func(t *testing.T) {
		a := []string{"hello"}
		b := []string{"bonjour"}
		c := []string{"guten tag"}
		len := len(a) + len(b) + len(c)
		ctx := context.Background()
		streamA := GenerateFromSlice(ctx, a)
		streamB := GenerateFromSlice(ctx, b)
		streamC := GenerateFromSlice(ctx, c)
		outStream := Combine(ctx, streamA, streamB, streamC)
		expectStreamLengthToBe(len, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when we provide many streams with multiple values, the resluting channel length should be the sum of all these streams content", func(t *testing.T) {
		a := []string{"apple", "pear", "grape"}
		b := []string{"chicken"}
		c := []string{"brocolli", "cheese", "turnip"}
		d := []string{}
		len := len(a) + len(b) + len(c)
		ctx := context.Background()
		streamA := GenerateFromSlice(ctx, a)
		streamB := GenerateFromSlice(ctx, b)
		streamC := GenerateFromSlice(ctx, c)
		streamD := GenerateFromSlice(ctx, d)
		outStream := Combine(ctx, streamA, streamB, streamC, streamD)
		expectStreamLengthToBe(len, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when we provide many streams with multiple values, the resluting channel should contain all these values", func(t *testing.T) {
		a := []string{"apple", "pear", "grape"}
		b := []string{"chicken", "peanuts"}
		ctx := context.Background()
		outStream := Combine(ctx, GenerateFromSlice(ctx, a), GenerateFromSlice(ctx, b))
		outList := make([]string, len(a)+len(b))
		var idx int
		for val := range outStream {
			outList[idx] = val
			idx++
		}

		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			for _, v := range a {
				if idx := getIndexOf(v, outList); idx < 0 {
					t.Errorf("could not find %s in outList", v)
				}
			}
		}()

		go func() {
			defer wg.Done()
			for _, v := range a {
				if idx := getIndexOf(v, outList); idx < 0 {
					t.Errorf("could not find %s in outList", v)
				}
			}
		}()

		wg.Wait()
		expectClosedChannel(true, outStream, t)
	})
}

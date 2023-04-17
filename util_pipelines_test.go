package pipelines

import (
	"context"
	"sync"
	"testing"
)

func TestTeeSplitter(t *testing.T) {
	t.Run("when the inStream has a nil value, then the method should panic", func(t *testing.T) {
		defer func() {
			if perr := recover(); perr == nil {
				t.Error("expected the method to panic")
			}
		}()
		_, _ = TeeSplitter[string](context.Background(), nil)
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
		var cnt1, cnt2 int
		wg := sync.WaitGroup{}

		wg.Add(2)
		go func() {
			defer wg.Done()
			for range o1 {
				cnt1++
			}
		}()

		go func() {
			defer wg.Done()
			for range o2 {
				cnt2++
			}
		}()

		wg.Wait()

		if cnt1 != len(list) || cnt1 != cnt2 {
			t.Errorf("expected both stream length to be %d but stream 1 length was %d and stream 2 length was %d", len(list), cnt1, cnt2)
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

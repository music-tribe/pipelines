package pipelines

import (
	"context"
	"testing"
)

func TestFanOut(t *testing.T) {
	var (
		pipeFunc WorkerFunc[string, string] = func(ctx context.Context, item string) string { return item }
		testFunc WorkerFunc[[2]int, int]    = func(ctx context.Context, in [2]int) int { return in[0] + in[1] }
	)

	t.Run("when we pass a nil value instean of a stream, we should receive a panic", func(t *testing.T) {
		defer func() {
			if perr := recover(); perr == nil {
				t.Errorf("expected FanOut to panic but got %v", perr)
			}
		}()
		_ = FanOut(context.Background(), nil, 1, pipeFunc)
	})

	t.Run("when we pass a nil value instean of a workerFunc, we should receive a panic", func(t *testing.T) {
		defer func() {
			if perr := recover(); perr == nil {
				t.Errorf("expected FanOut to panic but got %v", perr)
			}
		}()
		_ = FanOut[string, string](context.Background(), nil, 1, nil)
	})

	t.Run("when we request a concurrent processes, we should receive a closed stream containing one channel", func(t *testing.T) {
		ctx := context.Background()
		chanStream := FanOut(ctx, GenerateFromSlice(ctx, []string{"hello"}), 1, pipeFunc)
		expectStreamLengthToBe(1, chanStream, t)
		expectClosedChannel(true, chanStream, t)
		for chn := range chanStream {
			expectClosedChannel(true, chn, t)
		}
	})

	t.Run("when we request multiple concurrent processes, we should receive a closed stream containing multiple closed streams", func(t *testing.T) {
		ctx := context.Background()
		chanStream := FanOut(ctx, GenerateFromSlice(ctx, []string{"hello", "and", "welcome", "to", "testing"}), 4, pipeFunc)
		expectStreamLengthToBe(4, chanStream, t)
		expectClosedChannel(true, chanStream, t)
		for chn := range chanStream {
			expectClosedChannel(true, chn, t)
		}
	})

	t.Run("when we provide a workerFunc to a single stream, we expect to receive the correctly processed data back", func(t *testing.T) {
		ctx := context.Background()
		chanStream := FanOut(ctx, GenerateFromSlice(ctx, [][2]int{{1, 2}, {3, 4}}), 1, testFunc)
		expectClosedChannel(true, chanStream, t)
		for chn := range chanStream {
			expectOrderedResultsList([]int{3, 7}, chn, t)
			expectClosedChannel(true, chn, t)
		}
	})
}

func TestActionFactory(t *testing.T) {
	type (
		input struct {
			A, B int
		}
		output struct {
			id, res int
		}
	)
	var (
		pipeFunc WorkerFunc[string, string] = func(ctx context.Context, item string) string { return item }
		testFunc WorkerFunc[input, output]  = func(ctx context.Context, in input) output { return output{id: in.A, res: in.A + in.B} }
		inList                              = []input{{A: 0, B: 1}, {A: 4, B: 7}}
		wantList                            = []output{{id: 0, res: 1}, {id: 4, res: 11}}
	)

	t.Run("when the provided stream has a nil value, we should expect the worker stream to cancel the context", func(t *testing.T) {
		defer func() {
			if perr := recover(); perr == nil {
				t.Errorf("expected WorkerThread to panic but got %v", perr)
			}
		}()
		ctx := context.Background()
		_ = WorkerThread(ctx, nil, pipeFunc)
	})

	t.Run("when we put one item into the stream, we should get 1 item back before it closes", func(t *testing.T) {
		ctx := context.Background()
		outStream := WorkerThread(ctx, GenerateFromSlice(ctx, []string{"hello"}), pipeFunc)
		expectStreamLengthToBe(1, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when we put multiple items into the stream, we should get that many back before it closes", func(t *testing.T) {
		ctx := context.Background()
		outStream := WorkerThread(ctx, GenerateFromSlice(ctx, []string{"hello", "world"}), pipeFunc)
		expectStreamLengthToBe(2, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when we pass a WorkerFunc to perform a task on the elements of a stream, we should receive the expected results", func(t *testing.T) {
		ctx := context.Background()
		outStream := WorkerThread(ctx, GenerateFromSlice(ctx, inList), testFunc)
		expectOrderedResultsList(wantList, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when supply an empty stream, we should receive an empty stream from our worker", func(t *testing.T) {
		ctx := context.Background()
		outStream := WorkerThread(ctx, GenerateFromSlice(ctx, []input{}), testFunc)
		expectOrderedResultsList([]output{}, outStream, t)
		expectClosedChannel(true, outStream, t)
	})
}

func Benchmark(b *testing.B) {
	var pipeFunc WorkerFunc[string, string] = func(ctx context.Context, item string) string { return item }
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		_ = WorkerThread(ctx, GenerateFromSlice(ctx, []string{"hello"}), pipeFunc)
	}
}

func TestFanIn(t *testing.T) {
	var (
		pipeFunc WorkerFunc[string, string] = func(ctx context.Context, item string) string { return item }
	)

	t.Run("when the chanStream has a nil value, we should force a panic", func(t *testing.T) {
		defer func() {
			if perr := recover(); perr == nil {
				t.Errorf("expected FanIn to panic but got %v", perr)
			}
		}()
		_ = FanIn[string](context.Background(), nil)
	})

	t.Run("when we provide a single channel stream, we expect a closed channel in return", func(t *testing.T) {
		ctx := context.Background()
		chans := FanOut(ctx, GenerateFromSlice(ctx, []string{"thing"}), 1, pipeFunc)

		outStream := FanIn(context.Background(), chans)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when we provide a multi channel stream, we expect a closed channel in return", func(t *testing.T) {
		ctx := context.Background()
		chans := FanOut(ctx, GenerateFromSlice(ctx, []string{"thing"}), 5, pipeFunc)

		outStream := FanIn(context.Background(), chans)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when a single item is passed into the streams, we should expect a single item back", func(t *testing.T) {
		ctx := context.Background()
		list := []string{"single"}
		chans := FanOut(ctx, GenerateFromSlice(ctx, list), 5, pipeFunc)

		outStream := FanIn(context.Background(), chans)
		expectStreamLengthToBe(len(list), outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when a single item is passed into the streams, we should expect a single item back", func(t *testing.T) {
		ctx := context.Background()
		list := []string{"single"}
		chans := FanOut(ctx, GenerateFromSlice(ctx, list), 5, pipeFunc)

		outStream := FanIn(context.Background(), chans)
		expectStreamLengthToBe(len(list), outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when multiple items are passed into the streams, we should expect the same amount back", func(t *testing.T) {
		ctx := context.Background()
		list := []string{"this", "is", "a", "list", "of", "multiple", "items"}
		chans := FanOut(ctx, GenerateFromSlice(ctx, list), 5, pipeFunc)

		outStream := FanIn(context.Background(), chans)
		expectStreamLengthToBe(len(list), outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when multiple items are passed into the streams, we should check that the correct items are returned", func(t *testing.T) {
		ctx := context.Background()
		list := []string{"this", "is", "a", "list", "of", "multiple", "items"}
		chans := FanOut(ctx, GenerateFromSlice(ctx, list), 5, pipeFunc)

		outStream := FanIn(context.Background(), chans)
		expectStreamLengthToBe(len(list), outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	// t.Run("when we provide a stream containing a single stream, we should return that single closed stream", func(t *testing.T) {
	// 	ctx := context.Background()
	// 	list := []string{"hi", "there"}
	// 	chanStream := FanOut(ctx, GenerateFromSlice(ctx, list), 1, pipeFunc)
	// 	outStream := FanIn(ctx, chanStream)
	// 	expectStreamLengthToBe(len(list), outStream, t)
	// 	expectClosedChannel(true, outStream, t)
	// })
}

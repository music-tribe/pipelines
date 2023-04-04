package pipelines

import (
	"context"
	"testing"
)

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

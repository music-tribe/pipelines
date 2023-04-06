package pipelines

import "context"

type WorkerFunc[Arg, Res any] func(ctx context.Context, in Arg) Res

func FanOut[In, Out any](ctx context.Context, inStream <-chan In, maxProcs int, workerFunc WorkerFunc[In, Out]) <-chan (<-chan Out) {
	chanStream := make(chan (<-chan Out))
	if inStream == nil {
		close(chanStream)
		panic("WorkerThread: provided stream has nil value")
	}

	go func() {
		defer close(chanStream)
		for i := 0; i < maxProcs; i++ {
			select {
			case <-ctx.Done():
				return
			case chanStream <- WorkerThread(ctx, inStream, workerFunc):
			}
		}
	}()

	return chanStream
}

func WorkerThread[In, Out any](ctx context.Context, inStream <-chan In, workerFunc WorkerFunc[In, Out]) <-chan Out {
	resStream := make(chan Out)
	if inStream == nil {
		close(resStream)
		panic("WorkerThread: provided stream has nil value")
	}

	go func() {
		defer close(resStream)
		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-inStream:
				if !ok {
					return
				}
				select {
				case resStream <- workerFunc(ctx, item):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return resStream
}

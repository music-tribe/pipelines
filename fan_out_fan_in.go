package pipelines

import "context"

type WorkerFunc[Arg, Res any] func(ctx context.Context, in Arg) Res

func FanOut[In, Out any](ctx context.Context, inStream <-chan In, maxProcs int, workerFunc WorkerFunc[In, Out]) <-chan (<-chan Out) {
	chanStream := make(chan (<-chan Out))
	if inStream == nil {
		close(chanStream)
		panic("FanOut: inStream arg has nil value")
	}

	if workerFunc == nil {
		close(chanStream)
		panic("FanOut: workerFunc arg has nil value")
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

func FanIn[T any](ctx context.Context, chanStream <-chan (<-chan T)) chan T {
	outStream := make(chan T)
	if chanStream == nil {
		close(outStream)
		panic("FanIn: chanStream has nil value")
	}

	go func() {
		defer close(outStream)
		for {
			var possStream <-chan T
			select {
			case chn, ok := <-chanStream:
				if !ok {
					return
				}
				possStream = chn
			case <-ctx.Done():
				return
			}
			for t := range OrDone(ctx, possStream) {
				select {
				case outStream <- t:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

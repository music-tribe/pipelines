package pipelines

import "context"

type WorkerFunc[Arg, Res any] func(ctx context.Context, in Arg) Res

func WorkerThread[In, Out any](ctx context.Context, inStream <-chan In, workerFunc WorkerFunc[In, Out]) <-chan Out {
	resStream := make(chan Out)
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

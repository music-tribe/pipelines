package pipelines

import "context"

func GenerateFromSlice[T any](ctx context.Context, list []T) <-chan T {
	outStream := make(chan T)
	go generate(ctx, list, outStream)
	return outStream
}

func GenerateFromSliceBuffered[T any](ctx context.Context, list []T) <-chan T {
	outStream := make(chan T, len(list))
	go generate(ctx, list, outStream)
	return outStream
}

func generate[T any](ctx context.Context, list []T, outStream chan<- T) {
	defer close(outStream)
	for _, item := range list {
		select {
		case <-ctx.Done():
			return
		case outStream <- item:
		}
	}
}

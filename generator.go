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

func GenerateHashFromStream[K comparable, V any](ctx context.Context, inStream <-chan struct {
	Key K
	Val V
}) map[K]V {
	if inStream == nil {
		panic("GenerateHashFromStream: inStream has nil value")
	}
	out := make(map[K]V)
loop:
	for obj := range inStream {
		select {
		case <-ctx.Done():
			break loop
		default:
			out[obj.Key] = obj.Val
		}
	}

	return out
}

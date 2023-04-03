package pipelines

import "context"

func GenerateFromSlice[T any](ctx context.Context, list []T) <-chan T {
	outStream := make(chan T)
	go func() {
		defer close(outStream)
		for _, item := range list {
			select {
			case <-ctx.Done():
				return
			case outStream <- item:
			}
		}
	}()
	return outStream
}

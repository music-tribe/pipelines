package pipelines

import (
	"context"
	"sync"
)

func OrDone[R any](ctx context.Context, inStream <-chan R) <-chan R {
	outStream := make(chan R)
	if inStream == nil {
		close(outStream)
		panic("OrDone: the provided inStream argument has nil value")
	}

	go func() {
		defer close(outStream)
		for {
			select {
			case <-ctx.Done():
				return
			case res, ok := <-inStream:
				if !ok {
					return
				}
				select {
				case outStream <- res:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

// TeeSplitter should take a single stream and return 2 identical copies of it
func TeeSplitter[R any](ctx context.Context, inStream <-chan R) (_, _ <-chan R) {
	outStream1 := make(chan R)
	outStream2 := make(chan R)

	if inStream == nil {
		close(outStream1)
		close(outStream2)
		panic("TeeSplitter: the provided inStream argument has nil value")
	}

	go func() {
		defer close(outStream1)
		defer close(outStream2)

		for item := range OrDone(ctx, inStream) {
			var outStream1, outStream2 = outStream1, outStream2

			for i := 0; i < 2; i++ {
				select {
				case <-ctx.Done():
					return
				case outStream1 <- item:
					outStream1 = nil
				case outStream2 <- item:
					outStream2 = nil
				}
			}
		}
	}()

	return outStream1, outStream2
}

// Combine takes a context and any amount of channels of a type and combines them into one single channel of that same type.
func Combine[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	outStream := make(chan T)
	wg := sync.WaitGroup{}

	worker := func(inStream <-chan T) {
		defer wg.Done()
		if inStream == nil {
			return
		}

		for val := range OrDone(ctx, inStream) {
			select {
			case <-ctx.Done():
				return
			case outStream <- val:
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go worker(c)
	}

	go func() {
		wg.Wait()
		defer close(outStream)
	}()

	return outStream
}

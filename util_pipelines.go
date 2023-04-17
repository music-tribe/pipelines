package pipelines

import (
	"context"
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

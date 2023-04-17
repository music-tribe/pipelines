package pipelines

import (
	"context"
	"fmt"
)

func OrDone[R any](ctx context.Context, inStream <-chan R) <-chan R {
	outStream := make(chan R)

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
	if inStream == nil {
		panic("TeeSplitter: the provided inStream argument has nil value")
	}

	outStream1 := make(chan R)
	outStream2 := make(chan R)

	go func() {
		defer close(outStream1)
		defer close(outStream2)

		for item := range OrDone(ctx, inStream) {
			fmt.Println(item)
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

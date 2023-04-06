package pipelines

import "context"

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

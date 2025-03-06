package pipelines

import (
	"context"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
type HttpResponseHandler[T any] func(*http.Response, error) (T, error)
type HttpReqAsyncResponse[T any] struct {
	Res   T
	Error error
}

func HttpReqAsync[T any](ctx context.Context, httpClient HttpClient, req *http.Request, resHandler HttpResponseHandler[T]) <-chan HttpReqAsyncResponse[T] {
	res := make(chan HttpReqAsyncResponse[T])
	errStream := make(chan error)
	httpReqStream := make(chan T)

	go func() {
		defer close(res)
		for {
			select {
			case <-ctx.Done():
				res <- HttpReqAsyncResponse[T]{Error: ctx.Err()}
			case err, ok := <-errStream:
				if !ok {
					// errStream channel closed - handle this in way that suits your app
					return
				}
				res <- HttpReqAsyncResponse[T]{Error: err}
			case httpGetItem, ok := <-httpReqStream:
				if !ok {
					// httpReqStream channel closed - handle this in way that suits your app
					return
				}
				res <- HttpReqAsyncResponse[T]{Res: httpGetItem}
			}
		}
	}()

	go func() {
		defer close(errStream)
		defer close(httpReqStream)

		getItem, err := resHandler(httpClient.Do(req))
		if err != nil {
			errStream <- err
			return
		}
		httpReqStream <- getItem
	}()

	return res
}

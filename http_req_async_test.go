package pipelines

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_httpReqAsync(t *testing.T) {
	t.Parallel()

	resHandler := func(res *http.Response, err error) (sampleObject, error) {
		if err != nil {
			return sampleObject{}, err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return sampleObject{}, err
		}

		if res.StatusCode != http.StatusOK {
			return sampleObject{}, fmt.Errorf("failed to retrieve object. Response: %s, Body: %s", res.Status, body)
		}

		o := sampleObject{}
		err = json.Unmarshal(body, &o)
		if err != nil {
			return sampleObject{}, fmt.Errorf("failed to unmarshal object: %v. Response: %s, Body: %s", err, res.Status, body)
		}
		return o, nil
	}

	t.Run("Should return a channel with no error & response", func(t *testing.T) {
		ctx := context.TODO()

		expectedObject := createSampleObject()
		stubObjectBody, err := json.Marshal(expectedObject)
		require.NoError(t, err)

		client := &http.Client{
			Transport: newStubRoundTripper(
				&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(stubObjectBody)),
				}, nil),
		}

		req, err := http.NewRequest(http.MethodGet, "https://my-api/objects/v1/1/?expand=metadata", nil)
		require.NoError(t, err)
		h := http.Header{}
		h.Set("Authorization", "Bearer eyJhbGc")
		h.Set("Content-Type", "application/json")
		req.Header = h

		res := <-HttpReqAsync(ctx, client, req, resHandler)
		assert.NoError(t, res.Error)
		assert.Equal(t, expectedObject, res.Res)
	})

	t.Run("Should return an error if our http client errors", func(t *testing.T) {
		ctx := context.TODO()

		expectedErr := errors.New("error")
		client := &http.Client{
			Transport: newStubRoundTripper(nil, expectedErr),
		}

		req, err := http.NewRequest(http.MethodGet, "https://my-api/objects/v1/1/?expand=metadata", nil)
		require.NoError(t, err)
		h := http.Header{}
		h.Set("Authorization", "Bearer eyJhbGc")
		h.Set("Content-Type", "application/json")
		req.Header = h

		res := <-HttpReqAsync(ctx, client, req, resHandler)
		assert.Error(t, res.Error)
		assert.ErrorIs(t, res.Error, expectedErr)
	})

	t.Run("Should return an error if we can't read the body", func(t *testing.T) {
		ctx := context.TODO()

		client := &http.Client{
			Transport: newStubRoundTripper(
				&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(errReader(0)),
				}, nil),
		}

		req, err := http.NewRequest(http.MethodGet, "https://my-api/objects/v1/1/?expand=metadata", nil)
		require.NoError(t, err)
		h := http.Header{}
		h.Set("Authorization", "Bearer eyJhbGc")
		h.Set("Content-Type", "application/json")
		req.Header = h

		res := <-HttpReqAsync(ctx, client, req, resHandler)
		assert.Error(t, res.Error)
		assert.Equal(t, res.Error.Error(), "test error")
	})

	t.Run("Should return an error if we don't get a StatusOK", func(t *testing.T) {
		ctx := context.TODO()

		client := &http.Client{
			Transport: newStubRoundTripper(
				&http.Response{
					Status:     http.StatusText(http.StatusForbidden),
					StatusCode: http.StatusForbidden,
					Body:       io.NopCloser(bytes.NewReader([]byte("test"))),
				}, nil),
		}

		req, err := http.NewRequest(http.MethodGet, "https://my-api/objects/v1/1/?expand=metadata", nil)
		require.NoError(t, err)
		h := http.Header{}
		h.Set("Authorization", "Bearer eyJhbGc")
		h.Set("Content-Type", "application/json")
		req.Header = h

		res := <-HttpReqAsync(ctx, client, req, resHandler)
		assert.Error(t, res.Error)
		assert.Equal(t, res.Error.Error(), fmt.Sprintf("failed to retrieve object. Response: %s, Body: %s", http.StatusText(http.StatusForbidden), "test"))
	})

	t.Run("Should return an error if we can't unmarshal the device response", func(t *testing.T) {
		ctx := context.TODO()

		client := &http.Client{
			Transport: newStubRoundTripper(
				&http.Response{
					Status:     http.StatusText(http.StatusOK),
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte("test"))),
				}, nil),
		}

		req, err := http.NewRequest(http.MethodGet, "https://my-api/objects/v1/1/?expand=metadata", nil)
		require.NoError(t, err)
		h := http.Header{}
		h.Set("Authorization", "Bearer eyJhbGc")
		h.Set("Content-Type", "application/json")
		req.Header = h

		res := <-HttpReqAsync(ctx, client, req, resHandler)
		assert.Error(t, res.Error)
		assert.Equal(t, res.Error.Error(), fmt.Sprintf("failed to unmarshal object: %v. Response: %s, Body: %s", errors.New("invalid character 'e' in literal true (expecting 'r')"), http.StatusText(http.StatusOK), "test"))
	})
}

type sampleObject struct {
	Id   int    `json:"_id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func createSampleObject() sampleObject {
	return sampleObject{
		Id:   1,
		Name: "John",
		Age:  30,
	}
}

type stubRoundTripper struct {
	response *http.Response
	err      error
}

func newStubRoundTripper(response *http.Response, err error) *stubRoundTripper {
	return &stubRoundTripper{response, err}
}
func (sr *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return sr.response, sr.err
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

package pipelines

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type testType struct {
	OrigIdx int
	Name    string
}

func TestGenerateFromSlice(t *testing.T) {
	t.Run("when the slice is empty, it should just return a closed channel", func(t *testing.T) {
		list := populateList(0)
		ttStream := GenerateFromSlice(context.Background(), list)
		expectResults([]testType{}, ttStream, t)
		expectClosedChannel(true, ttStream, t)
	})

	t.Run("when we provide a slice, it should populate and close the channel, ensuring all items from that slice were added", func(t *testing.T) {
		list := populateList(200)
		ttStream := GenerateFromSlice(context.Background(), list)
		expectResults(list, ttStream, t)
		expectClosedChannel(true, ttStream, t)
	})

	t.Run("when we cancel the Generator, we should receive a truncated channel", func(t *testing.T) {
		// due to the random nature of the select statement, we could receive either 1 or 0 results via the channel
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ttStream := GenerateFromSlice(ctx, populateList(2))
		expectedResultsViaChannelToBeLessThan(2, ttStream, t)
		expectClosedChannel(true, ttStream, t)
	})
}

func expectClosedChannel[T any](expect bool, stream <-chan T, t *testing.T) {
	// read the stream first to remove race condition
	for range stream {
	}
	if _, ok := <-stream; ok && expect {
		t.Errorf("expected channel closure to be %v but it was %v", expect, ok)
	}
}

func populateList(count int) []testType {
	list := make([]testType, count)
	for i := 0; i < count; i++ {
		list[i] = testType{
			OrigIdx: i,
			Name:    fmt.Sprintf("test-%d", i),
		}
	}
	return list
}

func expectResults(expect []testType, stream <-chan testType, t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exp := make(map[int]testType)
	got := make(map[int]testType)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for _, tt := range expect {
			exp[tt.OrigIdx] = tt
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				t.Errorf("expectResults has errored internally: %v", ctx.Err())
				return
			case tt, ok := <-stream:
				if !ok {
					return
				}
				got[tt.OrigIdx] = tt
			}
		}
	}()

	wg.Wait()

	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected map representation of stream results to be \n%+v\n but got \n%+v\n", exp, got)
	}
}

func expectedResultsViaChannelToBeLessThan[T any](max int, stream <-chan T, t *testing.T) {
	var count int
	for range stream {
		count++
	}
	if count > max {
		t.Errorf("expected stream length to be less than %d but got %d\n", max, count)
	}
}

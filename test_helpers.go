package pipelines

import (
	"reflect"
	"testing"
)

func expectClosedChannel[T any](expect bool, stream <-chan T, t testing.TB) {
	// read the stream first to remove race condition
	for range stream {
	}

	if _, ok := <-stream; ok && expect {
		t.Errorf("expected channel closure to be %v but it was %v", expect, ok)
	}
}

func expectOrderedResultsList[W any](want []W, outStream <-chan W, t testing.TB) {
	got := make([]W, len(want))
	var idx int
	for item := range outStream {
		got[idx] = item
		idx++
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted resulting list to be %+v but got %+v\n", want, got)
	}
}

func expectStreamLengthToBe[T any](expect int, stream <-chan T, t testing.TB) {
	if cnt := lenStream(stream); cnt != expect {
		t.Errorf("expected stream length to be %d but it was %d", expect, cnt)
	}
}

func expectStreamLengthToBeLessThan[T any](max int, stream <-chan T, t testing.TB) {
	if cnt := lenStream(stream); cnt > max {
		t.Errorf("expected stream length to be less than %d but got %d\n", max, cnt)
	}
}

func lenStream[T any](inStream <-chan T) (count int) {
	for range inStream {
		count++
	}
	return
}

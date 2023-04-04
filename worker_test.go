package pipelines

import (
	"context"
	"reflect"
	"strconv"
	"testing"
)

type (
	input struct {
		A, B int
	}
	output struct {
		res string
		err error
	}
)

var (
	inList   = []input{{1, 0}, {1, 1}, {2, 2}, {4, 3}, {7, 4}}
	wantList = []output{
		{res: "1", err: nil},
		{res: "2", err: nil},
		{res: "4", err: nil},
		{res: "7", err: nil},
		{res: "11", err: nil},
	}
)

func TestActionFactory(t *testing.T) {
	t.Run("when we perform a task on the elements of a stream, we should receive the expected results", func(t *testing.T) {
		ctx := context.Background()
		inStream := GenerateFromSlice(ctx, inList)
		outStream := WorkerThread(ctx, inStream, testFunc)
		expectResultsList(wantList, outStream, t)
		expectClosedChannel(true, outStream, t)
	})

	t.Run("when supply an empty stream, we should receive an empty stream from our worker", func(t *testing.T) {
		ctx := context.Background()
		inStream := GenerateFromSlice(ctx, []input{})
		outStream := WorkerThread(ctx, inStream, testFunc)
		expectResultsList([]output{}, outStream, t)
		expectClosedChannel(true, outStream, t)
	})
}

var testFunc WorkerFunc[input, output] = func(ctx context.Context, in input) output {
	return output{res: strconv.Itoa(in.A + in.B)}
}

func expectResultsList[W any](want []W, outStream <-chan W, t *testing.T) {
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

# Pipelines
This package contains a small collection of generic, concurrent, reusable pipelines, that could prove useful in our projects.

## In use

### GenerateFromSlice
Generate a new reader channel from a slice of any type
```golang
import (
    "github.com/music-tribe/pipelines"
)

func main() {
	numList := []int{1, 4, 67, 843, 23}

	// in this situation, the content type within the slice is inferred
	genStream := pipelines.GenerateFromSlice(context.Background(), numList)

	...
	// do some funky concurrency stuff
}
```
### FanOut, WorkerFunc and FanIn
FanOut allows us to spread our work accross several workers. The WorkerFunc specifies the job that the worker must undertake. We then fan back in to a a single stream with FanIn.
This will be most useful with processor intensive or long running tasks.
```golang

func main() {
	// It's always worth benchmarking your functionality when using pipelines - concurrency is often not the best course of action.
	numList := []int{1, 4, 67, 843, 23, 57, 21, 68}
	maxProcs := 4
	var convToString WorkerFunc[int, string] = func(ctx context.Context, item int) string { return strconv.Itoa(item) }

	ctx := context.Background()
	numStream := pipelines.GenerateFromSlice(ctx, numList)
	chanStream := pipelines.FanOut(ctx, numStream, maxProcs, convToString)
	strStream := pipelines.FanIn(ctx, chanStream)

	for val := range strStream {
		fmt.Println(val)
	}

	    // This could also be written as...
	    stream := pipelines.FanIn(ctx, pipelines.FanOut(ctx, pipelines.GenerateFromSlice(ctx, numList), 4, convToString))

	for val := range strStream {
		fmt.Println(val)
	}
}

```
### TeeSplitter
TeeSplitter allows us to create 2 identical copies of one channel. This is useful when you require the same channel to perform two different tasks.
```golang
func main() {
	words := []string{"hello", "salut", "bonjour"}

	ctx := context.Background()
	splitMeStream := pipelines.GenerateFromSlice(ctx, words)
	out1, out2 := pipelines.TeeSplitter(ctx, splitMeStream)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for val := range out1 {
			fmt.Printf("stream out1:%s", val)
		}
	}()

	go func() {
		defer wg.Done()
		for val := range out2 {
			fmt.Printf("stream out2:%s", val)
		}
	}()

	wg.Wait()
}
```
### Combine
Combine allows us to combine any number of channels of the same type into one single channel of that type.
```golang
func main() {
	trees := []string{"ash", "oak", "beech"}
	bushes := []string{"laurel", "rhododendron"}
	greenery := make([]string, len(trees)+len(bushes))

	ctx := context.Background()
	treeStream := pipelines.GenerateFromSlice(ctx, trees)
	bushStream := pipelines.GenerateFromSlice(ctx, bushes)

	greenStream := Combine(ctx, treeStream, bushStream)

	var idx int
	for val := range greenStream {
		greenery[idx] = val
		idx++
	}
	// this is clearly not the most efficient way to proceed in this situation! :-)
}
```

### Heartbeats
DoWorkWithHeartbeats allows us to give a long running task a pulse - we can constantly monitor it's health and
watch for silent failures.
Simple example
```golang
var (
	timeout = time.Second*10
	pulseInterval = time.Second
)

type Response struct {
	Data []byte
	Err  error
}

func taskThatNeedsAHeartbeat(ctx context.Context) Response {
	select {
	case <-ctx.Done():
		return Response{Err: ctx.Err()}
	default:
	}

	data, err := myLongRunningTask(ctx)
	return Response{
		Data: data,
		Err:  err,
	}
}

res, err := DoWorkWithHeartbeats(
	context.Background(),
	taskThatNeedsAHeartbeat,
	func(ho *HeartbeatOptions) {
		ho.PulseInterval = pulseInterval
		ho.Timeout = timeout
	},
)
if err != nil {
	// this is any error that occured due to the DoWorkWithHeartbeats decorator
	log.Fatal(err)
}

if res.Err != nil {
	// this is the task error - we've used it in this case but it may not always be present due to the generic response
	log.Fatal(res.Err)
}

doSomethingWithData(res.Data)

...

```

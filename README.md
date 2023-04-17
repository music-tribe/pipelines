# Pipelines
This package contains a small collection of generic, concurrent, reusable pipelines, that could prove useful in our projects.

## In use

### Generate channels from slices
```golang
import (
    "github.com/music-tribe/pipelines"
)

func main() {
    numList := []int{1, 4, 67, 843, 23}

    // in this situation, the content type within the slice is inferred
    genStream := pipelines.GenerateFromSlice(context.Background(), numList)

    ...
}
```
### Fan out work concurrently
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
### Creating 2 identical copies of a channel
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
### Combine multiple channels into one
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

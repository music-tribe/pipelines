package pipelines

import (
	"context"
	"fmt"
	"time"
)

func DoWorkWithHeartbeats[R any](ctx context.Context, pulseInterval time.Duration, longRunningFunc func(context.Context) R) (<-chan interface{}, <-chan R) {
	if longRunningFunc == nil {
		panic("DoWorkWithHeartbeats: longRunningFunc arg has nil value")
	}

	heartbeat := make(chan interface{})
	results := make(chan R)

	go func() {
		defer close(heartbeat)
		defer close(results)

		pulse := time.NewTicker(pulseInterval)

		workChan := func() <-chan R {
			r := make(chan R)

			go func() {
				defer close(r)
				select {
				case <-ctx.Done():
					return
				case r <- longRunningFunc(ctx):
				}
			}()

			return r
		}()

		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
				fmt.Println("pulse")
			default:
				fmt.Println("channel is blocked")
			}
		}

		sendResult := func(res R) {
			for {
				select {
				case <-ctx.Done():
					return
				case <-pulse.C:
					sendPulse()
				case results <- res:
					return
				}
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-pulse.C:
				sendPulse()
			case res, ok := <-workChan:
				if !ok {
					return
				}
				fmt.Println("longRunningFunc complete")
				sendResult(res)
			}
		}
	}()

	return heartbeat, results
}

package pipelines

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func doWorkWithHeartbeats[R any](ctx context.Context, pulseInterval time.Duration, longRunningFunc func(context.Context) R) (<-chan interface{}, <-chan R) {
	if longRunningFunc == nil {
		panic("doWorkWithHeartbeats: longRunningFunc arg has nil value")
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

func DoWorkWithHeartbeats[R any](ctx context.Context, pulseInterval, timeout time.Duration, workFunc func(context.Context) R) (R, error) {
	ctx, cancel := context.WithCancel(ctx)
	time.AfterFunc(timeout, func() { cancel() })

	heartbeat, results := doWorkWithHeartbeats(ctx, pulseInterval, workFunc)
	var r R
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				return r, errors.New("DoWorkWithHeartbeats: heartbeat channel may be closed already")
			}
		case res, ok := <-results:
			if !ok {
				return r, errors.New("DoWorkWithHeartbeats: results channel may be closed already")
			}
			return res, nil
		case <-time.After(pulseInterval * 2):
			return r, errors.New("DoWorkWithHeartbeats: workFunc timed out")
		}
	}
}

package pipelines

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDoWorkWithHeartbeats(t *testing.T) {
	t.Run("when the longRunningFunc has a nil value, we should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			DoWorkWithHeartbeats[any](context.TODO(), time.Second, nil)
		})
	})

	t.Run("when the context is canceled before we expect a heartbeat of the work to finish, we should receive no response", func(t *testing.T) {
		pulseInterval := time.Second
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		go func() {
			time.Sleep(pulseInterval / 2)
			cancel()
		}()

		heartbeat, results := DoWorkWithHeartbeats(ctx, pulseInterval, func(ctx context.Context) error { time.Sleep(pulseInterval * 2); return nil })

		done := make(chan interface{})
		pulseCount := 0
		go func() {
			for range heartbeat {
				pulseCount++
			}
			defer close(done)
		}()

		res := <-results

		<-done
		assert.Nil(t, res)
		assert.Equal(t, 0, pulseCount)
	})

	t.Run("when the work takes 2 seconds to complete, we should receive to 1 second heartbeats", func(t *testing.T) {
		pulseInterval := time.Second
		ctx := context.TODO()

		heartbeat, results := DoWorkWithHeartbeats(
			ctx,
			pulseInterval,
			func(ctx context.Context) int { time.Sleep(pulseInterval * 2); return 3 },
		)

		done := make(chan interface{})
		pulseCount := 0
		go func() {
			for range heartbeat {
				pulseCount++
			}
			defer close(done)
		}()

		var res int
		for res = range results {

		}

		<-done
		assert.Equal(t, 3, res)
		assert.Equal(t, 2, pulseCount)
	})
}

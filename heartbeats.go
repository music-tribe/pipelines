package pipelines

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func doWorkWithHeartbeats[R any](ctx context.Context, pulseInterval time.Duration, task func(context.Context) R, logger HeartbeatsLogger) (<-chan interface{}, <-chan R) {
	if task == nil {
		panic("doWorkWithHeartbeats: task arg has nil value")
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
				case r <- task(ctx):
				}
			}()

			return r
		}()

		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
				logger.Debugf("doWorkWithHeartbeats: pulse\n")
			default:
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
				logger.Infof("doWorkWithHeartbeats: task completed")
				sendResult(res)
			}
		}
	}()

	return heartbeat, results
}

type HeartbeatsLogger interface {
	Debugf(fmt string, fields ...interface{})
	Infof(fmt string, fields ...interface{})
	Errorf(fmt string, fields ...interface{})
}

type HeartbeatsOptions struct {
	PulseInterval time.Duration
	Timeout       time.Duration

	Logger HeartbeatsLogger
	Debug  bool
}

type HeartbeatsOption func(*HeartbeatsOptions)

func DoWorkWithHeartbeats[R any](ctx context.Context, task func(ctx context.Context) R, options ...HeartbeatsOption) (R, error) {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	ops := HeartbeatsOptions{
		PulseInterval: time.Second,
		Timeout:       time.Second * 30,
		Logger:        l,
	}

	for _, optFunc := range options {
		optFunc(&ops)
	}

	ctx, cancel := context.WithCancel(ctx)
	time.AfterFunc(ops.Timeout, func() { cancel() })

	heartbeat, results := doWorkWithHeartbeats(ctx, ops.PulseInterval, task, ops.Logger)
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
		case <-time.After(ops.PulseInterval * 2):
			return r, errors.New("DoWorkWithHeartbeats: task timed out")
		}
	}
}

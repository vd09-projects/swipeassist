package utils

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	sharedRand *rand.Rand
	randOnce   sync.Once
)

func SharedRand() *rand.Rand {
	randOnce.Do(func() {
		sharedRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
	return sharedRand
}

var ErrInvalidSleepRange = errors.New("invalid sleep range")

func sleepCtxWithRand(
	ctx context.Context,
	r *rand.Rand,
	min, max time.Duration,
) error {
	if min < 0 || max < min {
		return ErrInvalidSleepRange
	}

	// Fast path
	if max == 0 {
		return ctx.Err()
	}

	var d time.Duration
	if min == max {
		d = min
	} else {
		d = min + time.Duration(r.Int63n(int64(max-min)+1))
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func SleepCtx(ctx context.Context, d time.Duration) error {
	return sleepCtxWithRand(ctx, SharedRand(), d, d)
}

func RandomSleepCtx(ctx context.Context, min, max time.Duration) error {
	return sleepCtxWithRand(ctx, SharedRand(), min, max)
}

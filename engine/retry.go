package engine

import (
	"context"
	"errors"
	"time"
)

func (e *Engine) retry(ctx context.Context, fn func() error) error {
	var last error
	for i := 0; i < e.cfg.RetryAttempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := fn()
		if err == nil {
			return nil
		}
		last = err
		_ = sleepCtx(ctx, e.cfg.RetryDelay)
	}
	if last == nil {
		last = errors.New("retry failed")
	}
	return last
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
package engine

import (
	"context"
	"fmt"
	"time"
)

func (e *Engine) findFirstVisible(ctx context.Context, selectors []string, timeout time.Duration) (Element, string, error) {
	e.MustHavePage()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nil, "", err
		}

		for _, sel := range selectors {
			el, _ := e.page.Timeout(600 * time.Millisecond).Element(sel)
			if el == nil {
				continue
			}
			ok, _ := el.EvalBool(`() => {
				const r = this.getBoundingClientRect();
				return !!(r && r.width > 0 && r.height > 0);
			}`)
			if ok {
				return el, sel, nil
			}
		}

		if err := sleepCtx(ctx, 120*time.Millisecond); err != nil {
			return nil, "", err
		}
	}

	return nil, "", fmt.Errorf("timeout waiting for any visible selector: %v", selectors)
}
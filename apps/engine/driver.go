package engine

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

type Driver struct {
	e *Engine
}

func NewDriver(e *Engine) *Driver { return &Driver{e: e} }

func (d *Driver) Close() { d.e.Close() }

func (d *Driver) Open(ctx context.Context, url string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	d.e.open(url)
	return nil
}

func (d *Driver) Screenshot(ctx context.Context, filePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.e.screenshot(filePath)
}

func (d *Driver) WaitAnyVisible(ctx context.Context, selectors []string) error {
	_, _, err := d.e.findFirstVisible(ctx, selectors, d.e.cfg.StepTimeout)
	return err
}

func (d *Driver) IsVisible(ctx context.Context, selectors []string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	d.e.MustHavePage()

	for _, sel := range selectors {
		el, _ := d.e.page.Timeout(600 * time.Millisecond).Element(sel)
		if el == nil {
			continue
		}
		ok, _ := el.EvalBool(`() => {
			const r = this.getBoundingClientRect();
			return !!(r && r.width > 0 && r.height > 0);
		}`)
		if ok {
			return true, nil
		}
	}

	return false, nil
}

func (d *Driver) ClickBySelectors(ctx context.Context, selectors []string) error {
	return d.e.retry(ctx, func() error {
		el, _, err := d.e.findFirstVisible(ctx, selectors, d.e.cfg.StepTimeout)
		if err != nil {
			return err
		}
		if el == nil {
			return fmt.Errorf("element not found for selectors: %v", selectors)
		}
		_ = el.ScrollIntoView()
		el.MustClick()
		return nil
	})
}

func (d *Driver) ScreenshotElement(ctx context.Context, selector string, filePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	d.e.MustHavePage()

	if err := ensureDir(filePath); err != nil {
		return err
	}

	el, _, err := d.e.findFirstVisible(ctx, []string{selector}, d.e.cfg.StepTimeout)
	if err != nil {
		return err
	}
	if el == nil {
		return fmt.Errorf("element not found for selector: %s", selector)
	}

	_ = el.ScrollIntoView()

	// ✅ element screenshot API: (format, quality)
	// quality is used for JPEG; for PNG it’s ignored but still required.
	buf, err := el.Screenshot(proto.PageCaptureScreenshotFormatPng, 100)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, buf, 0o644)
}

// compile-time check (optional)
var _ IDriver = (*Driver)(nil)

// You can expose a small helper if needed elsewhere.
func (d *Driver) SetStepTimeout(t time.Duration) {
	d.e.cfg.StepTimeout = t
}

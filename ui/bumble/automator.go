package bumble

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Action string

const (
	ActionPass       Action = "PASS"
	ActionLike       Action = "LIKE"
	ActionSuperSwipe Action = "SUPERSWIPE"
)

type Automator struct {
	browser     browser
	page        page
	ownsBrowser bool
	cfg         AutomatorConfig
}

type AutomatorConfig struct {
	Headless   bool
	ControlURL string

	// Internal waits/retries only (NOT client flow control)
	StepTimeout   time.Duration
	RetryAttempts int
	RetryDelay    time.Duration

	Selectors Selectors
}

type Selectors struct {
	// Album navigation
	NextImage []string

	// Action buttons
	Pass       []string
	SuperSwipe []string
	Like       []string

	// Optional primitive: click the picture (configurable; tune if needed)
	Picture []string

	// Optional page-ready hints
	ReadyHints []string
}

func DefaultConfig() AutomatorConfig {
	return AutomatorConfig{
		Headless:      false,
		StepTimeout:   6 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    250 * time.Millisecond,
		Selectors:     DefaultSelectors(),
	}
}

func DefaultSelectors() Selectors {
	return Selectors{
		NextImage: []string{
			"div.encounters-album__nav-item.encounters-album__nav-item--next[role='button']",
			"#main > div > div.page__layout > main > div.page__content-inner > div > div > span > div:nth-child(1) > article > div.encounters-album__nav > div.encounters-album__nav-item.encounters-album__nav-item--next",
		},
		Pass: []string{
			"div[data-qa-role='encounters-action-dislike'][role='button']",
			"div.encounters-action.encounters-action--dislike[role='button']",
			"#main > div > div.page__layout > main > div.page__content-inner > div > div > span > div.encounters-user__controls > div > div:nth-child(2) > div > div:nth-child(1) > div",
		},
		SuperSwipe: []string{
			"div[data-qa-role='encounters-action-superswipe'][role='button']",
			"div.encounters-action.encounters-action--superswipe[role='button']",
			"#main > div > div.page__layout > main > div.page__content-inner > div > div > span > div.encounters-user__controls > div > div:nth-child(2) > div > div:nth-child(2) > div",
		},
		Like: []string{
			"div[data-qa-role='encounters-action-like'][role='button']",
			"div.encounters-action.encounters-action--like[role='button']",
			"#main > div > div.page__layout > main > div.page__content-inner > div > div > span > div.encounters-user__controls > div > div:nth-child(2) > div > div:nth-child(3) > div",
		},
		// Broad defaults; client can override config if needed.
		Picture: []string{
			"article img",
			"article",
		},
		ReadyHints: []string{
			"div.encounters-user__controls",
			"article",
		},
	}
}

func NewAutomator(cfg AutomatorConfig) (*Automator, error) {
	if cfg.StepTimeout == 0 {
		cfg.StepTimeout = 6 * time.Second
	}
	if cfg.RetryAttempts <= 0 {
		cfg.RetryAttempts = 3
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = 250 * time.Millisecond
	}
	if len(cfg.Selectors.NextImage) == 0 {
		cfg.Selectors = DefaultSelectors()
	}

	var (
		url  string
		owns bool
	)

	if cfg.ControlURL != "" {
		url = cfg.ControlURL
	} else {
		url = launcher.New().Headless(cfg.Headless).MustLaunch()
		owns = true
	}

	b := rod.New().ControlURL(url).MustConnect()
	return &Automator{
		browser:     rodBrowser{inner: b},
		ownsBrowser: owns,
		cfg:         cfg,
	}, nil
}

func (a *Automator) Close() {
	if a.browser != nil && a.ownsBrowser {
		a.browser.MustClose()
	}
}

// Open a URL in the browser (client handles login/cookies as desired)
func (a *Automator) Open(ctx context.Context, url string) error {
	p := a.browser.MustPage(url)
	a.page = p
	_ = a.page.MustWaitLoad()
	return a.waitForAnyVisible(ctx, a.cfg.Selectors.ReadyHints)
}

// Click a picture (single step). Optional primitive.
func (a *Automator) ClickPicture(ctx context.Context) error {
	a.ensurePage()
	return a.clickAny(ctx, a.cfg.Selectors.Picture)
}

// Move to next picture (single step).
func (a *Automator) NextPicture(ctx context.Context) error {
	a.ensurePage()
	return a.clickAny(ctx, a.cfg.Selectors.NextImage)
}

// Screenshot the FULL current viewport (page) into filePath (PNG).
// Client controls sleeps / repetitions.
func (a *Automator) ScreenshotPage(ctx context.Context, filePath string) error {
	a.ensurePage()

	if err := ensureDir(filePath); err != nil {
		return err
	}

	buf, err := a.page.Screenshot(false, &proto.PageCaptureScreenshot{
		Format:      proto.PageCaptureScreenshotFormatPng,
		FromSurface: true,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, buf, 0o644)
}

// Click action button (PASS/LIKE/SUPERSWIPE)
func (a *Automator) ClickAction(ctx context.Context, action Action) error {
	a.ensurePage()

	var selectors []string
	switch action {
	case ActionPass:
		selectors = a.cfg.Selectors.Pass
	case ActionLike:
		selectors = a.cfg.Selectors.Like
	case ActionSuperSwipe:
		selectors = a.cfg.Selectors.SuperSwipe
	default:
		return fmt.Errorf("unsupported action %q", action)
	}

	return a.clickAny(ctx, selectors)
}

// -------------------------
// Internals
// -------------------------

func (a *Automator) ensurePage() {
	if a.page == nil {
		panic("bumble.Automator: page not initialized (call Open first)")
	}
}

func (a *Automator) clickAny(ctx context.Context, selectors []string) error {
	return a.retry(ctx, func() error {
		el, _, err := a.findFirstVisible(ctx, selectors, a.cfg.StepTimeout)
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

func (a *Automator) waitForAnyVisible(ctx context.Context, selectors []string) error {
	_, _, err := a.findFirstVisible(ctx, selectors, a.cfg.StepTimeout)
	return err
}

func (a *Automator) findFirstVisible(ctx context.Context, selectors []string, timeout time.Duration) (element, string, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nil, "", err
		}

		for _, sel := range selectors {
			el, _ := a.page.Timeout(600 * time.Millisecond).Element(sel)
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

func (a *Automator) retry(ctx context.Context, fn func() error) error {
	var last error
	for i := 0; i < a.cfg.RetryAttempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(); err == nil {
			return nil
		}
		last = fn() // (intentionally avoid sleep if fn is cheap? No—fix below)
		_ = sleepCtx(ctx, a.cfg.RetryDelay)
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

func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

// -------------------------
// Rod wrappers (testability)
// -------------------------

type browser interface {
	MustPage(url string) page
	MustClose()
}

type page interface {
	MustWaitLoad() page
	Timeout(d time.Duration) page
	Element(selector string) (element, error)
	Screenshot(fullPage bool, opt *proto.PageCaptureScreenshot) ([]byte, error)
	MustClose()
}

type element interface {
	MustClick()
	ScrollIntoView() error
	EvalBool(js string) (bool, error)
}

type rodBrowser struct{ inner *rod.Browser }

func (b rodBrowser) MustPage(url string) page { return rodPage{inner: b.inner.MustPage(url)} }
func (b rodBrowser) MustClose()               { b.inner.MustClose() }

type rodPage struct{ inner *rod.Page }

func (p rodPage) MustWaitLoad() page {
	p.inner.MustWaitLoad()
	return p
}

func (p rodPage) Timeout(d time.Duration) page { return rodPage{inner: p.inner.Timeout(d)} }

func (p rodPage) Element(selector string) (element, error) {
	el, err := p.inner.Element(selector)
	if err != nil || el == nil {
		return nil, err
	}
	return rodElement{inner: el}, nil
}

func (p rodPage) Screenshot(fullPage bool, opt *proto.PageCaptureScreenshot) ([]byte, error) {
	return p.inner.Screenshot(fullPage, opt)
}

func (p rodPage) MustClose() { _ = p.inner.Close() }

type rodElement struct{ inner *rod.Element }

func (e rodElement) MustClick()            { e.inner.MustClick() }
func (e rodElement) ScrollIntoView() error { return e.inner.ScrollIntoView() }

func (e rodElement) EvalBool(js string) (bool, error) {
	obj, err := e.inner.Eval(js)
	if err != nil {
		return false, err
	}
	if obj == nil {
		return false, nil
	}
	if obj.UnserializableValue != "" {
		return obj.UnserializableValue == "true", nil
	}
	return obj.Value.Bool(), nil
}

package engine

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Config struct {
	Headless   bool
	ControlURL string

	StepTimeout   time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
}

func DefaultConfig() Config {
	return Config{
		Headless:      false,
		StepTimeout:   6 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    250 * time.Millisecond,
	}
}

type Engine struct {
	browser     Browser
	page        Page
	ownsBrowser bool
	cfg         Config
}

func New(cfg Config) (*Engine, error) {
	if cfg.StepTimeout == 0 {
		cfg.StepTimeout = 6 * time.Second
	}
	if cfg.RetryAttempts <= 0 {
		cfg.RetryAttempts = 3
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = 250 * time.Millisecond
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
	return &Engine{
		browser:     RodBrowser{Inner: b},
		ownsBrowser: owns,
		cfg:         cfg,
	}, nil
}

func (e *Engine) Close() {
	if e.browser != nil && e.ownsBrowser {
		e.browser.MustClose()
	}
}

func (e *Engine) MustHavePage() {
	if e.page == nil {
		panic("engine: page not initialized (call Open first)")
	}
}

func (e *Engine) open(url string) {
	p := e.browser.MustPage(url)
	e.page = p
	_ = e.page.MustWaitLoad()
}

func (e *Engine) screenshot(filePath string) error {
	e.MustHavePage()
	if err := ensureDir(filePath); err != nil {
		return err
	}
	buf, err := e.page.Screenshot(false, &proto.PageCaptureScreenshot{
		Format:      proto.PageCaptureScreenshotFormatPng,
		FromSurface: true,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, buf, 0o644)
}

func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
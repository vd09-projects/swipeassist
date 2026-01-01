package bumble

import (
	"context"
	"fmt"
)

// ClientConfig describes how to connect to Bumble via the Automator.
type ClientConfig struct {
	LoginURL   string
	Headless   bool
	ControlURL string
}

// Client wraps Automator with a higher level workflow to test automation
// against a real Bumble URL.
type Client struct {
	cfg       ClientConfig
	automator *Automator
}

// NewClient constructs a Client backed by Automator.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.LoginURL == "" {
		return nil, fmt.Errorf("login URL is required")
	}

	// Use defaults (selectors/timeouts/retries) and only override connection bits.
	autoCfg := DefaultConfig()
	autoCfg.Headless = cfg.Headless
	autoCfg.ControlURL = cfg.ControlURL

	auto, err := NewAutomator(autoCfg)
	if err != nil {
		return nil, err
	}

	return &Client{cfg: cfg, automator: auto}, nil
}

func (c *Client) Close() {
	if c != nil && c.automator != nil {
		c.automator.Close()
	}
}

// Open navigates to the LoginURL.
func (c *Client) Open(ctx context.Context) error {
	return c.automator.Open(ctx, c.cfg.LoginURL)
}

// NextPicture delegates to automator (single step).
func (c *Client) NextPicture(ctx context.Context) error {
	return c.automator.NextPicture(ctx)
}

// ScreenshotPage delegates to automator (single step).
func (c *Client) ScreenshotPage(ctx context.Context, filePath string) error {
	return c.automator.ScreenshotPage(ctx, filePath)
}

// ClickAction delegates to automator (single step).
func (c *Client) ClickAction(ctx context.Context, action Action) error {
	return c.automator.ClickAction(ctx, action)
}
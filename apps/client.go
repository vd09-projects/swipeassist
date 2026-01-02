package apps

import (
	"context"
	"fmt"

	"github.com/vd09-projects/swipeassist/apps/engine"
	"github.com/vd09-projects/swipeassist/domain"
)

type Config struct {
	AppName    domain.AppName // "bumble", "tinder", ...
	EntryURL   string         // optional override; else adapter.DefaultEntryURL()
	Headless   bool
	ControlURL string
}

type GenericClient struct {
	cfg     Config
	adapter Adapter
	driver  engine.IDriver
}

func New(cfg Config) (*GenericClient, error) {
	ad, ok := GetAdapterRegistry()[cfg.AppName]
	if !ok {
		return nil, fmt.Errorf("unknown app %q", cfg.AppName)
	}

	ec := engine.DefaultConfig()
	ec.Headless = cfg.Headless
	ec.ControlURL = cfg.ControlURL

	eng, err := engine.New(ec)
	if err != nil {
		return nil, err
	}

	drv := engine.NewDriver(eng)

	return &GenericClient{
		cfg:     cfg,
		adapter: ad,
		driver:  drv,
	}, nil
}

func (c *GenericClient) Close() { c.driver.Close() }

func (c *GenericClient) Open(ctx context.Context) error {
	url := c.cfg.EntryURL
	if url == "" {
		url = c.adapter.DefaultEntryURL()
	}
	if err := c.driver.Open(ctx, url); err != nil {
		return err
	}
	return c.adapter.WaitReady(ctx, c.driver)
}

func (c *GenericClient) NextMedia(ctx context.Context) error {
	return c.adapter.NextMedia(ctx, c.driver)
}

func (c *GenericClient) Screenshot(ctx context.Context, filePath string) error {
	// return c.driver.Screenshot(ctx, filePath)
	return c.adapter.ScreenshotMedia(ctx, c.driver, filePath)
}

func (c *GenericClient) Act(ctx context.Context, action domain.AppAction) error {
	return c.adapter.Act(ctx, c.driver, action)
}

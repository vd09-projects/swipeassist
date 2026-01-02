package engine

import "context"

type IDriver interface {
	Open(ctx context.Context, url string) error
	Screenshot(ctx context.Context, filePath string) error
	ScreenshotElement(ctx context.Context, selector string, filePath string) error

	WaitAnyVisible(ctx context.Context, selectors []string) error
	// IsVisible returns true if any selector matches a visible element right now (no retries).
	IsVisible(ctx context.Context, selectors []string) (bool, error)
	ClickBySelectors(ctx context.Context, selectors []string) error

	Close()
}

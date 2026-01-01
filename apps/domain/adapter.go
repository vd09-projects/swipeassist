package domain

import (
	"context"

	"github.com/vd09-projects/swipeassist/engine"
)

type Adapter interface {
	Name() string
	DefaultEntryURL() string

	WaitReady(ctx context.Context, d engine.IDriver) error

	NextMedia(ctx context.Context, d engine.IDriver) error
	Act(ctx context.Context, d engine.IDriver, action Action) error

	ScreenshotMedia(ctx context.Context, d engine.IDriver, filePath string) error
}

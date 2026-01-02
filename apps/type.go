package apps

import (
	"context"

	"github.com/vd09-projects/swipeassist/domain"
	"github.com/vd09-projects/swipeassist/engine"
)

type Adapter interface {
	Name() string
	DefaultEntryURL() string

	WaitReady(ctx context.Context, d engine.IDriver) error

	NextMedia(ctx context.Context, d engine.IDriver) error
	Act(ctx context.Context, d engine.IDriver, action domain.AppAction) error

	ScreenshotMedia(ctx context.Context, d engine.IDriver, filePath string) error
}

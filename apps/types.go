package apps

import (
	"context"

	"github.com/vd09-projects/swipeassist/apps/engine"
	"github.com/vd09-projects/swipeassist/domain"
)

type Adapter interface {
	Name() string
	DefaultEntryURL() string

	WaitReady(ctx context.Context, d engine.IDriver) error

	GetProfileId(ctx context.Context) string
	NextMedia(ctx context.Context, d engine.IDriver) error
	Act(ctx context.Context, d engine.IDriver, action domain.AppAction) error

	ScreenshotMedia(ctx context.Context, d engine.IDriver, filePath string) error
}

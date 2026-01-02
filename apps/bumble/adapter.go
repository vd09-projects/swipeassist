package bumble

import (
	"context"
	"fmt"

	"github.com/vd09-projects/swipeassist/apps/engine"
	"github.com/vd09-projects/swipeassist/domain"
)

type Adapter struct {
	S Selectors
}

func NewAdapterFromDefaults() Adapter {
	return Adapter{S: DefaultSelectors()}
}

func (a Adapter) Name() string { return "bumble" }

func (a Adapter) DefaultEntryURL() string { return "https://bumble.com/app" }

func (a Adapter) WaitReady(ctx context.Context, d engine.IDriver) error {
	return d.WaitAnyVisible(ctx, a.S.ReadyHints)
}

func (a Adapter) NextMedia(ctx context.Context, d engine.IDriver) error {
	disabled, err := d.IsVisible(ctx, a.S.NextImageDisabled)
	if err != nil {
		return err
	}
	if disabled {
		return fmt.Errorf("next media navigation is disabled")
	}
	return d.ClickBySelectors(ctx, a.S.NextImage)
}

func (a Adapter) Act(ctx context.Context, d engine.IDriver, action domain.AppAction) error {
	switch action.Kind {
	case domain.AppActionPass:
		return d.ClickBySelectors(ctx, a.S.Pass)
	case domain.AppActionLike:
		return d.ClickBySelectors(ctx, a.S.Like)
	case domain.AppActionSuperSwipe:
		return d.ClickBySelectors(ctx, a.S.SuperSwipe)
	default:
		// ignore unknown actions; or return error if you prefer
		return nil
	}
}

func (a Adapter) ScreenshotMedia(
	ctx context.Context,
	d engine.IDriver,
	filePath string,
) error {
	return d.ScreenshotElement(ctx, a.S.AlbumNav, filePath)
}

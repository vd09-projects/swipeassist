package bumble

import (
	"context"

	"github.com/vd09-projects/swipeassist/apps/domain"
	"github.com/vd09-projects/swipeassist/engine"
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
	return d.ClickBySelectors(ctx, a.S.NextImage)
}

func (a Adapter) Act(ctx context.Context, d engine.IDriver, action domain.Action) error {
	switch action.AType {
	case domain.ActionPass:
		return d.ClickBySelectors(ctx, a.S.Pass)
	case domain.ActionLike:
		return d.ClickBySelectors(ctx, a.S.Like)
	case domain.ActionSuperSwipe:
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

// compile-time check (optional)
var _ domain.Adapter = Adapter{}

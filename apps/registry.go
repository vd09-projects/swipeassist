package apps

import (
	"github.com/vd09-projects/swipeassist/apps/bumble"
	"github.com/vd09-projects/swipeassist/apps/domain"
)

func GetAdapterRegistry() map[domain.AppName]domain.Adapter {
	return map[domain.AppName]domain.Adapter{
		domain.Bumble: bumble.NewAdapterFromDefaults(),
	}
}

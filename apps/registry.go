package apps

import (
	"github.com/vd09-projects/swipeassist/apps/bumble"
	"github.com/vd09-projects/swipeassist/domain"
)

func GetAdapterRegistry() map[domain.AppName]Adapter {
	return map[domain.AppName]Adapter{
		domain.Bumble: bumble.NewAdapterFromDefaults(),
	}
}

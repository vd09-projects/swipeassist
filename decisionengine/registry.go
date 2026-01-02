package decisionengine

import (
	"fmt"

	"github.com/vd09-projects/swipeassist/domain"
)

type Registry struct {
	byApp map[domain.AppName]Policy
}

func NewRegistry() *Registry {
	return &Registry{byApp: make(map[domain.AppName]Policy)}
}

func (r *Registry) Register(app domain.AppName, p Policy) {
	r.byApp[app] = p
}

func (r *Registry) Resolve(app domain.AppName) (Policy, error) {
	p, ok := r.byApp[app]
	if !ok {
		return nil, fmt.Errorf("no decision policy registered for app=%s", app)
	}
	return p, nil
}

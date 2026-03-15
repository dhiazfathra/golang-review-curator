package adapters

import "review-curator/pkg/module/scraper"

type Registry struct {
	adapters map[scraper.Platform]Adapter
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[scraper.Platform]Adapter)}
}

func (r *Registry) Register(adapter Adapter) {
	r.adapters[adapter.Platform()] = adapter
}

func (r *Registry) Get(p scraper.Platform) (Adapter, bool) {
	adapter, ok := r.adapters[p]
	return adapter, ok
}

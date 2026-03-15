package adapters

import (
	"review-curator/pkg/module/scraper"
	"review-curator/pkg/platform/browser"
	"review-curator/pkg/platform/captcha"
	"review-curator/pkg/platform/proxy"
	"review-curator/pkg/platform/ratelimit"
	"review-curator/pkg/platform/selector"
)

type Registry struct {
	adapters map[scraper.Platform]Adapter
}

func NewRegistry(
	pool *browser.Pool,
	rotator *proxy.Rotator,
	captchaResolver captcha.CaptchaResolver,
	selectorStore *selector.SelectorStore,
	limiter *ratelimit.Limiter,
) *Registry {
	base := BaseScraper{
		BrowserPool: pool,
		Rotator:     rotator,
		Captcha:     captchaResolver,
		Selectors:   selectorStore,
		Limiter:     limiter,
	}
	return &Registry{
		adapters: map[scraper.Platform]Adapter{
			scraper.PlatformShopee:    &ShopeeAdapter{BaseScraper: base},
			scraper.PlatformTokopedia: &TokopediaAdapter{BaseScraper: base},
			scraper.PlatformBlibli:    &BlibliAdapter{BaseScraper: base},
		},
	}
}

func (r *Registry) Get(p scraper.Platform) (Adapter, bool) {
	a, ok := r.adapters[p]
	return a, ok
}

func (r *Registry) Register(adapter Adapter) {
	r.adapters[adapter.Platform()] = adapter
}

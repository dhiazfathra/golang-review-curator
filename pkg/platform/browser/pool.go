package browser

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

type Pool struct {
	mu       sync.Mutex
	browsers []*rod.Browser
	idx      int
	size     int
}

func NewPool(size int, proxyURL string) (*Pool, error) {
	p := &Pool{size: size, browsers: make([]*rod.Browser, 0, size)}
	for i := 0; i < size; i++ {
		l := launcher.New().
			Headless(true).
			Set("disable-blink-features", "AutomationControlled").
			NoSandbox(true)
		if proxyURL != "" {
			l = l.Proxy(proxyURL)
		}
		u, err := l.Launch()
		if err != nil {
			return nil, fmt.Errorf("browser pool launch[%d]: %w", i, err)
		}
		b := rod.New().ControlURL(u).MustConnect()
		p.browsers = append(p.browsers, b)
	}
	return p, nil
}

func (p *Pool) Acquire(ctx context.Context) (*rod.Page, func(), error) {
	p.mu.Lock()
	b := p.browsers[p.idx%p.size]
	p.idx++
	p.mu.Unlock()

	page, err := stealth.Page(b)
	if err != nil {
		return nil, nil, fmt.Errorf("browser pool stealth page: %w", err)
	}
	page = page.Context(ctx)
	RandomiseFingerprint(page)
	release := func() { _ = page.Close() }
	return page, release, nil
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, b := range p.browsers {
		_ = b.Close()
	}
}

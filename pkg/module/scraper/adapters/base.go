package adapters

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-rod/rod"
	"review-curator/pkg/module/scraper"
	"review-curator/pkg/platform/browser"
	"review-curator/pkg/platform/captcha"
	"review-curator/pkg/platform/proxy"
	"review-curator/pkg/platform/ratelimit"
	"review-curator/pkg/platform/selector"
)

type Adapter interface {
	Platform() scraper.Platform
	FetchReviews(ctx context.Context, job scraper.CrawlJob) ([]scraper.CrawlResult, error)
}

type BaseScraper struct {
	BrowserPool *browser.Pool
	Rotator     *proxy.Rotator
	Captcha     captcha.CaptchaResolver
	Selectors   *selector.SelectorStore
	Limiter     *ratelimit.Limiter
}

func (b *BaseScraper) NavigateWithRetry(ctx context.Context, rawURL string) (*rod.Page, func(), string, error) {
	proxyURL, _ := b.Rotator.Next()

	page, release, err := b.BrowserPool.Acquire(ctx)
	if err != nil {
		b.Rotator.ReportFailure(proxyURL)
		return nil, nil, proxyURL, fmt.Errorf("base: acquire page: %w", err)
	}

	u, _ := url.Parse(rawURL)
	if err := b.Limiter.Wait(ctx, u.Host); err != nil {
		release()
		return nil, nil, proxyURL, fmt.Errorf("base: rate limit: %w", err)
	}

	if err := page.Navigate(rawURL); err != nil {
		release()
		b.Rotator.ReportFailure(proxyURL)
		return nil, nil, proxyURL, fmt.Errorf("base: navigate %s: %w", rawURL, err)
	}
	if err := page.WaitLoad(); err != nil {
		release()
		b.Rotator.ReportFailure(proxyURL)
		return nil, nil, proxyURL, fmt.Errorf("base: wait load: %w", err)
	}

	if err := b.resolveCaptchaIfPresent(ctx, page, rawURL); err != nil {
		release()
		return nil, nil, proxyURL, fmt.Errorf("base: captcha: %w", err)
	}

	return page, release, proxyURL, nil
}

func (b *BaseScraper) resolveCaptchaIfPresent(ctx context.Context, page *rod.Page, pageURL string) error {
	ctype := b.detectCaptchaType(page)
	switch ctype {
	case "recaptcha_v2":
		siteKey := b.extractSiteKey(page)
		token, err := b.Captcha.SolveRecaptchaV2(ctx, siteKey, pageURL)
		if err != nil {
			return fmt.Errorf("solve recaptcha v2: %w", err)
		}
		return b.injectRecaptchaToken(page, token)
	case "image":
		img := b.extractCaptchaImage(page)
		token, err := b.Captcha.SolveImage(ctx, img)
		if err != nil {
			return fmt.Errorf("solve image captcha: %w", err)
		}
		return b.submitCaptchaAnswer(page, token)
	case "cloudflare":
		_ = page.WaitIdle(5e9)
	}
	return nil
}

func (b *BaseScraper) detectCaptchaType(page *rod.Page) string {
	content, err := page.Element("html")
	if err != nil {
		return ""
	}
	html, err := content.HTML()
	if err != nil {
		return ""
	}
	switch {
	case strings.Contains(html, "g-recaptcha") || strings.Contains(html, "recaptcha/api.js"):
		return "recaptcha_v2"
	case strings.Contains(html, "cf-challenge") || strings.Contains(html, "cf_clearance"):
		return "cloudflare"
	case strings.Contains(html, "captcha-image") || strings.Contains(html, "image_captcha"):
		return "image"
	}
	return ""
}

func (b *BaseScraper) extractSiteKey(page *rod.Page) string {
	el, err := page.Element("[data-sitekey]")
	if err != nil {
		return ""
	}
	key, _ := el.Attribute("data-sitekey")
	if key == nil {
		return ""
	}
	return *key
}

func (b *BaseScraper) extractCaptchaImage(page *rod.Page) string {
	el, err := page.Element("img.captcha-image, img[id*=captcha]")
	if err != nil {
		return ""
	}
	src, _ := el.Attribute("src")
	if src == nil {
		return ""
	}
	return *src
}

func (b *BaseScraper) injectRecaptchaToken(page *rod.Page, token string) error {
	_, err := page.Eval(fmt.Sprintf(
		`() => { document.getElementById('g-recaptcha-response').innerHTML = '%s'; }`, token,
	))
	return err
}

func (b *BaseScraper) submitCaptchaAnswer(page *rod.Page, answer string) error {
	el, err := page.Element("input[name=captcha], input[id*=captcha]")
	if err != nil {
		return err
	}
	return el.Input(answer)
}

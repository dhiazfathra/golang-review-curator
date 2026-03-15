package captcha

import (
	"context"
	"fmt"
)

// CaptchaResolver defines methods for solving various types of captchas.
type CaptchaResolver interface {
	SolveImage(ctx context.Context, base64img string) (string, error)
	SolveRecaptchaV2(ctx context.Context, siteKey, pageURL string) (string, error)
	SolveRecaptchaV3(ctx context.Context, siteKey, pageURL, action string) (string, error)
}

// Dispatcher tries a primary captcha resolver, falling back to a secondary on failure.
type Dispatcher struct {
	primary   CaptchaResolver
	secondary CaptchaResolver
}

// NewDispatcher creates a new captcha dispatcher with primary and secondary resolvers.
func NewDispatcher(primary, secondary CaptchaResolver) *Dispatcher {
	return &Dispatcher{primary: primary, secondary: secondary}
}

// SolveImage attempts to solve an image captcha using primary, falling back to secondary.
func (d *Dispatcher) SolveImage(ctx context.Context, base64img string) (string, error) {
	token, err := d.primary.SolveImage(ctx, base64img)
	if err == nil {
		return token, nil
	}
	token, err2 := d.secondary.SolveImage(ctx, base64img)
	if err2 != nil {
		return "", fmt.Errorf("captcha: both providers failed: primary=%w secondary=%v", err, err2)
	}
	return token, nil
}

// SolveRecaptchaV2 attempts to solve a reCAPTCHA v2 challenge.
func (d *Dispatcher) SolveRecaptchaV2(ctx context.Context, siteKey, pageURL string) (string, error) {
	token, err := d.primary.SolveRecaptchaV2(ctx, siteKey, pageURL)
	if err == nil {
		return token, nil
	}
	token, err2 := d.secondary.SolveRecaptchaV2(ctx, siteKey, pageURL)
	if err2 != nil {
		return "", fmt.Errorf("captcha: both providers failed: primary=%w secondary=%v", err, err2)
	}
	return token, nil
}

// SolveRecaptchaV3 attempts to solve a reCAPTCHA v3 challenge.
func (d *Dispatcher) SolveRecaptchaV3(ctx context.Context, siteKey, pageURL, action string) (string, error) {
	token, err := d.primary.SolveRecaptchaV3(ctx, siteKey, pageURL, action)
	if err == nil {
		return token, nil
	}
	token, err2 := d.secondary.SolveRecaptchaV3(ctx, siteKey, pageURL, action)
	if err2 != nil {
		return "", fmt.Errorf("captcha: both providers failed: primary=%w secondary=%v", err, err2)
	}
	return token, nil
}

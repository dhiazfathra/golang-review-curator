package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const twoCaptchaBase = "https://2captcha.com"

type TwoCaptcha struct {
	apiKey string
	client *http.Client
}

func NewTwoCaptcha(apiKey string) *TwoCaptcha {
	return &TwoCaptcha{apiKey: apiKey, client: &http.Client{Timeout: 120 * time.Second}}
}

func (t *TwoCaptcha) SolveImage(ctx context.Context, base64img string) (string, error) {
	taskID, err := t.submitTask(ctx, url.Values{
		"key":    {t.apiKey},
		"method": {"base64"},
		"body":   {base64img},
		"json":   {"1"},
	})
	if err != nil {
		return "", err
	}
	return t.pollResult(ctx, taskID)
}

func (t *TwoCaptcha) SolveRecaptchaV2(ctx context.Context, siteKey, pageURL string) (string, error) {
	taskID, err := t.submitTask(ctx, url.Values{
		"key":       {t.apiKey},
		"method":    {"userrecaptcha"},
		"googlekey": {siteKey},
		"pageurl":   {pageURL},
		"json":      {"1"},
	})
	if err != nil {
		return "", err
	}
	return t.pollResult(ctx, taskID)
}

func (t *TwoCaptcha) SolveRecaptchaV3(ctx context.Context, siteKey, pageURL, action string) (string, error) {
	taskID, err := t.submitTask(ctx, url.Values{
		"key":       {t.apiKey},
		"method":    {"userrecaptcha"},
		"googlekey": {siteKey},
		"pageurl":   {pageURL},
		"version":   {"v3"},
		"action":    {action},
		"json":      {"1"},
	})
	if err != nil {
		return "", err
	}
	return t.pollResult(ctx, taskID)
}

func (t *TwoCaptcha) submitTask(ctx context.Context, params url.Values) (string, error) {
	resp, err := t.client.PostForm(twoCaptchaBase+"/in.php", params)
	if err != nil {
		return "", fmt.Errorf("2captcha submit: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	var result struct {
		Status  int    `json:"status"`
		Request string `json:"request"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("2captcha submit decode: %w", err)
	}
	if result.Status != 1 {
		return "", fmt.Errorf("2captcha submit error: %s", result.Request)
	}
	return result.Request, nil
}

func (t *TwoCaptcha) pollResult(ctx context.Context, taskID string) (string, error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			resp, err := t.client.Get(fmt.Sprintf(
				"%s/res.php?key=%s&action=get&id=%s&json=1",
				twoCaptchaBase, t.apiKey, taskID,
			))
			if err != nil {
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close() //nolint:errcheck
			var result struct {
				Status  int    `json:"status"`
				Request string `json:"request"`
			}
			if err := json.Unmarshal(body, &result); err != nil {
				continue
			}
			if result.Status == 1 {
				return result.Request, nil
			}
			if result.Request != "CAPCHA_NOT_READY" {
				return "", fmt.Errorf("2captcha poll error: %s", result.Request)
			}
		}
	}
}

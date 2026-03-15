package captcha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const antiCaptchaBase = "https://api.anti-captcha.com"

type AntiCaptcha struct {
	apiKey string
	client *http.Client
}

func NewAntiCaptcha(apiKey string) *AntiCaptcha {
	return &AntiCaptcha{apiKey: apiKey, client: &http.Client{Timeout: 120 * time.Second}}
}

func (a *AntiCaptcha) SolveImage(ctx context.Context, base64img string) (string, error) {
	return a.solve(ctx, map[string]any{
		"clientKey": a.apiKey,
		"task":      map[string]any{"type": "ImageToTextTask", "body": base64img},
	})
}

func (a *AntiCaptcha) SolveRecaptchaV2(ctx context.Context, siteKey, pageURL string) (string, error) {
	return a.solve(ctx, map[string]any{
		"clientKey": a.apiKey,
		"task": map[string]any{
			"type":       "NoCaptchaTaskProxyless",
			"websiteURL": pageURL,
			"websiteKey": siteKey,
		},
	})
}

func (a *AntiCaptcha) SolveRecaptchaV3(ctx context.Context, siteKey, pageURL, action string) (string, error) {
	return a.solve(ctx, map[string]any{
		"clientKey": a.apiKey,
		"task": map[string]any{
			"type":       "RecaptchaV3TaskProxyless",
			"websiteURL": pageURL,
			"websiteKey": siteKey,
			"pageAction": action,
		},
	})
}

func (a *AntiCaptcha) solve(ctx context.Context, payload map[string]any) (string, error) {
	body, _ := json.Marshal(payload)
	resp, err := a.client.Post(antiCaptchaBase+"/createTask", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("anticaptcha create: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	var createResp struct {
		ErrorID int `json:"errorId"`
		TaskID  int `json:"taskId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return "", fmt.Errorf("anticaptcha create decode: %w", err)
	}
	if createResp.ErrorID != 0 {
		return "", fmt.Errorf("anticaptcha create error: %d", createResp.ErrorID)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			pollBody, _ := json.Marshal(map[string]any{
				"clientKey": a.apiKey,
				"taskId":    createResp.TaskID,
			})
			pr, err := a.client.Post(antiCaptchaBase+"/getTaskResult", "application/json", bytes.NewReader(pollBody))
			if err != nil {
				continue
			}
			var pollResp struct {
				Status   string `json:"status"`
				Solution struct {
					GRecaptchaResponse string `json:"gRecaptchaResponse"`
					Text               string `json:"text"`
				} `json:"solution"`
			}
			_ = json.NewDecoder(pr.Body).Decode(&pollResp)
			pr.Body.Close() //nolint:errcheck
			if pollResp.Status == "ready" {
				if pollResp.Solution.GRecaptchaResponse != "" {
					return pollResp.Solution.GRecaptchaResponse, nil
				}
				return pollResp.Solution.Text, nil
			}
		}
	}
}

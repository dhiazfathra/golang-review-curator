package browser

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
)

func InjectStealth(page *rod.Page) (*rod.Page, error) {
	return stealth.Page(page.Browser())
}

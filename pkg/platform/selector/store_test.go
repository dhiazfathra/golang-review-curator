package selector_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"review-curator/pkg/platform/selector"
)

type mockExtractor struct {
	results map[string]string
}

func (m *mockExtractor) Element(css string) (string, bool) {
	v, ok := m.results[css]
	return v, ok && v != ""
}
func (m *mockExtractor) XPathElement(xpath string) (string, bool) {
	v, ok := m.results[xpath]
	return v, ok && v != ""
}

func TestExtractField_PrimaryHits(t *testing.T) {
	cfg := selector.SelectorConfig{
		Platform: "shopee",
		Field:    "review_text",
		Rules: []selector.SelectorRule{
			{Type: selector.RuleTypeCSS, Value: ".review"},
		},
	}
	e := &mockExtractor{results: map[string]string{".review": "Great product!"}}
	val, ok := selector.ExtractField(e, cfg)
	assert.True(t, ok)
	assert.Equal(t, "Great product!", val)
}

func TestExtractField_FallsBackToSecondary(t *testing.T) {
	cfg := selector.SelectorConfig{
		Platform: "shopee",
		Field:    "review_text",
		Rules: []selector.SelectorRule{
			{Type: selector.RuleTypeCSS, Value: ".dead"},
			{Type: selector.RuleTypeCSS, Value: ".alive"},
		},
	}
	e := &mockExtractor{results: map[string]string{".alive": "Good!"}}
	val, ok := selector.ExtractField(e, cfg)
	assert.True(t, ok)
	assert.Equal(t, "Good!", val)
}

func TestExtractField_AllFail(t *testing.T) {
	cfg := selector.SelectorConfig{
		Platform: "shopee",
		Field:    "review_text",
		Rules:    []selector.SelectorRule{{Type: selector.RuleTypeCSS, Value: ".dead"}},
	}
	e := &mockExtractor{results: map[string]string{}}
	_, ok := selector.ExtractField(e, cfg)
	assert.False(t, ok)
}

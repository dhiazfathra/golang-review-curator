package selector

// RuleType represents the type of selector rule.
type RuleType string

const (
	// RuleTypeCSS matches elements using CSS selectors.
	RuleTypeCSS RuleType = "css"
	// RuleTypeXPath matches elements using XPath expressions.
	RuleTypeXPath RuleType = "xpath"
	// RuleTypeJSONPath extracts data from JSON using JSONPath.
	RuleTypeJSONPath RuleType = "jsonpath"
	// RuleTypeRegex extracts data using regular expressions.
	RuleTypeRegex RuleType = "regex"
)

// SelectorRule defines a single selector rule with type and value.
type SelectorRule struct {
	Type  RuleType `json:"type"`
	Value string   `json:"value"`
}

// SelectorConfig holds configuration for a selector on a specific platform and field.
type SelectorConfig struct {
	Platform string
	Field    string
	Rules    []SelectorRule
}

// FallbackChain is a list of selector rules to try in order.
type FallbackChain = []SelectorRule

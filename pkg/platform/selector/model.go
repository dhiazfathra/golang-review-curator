package selector

type RuleType string

const (
	RuleTypeCSS      RuleType = "css"
	RuleTypeXPath    RuleType = "xpath"
	RuleTypeJSONPath RuleType = "jsonpath"
	RuleTypeRegex    RuleType = "regex"
)

type SelectorRule struct {
	Type  RuleType `json:"type"`
	Value string   `json:"value"`
}

type SelectorConfig struct {
	Platform string
	Field    string
	Rules    []SelectorRule
}

type FallbackChain = []SelectorRule

package sarif

type ruleDescriptor struct {
	ID                   string        `json:"id"`
	DefaultConfiguration configuration `json:"defaultConfiguration"`
}

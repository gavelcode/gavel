package sarif

type driver struct {
	Name  string           `json:"name"`
	Rules []ruleDescriptor `json:"rules"`
}

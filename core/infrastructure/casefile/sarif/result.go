package sarif

type result struct {
	RuleID              string            `json:"ruleId"`
	Level               string            `json:"level"`
	Message             message           `json:"message"`
	Locations           []location        `json:"locations"`
	Fingerprints        map[string]string `json:"fingerprints"`
	PartialFingerprints map[string]string `json:"partialFingerprints"`
}

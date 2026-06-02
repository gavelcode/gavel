package gateview

type GateResult struct {
	Passed     bool
	Conditions []GateCondition
}

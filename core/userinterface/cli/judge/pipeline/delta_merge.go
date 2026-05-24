package pipeline

func MergeDelta(findings, arch Delta) Delta {
	findings.NewViolationsCount = arch.NewViolationsCount
	findings.FixedViolationsCount = arch.FixedViolationsCount
	findings.ExistingViolationsCount = arch.ExistingViolationsCount
	findings.NewViolationIDs = arch.NewViolationIDs
	findings.HasArchPrevious = arch.HasArchPrevious
	return findings
}

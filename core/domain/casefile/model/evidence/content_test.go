package evidence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/architecture"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/license"
)

var (
	_ evidence.Content = finding.Content{}
	_ evidence.Content = coverage.Content{}
	_ evidence.Content = license.Content{}
	_ evidence.Content = architecture.Content{}
	_ evidence.Content = coverage.PatchContent{}
)

func validFinding(t *testing.T, tool, ruleID string) finding.Finding {
	t.Helper()
	fp, err := finding.NewFingerprintID(tool + "-" + ruleID)
	require.NoError(t, err)
	f, err := finding.NewFinding(tool, ruleID, finding.SeverityError, "src/Main.java", 10, "msg", fp)
	require.NoError(t, err)
	return f
}

func validLanguageCoverage(t *testing.T, lang string, total, covered int) coverage.LanguageStats {
	t.Helper()
	l, err := coverage.NewLanguage(lang)
	require.NoError(t, err)
	langCov, err := coverage.NewLanguageStats(l, total, covered)
	require.NoError(t, err)
	return langCov
}

func validDependencyLicense(t *testing.T, name, version, licenseSPDX string) license.Dependency {
	t.Helper()
	depLic, err := license.NewDependency(name, version, licenseSPDX)
	require.NoError(t, err)
	return depLic
}

func TestNewFindingsContent(t *testing.T) {
	find1 := validFinding(t, "pmd", "UnusedVar")
	find2 := validFinding(t, "spotbugs", "NullDeref")

	tests := []struct {
		name     string
		subtype  evidence.Subtype
		findings []finding.Finding
		wantErr  bool
	}{
		{
			name:     "shouldCreateWithCodeQualitySubtype",
			subtype:  evidence.SubtypeCodeQuality,
			findings: []finding.Finding{find1, find2},
		},
		{
			name:     "shouldCreateWithSASTSubtype",
			subtype:  evidence.SubtypeSAST,
			findings: []finding.Finding{find1},
		},
		{
			name:     "shouldCreateWithEmptyFindings",
			subtype:  evidence.SubtypeCodeQuality,
			findings: []finding.Finding{},
		},
		{
			name:     "shouldCreateWithNilFindings",
			subtype:  evidence.SubtypeCodeQuality,
			findings: nil,
		},
		{
			name:     "shouldRejectCoverageSubtype",
			subtype:  evidence.SubtypeCoverage,
			findings: []finding.Finding{find1},
			wantErr:  true,
		},
		{
			name:     "shouldRejectLicenseSubtype",
			subtype:  evidence.SubtypeLicense,
			findings: []finding.Finding{find1},
			wantErr:  true,
		},
		{
			name:     "shouldRejectArchitectureSubtype",
			subtype:  evidence.SubtypeArchitecture,
			findings: []finding.Finding{find1},
			wantErr:  true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			findCont, err := finding.NewContent(tcase.subtype, tcase.findings)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, finding.ErrInvalidContent)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tcase.subtype, findCont.Subtype())
			assert.Equal(t, tcase.subtype.Type(), findCont.Type())
			assert.Equal(t, len(tcase.findings), len(findCont.Findings()))
		})
	}
}

func TestFindingsContentType(t *testing.T) {
	f := validFinding(t, "pmd", "Rule1")
	findCont, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f})
	require.NoError(t, err)

	assert.Equal(t, evidence.TypeSourceCode, findCont.Type())
}

func TestFindingsContentSubtype(t *testing.T) {
	f := validFinding(t, "pmd", "Rule1")
	findCont, err := finding.NewContent(evidence.SubtypeComplexity, []finding.Finding{f})
	require.NoError(t, err)

	assert.Equal(t, evidence.SubtypeComplexity, findCont.Subtype())
}

func TestFindingsContentDefensiveCopy(t *testing.T) {
	find1 := validFinding(t, "pmd", "Rule1")
	find2 := validFinding(t, "pmd", "Rule2")
	original := []finding.Finding{find1}

	findCont, err := finding.NewContent(evidence.SubtypeCodeQuality, original)
	require.NoError(t, err)

	original[0] = find2
	assert.NotEqual(t, find2, findCont.Findings()[0])

	returned := findCont.Findings()
	returned[0] = find2
	assert.NotEqual(t, find2, findCont.Findings()[0])
}

func TestFindingsContentMerge(t *testing.T) {
	find1 := validFinding(t, "pmd", "Rule1")
	find2 := validFinding(t, "spotbugs", "Rule2")
	f3 := validFinding(t, "errorprone", "Rule3")

	fc1, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{find1, find2})
	require.NoError(t, err)
	fc2, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f3})
	require.NoError(t, err)

	merged, err := fc1.Merge(fc2)
	require.NoError(t, err)
	mergedFC, ok := merged.(finding.Content)
	require.True(t, ok)

	assert.Equal(t, 3, len(mergedFC.Findings()))
	assert.Equal(t, evidence.SubtypeCodeQuality, mergedFC.Subtype())
}

func TestFindingsContentMergeSubtypeMismatchRejected(t *testing.T) {
	find1 := validFinding(t, "pmd", "Rule1")
	find2 := validFinding(t, "semgrep", "Rule2")

	fc1, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{find1})
	require.NoError(t, err)
	fc2, err := finding.NewContent(evidence.SubtypeSAST, []finding.Finding{find2})
	require.NoError(t, err)

	_, err = fc1.Merge(fc2)
	require.Error(t, err)
	assert.ErrorIs(t, err, finding.ErrInvalidContent)
}

func TestNewCoverageContent(t *testing.T) {
	langCov := validLanguageCoverage(t, "go", 100, 80)

	tests := []struct {
		name         string
		totalLines   int
		coveredLines int
		byLanguage   []coverage.LanguageStats
		wantErr      bool
	}{
		{
			name:         "shouldCreateValidCoverageContent",
			totalLines:   100,
			coveredLines: 80,
			byLanguage:   []coverage.LanguageStats{langCov},
		},
		{
			name:         "shouldCreateWithZeroLines",
			totalLines:   0,
			coveredLines: 0,
			byLanguage:   nil,
		},
		{
			name:         "shouldRejectNegativeTotalLines",
			totalLines:   -1,
			coveredLines: 0,
			byLanguage:   nil,
			wantErr:      true,
		},
		{
			name:         "shouldRejectNegativeCoveredLines",
			totalLines:   100,
			coveredLines: -1,
			byLanguage:   nil,
			wantErr:      true,
		},
		{
			name:         "shouldRejectCoveredLinesGreaterThanTotalLines",
			totalLines:   50,
			coveredLines: 51,
			byLanguage:   nil,
			wantErr:      true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			codeCov, err := coverage.NewContent(tcase.totalLines, tcase.coveredLines, tcase.byLanguage)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, coverage.ErrInvalidContent)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tcase.totalLines, codeCov.TotalLines())
			assert.Equal(t, tcase.coveredLines, codeCov.CoveredLines())
		})
	}
}

func TestCoverageContentType(t *testing.T) {
	codeCov, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	assert.Equal(t, evidence.TypeCoverage, codeCov.Type())
}

func TestCoverageContentSubtype(t *testing.T) {
	codeCov, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	assert.Equal(t, evidence.SubtypeCoverage, codeCov.Subtype())
}

func TestCoverageContentPercent(t *testing.T) {
	tests := []struct {
		name         string
		totalLines   int
		coveredLines int
		expected     float64
	}{
		{
			name:         "shouldCalculatePercentNormally",
			totalLines:   200,
			coveredLines: 150,
			expected:     75.0,
		},
		{
			name:         "shouldReturnZeroWhenTotalLinesIsZero",
			totalLines:   0,
			coveredLines: 0,
			expected:     0.0,
		},
		{
			name:         "shouldReturn100WhenFullyCovered",
			totalLines:   100,
			coveredLines: 100,
			expected:     100.0,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			codeCov, err := coverage.NewContent(tcase.totalLines, tcase.coveredLines, nil)
			require.NoError(t, err)

			assert.InDelta(t, tcase.expected, codeCov.Percent(), 0.001)
		})
	}
}

func TestCoverageContentByLanguageDefensiveCopy(t *testing.T) {
	lc1 := validLanguageCoverage(t, "go", 100, 80)
	lc2 := validLanguageCoverage(t, "java", 200, 100)
	original := []coverage.LanguageStats{lc1}

	codeCov, err := coverage.NewContent(100, 80, original)
	require.NoError(t, err)

	original[0] = lc2
	assert.NotEqual(t, lc2, codeCov.ByLanguage()[0])

	returned := codeCov.ByLanguage()
	returned[0] = lc2
	assert.NotEqual(t, lc2, codeCov.ByLanguage()[0])
}

func TestCoverageContentMerge(t *testing.T) {
	lc1 := validLanguageCoverage(t, "go", 100, 80)
	lc2 := validLanguageCoverage(t, "java", 200, 150)

	cc1, err := coverage.NewContent(100, 80, []coverage.LanguageStats{lc1})
	require.NoError(t, err)
	cc2, err := coverage.NewContent(200, 150, []coverage.LanguageStats{lc2})
	require.NoError(t, err)

	merged, err := cc1.Merge(cc2)
	require.NoError(t, err)
	mergedCC, ok := merged.(coverage.Content)
	require.True(t, ok)

	assert.Equal(t, 300, mergedCC.TotalLines())
	assert.Equal(t, 230, mergedCC.CoveredLines())
	assert.Equal(t, 2, len(mergedCC.ByLanguage()))
}

func TestNewLicenseContent(t *testing.T) {
	depLic := validDependencyLicense(t, "chi", "v5.0.0", "MIT")

	tests := []struct {
		name         string
		dependencies []license.Dependency
	}{
		{
			name:         "shouldCreateWithDependencies",
			dependencies: []license.Dependency{depLic},
		},
		{
			name:         "shouldCreateWithEmptyDependencies",
			dependencies: []license.Dependency{},
		},
		{
			name:         "shouldCreateWithNilDependencies",
			dependencies: nil,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			langCov, err := license.NewContent(tcase.dependencies)

			require.NoError(t, err)
			if tcase.dependencies == nil {
				assert.Empty(t, langCov.Dependencies())
			} else {
				assert.Equal(t, len(tcase.dependencies), len(langCov.Dependencies()))
			}
		})
	}
}

func TestLicenseContentType(t *testing.T) {
	langCov, err := license.NewContent(nil)
	require.NoError(t, err)

	assert.Equal(t, evidence.TypeSupplyChain, langCov.Type())
}

func TestLicenseContentSubtype(t *testing.T) {
	langCov, err := license.NewContent(nil)
	require.NoError(t, err)

	assert.Equal(t, evidence.SubtypeLicense, langCov.Subtype())
}

func TestLicenseContentDependenciesDefensiveCopy(t *testing.T) {
	dl1 := validDependencyLicense(t, "chi", "v5.0.0", "MIT")
	dl2 := validDependencyLicense(t, "testify", "viol1.9.0", "MIT")
	original := []license.Dependency{dl1}

	langCov, err := license.NewContent(original)
	require.NoError(t, err)

	original[0] = dl2
	assert.NotEqual(t, dl2, langCov.Dependencies()[0])

	returned := langCov.Dependencies()
	returned[0] = dl2
	assert.NotEqual(t, dl2, langCov.Dependencies()[0])
}

func TestNewNewCodeCoverageContent(t *testing.T) {
	tests := []struct {
		name           string
		coveredLines   int
		coverableLines int
		wantErr        bool
	}{
		{name: "valid", coveredLines: 80, coverableLines: 100},
		{name: "zero lines", coveredLines: 0, coverableLines: 0},
		{name: "fully covered", coveredLines: 100, coverableLines: 100},
		{name: "negative covered rejected", coveredLines: -1, coverableLines: 100, wantErr: true},
		{name: "negative coverable rejected", coveredLines: 0, coverableLines: -1, wantErr: true},
		{name: "covered exceeds coverable rejected", coveredLines: 101, coverableLines: 100, wantErr: true},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			ncc, err := coverage.NewPatchContent(tcase.coveredLines, tcase.coverableLines)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, coverage.ErrInvalidPatchContent)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tcase.coveredLines, ncc.CoveredLines())
			assert.Equal(t, tcase.coverableLines, ncc.CoverableLines())
		})
	}
}

func TestNewCodeCoverageContentTypeAndSubtype(t *testing.T) {
	ncc, err := coverage.NewPatchContent(80, 100)
	require.NoError(t, err)

	assert.Equal(t, evidence.TypeCoverage, ncc.Type())
	assert.Equal(t, evidence.SubtypeNewCodeCoverage, ncc.Subtype())
}

func TestNewCodeCoverageContentPercent(t *testing.T) {
	tests := []struct {
		name     string
		covered  int
		total    int
		expected float64
	}{
		{name: "normal", covered: 75, total: 100, expected: 75.0},
		{name: "zero total", covered: 0, total: 0, expected: 0.0},
		{name: "fully covered", covered: 100, total: 100, expected: 100.0},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			ncc, err := coverage.NewPatchContent(tcase.covered, tcase.total)
			require.NoError(t, err)
			assert.InDelta(t, tcase.expected, ncc.Percent(), 0.001)
		})
	}
}

func TestNewCodeCoverageContentMerge(t *testing.T) {
	ncc1, err := coverage.NewPatchContent(50, 100)
	require.NoError(t, err)
	ncc2, err := coverage.NewPatchContent(30, 50)
	require.NoError(t, err)

	merged, err := ncc1.Merge(ncc2)
	require.NoError(t, err)
	mergedPC, ok := merged.(coverage.PatchContent)
	require.True(t, ok)

	assert.Equal(t, 80, mergedPC.CoveredLines())
	assert.Equal(t, 150, mergedPC.CoverableLines())
}

func TestFindingsContentMergeRejectsWrongType(t *testing.T) {
	findCont, err := finding.NewContent(evidence.SubtypeCodeQuality, nil)
	require.NoError(t, err)
	codeCov, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	_, err = findCont.Merge(codeCov)
	require.Error(t, err)
	assert.ErrorIs(t, err, finding.ErrInvalidContent)
}

func TestCoverageContentMergeRejectsWrongType(t *testing.T) {
	codeCov, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)
	findCont, err := finding.NewContent(evidence.SubtypeCodeQuality, nil)
	require.NoError(t, err)

	_, err = codeCov.Merge(findCont)
	require.Error(t, err)
	assert.ErrorIs(t, err, coverage.ErrInvalidContent)
}

func TestPatchContentMergeRejectsWrongType(t *testing.T) {
	pc, err := coverage.NewPatchContent(50, 100)
	require.NoError(t, err)
	codeCov, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	_, err = pc.Merge(codeCov)
	require.Error(t, err)
	assert.ErrorIs(t, err, coverage.ErrInvalidPatchContent)
}

func TestArchitectureContentMergeRejectsWrongType(t *testing.T) {
	archCont, err := architecture.NewContent(nil)
	require.NoError(t, err)
	codeCov, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	_, err = archCont.Merge(codeCov)
	require.Error(t, err)
}

func TestLicenseContentMergeRejectsWrongType(t *testing.T) {
	langCov, err := license.NewContent(nil)
	require.NoError(t, err)
	codeCov, err := coverage.NewContent(100, 80, nil)
	require.NoError(t, err)

	_, err = langCov.Merge(codeCov)
	require.Error(t, err)
}

func TestLicenseContentMerge(t *testing.T) {
	dl1 := validDependencyLicense(t, "chi", "v5.0.0", "MIT")
	dl2 := validDependencyLicense(t, "testify", "viol1.9.0", "MIT")
	dl3 := validDependencyLicense(t, "slog", "viol1.0.0", "BSD-3")

	lc1, err := license.NewContent([]license.Dependency{dl1, dl2})
	require.NoError(t, err)
	lc2, err := license.NewContent([]license.Dependency{dl3})
	require.NoError(t, err)

	merged, err := lc1.Merge(lc2)
	require.NoError(t, err)
	mergedLC, ok := merged.(license.Content)
	require.True(t, ok)

	assert.Equal(t, 3, len(mergedLC.Dependencies()))
}

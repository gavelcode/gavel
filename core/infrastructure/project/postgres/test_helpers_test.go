package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	casefile "github.com/usegavel/gavel/core/domain/casefile/model"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence"
	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/finding"
	"github.com/usegavel/gavel/core/domain/casefile/model/verdict"
	projectmodel "github.com/usegavel/gavel/core/domain/project/model"
	"github.com/usegavel/gavel/core/infrastructure/platform/database"
	"github.com/usegavel/gavel/core/infrastructure/platform/database/testkit"
)

func setupDB(t *testing.T) *database.DB { return testkit.TestDB(t) }

func mustGenerateProjectID(t *testing.T) projectmodel.ProjectID {
	t.Helper()
	id, err := projectmodel.ParseProjectID(uuid.NewString())
	require.NoError(t, err)
	return id
}

func newTestCaseFile(t *testing.T, projectID projectmodel.ProjectID, commitSHA, branch string) casefile.CaseFile {
	t.Helper()
	startedAt := time.Now().UTC()
	cf, err := casefile.NewCaseFile(projectID, commitSHA, branch, startedAt, startedAt)
	require.NoError(t, err)
	return cf
}

func newFindingsEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	fp1, err := finding.NewFingerprintID("fp-aaa")
	require.NoError(t, err)
	fp2, err := finding.NewFingerprintID("fp-bbb")
	require.NoError(t, err)

	f1, err := finding.NewFinding("pmd", "UnusedVariable", finding.SeverityWarning, "src/Foo.java", 10, "unused var x", fp1)
	require.NoError(t, err)
	f2, err := finding.NewFinding("spotbugs", "NP_NULL_DEREF", finding.SeverityError, "src/Bar.java", 25, "null deref", fp2)
	require.NoError(t, err)

	fc, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{f1, f2})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func newMultiSeverityEvidence(t *testing.T) evidence.Evidence {
	t.Helper()
	fp1, err := finding.NewFingerprintID("fp-err-1")
	require.NoError(t, err)
	fp2, err := finding.NewFingerprintID("fp-warn-1")
	require.NoError(t, err)
	fp3, err := finding.NewFingerprintID("fp-warn-2")
	require.NoError(t, err)
	fp4, err := finding.NewFingerprintID("fp-note-1")
	require.NoError(t, err)

	firstFinding, err := finding.NewFinding("pmd", "CriticalBug", finding.SeverityError, "src/A.java", 1, "critical bug", fp1)
	require.NoError(t, err)
	secondFinding, err := finding.NewFinding("pmd", "UnusedVar", finding.SeverityWarning, "src/B.java", 2, "unused var", fp2)
	require.NoError(t, err)
	f3, err := finding.NewFinding("pmd", "LongMethod", finding.SeverityWarning, "src/C.java", 3, "long method", fp3)
	require.NoError(t, err)
	f4, err := finding.NewFinding("pmd", "NamingConvention", finding.SeverityNote, "src/D.java", 4, "naming", fp4)
	require.NoError(t, err)

	fc, err := finding.NewContent(evidence.SubtypeCodeQuality, []finding.Finding{firstFinding, secondFinding, f3, f4})
	require.NoError(t, err)
	ev, err := evidence.NewEvidence(evidence.SubtypeCodeQuality, "pmd", fc, time.Now().UTC())
	require.NoError(t, err)
	return ev
}

func newTestVerdict(t *testing.T) verdict.Result {
	t.Helper()
	r1 := verdict.NewRuling(evidence.SubtypeCodeQuality, true, "0 errors found")
	r2 := verdict.NewRuling(evidence.SubtypeCoverage, true, "")
	v, err := verdict.Compose([]verdict.Ruling{r1, r2}, time.Now().UTC())
	require.NoError(t, err)
	return v
}

package v1integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	pleadingget "github.com/usegavel/gavel/core/application/pleading/get"
)

func seedPleading(f *testFixture, projectID string) string {
	pleadingID := uuid.NewString()
	f.pleadings.putDetail(&pleadingget.PleadingDetail{
		ID:           pleadingID,
		ProjectID:    projectID,
		Number:       42,
		Title:        "Fix bug",
		Petitioner:   "alice",
		SourceBranch: "feature",
		TargetBranch: "main",
		CommitSHA:    "abc123",
		Status:       "open",
		CreatedAt:    time.Date(2026, time.June, 6, 9, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, time.June, 6, 9, 0, 0, 0, time.UTC),
	})
	return pleadingID
}

func TestGetPleading_Success(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)
	pleadingID := seedPleading(f, uuid.NewString())

	res := f.do(t, http.MethodGet, "/pleadings/"+pleadingID, nil, cookie)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var body struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	mustDecode(t, res.Body.Bytes(), &body)
	require.Equal(t, pleadingID, body.ID)
	require.Equal(t, "Fix bug", body.Title)
}

func TestGetPleading_NotFound(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodGet, "/pleadings/"+uuid.NewString(), nil, cookie)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestResolvePleading_RequiresAdmin(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, viewerEmail, viewerPassword)

	res := f.do(t, http.MethodPatch, "/pleadings/"+uuid.NewString(), map[string]string{"outcome": "merged"}, cookie)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestResolvePleading_NotFound(t *testing.T) {
	f := newTestFixture(t)
	cookie := f.loginCookie(t, adminEmail, adminPassword)

	res := f.do(t, http.MethodPatch, "/pleadings/"+uuid.NewString(), map[string]string{"outcome": "merged"}, cookie)
	require.Equal(t, http.StatusNotFound, res.Code)
}

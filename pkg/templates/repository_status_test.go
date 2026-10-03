package templates

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestReportRendersRepositoryStatus(t *testing.T) {
	var rendered strings.Builder
	status := &models.RepositoryStatus{
		Branch:       "main",
		LastCommit:   "Add project report",
		LastCommitAt: time.Now().Add(-time.Hour),
		Staged:       1,
		Modified:     2,
		Untracked:    3,
	}
	meta := models.ReportMeta{IsServerMode: true, RepositoryStatus: status}
	if err := Report("view", &models.Config{}, nil, t.TempDir(), nil, nil, nil, meta, 0, nil).Render(context.Background(), &rendered); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Git", "main", "Last commit", "Not saved: 6", "1 staged", "2 modified", "3 new"} {
		if !strings.Contains(rendered.String(), expected) {
			t.Errorf("report missing %q", expected)
		}
	}
}

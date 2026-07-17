package templates

import (
	"context"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestChatPageCapturesFormDataBeforeDisablingForm(t *testing.T) {
	var rendered strings.Builder
	if err := ChatPage(models.ChatPageData{
		Title:        "AI Chat",
		HistoryJSON:  "[]",
		Provider:     "openai",
		Model:        "gpt-4.1",
		ActionURL:    "/chat",
		RunURL:       "/api/chat/run",
		StatusURL:    "/api/chat/status",
		AsyncEnabled: true,
	}).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render chat page: %v", err)
	}

	body := rendered.String()
	formDataIdx := strings.Index(body, "const formData = new FormData(form);")
	if formDataIdx == -1 {
		t.Fatalf("expected rendered chat page to capture form data before submit\nbody=%s", body)
	}

	setBusyIdx := strings.Index(body, "setBusy(true);")
	if setBusyIdx == -1 {
		t.Fatalf("expected rendered chat page to toggle busy state\nbody=%s", body)
	}

	if formDataIdx > setBusyIdx {
		t.Fatalf("expected form data capture before disabling form controls\nbody=%s", body)
	}

	if !strings.Contains(body, "body: formData,") {
		t.Fatalf("expected rendered chat page to submit the captured form data\nbody=%s", body)
	}
}

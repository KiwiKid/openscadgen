package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleChatStatusReturnsEventsAndResult(t *testing.T) {
	chatJobsMu.Lock()
	chatJobs = map[string]*chatJobState{
		"job-1": {
			ID:               "job-1",
			Provider:         providerOpenAI,
			Model:            "gpt-4.1",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			Done:             true,
			AssistantMessage: "Updated the config and reviewed the renders.",
			HistoryJSON:      `[{"role":"user","content":"hi"},{"role":"assistant","content":"done"}]`,
			Notice:           "Included config.toml as request context.",
			Events: []chatJobEvent{
				{Seq: 1, Message: "Applying edits"},
				{Seq: 2, Message: "Running OpenSCAD generation"},
			},
			FinishedAt: time.Now(),
		},
	}
	chatJobsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/chat/status?id=job-1&after=1", nil)
	rr := httptest.NewRecorder()

	handleChatStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp chatStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Done {
		t.Fatalf("expected done response, got %#v", resp)
	}
	if resp.AssistantMessage == "" {
		t.Fatalf("expected assistant message, got %#v", resp)
	}
	if len(resp.Events) != 1 || resp.Events[0].Seq != 2 {
		t.Fatalf("expected only seq>1 events, got %#v", resp.Events)
	}
}

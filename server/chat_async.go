package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kiwikid/openscadgen/pkg/models"
)

const (
	maxChatJobEvents = 200
	chatJobRetention = 10 * time.Minute
)

type chatJobEvent struct {
	Seq       int    `json:"seq"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at,omitempty"`
}

type chatJobState struct {
	ID               string
	Provider         string
	Model            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	FinishedAt       time.Time
	NextSeq          int
	Events           []chatJobEvent
	Done             bool
	Error            string
	AssistantMessage string
	HistoryJSON      string
	Notice           string
}

type chatRunResponse struct {
	JobID string `json:"job_id"`
}

type chatRunErrorResponse struct {
	Error  string `json:"error"`
	Notice string `json:"notice,omitempty"`
}

type chatStatusResponse struct {
	Events           []chatJobEvent `json:"events,omitempty"`
	Done             bool           `json:"done"`
	Error            string         `json:"error,omitempty"`
	AssistantMessage string         `json:"assistant_message,omitempty"`
	HistoryJSON      string         `json:"history_json,omitempty"`
	Notice           string         `json:"notice,omitempty"`
	Provider         string         `json:"provider,omitempty"`
	Model            string         `json:"model,omitempty"`
}

type chatStatusSink func(string)

var (
	chatJobsMu sync.Mutex
	chatJobs   = make(map[string]*chatJobState)
)

func handleChatRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	submission, historyErr, err := buildChatSubmission(r)
	if err != nil {
		writeChatRunError(w, http.StatusBadRequest, err.Error(), "")
		return
	}
	if historyErr != nil {
		writeChatRunError(w, http.StatusBadRequest, "Conversation state was invalid. Start a new chat.", "")
		return
	}
	if submission.Draft == "" {
		writeChatRunError(w, http.StatusBadRequest, "Enter a message before sending.", submission.Notice)
		return
	}
	if submission.Auth.Token == "" {
		writeChatRunError(w, http.StatusBadRequest, "Neither DEEPSEEK_API_TOKEN nor OPENAI_API_KEY is set.", submission.Notice)
		return
	}

	history := append(append([]models.ChatMessage(nil), submission.History...), models.ChatMessage{
		Role:    "user",
		Content: submission.Draft,
	})
	history = normalizeChatHistory(history)

	jobID := createChatJob(submission.Auth.SelectedProvider, submission.Model)
	writeJSON(w, http.StatusAccepted, chatRunResponse{JobID: jobID})

	go runChatJob(jobID, submission, history)
}

func handleChatStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		http.Error(w, "Missing job id", http.StatusBadRequest)
		return
	}

	after, _ := strconv.Atoi(r.URL.Query().Get("after"))
	resp, ok := getChatJobSnapshot(jobID, after)
	if !ok {
		http.Error(w, "Chat job not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func runChatJob(jobID string, submission chatSubmission, history []models.ChatMessage) {
	report := func(message string) {
		appendChatJobEvent(jobID, message)
	}

	report("Preparing chat request")
	report("Calling " + submission.Auth.SelectedProvider + " model " + submission.Model)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := createProviderChatResultWithContextAndStatus(
		ctx,
		submission.Auth,
		submission.Model,
		submission.Instructions,
		history,
		submission.CtxInfo,
		report,
	)
	if err != nil {
		completeChatJobError(jobID, err.Error(), submission.Notice)
		return
	}

	report("Received final assistant response")

	finalHistory := append(history, models.ChatMessage{
		Role:    "assistant",
		Content: result.OutputText,
	})
	finalHistory = normalizeChatHistory(finalHistory)

	completeChatJobSuccess(jobID, result.OutputText, finalHistory, submission.Notice)
}

func buildChatSubmission(r *http.Request) (chatSubmission, error, error) {
	if err := r.ParseForm(); err != nil {
		return chatSubmission{}, nil, err
	}

	history, historyErr := parseChatHistory(r.FormValue("history_json"))
	draft := strings.TrimSpace(r.FormValue("message"))
	ctxInfo := buildChatContext(r, true)
	auth := getChatAuthConfig(readRequestedProvider(r))
	model := readChatModel(r, auth.SelectedProvider)

	instructions, notice := buildChatInstructions(ctxInfo)
	if ctxInfo.LoadErr != nil {
		if notice != "" {
			notice += " "
		}
		notice += "Current file context could not be loaded: " + ctxInfo.LoadErr.Error()
	}
	if ctxInfo.ProjectErr != nil {
		if notice != "" {
			notice += " "
		}
		notice += "Project context could not be loaded: " + ctxInfo.ProjectErr.Error()
	}

	return chatSubmission{
		History:      history,
		Draft:        draft,
		Auth:         auth,
		Model:        model,
		CtxInfo:      ctxInfo,
		Instructions: instructions,
		Notice:       notice,
	}, historyErr, nil
}

func createChatJob(provider string, model string) string {
	chatJobsMu.Lock()
	defer chatJobsMu.Unlock()

	cleanupExpiredChatJobsLocked(time.Now())

	id := uuid.NewString()
	chatJobs[id] = &chatJobState{
		ID:        id,
		Provider:  provider,
		Model:     model,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return id
}

func appendChatJobEvent(jobID string, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}

	chatJobsMu.Lock()
	defer chatJobsMu.Unlock()

	job, ok := chatJobs[jobID]
	if !ok || job.Done {
		return
	}

	job.NextSeq++
	job.UpdatedAt = time.Now()
	job.Events = append(job.Events, chatJobEvent{
		Seq:       job.NextSeq,
		Message:   message,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if len(job.Events) > maxChatJobEvents {
		job.Events = append([]chatJobEvent(nil), job.Events[len(job.Events)-maxChatJobEvents:]...)
	}
}

func completeChatJobSuccess(jobID string, assistantMessage string, history []models.ChatMessage, notice string) {
	historyJSON, err := json.Marshal(history)
	if err != nil {
		historyJSON = []byte("[]")
	}

	chatJobsMu.Lock()
	defer chatJobsMu.Unlock()

	job, ok := chatJobs[jobID]
	if !ok {
		return
	}

	job.Done = true
	job.AssistantMessage = assistantMessage
	job.HistoryJSON = string(historyJSON)
	job.Notice = notice
	job.FinishedAt = time.Now()
	job.UpdatedAt = job.FinishedAt
}

func completeChatJobError(jobID string, errMsg string, notice string) {
	chatJobsMu.Lock()
	defer chatJobsMu.Unlock()

	job, ok := chatJobs[jobID]
	if !ok {
		return
	}

	job.Done = true
	job.Error = errMsg
	job.Notice = notice
	job.FinishedAt = time.Now()
	job.UpdatedAt = job.FinishedAt
}

func getChatJobSnapshot(jobID string, after int) (chatStatusResponse, bool) {
	chatJobsMu.Lock()
	defer chatJobsMu.Unlock()

	cleanupExpiredChatJobsLocked(time.Now())

	job, ok := chatJobs[jobID]
	if !ok {
		return chatStatusResponse{}, false
	}

	events := make([]chatJobEvent, 0, len(job.Events))
	for _, event := range job.Events {
		if event.Seq > after {
			events = append(events, event)
		}
	}

	return chatStatusResponse{
		Events:           events,
		Done:             job.Done,
		Error:            job.Error,
		AssistantMessage: job.AssistantMessage,
		HistoryJSON:      job.HistoryJSON,
		Notice:           job.Notice,
		Provider:         job.Provider,
		Model:            job.Model,
	}, true
}

func cleanupExpiredChatJobsLocked(now time.Time) {
	for id, job := range chatJobs {
		if !job.Done {
			continue
		}
		if now.Sub(job.FinishedAt) > chatJobRetention {
			delete(chatJobs, id)
		}
	}
}

func writeChatRunError(w http.ResponseWriter, status int, errMsg string, notice string) {
	writeJSON(w, status, chatRunErrorResponse{
		Error:  errMsg,
		Notice: notice,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

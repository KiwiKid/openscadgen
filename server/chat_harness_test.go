package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/kiwikid/openscadgen/pkg/models"
)

const (
	chatHarnessEnvMode  = "OPENSCADGEN_CHAT_HARNESS_MODE"
	chatHarnessEnvCase  = "OPENSCADGEN_CHAT_HARNESS_CASE"
	chatHarnessModeLive = "live"
	chatHarnessModeTest = "test-chat-tools"
)

type chatHarnessCase struct {
	Name         string               `json:"name,omitempty"`
	Provider     string               `json:"provider,omitempty"`
	Model        string               `json:"model,omitempty"`
	Prompt       string               `json:"prompt,omitempty"`
	History      []models.ChatMessage `json:"history,omitempty"`
	Instructions string               `json:"instructions,omitempty"`
	ContextPath  string               `json:"context_path,omitempty"`
	ServerFolder string               `json:"server_folder,omitempty"`
	Expect       chatHarnessExpect    `json:"expect,omitempty"`
}

type chatHarnessExpect struct {
	RequireText  bool     `json:"require_text,omitempty"`
	MinToolCalls int      `json:"min_tool_calls,omitempty"`
	ToolNames    []string `json:"tool_names,omitempty"`
}

type chatHarnessCapture struct {
	SchemaVersion int                         `json:"schema_version"`
	CaseName      string                      `json:"case_name"`
	CapturedAt    string                      `json:"captured_at"`
	Request       chatHarnessRecordedRequest  `json:"request"`
	Response      chatHarnessRecordedResponse `json:"response"`
}

type chatHarnessRecordedRequest struct {
	Provider     string               `json:"provider"`
	Model        string               `json:"model"`
	Instructions string               `json:"instructions"`
	History      []models.ChatMessage `json:"history"`
	ContextPath  string               `json:"context_path,omitempty"`
	ServerFolder string               `json:"server_folder,omitempty"`
}

type chatHarnessRecordedResponse struct {
	Provider   string         `json:"provider"`
	Model      string         `json:"model"`
	OutputText string         `json:"output_text"`
	ToolCalls  []chatToolCall `json:"tool_calls,omitempty"`
}

type chatHarnessToolChecker func(chatToolCall, chatHarnessCapture) error

func TestChatHarnessModes(t *testing.T) {
	mode := strings.TrimSpace(os.Getenv(chatHarnessEnvMode))
	if mode == "" {
		t.Skipf("set %s to %q or %q to run the chat harness", chatHarnessEnvMode, chatHarnessModeLive, chatHarnessModeTest)
	}

	repoRoot := chatHarnessRepoRoot(t)
	switch mode {
	case chatHarnessModeLive:
		runChatHarnessLive(t, repoRoot)
	case chatHarnessModeTest:
		runChatHarnessReplay(t, repoRoot)
	default:
		t.Fatalf("unknown %s %q", chatHarnessEnvMode, mode)
	}
}

func TestReplayChatHarnessToolCalls(t *testing.T) {
	capture := chatHarnessCapture{
		CaseName: "demo",
		Response: chatHarnessRecordedResponse{
			ToolCalls: []chatToolCall{
				{
					Type:      "function_call",
					Name:      "read_config_file",
					Arguments: `{"path":"examples/stick_hinge/config.toml"}`,
				},
			},
		},
	}

	registry := map[string]chatHarnessToolChecker{
		"read_config_file": func(call chatToolCall, capture chatHarnessCapture) error {
			if capture.CaseName != "demo" {
				return fmt.Errorf("unexpected case %q", capture.CaseName)
			}
			if call.Arguments != `{"path":"examples/stick_hinge/config.toml"}` {
				return fmt.Errorf("unexpected args %q", call.Arguments)
			}
			return nil
		},
	}

	if err := replayChatHarnessToolCalls(capture, registry); err != nil {
		t.Fatalf("replayChatHarnessToolCalls error: %v", err)
	}
}

func runChatHarnessLive(t *testing.T, repoRoot string) {
	cases := loadChatHarnessCases(t, repoRoot, true)
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			auth := getChatAuthConfig(tc.Provider)
			if auth.Token == "" {
				t.Fatalf("provider %q is not configured; set OPENAI_API_KEY or DEEPSEEK_API_TOKEN", tc.Provider)
			}

			ctxInfo := buildChatContextFromPaths(resolveChatHarnessContextPath(repoRoot, tc.ContextPath), resolveChatHarnessServerFolder(repoRoot, tc.ServerFolder, tc.ContextPath), true)
			if tc.ContextPath != "" && ctxInfo.LoadErr != nil {
				t.Fatalf("load context %q: %v", tc.ContextPath, ctxInfo.LoadErr)
			}

			baseInstructions := strings.TrimSpace(tc.Instructions)
			if baseInstructions == "" {
				baseInstructions = strings.TrimSpace(os.Getenv("OPENAI_CHAT_INSTRUCTIONS"))
			}
			instructions, notice := buildChatInstructionsWithBase(baseInstructions, ctxInfo)
			if notice != "" {
				t.Log(notice)
			}

			history := chatHarnessHistory(t, tc)
			model := strings.TrimSpace(tc.Model)
			if model == "" {
				model = defaultModel(auth.SelectedProvider)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			result, err := createProviderChatResult(ctx, auth, model, instructions, history)
			if err != nil {
				t.Fatalf("createProviderChatResult error: %v", err)
			}

			capture := chatHarnessCapture{
				SchemaVersion: 1,
				CaseName:      tc.Name,
				CapturedAt:    time.Now().UTC().Format(time.RFC3339),
				Request: chatHarnessRecordedRequest{
					Provider:     auth.SelectedProvider,
					Model:        model,
					Instructions: instructions,
					History:      history,
					ContextPath:  resolveChatHarnessContextPath(repoRoot, tc.ContextPath),
					ServerFolder: resolveChatHarnessServerFolder(repoRoot, tc.ServerFolder, tc.ContextPath),
				},
				Response: chatHarnessRecordedResponse{
					Provider:   result.Provider,
					Model:      result.Model,
					OutputText: result.OutputText,
					ToolCalls:  result.ToolCalls,
				},
			}

			if err := validateChatHarnessExpectations(tc, capture.Response); err != nil {
				t.Fatalf("capture did not match expectations: %v", err)
			}
			if err := writeChatHarnessCapture(repoRoot, capture); err != nil {
				t.Fatalf("write capture: %v", err)
			}
			t.Logf("wrote %s", chatHarnessCapturePath(repoRoot, tc.Name))
		})
	}
}

func runChatHarnessReplay(t *testing.T, repoRoot string) {
	cases := loadChatHarnessCases(t, repoRoot, false)
	registry := chatHarnessToolCheckers(repoRoot)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			capture, err := loadChatHarnessCapture(repoRoot, tc.Name)
			if err != nil {
				t.Fatalf("load capture: %v", err)
			}
			if err := validateChatHarnessExpectations(tc, capture.Response); err != nil {
				t.Fatalf("capture did not match expectations: %v", err)
			}
			if len(capture.Response.ToolCalls) == 0 {
				t.Log("no tool calls captured for this case")
				return
			}
			if err := replayChatHarnessToolCalls(capture, registry); err != nil {
				t.Fatalf("tool replay failed: %v", err)
			}
		})
	}
}

func loadChatHarnessCases(t *testing.T, repoRoot string, requireExplicitCase bool) []chatHarnessCase {
	t.Helper()

	selectedCase := strings.TrimSpace(os.Getenv(chatHarnessEnvCase))
	if requireExplicitCase && selectedCase == "" {
		t.Fatalf("set %s to the case filename without .json to avoid accidental live API fan-out", chatHarnessEnvCase)
	}

	casesDir := filepath.Join(repoRoot, "server", "testdata", "chat_harness_cases")
	if selectedCase != "" {
		return []chatHarnessCase{loadChatHarnessCaseFile(t, filepath.Join(casesDir, selectedCase+".json"))}
	}

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("read chat harness cases dir: %v", err)
	}

	cases := make([]chatHarnessCase, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		cases = append(cases, loadChatHarnessCaseFile(t, filepath.Join(casesDir, entry.Name())))
	}
	sort.Slice(cases, func(i, j int) bool {
		return cases[i].Name < cases[j].Name
	})
	if len(cases) == 0 {
		t.Fatalf("no chat harness cases found in %s", casesDir)
	}
	return cases
}

func loadChatHarnessCaseFile(t *testing.T, path string) chatHarnessCase {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read case %s: %v", path, err)
	}

	var tc chatHarnessCase
	if err := json.Unmarshal(body, &tc); err != nil {
		t.Fatalf("decode case %s: %v", path, err)
	}
	if strings.TrimSpace(tc.Name) == "" {
		tc.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	return tc
}

func writeChatHarnessCapture(repoRoot string, capture chatHarnessCapture) error {
	path := chatHarnessCapturePath(repoRoot, capture.CaseName)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	body, err := json.MarshalIndent(capture, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(path, body, 0o644)
}

func loadChatHarnessCapture(repoRoot string, caseName string) (chatHarnessCapture, error) {
	body, err := os.ReadFile(chatHarnessCapturePath(repoRoot, caseName))
	if err != nil {
		return chatHarnessCapture{}, err
	}

	var capture chatHarnessCapture
	if err := json.Unmarshal(body, &capture); err != nil {
		return chatHarnessCapture{}, err
	}
	return capture, nil
}

func validateChatHarnessExpectations(tc chatHarnessCase, response chatHarnessRecordedResponse) error {
	if tc.Expect.RequireText && strings.TrimSpace(response.OutputText) == "" {
		return fmt.Errorf("expected assistant text output")
	}
	if got := len(response.ToolCalls); got < tc.Expect.MinToolCalls {
		return fmt.Errorf("expected at least %d tool calls, got %d", tc.Expect.MinToolCalls, got)
	}
	for _, toolName := range tc.Expect.ToolNames {
		if !chatHarnessHasToolCall(response.ToolCalls, toolName) {
			return fmt.Errorf("expected tool call %q", toolName)
		}
	}
	return nil
}

func chatHarnessHasToolCall(calls []chatToolCall, want string) bool {
	for _, call := range calls {
		if call.Name == want || call.Type == want {
			return true
		}
	}
	return false
}

func replayChatHarnessToolCalls(capture chatHarnessCapture, registry map[string]chatHarnessToolChecker) error {
	for _, call := range capture.Response.ToolCalls {
		key := chatHarnessToolKey(call)
		if key == "" {
			return fmt.Errorf("captured tool call is missing both name and type")
		}
		if strings.TrimSpace(call.Arguments) != "" && !json.Valid([]byte(call.Arguments)) {
			return fmt.Errorf("%s has invalid JSON arguments: %q", key, call.Arguments)
		}
		checker, ok := registry[key]
		if !ok {
			return fmt.Errorf("no chat harness checker registered for %q", key)
		}
		if err := checker(call, capture); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}
	return nil
}

func chatHarnessToolKey(call chatToolCall) string {
	if strings.TrimSpace(call.Name) != "" {
		return call.Name
	}
	return strings.TrimSpace(call.Type)
}

func chatHarnessToolCheckers(repoRoot string) map[string]chatHarnessToolChecker {
	_ = repoRoot
	return map[string]chatHarnessToolChecker{}
}

func chatHarnessHistory(t *testing.T, tc chatHarnessCase) []models.ChatMessage {
	t.Helper()

	history := normalizeChatHistory(tc.History)
	if len(history) > 0 {
		return history
	}

	prompt := strings.TrimSpace(tc.Prompt)
	if prompt == "" {
		t.Fatalf("case %q must define either history or prompt", tc.Name)
	}
	return []models.ChatMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}
}

func resolveChatHarnessContextPath(repoRoot string, contextPath string) string {
	contextPath = strings.TrimSpace(contextPath)
	if contextPath == "" || filepath.IsAbs(contextPath) {
		return contextPath
	}
	return filepath.Join(repoRoot, contextPath)
}

func resolveChatHarnessServerFolder(repoRoot string, serverFolder string, contextPath string) string {
	serverFolder = strings.TrimSpace(serverFolder)
	if serverFolder != "" {
		if filepath.IsAbs(serverFolder) {
			return serverFolder
		}
		return filepath.Join(repoRoot, serverFolder)
	}
	if strings.TrimSpace(contextPath) != "" && !filepath.IsAbs(contextPath) {
		return repoRoot
	}
	return ""
}

func chatHarnessCapturePath(repoRoot string, caseName string) string {
	return filepath.Join(repoRoot, "tmp", "chat-harness", caseName+".json")
}

func chatHarnessRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(file))
}

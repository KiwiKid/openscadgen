package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestCreateOpenAIChatResponse(t *testing.T) {
	var got openAIResponsesRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if gotAuth := r.Header.Get("Authorization"); gotAuth != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %q", gotAuth)
		}

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"model":"gpt-4.1",
			"output":[
				{
					"type":"message",
					"role":"assistant",
					"content":[{"type":"output_text","text":"Use a thicker wall and round the corner."}]
				}
			]
		}`))
	}))
	defer srv.Close()

	answer, err := createOpenAIChatResponse(
		context.Background(),
		"test-key",
		srv.URL,
		"gpt-4.1",
		"Be helpful.",
		[]models.ChatMessage{
			{Role: "user", Content: "How can I make this stronger?"},
		},
	)
	if err != nil {
		t.Fatalf("createOpenAIChatResponse error: %v", err)
	}

	if answer != "Use a thicker wall and round the corner." {
		t.Fatalf("unexpected answer: %q", answer)
	}
	if got.Model != "gpt-4.1" {
		t.Fatalf("unexpected model: %q", got.Model)
	}
	if got.Instructions != "Be helpful." {
		t.Fatalf("unexpected instructions: %q", got.Instructions)
	}
	if len(got.Input) != 1 {
		t.Fatalf("expected 1 input message, got %d", len(got.Input))
	}
	var inputItem openAIInputMessage
	if err := json.Unmarshal(got.Input[0], &inputItem); err != nil {
		t.Fatalf("decode input item: %v", err)
	}
	if inputItem.Role != "user" {
		t.Fatalf("unexpected role: %q", inputItem.Role)
	}
	if len(inputItem.Content) != 1 || inputItem.Content[0].Text != "How can I make this stronger?" {
		t.Fatalf("unexpected content: %#v", inputItem.Content)
	}
}

func TestCreateOpenAIChatResultCapturesToolCallsWithoutAssistantText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"model":"gpt-4.1",
			"output":[
				{
					"id":"fc_123",
					"type":"function_call",
					"call_id":"call_123",
					"name":"read_config_file",
					"arguments":"{\"path\":\"examples/stick_hinge/config.toml\"}",
					"status":"completed"
				}
			]
		}`))
	}))
	defer srv.Close()

	result, err := createOpenAIChatResult(
		context.Background(),
		"test-key",
		srv.URL,
		"gpt-4.1",
		"Be helpful.",
		[]models.ChatMessage{
			{Role: "user", Content: "Read the config before answering."},
		},
	)
	if err != nil {
		t.Fatalf("createOpenAIChatResult error: %v", err)
	}

	if result.OutputText != "" {
		t.Fatalf("expected empty assistant text, got %q", result.OutputText)
	}
	if len(result.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(result.ToolCalls))
	}
	if result.ToolCalls[0].Name != "read_config_file" {
		t.Fatalf("unexpected tool name: %#v", result.ToolCalls[0])
	}
	if result.ToolCalls[0].Arguments != `{"path":"examples/stick_hinge/config.toml"}` {
		t.Fatalf("unexpected tool args: %q", result.ToolCalls[0].Arguments)
	}
	if len(result.RawResponse) == 0 {
		t.Fatal("expected raw response to be captured")
	}
}

func TestCreateDeepSeekChatResponse(t *testing.T) {
	var got deepSeekChatCompletionsRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if gotAuth := r.Header.Get("Authorization"); gotAuth != "Bearer deepseek-token" {
			t.Fatalf("unexpected auth header: %q", gotAuth)
		}

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"model":"deepseek-v4-flash",
			"choices":[
				{
					"message":{"role":"assistant","content":"Try increasing the fillet radius and wall thickness."}
				}
			]
		}`))
	}))
	defer srv.Close()

	answer, err := createDeepSeekChatResponse(
		context.Background(),
		"deepseek-token",
		srv.URL,
		"deepseek-v4-flash",
		"Be helpful.",
		[]models.ChatMessage{
			{Role: "user", Content: "How can I make this stronger?"},
		},
	)
	if err != nil {
		t.Fatalf("createDeepSeekChatResponse error: %v", err)
	}

	if answer != "Try increasing the fillet radius and wall thickness." {
		t.Fatalf("unexpected answer: %q", answer)
	}
	if got.Model != "deepseek-v4-flash" {
		t.Fatalf("unexpected model: %q", got.Model)
	}
	if len(got.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(got.Messages))
	}
	if got.Messages[0].Role != "system" || got.Messages[0].Content != "Be helpful." {
		t.Fatalf("unexpected system message: %#v", got.Messages[0])
	}
	if got.Messages[1].Role != "user" || got.Messages[1].Content != "How can I make this stronger?" {
		t.Fatalf("unexpected user message: %#v", got.Messages[1])
	}
}

func TestBuildChatContextLoadsRelativeConfigFromServerFolder(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("name = \"demo\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	configEncoded := base64.StdEncoding.EncodeToString([]byte("config.toml"))
	serverFolderEncoded := base64.StdEncoding.EncodeToString([]byte(tempDir))
	req := httptest.NewRequest(http.MethodGet, "/chat?config="+configEncoded+"&server_folder="+serverFolderEncoded, nil)

	ctxInfo := buildChatContext(req, true)

	if ctxInfo.ResolvedPath != configPath {
		t.Fatalf("expected resolved path %q, got %q", configPath, ctxInfo.ResolvedPath)
	}
	if ctxInfo.ProjectRoot != tempDir {
		t.Fatalf("expected project root %q, got %q", tempDir, ctxInfo.ProjectRoot)
	}
	if ctxInfo.ContextLabel != "config.toml" {
		t.Fatalf("unexpected label: %q", ctxInfo.ContextLabel)
	}
	if ctxInfo.LoadErr != nil {
		t.Fatalf("unexpected load err: %v", ctxInfo.LoadErr)
	}
	if !strings.Contains(ctxInfo.Content, `name = "demo"`) {
		t.Fatalf("unexpected content: %q", ctxInfo.Content)
	}
	if !strings.Contains(ctxInfo.ProjectSummary, "Top-level project entries:") {
		t.Fatalf("expected project summary, got %q", ctxInfo.ProjectSummary)
	}
}

func TestBuildChatContextUsesServerFolderAsProjectRootWithoutConfig(t *testing.T) {
	tempDir := t.TempDir()

	ctxInfo := buildChatContextFromPaths("", tempDir, true)

	if ctxInfo.ProjectRoot != tempDir {
		t.Fatalf("expected project root %q, got %q", tempDir, ctxInfo.ProjectRoot)
	}
	if ctxInfo.ProjectLabel != filepath.Base(tempDir) {
		t.Fatalf("unexpected project label: %q", ctxInfo.ProjectLabel)
	}
	if ctxInfo.ContextLabel != filepath.Base(tempDir) {
		t.Fatalf("unexpected context label: %q", ctxInfo.ContextLabel)
	}
	if ctxInfo.ProjectErr != nil {
		t.Fatalf("unexpected project err: %v", ctxInfo.ProjectErr)
	}
	if !strings.Contains(ctxInfo.ProjectSummary, "Project root:") {
		t.Fatalf("expected project summary, got %q", ctxInfo.ProjectSummary)
	}
}

func TestBuildChatInstructionsWithBaseAppendsContext(t *testing.T) {
	ctxInfo := chatContext{
		ContextLabel: "config.toml",
		ResolvedPath: "/tmp/config.toml",
		Content:      "name = \"demo\"\n",
	}

	instructions, notice := buildChatInstructionsWithBase("Custom instructions.", ctxInfo)

	if !strings.Contains(instructions, "Custom instructions.") {
		t.Fatalf("expected custom instructions, got %q", instructions)
	}
	if !strings.Contains(instructions, "Path: /tmp/config.toml") {
		t.Fatalf("expected path in instructions, got %q", instructions)
	}
	if !strings.Contains(instructions, "name = \"demo\"") {
		t.Fatalf("expected content in instructions, got %q", instructions)
	}
	if notice != "Included config.toml as request context." {
		t.Fatalf("unexpected notice: %q", notice)
	}
}

func TestBuildChatInstructionsWithBaseAppendsProjectContext(t *testing.T) {
	ctxInfo := chatContext{
		ProjectRoot:    "/tmp/stick_hinge",
		ProjectLabel:   "stick_hinge",
		ProjectSummary: "Project root: /tmp/stick_hinge\nTop-level project entries:\n- config.toml\n- stick_hinge.scad",
	}

	instructions, notice := buildChatInstructionsWithBase("Custom instructions.", ctxInfo)

	if !strings.Contains(instructions, "Current project workspace context follows.") {
		t.Fatalf("expected project workspace section, got %q", instructions)
	}
	if !strings.Contains(instructions, "Top-level project entries:") {
		t.Fatalf("expected project entries, got %q", instructions)
	}
	if !strings.Contains(instructions, "Never commit or push unless the user explicitly asks") {
		t.Fatalf("expected git safety instruction, got %q", instructions)
	}
	if notice != "Included project context for stick_hinge." {
		t.Fatalf("unexpected notice: %q", notice)
	}
}

func TestBuildOpenAIProjectToolsIncludesGitAndFileTools(t *testing.T) {
	tools := buildOpenAIProjectTools(chatContext{ProjectRoot: "/tmp/demo"})
	if len(tools) == 0 {
		t.Fatal("expected project tools")
	}

	names := make(map[string]bool, len(tools))
	for _, tool := range tools {
		names[tool.Name] = true
	}

	for _, want := range []string{
		"list_project_files",
		"read_project_file",
		"edit_project_file",
		"write_project_file",
		"git_status",
		"git_diff",
		"git_commit",
		"git_push",
	} {
		if !names[want] {
			t.Fatalf("expected tool %q to be present", want)
		}
	}
}

func TestCreateOpenAIChatResultWithContextExecutesProjectToolLoop(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "config.toml"), []byte("name = \"demo\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	requests := make([]openAIResponsesRequest, 0, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var got openAIResponsesRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, got)

		w.Header().Set("Content-Type", "application/json")
		switch len(requests) {
		case 1:
			foundReadTool := false
			for _, tool := range got.Tools {
				if tool.Name == "read_project_file" {
					foundReadTool = true
					break
				}
			}
			if !foundReadTool {
				t.Fatalf("expected read_project_file tool in first request, got %#v", got.Tools)
			}
			_, _ = w.Write([]byte(`{
				"model":"gpt-4.1",
				"output":[
					{
						"id":"fc_123",
						"type":"function_call",
						"call_id":"call_123",
						"name":"read_project_file",
						"arguments":"{\"path\":\"config.toml\"}",
						"status":"completed"
					}
				]
			}`))
		case 2:
			foundToolOutput := false
			for _, raw := range got.Input {
				var item map[string]any
				if err := json.Unmarshal(raw, &item); err != nil {
					t.Fatalf("decode input item: %v", err)
				}
				if item["type"] != "function_call_output" {
					continue
				}
				foundToolOutput = true
				output, _ := item["output"].(string)
				if !strings.Contains(output, `name = \"demo\"`) {
					t.Fatalf("expected file content in tool output, got %q", output)
				}
			}
			if !foundToolOutput {
				t.Fatal("expected function_call_output item in second request")
			}
			_, _ = w.Write([]byte(`{
				"model":"gpt-4.1",
				"output":[
					{
						"type":"message",
						"role":"assistant",
						"content":[{"type":"output_text","text":"The config is named demo."}]
					}
				]
			}`))
		default:
			t.Fatalf("unexpected request count: %d", len(requests))
		}
	}))
	defer srv.Close()

	result, err := createOpenAIChatResultWithContext(
		context.Background(),
		"test-key",
		srv.URL,
		"gpt-4.1",
		"Be helpful.",
		[]models.ChatMessage{
			{Role: "user", Content: "Read the config then summarize it."},
		},
		chatContext{
			ProjectRoot:  tempDir,
			ProjectLabel: "demo",
		},
	)
	if err != nil {
		t.Fatalf("createOpenAIChatResultWithContext error: %v", err)
	}

	if result.OutputText != "The config is named demo." {
		t.Fatalf("unexpected output text: %q", result.OutputText)
	}
	if len(result.ToolCalls) != 1 || result.ToolCalls[0].Name != "read_project_file" {
		t.Fatalf("unexpected tool calls: %#v", result.ToolCalls)
	}
	if len(requests) != 2 {
		t.Fatalf("expected 2 OpenAI requests, got %d", len(requests))
	}
}

func TestParseChatHistoryNormalizesAndClamps(t *testing.T) {
	input := `[{"role":"USER","content":" first "},{"role":"assistant","content":"second"},{"role":"system","content":"ignore"}]`

	history, err := parseChatHistory(input)
	if err != nil {
		t.Fatalf("parseChatHistory error: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(history))
	}
	if history[0].Role != "user" || history[0].Content != "first" {
		t.Fatalf("unexpected first message: %#v", history[0])
	}
	if history[1].Role != "assistant" || history[1].Content != "second" {
		t.Fatalf("unexpected second message: %#v", history[1])
	}
}

func TestGetChatAuthConfig(t *testing.T) {
	t.Run("defaults to deepseek when it is the only configured provider", func(t *testing.T) {
		t.Setenv("DEEPSEEK_API_TOKEN", "deepseek-token")
		t.Setenv("OPENAI_API_KEY", "")

		auth := getChatAuthConfig("")
		if auth.SelectedProvider != providerDeepSeek {
			t.Fatalf("expected deepseek provider, got %q", auth.SelectedProvider)
		}
		if auth.Token != "deepseek-token" {
			t.Fatalf("expected deepseek token, got %q", auth.Token)
		}
		if auth.Label != "Using DEEPSEEK_API_TOKEN" {
			t.Fatalf("unexpected label: %q", auth.Label)
		}
	})

	t.Run("defaults to openai when it is the only configured provider", func(t *testing.T) {
		t.Setenv("DEEPSEEK_API_TOKEN", "")
		t.Setenv("OPENAI_API_KEY", "openai-token")

		auth := getChatAuthConfig("")
		if auth.SelectedProvider != providerOpenAI {
			t.Fatalf("expected openai provider, got %q", auth.SelectedProvider)
		}
		if auth.Token != "openai-token" {
			t.Fatalf("expected openai token, got %q", auth.Token)
		}
		if auth.Label != "Using OPENAI_API_KEY" {
			t.Fatalf("unexpected label: %q", auth.Label)
		}
	})

	t.Run("allows switching providers when both are set", func(t *testing.T) {
		t.Setenv("DEEPSEEK_API_TOKEN", "deepseek-token")
		t.Setenv("OPENAI_API_KEY", "openai-token")

		auth := getChatAuthConfig(providerOpenAI)
		if auth.SelectedProvider != providerOpenAI {
			t.Fatalf("expected openai provider, got %q", auth.SelectedProvider)
		}
		if auth.Token != "openai-token" {
			t.Fatalf("expected openai token, got %q", auth.Token)
		}
		if auth.Label != "Using OPENAI_API_KEY" {
			t.Fatalf("unexpected label: %q", auth.Label)
		}
	})

	t.Run("shows missing state when neither is set", func(t *testing.T) {
		t.Setenv("DEEPSEEK_API_TOKEN", "")
		t.Setenv("OPENAI_API_KEY", "")

		auth := getChatAuthConfig("")
		if auth.Token != "" {
			t.Fatalf("expected empty token, got %q", auth.Token)
		}
		if auth.Label != "No API token configured" {
			t.Fatalf("unexpected label: %q", auth.Label)
		}
	})
}

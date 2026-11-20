package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testMCPConfig(t *testing.T, roots ...string) mcpConfig {
	t.Helper()
	return mcpConfig{Enabled: true, Detail: mcpDetailNormal, Roots: roots, BindAddress: "127.0.0.1"}
}

func callMCP(t *testing.T, server *mcpServer, method string, params any) rpcResponse {
	t.Helper()
	var raw json.RawMessage
	if params != nil {
		encoded, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("marshal params: %v", err)
		}
		raw = encoded
	}
	request := map[string]any{"jsonrpc": "2.0", "id": 1, "method": method}
	if raw != nil {
		request["params"] = raw
	}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	response, hasResponse := server.handlePayload(context.Background(), payload)
	if !hasResponse {
		t.Fatalf("%s produced no response", method)
	}
	var parsed rpcResponse
	if err := json.Unmarshal(response, &parsed); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return parsed
}

func callTool(t *testing.T, server *mcpServer, name string, args any) rpcResponse {
	t.Helper()
	return callMCP(t, server, "tools/call", map[string]any{"name": name, "arguments": args})
}

func TestMCPInitialize(t *testing.T) {
	root := t.TempDir()
	server := newMCPServer(testMCPConfig(t, root))

	response := callMCP(t, server, "initialize", map[string]any{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test", "version": "1"},
	})
	if response.Error != nil {
		t.Fatalf("initialize error: %+v", response.Error)
	}

	var result struct {
		ProtocolVersion string `json:"protocolVersion"`
		ServerInfo      struct {
			Name string `json:"name"`
		} `json:"serverInfo"`
		Capabilities map[string]json.RawMessage `json:"capabilities"`
		Instructions string                     `json:"instructions"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result.ProtocolVersion != "2025-06-18" {
		t.Fatalf("protocolVersion = %q", result.ProtocolVersion)
	}
	if result.ServerInfo.Name != "openscadgen" {
		t.Fatalf("serverInfo.name = %q", result.ServerInfo.Name)
	}
	if _, ok := result.Capabilities["tools"]; !ok {
		t.Fatalf("tools capability missing: %v", result.Capabilities)
	}
	if !strings.Contains(result.Instructions, root) {
		t.Fatalf("instructions should mention the root: %q", result.Instructions)
	}
}

func TestMCPProtocolVersionFallback(t *testing.T) {
	server := newMCPServer(testMCPConfig(t, t.TempDir()))
	response := callMCP(t, server, "initialize", map[string]any{"protocolVersion": "1999-01-01"})
	var result struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result.ProtocolVersion != mcpProtocolVersion {
		t.Fatalf("expected fallback to %q, got %q", mcpProtocolVersion, result.ProtocolVersion)
	}
}

func TestMCPToolsListGating(t *testing.T) {
	root := t.TempDir()

	names := func(cfg mcpConfig) []string {
		server := newMCPServer(cfg)
		response := callMCP(t, server, "tools/list", nil)
		var result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		}
		if err := json.Unmarshal(response.Result, &result); err != nil {
			t.Fatalf("decode tools: %v", err)
		}
		out := make([]string, 0, len(result.Tools))
		for _, tool := range result.Tools {
			out = append(out, tool.Name)
		}
		return out
	}

	def := testMCPConfig(t, root)
	if got := len(names(def)); got != 10 {
		t.Fatalf("default tools = %d, want 10: %v", got, names(def))
	}

	readOnly := def
	readOnly.ReadOnly = true
	readOnlyNames := names(readOnly)
	if len(readOnlyNames) != 6 {
		t.Fatalf("read-only tools = %d, want 6: %v", len(readOnlyNames), readOnlyNames)
	}
	for _, name := range readOnlyNames {
		switch name {
		case "list_projects", "list_project_files", "read_project_file", "git_status", "git_diff", "process_status":
		default:
			t.Fatalf("read-only mode advertised write tool %q", name)
		}
	}

	withPush := def
	withPush.AllowPush = true
	if got := len(names(withPush)); got != 11 {
		t.Fatalf("allow-push tools = %d, want 11: %v", got, names(withPush))
	}
}

func TestMCPToolCallReadsFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello mcp"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	server := newMCPServer(testMCPConfig(t, root))

	response := callTool(t, server, "read_project_file", map[string]any{"path": "hello.txt"})
	if response.Error != nil {
		t.Fatalf("tool call error: %+v", response.Error)
	}
	if !strings.Contains(string(response.Result), "hello mcp") {
		t.Fatalf("result missing file content: %s", response.Result)
	}

	var result struct {
		IsError bool `json:"isError"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Structured map[string]any `json:"structuredContent"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("decode tool result: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected isError: %s", result.Content)
	}
	if result.Structured["ok"] != true {
		t.Fatalf("structuredContent.ok = %v", result.Structured["ok"])
	}
	if len(result.Content) != 1 || !strings.Contains(result.Content[0].Text, "hello mcp") {
		t.Fatalf("unexpected content: %+v", result.Content)
	}
}

func TestMCPUnknownAndGatedTools(t *testing.T) {
	server := newMCPServer(testMCPConfig(t, t.TempDir()))

	unknown := callTool(t, server, "does_not_exist", map[string]any{})
	if unknown.Error == nil || unknown.Error.Code != rpcInvalidParams {
		t.Fatalf("unknown tool error = %+v", unknown.Error)
	}

	gated := callTool(t, server, "git_push", map[string]any{})
	if gated.Error == nil || gated.Error.Code != rpcInvalidParams {
		t.Fatalf("gated tool error = %+v", gated.Error)
	}
	if !strings.Contains(gated.Error.Message, "-mcp-allow-push") {
		t.Fatalf("gated tool message should explain the flag: %q", gated.Error.Message)
	}
}

func TestMCPNotificationProducesNoResponse(t *testing.T) {
	server := newMCPServer(testMCPConfig(t, t.TempDir()))
	payload := []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	response, hasResponse := server.handlePayload(context.Background(), payload)
	if hasResponse {
		t.Fatalf("notification produced a response: %s", response)
	}
}

func TestMCPBatchAndProtocolErrors(t *testing.T) {
	server := newMCPServer(testMCPConfig(t, t.TempDir()))

	batch := []byte(`[
		{"jsonrpc":"2.0","id":1,"method":"ping"},
		{"jsonrpc":"2.0","method":"notifications/initialized"}
	]`)
	response, hasResponse := server.handlePayload(context.Background(), batch)
	if !hasResponse {
		t.Fatal("batch produced no response")
	}
	var responses []rpcResponse
	if err := json.Unmarshal(response, &responses); err != nil {
		t.Fatalf("decode batch: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("batch responses = %d, want 1", len(responses))
	}

	badVersion, hasResponse := server.handlePayload(context.Background(), []byte(`{"jsonrpc":"1.0","id":2,"method":"ping"}`))
	if !hasResponse || !strings.Contains(string(badVersion), "-32600") {
		t.Fatalf("expected invalid request error, got %s", badVersion)
	}

	invalidJSON, hasResponse := server.handlePayload(context.Background(), []byte(`{`))
	if !hasResponse || !strings.Contains(string(invalidJSON), "-32700") {
		t.Fatalf("expected parse error, got %s", invalidJSON)
	}

	missingMethod, hasResponse := server.handlePayload(context.Background(), []byte(`{"jsonrpc":"2.0","id":3}`))
	if !hasResponse || !strings.Contains(string(missingMethod), "-32600") {
		t.Fatalf("expected invalid request, got %s", missingMethod)
	}

	unknownMethod, _ := server.handlePayload(context.Background(), []byte(`{"jsonrpc":"2.0","id":4,"method":"tools/unknown"}`))
	if !strings.Contains(string(unknownMethod), "-32601") {
		t.Fatalf("expected method not found, got %s", unknownMethod)
	}
}

func TestMCPHTTPHandler(t *testing.T) {
	root := t.TempDir()
	previousEnabled := mcpToolsEnabled
	previousConfig := mcpServerConfig
	t.Cleanup(func() {
		mcpToolsEnabled = previousEnabled
		mcpServerConfig = previousConfig
	})

	mcpToolsEnabled = false
	disabled := httptest.NewRecorder()
	handleMCPHTTP(disabled, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{}`)))
	if disabled.Code != http.StatusNotFound {
		t.Fatalf("disabled status = %d, want 404", disabled.Code)
	}

	mcpToolsEnabled = true
	mcpServerConfig = testMCPConfig(t, root)

	getRecorder := httptest.NewRecorder()
	handleMCPHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/mcp", nil))
	if getRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET status = %d, want 405", getRecorder.Code)
	}
	if allow := getRecorder.Header().Get("Allow"); allow != http.MethodPost {
		t.Fatalf("Allow = %q", allow)
	}

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	postRecorder := httptest.NewRecorder()
	handleMCPHTTP(postRecorder, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body)))
	if postRecorder.Code != http.StatusOK {
		t.Fatalf("POST status = %d, want 200: %s", postRecorder.Code, postRecorder.Body.String())
	}
	if !strings.Contains(postRecorder.Body.String(), `"list_projects"`) {
		t.Fatalf("response missing tools: %s", postRecorder.Body.String())
	}

	notification := httptest.NewRecorder()
	handleMCPHTTP(notification, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)))
	if notification.Code != http.StatusAccepted {
		t.Fatalf("notification status = %d, want 202", notification.Code)
	}
}

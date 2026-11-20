package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

// MCP protocol versions this server can speak. The newest is advertised when a
// client asks for something unknown.
const (
	mcpProtocolVersion       = "2025-06-18"
	mcpProtocolVersionLegacy = "2024-11-05"
	maxMCPBodyBytes          = 1 << 20
	maxMCPStdioLineBytes     = 4 << 20
)

var mcpSupportedVersions = map[string]bool{
	mcpProtocolVersion:       true,
	"2025-03-26":             true,
	mcpProtocolVersionLegacy: true,
}

// JSON-RPC 2.0 error codes used by this server.
const (
	rpcParseError     = -32700
	rpcInvalidRequest = -32600
	rpcMethodNotFound = -32601
	rpcInvalidParams  = -32602
	rpcInternalError  = -32603
)

type rpcRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// mcpServer is the transport-independent JSON-RPC handler. Both HTTP and stdio
// feed it raw frames and write the returned bytes.
type mcpServer struct {
	cfg mcpConfig
}

func newMCPServer(cfg mcpConfig) *mcpServer {
	return &mcpServer{cfg: cfg}
}

// handlePayload accepts one JSON-RPC message, or a batch array, and returns the
// response bytes. hasResponse is false for pure notifications, which must not
// produce a body.
func (s *mcpServer) handlePayload(ctx context.Context, payload []byte) ([]byte, bool) {
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 {
		return mustMarshal(rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: rpcParseError, Message: "empty request"}}), true
	}

	if trimmed[0] == '[' {
		var batch []json.RawMessage
		if err := json.Unmarshal(trimmed, &batch); err != nil {
			return mustMarshal(rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: rpcParseError, Message: "invalid JSON: " + err.Error()}}), true
		}
		if len(batch) == 0 {
			return mustMarshal(rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: rpcInvalidRequest, Message: "empty batch"}}), true
		}
		responses := make([]rpcResponse, 0, len(batch))
		for _, item := range batch {
			if response, ok := s.handleOne(ctx, item); ok {
				responses = append(responses, response)
			}
		}
		if len(responses) == 0 {
			return nil, false
		}
		return mustMarshal(responses), true
	}

	response, ok := s.handleOne(ctx, trimmed)
	if !ok {
		return nil, false
	}
	return mustMarshal(response), true
}

// handleOne dispatches a single JSON-RPC message. A false return means the
// message was a notification and nothing should be written.
func (s *mcpServer) handleOne(ctx context.Context, raw json.RawMessage) (rpcResponse, bool) {
	var request rpcRequest
	if err := json.Unmarshal(raw, &request); err != nil {
		return rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: rpcParseError, Message: "invalid JSON: " + err.Error()}}, true
	}

	notification := request.ID == nil
	id := json.RawMessage("null")
	if request.ID != nil {
		id = *request.ID
	}

	if request.JSONRPC != "2.0" {
		return rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: rpcInvalidRequest, Message: `jsonrpc must be "2.0"`}}, true
	}
	if strings.TrimSpace(request.Method) == "" {
		return rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: rpcInvalidRequest, Message: "method is required"}}, true
	}

	// Notifications never get a response, including unknown ones.
	if notification {
		return rpcResponse{}, false
	}

	result, rpcErr := s.dispatch(ctx, request)
	if rpcErr != nil {
		return rpcResponse{JSONRPC: "2.0", ID: id, Error: rpcErr}, true
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: rpcInternalError, Message: "failed to encode result: " + err.Error()}}, true
	}
	return rpcResponse{JSONRPC: "2.0", ID: id, Result: encoded}, true
}

func (s *mcpServer) dispatch(ctx context.Context, request rpcRequest) (any, *rpcError) {
	switch request.Method {
	case "initialize":
		return s.initialize(request.Params)
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return s.listTools(), nil
	case "tools/call":
		return s.callTool(ctx, request.Params)
	case "resources/list":
		return map[string]any{"resources": []any{}}, nil
	case "prompts/list":
		return map[string]any{"prompts": []any{}}, nil
	default:
		return nil, &rpcError{Code: rpcMethodNotFound, Message: fmt.Sprintf("method %q is not supported", request.Method)}
	}
}

func (s *mcpServer) initialize(params json.RawMessage) (any, *rpcError) {
	var request struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	_ = json.Unmarshal(params, &request)

	version := mcpProtocolVersion
	if mcpSupportedVersions[request.ProtocolVersion] {
		version = request.ProtocolVersion
	}

	return map[string]any{
		"protocolVersion": version,
		"capabilities": map[string]any{
			"tools": map[string]any{"listChanged": false},
		},
		"serverInfo": map[string]any{
			"name":    "openscadgen",
			"version": mcpServerVersion(),
		},
		"instructions": s.instructions(),
	}, nil
}

// instructions tells the client how to work within this server's safety gates.
func (s *mcpServer) instructions() string {
	var b strings.Builder
	b.WriteString("OpenSCADGen MCP server. Tools operate on OpenSCADGen project folders.\n")
	b.WriteString("Allowed roots:\n")
	for _, root := range s.cfg.Roots {
		fmt.Fprintf(&b, "- %s\n", root)
	}
	b.WriteString("All paths are relative to a root and are refused if they escape it.\n")
	fmt.Fprintf(&b, "Default response_detail is %q. Use response_detail=summary on successful calls to save tokens; failures always return full detail.\n", s.cfg.Detail)
	if s.cfg.ReadOnly {
		b.WriteString("This server is read-only: write, edit, commit, and push tools are not available.\n")
	} else if !s.cfg.AllowPush {
		b.WriteString("git_push is disabled; git_commit is available.\n")
	}
	b.WriteString("process_config validates synchronously and then renders in the background. It returns job_id plus progress_url (HTTP) and progress_tool; poll either for started_at, elapsed_ms, current_instance, completed_instances, and total_instances. Pass wait_seconds to block for the finished result instead.\n")
	return b.String()
}

func (s *mcpServer) listTools() any {
	tools := availableMCPTools(s.cfg)
	items := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		items = append(items, map[string]any{
			"name":        tool.Name,
			"title":       tool.Title,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
			"annotations": map[string]any{
				"title":           tool.Title,
				"readOnlyHint":    tool.ReadOnly,
				"destructiveHint": !tool.ReadOnly,
				"idempotentHint":  tool.ReadOnly,
				"openWorldHint":   false,
			},
		})
	}
	return map[string]any{"tools": items}
}

func (s *mcpServer) callTool(ctx context.Context, params json.RawMessage) (any, *rpcError) {
	var request struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, &rpcError{Code: rpcInvalidParams, Message: "invalid tools/call params: " + err.Error()}
	}
	if strings.TrimSpace(request.Name) == "" {
		return nil, &rpcError{Code: rpcInvalidParams, Message: "tools/call requires a tool name"}
	}

	tool, ok := findMCPTool(s.cfg, request.Name)
	if !ok {
		// Distinguish "unknown" from "disabled" so clients get actionable text.
		for _, candidate := range allMCPTools() {
			if candidate.Name != request.Name {
				continue
			}
			reason := "this tool is disabled by the current server configuration"
			if candidate.RequiresPush {
				reason = "git_push is disabled; restart the server with -mcp-allow-push to enable it"
			} else if s.cfg.ReadOnly {
				reason = "the server is running read-only; restart without -mcp-read-only to enable write tools"
			}
			return nil, &rpcError{Code: rpcInvalidParams, Message: fmt.Sprintf("tool %q is not available: %s", request.Name, reason)}
		}
		return nil, &rpcError{Code: rpcInvalidParams, Message: fmt.Sprintf("unknown tool %q", request.Name)}
	}

	payload, err := tool.Handler(ctx, s.cfg, request.Arguments)
	if err != nil {
		return toolErrorContent(err), nil
	}
	if payload == nil {
		return toolErrorContent(fmt.Errorf("tool %q returned no payload", tool.Name)), nil
	}

	encoded, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return toolErrorContent(fmt.Errorf("encode tool result: %w", marshalErr)), nil
	}

	isError := false
	if okValue, exists := payload["ok"]; exists {
		if boolValue, isBool := okValue.(bool); isBool && !boolValue {
			isError = true
		}
	}

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(encoded)},
		},
		"structuredContent": payload,
		"isError":           isError,
	}, nil
}

// toolErrorContent renders a tool failure as an MCP tool error result. Tool
// failures stay inside tools/call (isError) per the MCP spec; they are not
// JSON-RPC protocol errors.
func toolErrorContent(err error) map[string]any {
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": err.Error()},
		},
		"isError": true,
	}
}

func mustMarshal(value any) []byte {
	encoded, err := json.Marshal(value)
	if err != nil {
		return []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"failed to encode response"}}`)
	}
	return encoded
}

// handleMCPHTTP serves the Streamable HTTP transport: one POST endpoint that
// accepts a single JSON-RPC message or a batch. It is mounted only when MCP is
// enabled.
func handleMCPHTTP(w http.ResponseWriter, r *http.Request) {
	if !mcpToolsEnabled {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "MCP endpoint only accepts POST", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxMCPBodyBytes))
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	server := newMCPServer(mcpServerConfig)
	response, hasResponse := server.handlePayload(r.Context(), body)
	w.Header().Set("Cache-Control", "no-store")
	if !hasResponse {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

// StartMCPStdio resolves the effective configuration and serves MCP over
// stdin/stdout until the input stream closes or ctx is cancelled.
func StartMCPStdio(ctx context.Context, flags models.CmdFlags) error {
	cfg := newMCPConfig(flags, flags.ServerFolder)
	if !cfg.Enabled {
		return fmt.Errorf("stdio MCP requires -ai or OPENSCADGEN_AI_ON=true")
	}
	pkg.LogInfof("MCP stdio server starting: %d tool(s), default detail %q, %d root(s)",
		len(availableMCPTools(cfg)), cfg.Detail, len(cfg.Roots))
	for _, notice := range cfg.Notices {
		pkg.LogWarnf("%s", notice)
	}
	return RunMCPStdio(ctx, cfg)
}

// RunMCPStdio serves MCP over stdin/stdout using newline-delimited JSON-RPC.
// stdout is reserved for protocol frames; main routes logs to stderr first.
func RunMCPStdio(ctx context.Context, cfg mcpConfig) error {
	// Belt and braces: any stdlib log call in the render path must reach stderr,
	// never the protocol stream.
	log.SetOutput(os.Stderr)

	server := newMCPServer(cfg)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 64*1024), maxMCPStdioLineBytes)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		response, hasResponse := server.handlePayload(ctx, line)
		if !hasResponse {
			continue
		}
		if _, err := writer.Write(response); err != nil {
			return err
		}
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
		if err := writer.Flush(); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stdio frame: %w", err)
	}
	return nil
}

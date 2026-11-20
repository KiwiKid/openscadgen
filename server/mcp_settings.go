package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"
)

// Settings page and MCP handler state. mcpToolsEnabled replaces the old chat
// flag and is the single source of truth for "is AI/MCP on".
var (
	mcpServerConfig mcpConfig
	mcpHTTPPort     int
)

// handleAISettingsPage renders the read-only AI/MCP settings page. Every value
// shown is derived from flags and environment variables; the page never writes
// configuration back to disk.
func handleAISettingsPage(w http.ResponseWriter, r *http.Request) {
	if !mcpToolsEnabled {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := buildAISettingsPageData(mcpServerConfig, mcpHTTPPort)
	if err := templates.AISettingsPage(data, openscadNavView()).Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render AI settings page: "+err.Error(), http.StatusInternalServerError)
	}
}

// buildAISettingsPageData assembles the settings view model.
func buildAISettingsPageData(cfg mcpConfig, port int) models.AISettingsPageData {
	if port == 0 {
		port = 7203
	}
	endpoint := cfg.mcpEndpointURL(port)

	data := models.AISettingsPageData{
		Title:           "AI & MCP",
		Enabled:         true,
		HTTPMounted:     !cfg.Stdio,
		StdioMode:       cfg.Stdio,
		EndpointURL:     endpoint,
		ProgressPath:    "/api/mcp/process/{job_id}",
		StdioCommand:    cfg.stdioCommand(),
		ProtocolVersion: mcpProtocolVersion,
		DetailLevel:     cfg.Detail,
		ReadOnly:        cfg.ReadOnly,
		AllowPush:       cfg.AllowPush,
		Roots:           cfg.Roots,
		BindAddress:     cfg.BindAddress,
		LoopbackOnly:    isLoopbackAddress(cfg.BindAddress),
		Tools:           mcpToolDescriptions(cfg),
		Sources:         cfg.describeSources(),
		Notices:         cfg.Notices,
	}
	data.Snippets = buildMCPClientSnippets(cfg, port)
	return data
}

// buildMCPClientSnippets produces copy-ready client configuration blocks.
func buildMCPClientSnippets(cfg mcpConfig, port int) []models.MCPClientSnippet {
	endpoint := cfg.mcpEndpointURL(port)

	httpConfig := map[string]any{
		"mcpServers": map[string]any{
			"openscadgen": map[string]any{
				"url": endpoint,
			},
		},
	}

	args := []string{"-ai", "-mcp-stdio"}
	if len(cfg.Roots) > 0 {
		args = append(args, "-mcp-root", strings.Join(cfg.Roots, ","))
	}
	if cfg.ReadOnly {
		args = append(args, "-mcp-read-only")
	}
	if cfg.AllowPush {
		args = append(args, "-mcp-allow-push")
	}
	if cfg.Detail != "" && cfg.Detail != mcpDetailNormal {
		args = append(args, "-mcp-detail", cfg.Detail)
	}
	exe, err := os.Executable()
	if err != nil || exe == "" {
		exe = "openscadgen"
	}
	stdioConfig := map[string]any{
		"mcpServers": map[string]any{
			"openscadgen": map[string]any{
				"command": exe,
				"args":    args,
			},
		},
	}

	return []models.MCPClientSnippet{
		{
			Title:   "HTTP client",
			Client:  "Cursor, VS Code, MCP Inspector",
			Format:  "json",
			Content: prettyJSON(httpConfig),
		},
		{
			Title:   "Stdio client",
			Client:  "Claude Desktop, any client that spawns a process",
			Format:  "json",
			Content: prettyJSON(stdioConfig),
		},
		{
			Title:   "Bridge an HTTP-only client",
			Client:  "Clients without streamable HTTP support",
			Format:  "shell",
			Content: "npx -y mcp-remote " + endpoint,
		},
	}
}

func prettyJSON(value any) string {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

// openscadNavView probes OpenSCAD for the shared nav widget.
func openscadNavView() templates.OpenSCADNavView {
	info, err := pkg.ProbeOpenSCAD()
	if err != nil {
		return templates.OpenSCADNavView{Error: err.Error()}
	}
	return templates.OpenSCADNavView{
		Available: true,
		Path:      info.Path,
		Version:   info.Version,
	}
}

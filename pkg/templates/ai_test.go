package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestAISettingsPageRenders(t *testing.T) {
	data := models.AISettingsPageData{
		Title:           "AI & MCP",
		Enabled:         true,
		HTTPMounted:     true,
		EndpointURL:     "http://127.0.0.1:7203/mcp",
		ProgressPath:    "/api/mcp/process/{job_id}",
		StdioCommand:    `openscadgen -ai -mcp-stdio`,
		ProtocolVersion: "2025-06-18",
		DetailLevel:     "summary",
		ReadOnly:        false,
		AllowPush:       false,
		Roots:           []string{`C:\projects\alpha`},
		BindAddress:     "127.0.0.1",
		LoopbackOnly:    true,
		Sources: []models.AISettingSource{
			{Name: "Default detail level", Value: "summary", Flag: "-mcp-detail", Env: "OPENSCADGEN_MCP_DETAIL", Default: "normal"},
		},
		Tools: []models.MCPToolInfo{
			{
				Name: "read_project_file", Title: "Read project file", Category: "Files",
				Description: "Read a file.", ReadOnly: true,
				Params: []models.MCPParamInfo{{Name: "path", Type: "string", Required: true, Description: "File path."}},
			},
			{
				Name: "git_push", Title: "Git push", Category: "Git",
				Description: "Push a branch.", RequiresPush: true, Gated: true,
			},
		},
		Snippets: []models.MCPClientSnippet{
			{Title: "HTTP client", Client: "Cursor", Format: "json", Content: `{"mcpServers":{"openscadgen":{"url":"http://127.0.0.1:7203/mcp"}}}`},
		},
	}

	var buffer bytes.Buffer
	if err := AISettingsPage(data, OpenSCADNavView{}).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buffer.String()

	for _, want := range []string{
		"http://127.0.0.1:7203/mcp",
		"read_project_file",
		"git_push",
		"disabled",
		"HTTP client",
		"WHERE EACH SETTING COMES FROM",
		"-mcp-detail",
		"path: string (required)",
		"/api/mcp/process/{job_id}",
	} {
		if !strings.Contains(strings.ToUpper(html), strings.ToUpper(want)) {
			t.Fatalf("rendered page missing %q", want)
		}
	}
	if strings.Contains(html, "AI Chat") {
		t.Fatal("settings page should not mention the removed chat UI")
	}
}

func TestAISettingsPageEmptyRoots(t *testing.T) {
	data := models.AISettingsPageData{
		Title:           "AI & MCP",
		Enabled:         true,
		HTTPMounted:     true,
		EndpointURL:     "http://127.0.0.1:7203/mcp",
		ProtocolVersion: "2025-06-18",
		DetailLevel:     "normal",
		BindAddress:     "127.0.0.1",
		LoopbackOnly:    true,
		Notices:         []string{"No MCP roots resolved"},
	}
	var buffer bytes.Buffer
	if err := AISettingsPage(data, OpenSCADNavView{}).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(buffer.String(), "No MCP roots resolved") {
		t.Fatal("notice not rendered")
	}
}

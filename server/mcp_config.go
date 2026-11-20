package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

// Detail levels accepted by response_detail / -mcp-detail. The level only
// trims successful payloads; failures are always reported in full so a client
// can act on them.
const (
	mcpDetailSummary = "summary"
	mcpDetailNormal  = "normal"
	mcpDetailFull    = "full"
)

// mcpConfig is the effective AI/MCP configuration. Everything here comes from
// flags or environment variables; nothing is persisted to disk. The AI settings
// page renders exactly this struct so the page cannot drift from behaviour.
type mcpConfig struct {
	Enabled      bool
	Stdio        bool
	Detail       string
	ReadOnly     bool
	AllowPush    bool
	Roots        []string
	BindAddress  string
	ServerFolder string
	Notices      []string
}

// newMCPConfig resolves flags and their environment fallbacks into the
// effective configuration. Precedence is flag, then environment, then default,
// which mirrors the rest of the CLI.
func newMCPConfig(flags models.CmdFlags, serverFolder string) mcpConfig {
	cfg := mcpConfig{
		Enabled:      flags.ShowAI,
		Stdio:        flags.MCPStdio,
		BindAddress:  strings.TrimSpace(flags.ServerAddress),
		ServerFolder: strings.TrimSpace(serverFolder),
	}
	if cfg.Stdio {
		cfg.Enabled = true
	}
	if cfg.BindAddress == "" {
		cfg.BindAddress = "127.0.0.1"
	}

	detail := firstNonEmpty(flags.MCPDetail, os.Getenv("OPENSCADGEN_MCP_DETAIL"), mcpDetailNormal)
	if normalized, ok := normalizeMCPDetail(detail); ok {
		cfg.Detail = normalized
	} else {
		cfg.Detail = mcpDetailNormal
		cfg.Notices = append(cfg.Notices, fmt.Sprintf(
			"Unknown detail level %q ignored; using %q. Valid values: summary, normal, full.", detail, mcpDetailNormal))
	}

	cfg.ReadOnly = flags.MCPReadOnly || envBool("OPENSCADGEN_MCP_READ_ONLY")
	cfg.AllowPush = flags.MCPAllowPush || envBool("OPENSCADGEN_MCP_ALLOW_PUSH")

	rawRoots := firstNonEmpty(flags.MCPRoots, os.Getenv("OPENSCADGEN_MCP_ROOT"))
	if rawRoots == "" && cfg.ServerFolder != "" {
		rawRoots = cfg.ServerFolder
	}
	if rawRoots == "" {
		if wd, err := os.Getwd(); err == nil {
			rawRoots = wd
		}
	}

	roots, err := normalizeMCPRoots(rawRoots)
	if err != nil {
		cfg.Notices = append(cfg.Notices, err.Error())
	}
	cfg.Roots = roots
	if len(cfg.Roots) == 0 {
		cfg.Notices = append(cfg.Notices, "No MCP roots resolved; project tools will refuse every path until -mcp-root or -sf is set.")
	}

	if !isLoopbackAddress(cfg.BindAddress) {
		cfg.Notices = append(cfg.Notices, fmt.Sprintf(
			"MCP is bound to %s, which is reachable from other machines. This server has no authentication; keep the bind address on loopback or put it behind an authenticating proxy.",
			cfg.BindAddress))
	}

	return cfg
}

// normalizeMCPDetail maps aliases onto the three supported levels.
func normalizeMCPDetail(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", mcpDetailNormal:
		return mcpDetailNormal, true
	case mcpDetailSummary, "brief", "minimal", "short":
		return mcpDetailSummary, true
	case mcpDetailFull, "verbose", "all":
		return mcpDetailFull, true
	default:
		return mcpDetailNormal, false
	}
}

// normalizeMCPRoots splits a comma-separated list, makes every entry absolute,
// drops duplicates, and keeps the first-seen order.
func normalizeMCPRoots(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	roots := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	missing := make([]string, 0)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		abs, err := filepath.Abs(part)
		if err != nil {
			return roots, fmt.Errorf("could not resolve MCP root %q: %w", part, err)
		}
		abs = filepath.Clean(abs)
		key := strings.ToLower(abs)
		if seen[key] {
			continue
		}
		seen[key] = true
		if info, statErr := os.Stat(abs); statErr != nil || !info.IsDir() {
			missing = append(missing, abs)
		}
		roots = append(roots, abs)
	}

	if len(missing) > 0 {
		return roots, fmt.Errorf("MCP root is not a directory: %s", strings.Join(missing, ", "))
	}
	return roots, nil
}

// detailLevel resolves the per-call response_detail argument against the server
// default.
func (c mcpConfig) detailLevel(requested string) string {
	if level, ok := normalizeMCPDetail(requested); ok && strings.TrimSpace(requested) != "" {
		return level
	}
	if c.Detail != "" {
		return c.Detail
	}
	return mcpDetailNormal
}

// allowsTool reports whether a tool's safety gates are satisfied.
func (c mcpConfig) allowsTool(tool mcpTool) bool {
	if tool.RequiresPush && !c.AllowPush {
		return false
	}
	if !tool.ReadOnly && c.ReadOnly {
		return false
	}
	return true
}

func envBool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// isLoopbackAddress reports whether a bind address is loopback-only.
func isLoopbackAddress(address string) bool {
	switch strings.ToLower(strings.TrimSpace(address)) {
	case "", "localhost", "127.0.0.1", "::1", "[::1]":
		return true
	default:
		return false
	}
}

// mcpEndpointURL is the URL an HTTP client should connect to. Wildcard binds are
// rendered as loopback: 0.0.0.0 is not a connectable client address.
func (c mcpConfig) mcpEndpointURL(port int) string {
	host := c.BindAddress
	switch host {
	case "", "localhost", "::1", "[::1]", "0.0.0.0", "::", "[::]":
		host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%d/mcp", host, port)
}

// stdioCommand is the command a client should spawn for the stdio transport.
func (c mcpConfig) stdioCommand() string {
	exe, err := os.Executable()
	if err != nil || exe == "" {
		exe = "openscadgen"
	}
	parts := []string{quoteCommandArg(exe), "-ai", "-mcp-stdio"}
	if c.ReadOnly {
		parts = append(parts, "-mcp-read-only")
	}
	if c.AllowPush {
		parts = append(parts, "-mcp-allow-push")
	}
	if c.Detail != "" && c.Detail != mcpDetailNormal {
		parts = append(parts, "-mcp-detail", c.Detail)
	}
	if len(c.Roots) == 1 {
		parts = append(parts, "-mcp-root", quoteCommandArg(c.Roots[0]))
	} else if len(c.Roots) > 1 {
		parts = append(parts, "-mcp-root", quoteCommandArg(strings.Join(c.Roots, ",")))
	}
	return strings.Join(parts, " ")
}

func quoteCommandArg(value string) string {
	if value == "" {
		return `""`
	}
	if strings.ContainsAny(value, " \t\"") {
		return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return value
}

// describeSources renders the flag/env/default provenance table for the
// settings page.
func (c mcpConfig) describeSources() []models.AISettingSource {
	detailSource := models.AISettingSource{
		Name: "Default detail level", Value: c.Detail,
		Flag: "-mcp-detail", Env: "OPENSCADGEN_MCP_DETAIL", Default: mcpDetailNormal,
	}
	readOnly := models.AISettingSource{
		Name: "Read-only mode", Value: fmt.Sprintf("%t", c.ReadOnly),
		Flag: "-mcp-read-only", Env: "OPENSCADGEN_MCP_READ_ONLY", Default: "false",
	}
	push := models.AISettingSource{
		Name: "Allow git push", Value: fmt.Sprintf("%t", c.AllowPush),
		Flag: "-mcp-allow-push", Env: "OPENSCADGEN_MCP_ALLOW_PUSH", Default: "false",
	}
	root := models.AISettingSource{
		Name: "Tool roots", Value: strings.Join(c.Roots, ", "),
		Flag: "-mcp-root", Env: "OPENSCADGEN_MCP_ROOT", Default: "server folder, else working directory",
	}
	enabled := models.AISettingSource{
		Name: "MCP enabled", Value: fmt.Sprintf("%t", c.Enabled),
		Flag: "-ai", Env: "OPENSCADGEN_AI_ON", Default: "false",
	}
	return []models.AISettingSource{enabled, detailSource, readOnly, push, root}
}

// mcpServerVersion is the reported implementation version.
func mcpServerVersion() string {
	return pkg.GetVersion().OpenSCADGen
}

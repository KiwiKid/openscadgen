package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

const (
	defaultMCPReadBytes   = 20000
	maxMCPReadBytes       = 200000
	maxMCPEditBytes       = 4 << 20
	maxMCPListEntries     = 400
	maxMCPDiffBytes       = 24000
	maxMCPCommandOutput   = 24000
	maxMCPProcessDuration = 10 * time.Minute
)

// mcpTool is one advertised MCP tool. ReadOnly and RequiresPush are the safety
// gates; the AI settings page renders Gated so operators can see what a flag
// would unlock.
type mcpTool struct {
	Name         string
	Title        string
	Category     string
	Description  string
	ReadOnly     bool
	RequiresPush bool

	InputSchema map[string]any
	Handler     func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error)
}

// mcpResult is a detail-aware payload. Summary is what a successful call costs
// in tokens; Normal and Full are merged in as the requested detail level rises.
// Failures always render every section.
type mcpResult struct {
	OK      bool
	Summary map[string]any
	Normal  map[string]any
	Full    map[string]any
}

func (r mcpResult) render(level string) map[string]any {
	out := make(map[string]any, 8)
	for key, value := range r.Summary {
		out[key] = value
	}
	if !r.OK || level == mcpDetailNormal || level == mcpDetailFull {
		for key, value := range r.Normal {
			out[key] = value
		}
	}
	if !r.OK || level == mcpDetailFull {
		for key, value := range r.Full {
			out[key] = value
		}
	}
	out["ok"] = r.OK
	return out
}

// allMCPTools returns every tool the build supports, including gated ones. The
// settings page uses this so operators can see what is available but disabled.
func allMCPTools() []mcpTool {
	return []mcpTool{
		toolListProjects(),
		toolListProjectFiles(),
		toolReadProjectFile(),
		toolEditProjectFile(),
		toolWriteProjectFile(),
		toolProcessConfig(),
		toolProcessStatus(),
		toolGitStatus(),
		toolGitDiff(),
		toolGitCommit(),
		toolGitPush(),
	}
}

// availableMCPTools filters the registry by the effective safety configuration.
func availableMCPTools(cfg mcpConfig) []mcpTool {
	tools := make([]mcpTool, 0, len(allMCPTools()))
	for _, tool := range allMCPTools() {
		if cfg.allowsTool(tool) {
			tools = append(tools, tool)
		}
	}
	return tools
}

func findMCPTool(cfg mcpConfig, name string) (mcpTool, bool) {
	for _, tool := range availableMCPTools(cfg) {
		if tool.Name == name {
			return tool, true
		}
	}
	return mcpTool{}, false
}

// mcpToolDescriptions renders the tool list for the settings page.
func mcpToolDescriptions(cfg mcpConfig) []models.MCPToolInfo {
	infos := make([]models.MCPToolInfo, 0, len(allMCPTools()))
	for _, tool := range allMCPTools() {
		infos = append(infos, models.MCPToolInfo{
			Name:         tool.Name,
			Title:        tool.Title,
			Category:     tool.Category,
			Description:  tool.Description,
			ReadOnly:     tool.ReadOnly,
			RequiresPush: tool.RequiresPush,
			Gated:        !cfg.allowsTool(tool),
			Params:       mcpToolParams(tool.InputSchema),
		})
	}
	return infos
}

// mcpToolParams flattens a JSON schema's top-level properties for display.
func mcpToolParams(schema map[string]any) []models.MCPParamInfo {
	properties, _ := schema["properties"].(map[string]any)
	required := make(map[string]bool)
	if list, ok := schema["required"].([]string); ok {
		for _, name := range list {
			required[name] = true
		}
	}

	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)

	params := make([]models.MCPParamInfo, 0, len(names))
	for _, name := range names {
		prop, _ := properties[name].(map[string]any)
		paramType, _ := prop["type"].(string)
		description, _ := prop["description"].(string)
		if enum, ok := prop["enum"].([]string); ok && len(enum) > 0 {
			paramType = strings.Join(enum, "|")
		}
		params = append(params, models.MCPParamInfo{
			Name:        name,
			Type:        paramType,
			Required:    required[name],
			Description: description,
		})
	}
	return params
}

func mcpStringProp(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func mcpIntProp(description string, minimum int, maximum int) map[string]any {
	prop := map[string]any{"type": "integer", "description": description, "minimum": minimum}
	if maximum > 0 {
		prop["maximum"] = maximum
	}
	return prop
}

func mcpBoolProp(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}

func mcpDetailProp() map[string]any {
	return map[string]any{
		"type":        "string",
		"enum":        []string{mcpDetailSummary, mcpDetailNormal, mcpDetailFull},
		"description": "How much detail to return on success. summary keeps token cost low, normal adds diagnostics, full includes raw command output. Failures always return full detail.",
	}
}

func mcpObject(properties map[string]any, required ...string) map[string]any {
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// ---- tool definitions -------------------------------------------------------

func toolListProjects() mcpTool {
	return mcpTool{
		Name:        "list_projects",
		Title:       "List projects",
		Category:    "Discovery",
		Description: "List the OpenSCADGen project folders (folders containing config.toml) under the configured MCP roots.",
		ReadOnly:    true,
		InputSchema: mcpObject(map[string]any{
			"response_detail": mcpDetailProp(),
		}),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)
			projects := discoverProjects(cfg.Roots)

			labels := make([]string, 0, len(projects))
			for _, project := range projects {
				labels = append(labels, project.Label)
			}
			normal := make([]models.MCPProjectInfo, 0, len(projects))
			for _, project := range projects {
				normal = append(normal, projectSummaryFor(project))
			}
			configs := make([]string, 0, len(projects))
			for _, project := range projects {
				configs = append(configs, project.ConfigPath)
			}

			return mcpResult{
				OK:      true,
				Summary: map[string]any{"count": len(projects), "projects": labels},
				Normal:  map[string]any{"roots": cfg.Roots, "projects": normal},
				Full:    map[string]any{"config_paths": configs},
			}.render(level), nil
		},
	}
}

func toolListProjectFiles() mcpTool {
	return mcpTool{
		Name:        "list_project_files",
		Title:       "List project files",
		Category:    "Files",
		Description: "List files and folders under a project path. Depth is limited to 4 and the result is capped; use small depths first.",
		ReadOnly:    true,
		InputSchema: mcpObject(map[string]any{
			"path":            mcpStringProp("Path relative to the MCP root. Empty lists the root itself."),
			"depth":           mcpIntProp("How many folder levels to descend (0-4, default 2).", 0, 4),
			"response_detail": mcpDetailProp(),
		}),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				Path           string `json:"path"`
				Depth          int    `json:"depth"`
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)

			// Listing at the root of a multi-root server shows every root rather
			// than silently picking the first one.
			if strings.TrimSpace(in.Path) == "" && len(cfg.Roots) > 1 {
				roots := make([]map[string]any, 0, len(cfg.Roots))
				paths := make([]string, 0)
				truncated := false
				for _, candidateRoot := range cfg.Roots {
					entries, rootTruncated, walkErr := walkProjectFiles(candidateRoot, ".", in.Depth)
					if walkErr != nil {
						return nil, walkErr
					}
					truncated = truncated || rootTruncated
					label := mcpProjectLabel(candidateRoot, candidateRoot)
					roots = append(roots, map[string]any{"root": candidateRoot, "entries": entries})
					for _, entry := range entries {
						paths = append(paths, filepath.ToSlash(label+"/"+entry.Path))
					}
				}
				return mcpResult{
					OK:      true,
					Summary: map[string]any{"roots": len(cfg.Roots), "entries": paths, "truncated": truncated},
					Normal:  map[string]any{"listings": roots},
					Full:    map[string]any{"depth": normalizedDepth(in.Depth)},
				}.render(level), nil
			}

			abs, rel, root, err := resolveToolPath(cfg.Roots, in.Path, true)
			if err != nil {
				return nil, err
			}
			entries, truncated, err := walkProjectFiles(abs, rel, in.Depth)
			if err != nil {
				return nil, err
			}
			paths := make([]string, 0, len(entries))
			for _, entry := range entries {
				paths = append(paths, entry.Path)
			}
			return mcpResult{
				OK:      true,
				Summary: map[string]any{"path": filepath.ToSlash(rel), "entries": paths, "truncated": truncated},
				Normal:  map[string]any{"entries": entries},
				Full:    map[string]any{"root": root, "depth": normalizedDepth(in.Depth)},
			}.render(level), nil
		},
	}
}

func toolReadProjectFile() mcpTool {
	return mcpTool{
		Name:        "read_project_file",
		Title:       "Read project file",
		Category:    "Files",
		Description: "Read a UTF-8 text file from the project roots. Reads are bounded; use offset and max_bytes to page through large files.",
		ReadOnly:    true,
		InputSchema: mcpObject(map[string]any{
			"path":            mcpStringProp("File path relative to the MCP root."),
			"offset":          mcpIntProp("Byte offset to start reading from (default 0).", 0, 0),
			"max_bytes":       mcpIntProp("Maximum bytes to return (default 20000, max 200000).", 1, maxMCPReadBytes),
			"response_detail": mcpDetailProp(),
		}, "path"),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				Path           string `json:"path"`
				Offset         int    `json:"offset"`
				MaxBytes       int    `json:"max_bytes"`
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)
			abs, rel, root, err := resolveToolPath(cfg.Roots, in.Path, false)
			if err != nil {
				return nil, err
			}
			content, truncated, total, err := boundedRead(abs, in.Offset, in.MaxBytes)
			if err != nil {
				return nil, err
			}
			return mcpResult{
				OK: true,
				Summary: map[string]any{
					"path": filepath.ToSlash(rel), "content": content, "truncated": truncated,
				},
				Normal: map[string]any{"total_size": total, "offset": in.Offset},
				Full:   map[string]any{"root": root, "encoding": "utf-8"},
			}.render(level), nil
		},
	}
}

func toolEditProjectFile() mcpTool {
	return mcpTool{
		Name:        "edit_project_file",
		Title:       "Edit project file",
		Category:    "Files",
		Description: "Replace an exact text snippet in a project file. For config.toml the edited TOML is validated before it is written, so an invalid edit is refused rather than saved.",
		ReadOnly:    false,
		InputSchema: mcpObject(map[string]any{
			"path":            mcpStringProp("File path relative to the MCP root."),
			"old_text":        mcpStringProp("Exact text to replace. Must be non-empty and present in the file."),
			"new_text":        mcpStringProp("Replacement text."),
			"replace_all":     mcpBoolProp("Replace every occurrence instead of the first (default false)."),
			"validate":        mcpBoolProp("Validate config.toml content before writing (default true)."),
			"response_detail": mcpDetailProp(),
		}, "path", "old_text", "new_text"),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				Path           string `json:"path"`
				OldText        string `json:"old_text"`
				NewText        string `json:"new_text"`
				ReplaceAll     bool   `json:"replace_all"`
				Validate       *bool  `json:"validate"`
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)
			abs, rel, _, err := resolveToolPath(cfg.Roots, in.Path, false)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(in.OldText) == "" {
				return nil, fmt.Errorf("old_text must be non-empty")
			}
			info, err := os.Stat(abs)
			if err != nil {
				return nil, err
			}
			if info.IsDir() {
				return nil, fmt.Errorf("%s is a directory", filepath.ToSlash(rel))
			}
			if info.Size() > maxMCPEditBytes {
				return nil, fmt.Errorf("%s is %d bytes; edits are limited to %d bytes", filepath.ToSlash(rel), info.Size(), maxMCPEditBytes)
			}
			body, err := os.ReadFile(abs)
			if err != nil {
				return nil, err
			}
			content := string(body)
			if !strings.Contains(content, in.OldText) {
				return nil, fmt.Errorf("old_text was not found in %s", filepath.ToSlash(rel))
			}
			replacements := 1
			if in.ReplaceAll {
				replacements = -1
			}
			updated := strings.Replace(content, in.OldText, in.NewText, replacements)
			if updated == content {
				return nil, fmt.Errorf("edit produced no change for %s", filepath.ToSlash(rel))
			}

			validate := in.Validate == nil || *in.Validate
			warning := ""
			if validate && isConfigToml(rel) {
				if warning, err = validateConfigContent(cfg, abs, []byte(updated)); err != nil {
					return mcpResult{
						OK: false,
						Summary: map[string]any{
							"path": filepath.ToSlash(rel), "error": "validation failed: " + err.Error(),
						},
					}.render(mcpDetailFull), nil
				}
			}

			if err := os.WriteFile(abs, []byte(updated), info.Mode().Perm()); err != nil {
				return nil, err
			}
			if isConfigToml(rel) {
				if err := pkg.RecordUpdateJournal(filepath.Dir(abs), filepath.Base(abs), "file edited", "file saved",
					fmt.Sprintf("tool=mcp/edit_project_file path=%s replace_all=%t", filepath.ToSlash(rel), in.ReplaceAll), true); err != nil {
					pkg.LogWarnf("update journal write failed: %v", err)
				}
			}

			count := strings.Count(content, in.OldText)
			if !in.ReplaceAll && count > 1 {
				count = 1
			}
			return mcpResult{
				OK: true,
				Summary: map[string]any{
					"path": filepath.ToSlash(rel), "replacements": count,
				},
				Normal: map[string]any{"bytes_written": len(updated), "validated": validate && isConfigToml(rel)},
				Full:   map[string]any{"validation_warning": warning, "size_before": len(content)},
			}.render(level), nil
		},
	}
}

func toolWriteProjectFile() mcpTool {
	return mcpTool{
		Name:        "write_project_file",
		Title:       "Write project file",
		Category:    "Files",
		Description: "Create or fully overwrite a text file inside the project roots. For config.toml the content is validated before it is written.",
		ReadOnly:    false,
		InputSchema: mcpObject(map[string]any{
			"path":            mcpStringProp("File path relative to the MCP root. Parent folders are created."),
			"content":         mcpStringProp("Full file content."),
			"validate":        mcpBoolProp("Validate config.toml content before writing (default true)."),
			"response_detail": mcpDetailProp(),
		}, "path", "content"),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				Path           string `json:"path"`
				Content        string `json:"content"`
				Validate       *bool  `json:"validate"`
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)
			abs, rel, _, err := resolveToolPath(cfg.Roots, in.Path, false)
			if err != nil {
				return nil, err
			}

			validate := in.Validate == nil || *in.Validate
			warning := ""
			if validate && isConfigToml(rel) {
				if warning, err = validateConfigContent(cfg, abs, []byte(in.Content)); err != nil {
					return mcpResult{
						OK: false,
						Summary: map[string]any{
							"path": filepath.ToSlash(rel), "error": "validation failed: " + err.Error(),
						},
					}.render(mcpDetailFull), nil
				}
			}

			if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
				return nil, err
			}
			mode := fs.FileMode(0o644)
			_, statErr := os.Stat(abs)
			created := os.IsNotExist(statErr)
			if statErr == nil {
				if info, err := os.Stat(abs); err == nil {
					mode = info.Mode().Perm()
				}
			}
			if err := os.WriteFile(abs, []byte(in.Content), mode); err != nil {
				return nil, err
			}
			if isConfigToml(rel) {
				if err := pkg.RecordUpdateJournal(filepath.Dir(abs), filepath.Base(abs), "file written", "file saved",
					fmt.Sprintf("tool=mcp/write_project_file path=%s", filepath.ToSlash(rel)), true); err != nil {
					pkg.LogWarnf("update journal write failed: %v", err)
				}
			}

			return mcpResult{
				OK: true,
				Summary: map[string]any{
					"path": filepath.ToSlash(rel), "bytes_written": len(in.Content), "created": created,
				},
				Normal: map[string]any{"validated": validate && isConfigToml(rel)},
				Full:   map[string]any{"validation_warning": warning},
			}.render(level), nil
		},
	}
}

func toolProcessConfig() mcpTool {
	return mcpTool{
		Name:        "process_config",
		Title:       "Process config",
		Category:    "Render",
		Description: "Validate a config.toml and start OpenSCAD generation. Validation and config loading are synchronous; rendering runs as a background job and the result returns a job_id and progress_url. Poll progress_url (HTTP) or call process_status (any transport) for start time, elapsed time, the instance being rendered, and completed/total counts. Set wait_seconds to block for the final result instead.",
		ReadOnly:    false,
		InputSchema: mcpObject(map[string]any{
			"path":            mcpStringProp("Path to config.toml, relative to the MCP root."),
			"validate":        mcpBoolProp("Run the validation stage before rendering (default true)."),
			"wait_seconds":    mcpIntProp("Block up to this many seconds for the render to finish (default 0, max 300).", 0, maxMCPProcessWait),
			"response_detail": mcpDetailProp(),
		}, "path"),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				Path           string `json:"path"`
				Validate       *bool  `json:"validate"`
				WaitSeconds    int    `json:"wait_seconds"`
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)
			abs, rel, root, err := resolveToolPath(cfg.Roots, in.Path, false)
			if err != nil {
				return nil, err
			}
			if !isConfigToml(rel) {
				return nil, fmt.Errorf("path %q is not a config.toml file", filepath.ToSlash(rel))
			}

			validate := in.Validate == nil || *in.Validate
			warning := ""
			if validate {
				body, readErr := os.ReadFile(abs)
				if readErr != nil {
					return nil, readErr
				}
				if warning, err = validateConfigContent(cfg, abs, body); err != nil {
					return mcpResult{
						OK: false,
						Summary: map[string]any{
							"config_path":  filepath.ToSlash(rel),
							"config_valid": false,
							"error":        err.Error(),
						},
					}.render(mcpDetailFull), nil
				}
			}

			config, warn, loadErr := pkg.LoadConfigFromFile(models.CmdFlags{
				ConfigFile:   abs,
				Server:       true,
				ServerFolder: cfg.ServerFolder,
			})
			if loadErr != nil {
				return mcpResult{
					OK: false,
					Summary: map[string]any{
						"config_path":  filepath.ToSlash(rel),
						"config_valid": false,
						"error":        loadErr.Error(),
					},
				}.render(mcpDetailFull), nil
			}
			if warning == "" && warn != nil {
				warning = warn.Error()
			}

			job := newProcessJob(filepath.ToSlash(rel), root, level)
			go runProcessJob(job, config, root, warning)
			waitForProcessStart(job, 300*time.Millisecond)

			waitSeconds := in.WaitSeconds
			if waitSeconds < 0 {
				waitSeconds = 0
			}
			if waitSeconds > maxMCPProcessWait {
				waitSeconds = maxMCPProcessWait
			}
			waitForProcessJob(job, time.Duration(waitSeconds)*time.Second)

			payload := withProgressLinks(job.snapshot(level))
			payload["config_valid"] = true
			return payload, nil
		},
	}
}

func toolProcessStatus() mcpTool {
	return mcpTool{
		Name:        "process_status",
		Title:       "Process status",
		Category:    "Render",
		Description: "Report progress for a process_config job: start time, elapsed time, the instance being rendered, and completed/total counts. When the job has finished the final render result is included. Jobs are kept for 15 minutes after completion.",
		ReadOnly:    true,
		InputSchema: mcpObject(map[string]any{
			"job_id":          mcpStringProp("The job_id returned by process_config."),
			"response_detail": mcpDetailProp(),
		}, "job_id"),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				JobID          string `json:"job_id"`
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			job, ok := lookupProcessJob(in.JobID)
			if !ok {
				return nil, fmt.Errorf("unknown or expired process job %q; jobs are kept for %d minutes after completion", in.JobID, int(mcpProcessJobRetention.Minutes()))
			}
			return withProgressLinks(job.snapshot(cfg.detailLevel(in.ResponseDetail))), nil
		},
	}
}

func toolGitStatus() mcpTool {
	return mcpTool{
		Name:        "git_status",
		Title:       "Git status",
		Category:    "Git",
		Description: "Show git status for the project root, scoped to that root.",
		ReadOnly:    true,
		InputSchema: mcpObject(map[string]any{
			"response_detail": mcpDetailProp(),
		}),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)
			root, err := gitRoot(cfg.Roots)
			if err != nil {
				return nil, err
			}
			output, err := runGitClean(ctx, root, 20*time.Second, "status", "--short", "--branch", "--", ".")
			if err != nil {
				return nil, err
			}
			branch, changed := parseGitStatus(output)
			return mcpResult{
				OK:      true,
				Summary: map[string]any{"branch": branch, "changed_files": changed},
				Normal:  map[string]any{"output": output},
				Full:    map[string]any{"root": root},
			}.render(level), nil
		},
	}
}

func toolGitDiff() mcpTool {
	return mcpTool{
		Name:        "git_diff",
		Title:       "Git diff",
		Category:    "Git",
		Description: "Show the working-tree or staged diff for the project root, optionally limited to one path. summary returns the stat line; normal returns the patch.",
		ReadOnly:    true,
		InputSchema: mcpObject(map[string]any{
			"path":            mcpStringProp("Optional path relative to the project root."),
			"cached":          mcpBoolProp("Show the staged diff instead of the working tree (default false)."),
			"response_detail": mcpDetailProp(),
		}),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				Path           string `json:"path"`
				Cached         bool   `json:"cached"`
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)
			root, err := gitRoot(cfg.Roots)
			if err != nil {
				return nil, err
			}
			target := "."
			if strings.TrimSpace(in.Path) != "" {
				_, rel, _, resolveErr := resolveToolPath(cfg.Roots, in.Path, true)
				if resolveErr != nil {
					return nil, resolveErr
				}
				target = rel
			}

			baseArgs := []string{"diff"}
			if in.Cached {
				baseArgs = append(baseArgs, "--cached")
			}
			statArgs := append(append([]string{}, baseArgs...), "--stat", "--", target)
			stat, err := runGitClean(ctx, root, 20*time.Second, statArgs...)
			if err != nil {
				return nil, err
			}
			nameArgs := append(append([]string{}, baseArgs...), "--name-only", "--", target)
			names, err := runGitClean(ctx, root, 20*time.Second, nameArgs...)
			if err != nil {
				return nil, err
			}
			files := splitNonEmpty(names)

			result := mcpResult{
				OK:      true,
				Summary: map[string]any{"files_changed": len(files), "stat": stat},
				Normal:  map[string]any{"files": files},
				Full:    map[string]any{"cached": in.Cached, "root": root},
			}
			if level != mcpDetailSummary {
				patchArgs := append(append([]string{}, baseArgs...), "--", target)
				patch, err := runGitClean(ctx, root, 20*time.Second, patchArgs...)
				if err != nil {
					return nil, err
				}
				trimmed, truncated := trimMCPOutput(patch, maxMCPDiffBytes)
				result.Normal["patch"] = trimmed
				result.Normal["truncated"] = truncated
			}
			return result.render(level), nil
		},
	}
}

func toolGitCommit() mcpTool {
	return mcpTool{
		Name:        "git_commit",
		Title:       "Git commit",
		Category:    "Git",
		Description: "Stage and commit changes inside the project root with the supplied message. Without paths it stages tracked modifications only; pass paths to include new files. Refuses to run when changes are staged outside the root.",
		ReadOnly:    false,
		InputSchema: mcpObject(map[string]any{
			"message":         mcpStringProp("Commit message. Must be non-empty."),
			"paths":           map[string]any{"type": "array", "items": mcpStringProp("Path relative to the project root."), "description": "Optional paths to stage. Defaults to everything under the project root."},
			"response_detail": mcpDetailProp(),
		}, "message"),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				Message        string   `json:"message"`
				Paths          []string `json:"paths"`
				ResponseDetail string   `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)
			message := strings.TrimSpace(in.Message)
			if message == "" {
				return nil, fmt.Errorf("commit message must be non-empty")
			}
			root, err := gitRoot(cfg.Roots)
			if err != nil {
				return nil, err
			}
			if err := ensureNoStagedChangesOutsideRoot(ctx, root); err != nil {
				return nil, err
			}

			// Default to tracked changes only: sweeping untracked files into an
			// agent-authored commit is surprising. Callers that want new files
			// committed pass them explicitly in paths.
			addArgs := []string{"add", "-u", "--"}
			if len(in.Paths) == 0 {
				addArgs = append(addArgs, ".")
			} else {
				addArgs = []string{"add", "-A", "--"}
				for _, requestedPath := range in.Paths {
					_, rel, _, resolveErr := resolveToolPath(cfg.Roots, requestedPath, false)
					if resolveErr != nil {
						return nil, resolveErr
					}
					addArgs = append(addArgs, rel)
				}
			}
			if _, err := runGit(ctx, root, 20*time.Second, addArgs...); err != nil {
				return nil, err
			}
			if err := ensureNoStagedChangesOutsideRoot(ctx, root); err != nil {
				return nil, err
			}

			staged, err := runGitClean(ctx, root, 10*time.Second, "diff", "--cached", "--name-only", "--", ".")
			if err != nil {
				return nil, err
			}
			files := splitNonEmpty(staged)
			if len(files) == 0 {
				return nil, fmt.Errorf("no staged project changes to commit")
			}

			commitOutput, err := runGit(ctx, root, 30*time.Second, "commit", "-m", message)
			if err != nil {
				return nil, err
			}
			head, err := runGitClean(ctx, root, 10*time.Second, "rev-parse", "HEAD")
			if err != nil {
				return nil, err
			}

			return mcpResult{
				OK: true,
				Summary: map[string]any{
					"commit": strings.TrimSpace(head), "files_committed": len(files),
				},
				Normal: map[string]any{"message": message, "output": commitOutput},
				Full:   map[string]any{"files": files, "root": root},
			}.render(level), nil
		},
	}
}

func toolGitPush() mcpTool {
	return mcpTool{
		Name:         "git_push",
		Title:        "Git push",
		Category:     "Git",
		Description:  "Push the current branch, or the supplied remote and branch. Only advertised when -mcp-allow-push is set.",
		ReadOnly:     false,
		RequiresPush: true,
		InputSchema: mcpObject(map[string]any{
			"remote":          mcpStringProp("Remote name (default: the branch's upstream)."),
			"branch":          mcpStringProp("Branch name (default: the current branch)."),
			"set_upstream":    mcpBoolProp("Pass -u and set the upstream (requires remote and branch)."),
			"response_detail": mcpDetailProp(),
		}),
		Handler: func(ctx context.Context, cfg mcpConfig, args json.RawMessage) (map[string]any, error) {
			var in struct {
				Remote         string `json:"remote"`
				Branch         string `json:"branch"`
				SetUpstream    bool   `json:"set_upstream"`
				ResponseDetail string `json:"response_detail"`
			}
			if err := decodeMCPArgs(args, &in); err != nil {
				return nil, err
			}
			level := cfg.detailLevel(in.ResponseDetail)
			if !cfg.AllowPush {
				return nil, fmt.Errorf("git_push is disabled; start the server with -mcp-allow-push to enable it")
			}
			remote := strings.TrimSpace(in.Remote)
			branch := strings.TrimSpace(in.Branch)
			if in.SetUpstream && (remote == "" || branch == "") {
				return nil, fmt.Errorf("set_upstream requires both remote and branch")
			}
			if remote == "" && branch != "" {
				return nil, fmt.Errorf("branch cannot be set without remote")
			}
			root, err := gitRoot(cfg.Roots)
			if err != nil {
				return nil, err
			}

			pushArgs := []string{"push"}
			if in.SetUpstream {
				pushArgs = append(pushArgs, "-u")
			}
			if remote != "" {
				pushArgs = append(pushArgs, remote)
			}
			if branch != "" {
				pushArgs = append(pushArgs, branch)
			}

			output, err := runGit(ctx, root, 2*time.Minute, pushArgs...)
			if err != nil {
				return nil, err
			}
			return mcpResult{
				OK:      true,
				Summary: map[string]any{"remote": remote, "branch": branch},
				Normal:  map[string]any{"output": output, "upstream": in.SetUpstream},
				Full:    map[string]any{"root": root},
			}.render(level), nil
		},
	}
}

// ---- helpers ----------------------------------------------------------------

func decodeMCPArgs(raw json.RawMessage, dest any) error {
	if len(strings.TrimSpace(string(raw))) == 0 || strings.TrimSpace(string(raw)) == "null" {
		raw = json.RawMessage("{}")
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return fmt.Errorf("invalid tool arguments: %w", err)
	}
	return nil
}

func isConfigToml(relPath string) bool {
	base := filepath.Base(relPath)
	return strings.EqualFold(filepath.Ext(relPath), ".toml") || strings.EqualFold(base, "config.toml")
}

// validateConfigContent parses and validates TOML content without touching
// disk. It returns any non-fatal warning plus a fatal error.
func validateConfigContent(cfg mcpConfig, absPath string, content []byte) (string, error) {
	_, warn, err := pkg.LoadConfig(string(content), models.CmdFlags{
		ConfigFile:   absPath,
		Server:       true,
		ServerFolder: cfg.ServerFolder,
	}, absPath)
	if err != nil {
		return "", err
	}
	if warn != nil {
		return warn.Error(), nil
	}
	return "", nil
}

func normalizedDepth(depth int) int {
	if depth <= 0 {
		return 2
	}
	if depth > 4 {
		return 4
	}
	return depth
}

type mcpFileEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

// walkProjectFiles lists entries under abs without following symlinked folders.
func walkProjectFiles(abs string, rel string, depth int) ([]mcpFileEntry, bool, error) {
	depth = normalizedDepth(depth)
	entries := make([]mcpFileEntry, 0)
	truncated := false

	var walk func(currentAbs string, currentRel string, remaining int) error
	walk = func(currentAbs string, currentRel string, remaining int) error {
		dirEntries, err := os.ReadDir(currentAbs)
		if err != nil {
			return err
		}
		sort.Slice(dirEntries, func(i, j int) bool {
			if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
				return dirEntries[i].Name() < dirEntries[j].Name()
			}
			return dirEntries[i].IsDir()
		})
		for _, entry := range dirEntries {
			if len(entries) >= maxMCPListEntries {
				truncated = true
				return nil
			}
			if entry.IsDir() && shouldSkipDir(entry.Name()) {
				continue
			}
			childRel := entry.Name()
			if currentRel != "." {
				childRel = filepath.Join(currentRel, childRel)
			}
			entryType := "file"
			if entry.IsDir() {
				entryType = "dir"
			}
			entries = append(entries, mcpFileEntry{Path: filepath.ToSlash(childRel), Type: entryType})
			if entry.IsDir() && remaining > 0 && !isSymlink(entry) {
				if err := walk(filepath.Join(currentAbs, entry.Name()), childRel, remaining-1); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := walk(abs, rel, depth-1); err != nil {
		return nil, false, err
	}
	return entries, truncated, nil
}

// boundedRead reads at most maxBytes from offset without loading the whole file.
func boundedRead(abs string, offset int, maxBytes int) (string, bool, int64, error) {
	file, err := os.Open(abs)
	if err != nil {
		return "", false, 0, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", false, 0, err
	}
	if info.IsDir() {
		return "", false, 0, fmt.Errorf("%s is a directory", abs)
	}
	switch {
	case maxBytes <= 0:
		maxBytes = defaultMCPReadBytes
	case maxBytes > maxMCPReadBytes:
		maxBytes = maxMCPReadBytes
	}
	if offset < 0 {
		offset = 0
	}
	if int64(offset) > info.Size() {
		offset = int(info.Size())
	}
	if _, err := file.Seek(int64(offset), io.SeekStart); err != nil {
		return "", false, 0, err
	}
	buf, err := io.ReadAll(io.LimitReader(file, int64(maxBytes)))
	if err != nil {
		return "", false, 0, err
	}
	truncated := int64(offset)+int64(len(buf)) < info.Size()
	return string(buf), truncated, info.Size(), nil
}

// runGit runs a git command in dir with prompts disabled so a chat-less client
// never hangs on credentials. Output is stdout+stderr, which is what callers
// that surface raw git text to a client want.
func runGit(ctx context.Context, dir string, timeout time.Duration, args ...string) (string, error) {
	output, stderr, err := runGitSplit(ctx, dir, timeout, args...)
	if err != nil {
		return output, err
	}
	if strings.TrimSpace(stderr) != "" {
		output = strings.TrimSpace(output + "\n" + stderr)
	}
	return output, nil
}

// runGitClean returns stdout only. Use it whenever the output is parsed
// (status, name-only, rev-parse) or shown as a patch: git writes cosmetic
// warnings such as CRLF conversion notices to stderr, and those must not be
// mistaken for filenames or diff content.
func runGitClean(ctx context.Context, dir string, timeout time.Duration, args ...string) (string, error) {
	output, _, err := runGitSplit(ctx, dir, timeout, args...)
	return output, err
}

// runGitSplit runs git with separate stdout/stderr capture.
func runGitSplit(ctx context.Context, dir string, timeout time.Duration, args ...string) (string, string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"GIT_OPTIONAL_LOCKS=0",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	output := strings.TrimSpace(stdout.String())
	errOutput := strings.TrimSpace(stderr.String())
	output, _ = trimMCPOutput(output, maxMCPCommandOutput)
	errOutput, _ = trimMCPOutput(errOutput, maxMCPCommandOutput)

	if cmdCtx.Err() == context.DeadlineExceeded {
		return output, errOutput, fmt.Errorf("git %s timed out after %s", strings.Join(args, " "), timeout)
	}
	if err != nil {
		detail := errOutput
		if detail == "" {
			detail = output
		}
		if detail == "" {
			detail = err.Error()
		}
		return output, errOutput, fmt.Errorf("git %s failed: %s", strings.Join(args, " "), detail)
	}
	return output, errOutput, nil
}

func trimMCPOutput(output string, limit int) (string, bool) {
	if limit <= 0 || len(output) <= limit {
		return output, false
	}
	return output[:limit] + "\n... (truncated)", true
}

func splitNonEmpty(output string) []string {
	parts := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			parts = append(parts, filepath.ToSlash(trimmed))
		}
	}
	return parts
}

func parseGitStatus(output string) (string, int) {
	branch := ""
	changed := 0
	for index, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if index == 0 && strings.HasPrefix(line, "## ") {
			branch = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}
		changed++
	}
	return branch, changed
}

func ensureNoStagedChangesOutsideRoot(ctx context.Context, root string) error {
	repoRoot, err := runGitClean(ctx, root, 10*time.Second, "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	repoRoot = strings.TrimSpace(repoRoot)

	projectRel, err := filepath.Rel(repoRoot, root)
	if err != nil {
		return err
	}
	projectRel = filepath.ToSlash(filepath.Clean(projectRel))

	staged, err := runGitClean(ctx, root, 10*time.Second, "diff", "--cached", "--name-only")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(staged, "\n") {
		line = filepath.ToSlash(strings.TrimSpace(line))
		if line == "" || projectRel == "." {
			continue
		}
		if line != projectRel && !strings.HasPrefix(line, projectRel+"/") {
			return fmt.Errorf("staged changes outside the current project root would be included in the commit: %s", line)
		}
	}
	return nil
}

// buildProcessPayload converts a ProcessResult into a detail-aware result.
func buildProcessPayload(root string, result models.ProcessResult, processErr error, warning string) mcpResult {
	totalInstances := len(result.Instances)
	successful := 0
	failed := 0
	skipped := 0
	for _, instance := range result.Instances {
		if strings.TrimSpace(instance.SkippedReason) != "" {
			skipped++
			continue
		}
		if instance.IsSuccessful && strings.TrimSpace(instance.ConfigError) == "" {
			successful++
			continue
		}
		if strings.TrimSpace(instance.ConfigError) != "" || !instance.IsSuccessful {
			failed++
		}
	}

	artifacts := make([]string, 0, len(result.STLResults)+len(result.ImageResults))
	missingArtifacts := make([]string, 0)
	for _, stl := range result.STLResults {
		if stl.Skipped {
			continue
		}
		absPath := stlOutputPath(result, stl)
		if absPath == "" {
			continue
		}
		rel := relativeToRoot(root, absPath)
		artifacts = append(artifacts, rel)
		if info, err := os.Stat(absPath); err != nil || info.IsDir() || info.Size() == 0 {
			missingArtifacts = append(missingArtifacts, rel)
		}
	}
	for _, image := range result.ImageResults {
		if strings.TrimSpace(image.OutputPath) == "" || image.Skipped {
			continue
		}
		rel := relativeToRoot(root, image.OutputPath)
		artifacts = append(artifacts, rel)
		if info, err := os.Stat(image.OutputPath); err != nil || info.IsDir() || info.Size() == 0 {
			missingArtifacts = append(missingArtifacts, rel)
		}
	}

	issueList := make([]map[string]any, 0)
	for _, instance := range result.Instances {
		if strings.TrimSpace(instance.SkippedReason) != "" {
			issueList = append(issueList, map[string]any{
				"instance": instance.AutoName, "type": "skipped", "message": instance.SkippedReason,
			})
			continue
		}
		if strings.TrimSpace(instance.ConfigError) != "" {
			issueList = append(issueList, map[string]any{
				"instance": instance.AutoName, "type": "render_error", "message": instance.ConfigError,
				"output_excerpt": instanceOutputExcerpt(instance),
			})
		}
	}

	summary := map[string]any{
		"instances":         totalInstances,
		"successful":        successful,
		"failed":            failed,
		"skipped":           skipped,
		"artifacts":         len(artifacts),
		"missing_artifacts": len(missingArtifacts),
		"duration_ms":       result.TotalTimeTaken.Milliseconds(),
	}

	// Success means the run finished *and* produced something. A run where every
	// instance failed, or where expected artifacts are missing, must not report
	// ok=true just because the process returned nil.
	ok := processErr == nil && failed == 0 && successful > 0 && len(missingArtifacts) == 0
	switch {
	case processErr != nil:
		summary["error"] = processErr.Error()
	case totalInstances == 0:
		summary["error"] = "no instances matched the config"
	case successful == 0 && failed > 0:
		summary["error"] = "all instances failed to render"
	case failed > 0:
		summary["error"] = "some instances failed to render"
	case len(missingArtifacts) > 0:
		summary["error"] = "expected artifacts are missing or empty"
	}
	if len(missingArtifacts) > 0 {
		summary["missing"] = missingArtifacts
	}

	perInstance := make([]map[string]any, 0, totalInstances)
	for _, instance := range result.Instances {
		perInstance = append(perInstance, map[string]any{
			"name":       instance.AutoName,
			"successful": instance.IsSuccessful,
			"skipped":    strings.TrimSpace(instance.SkippedReason) != "",
			"complete":   instance.IsComplete,
		})
	}

	return mcpResult{
		OK:      ok,
		Summary: summary,
		Normal: map[string]any{
			"artifact_paths": artifacts,
			"issues":         issueList,
			"warning":        warning,
		},
		Full: map[string]any{
			"instances":       perInstance,
			"export_location": relativeToRoot(root, result.ExportLocation),
			"config_file":     relativeToRoot(root, result.ConfigFile),
		},
	}
}

func instanceOutputExcerpt(instance models.InstanceConfig) string {
	for _, stl := range instance.STLResults {
		if trimmed := strings.TrimSpace(stl.OutputLog); trimmed != "" {
			excerpt, _ := trimMCPOutput(trimmed, 2000)
			return excerpt
		}
		if trimmed := strings.TrimSpace(stl.CommandOutput); trimmed != "" {
			excerpt, _ := trimMCPOutput(trimmed, 2000)
			return excerpt
		}
	}
	return ""
}

func relativeToRoot(root string, path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

// stlOutputPath resolves the file a render actually wrote.
// GenerateSTLResult.OutputPath is not populated by the processor, so fall back
// to the instance run path, which is relative to the config file's folder.
func stlOutputPath(result models.ProcessResult, stl models.GenerateSTLResult) string {
	if path := strings.TrimSpace(stl.OutputPath); path != "" {
		return path
	}
	path := strings.TrimSpace(stl.InstanceConfig.RunOutputPathV3)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(filepath.Dir(result.ConfigFile), path)
}

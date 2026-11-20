package server

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func flagsForMCPTest(root string) models.CmdFlags {
	return models.CmdFlags{ShowAI: true, MCPRoots: root, ServerAddress: "127.0.0.1", ServerPort: 7203}
}

const testValidConfig = `[openscadgen]
name = "demo"
version = "v0.1"
export_name_format = "{designFileName}"

[[openscadgen.input_paths]]
path = "./demo.scad"

[[openscadgen.images]]
name = "nice"
`

func toolPayload(t *testing.T, server *mcpServer, name string, args map[string]any) map[string]any {
	t.Helper()
	response := callTool(t, server, name, args)
	if response.Error != nil {
		t.Fatalf("%s protocol error: %+v", name, response.Error)
	}
	var result struct {
		Structured map[string]any `json:"structuredContent"`
		Content    []struct {
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("decode %s result: %v", name, err)
	}
	if result.Structured == nil {
		if len(result.Content) > 0 {
			if err := json.Unmarshal([]byte(result.Content[0].Text), &result.Structured); err != nil {
				t.Fatalf("decode %s text payload: %v (%s)", name, err, result.Content[0].Text)
			}
		}
	}
	return result.Structured
}

func TestResolveToolPathRejectsEscapes(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	outside := filepath.Join(base, "outside")
	for _, dir := range []string{root, outside, filepath.Join(root, "sub")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	if _, rel, _, err := resolveToolPath([]string{root}, "sub/file.txt", false); err != nil || rel != filepath.Join("sub", "file.txt") {
		t.Fatalf("relative path rejected: rel=%q err=%v", rel, err)
	}
	if _, _, _, err := resolveToolPath([]string{root}, "../outside/secret.txt", false); err == nil {
		t.Fatal("expected .. escape to be refused")
	}
	if _, _, _, err := resolveToolPath([]string{root}, filepath.Join(outside, "secret.txt"), false); err == nil {
		t.Fatal("expected absolute outside path to be refused")
	}
	if _, _, _, err := resolveToolPath([]string{root}, filepath.Join(root, "sub", "ok.txt"), false); err != nil {
		t.Fatalf("absolute inside path refused: %v", err)
	}
	if _, _, _, err := resolveToolPath(nil, "anything", false); err == nil {
		t.Fatal("expected missing roots to be refused")
	}
}

func TestResolveToolPathRejectsSymlinkEscape(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	outside := filepath.Join(base, "outside")
	for _, dir := range []string{root, outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}

	if _, _, _, err := resolveToolPath([]string{root}, "link/secret.txt", false); err == nil {
		t.Fatal("expected symlinked escape to be refused")
	}
}

func TestResolveToolPathRejectsGitInternals(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "config"), []byte("[core]"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	for _, path := range []string{".git/config", ".git/hooks/pre-commit"} {
		if _, _, _, err := resolveToolPath([]string{root}, path, false); err == nil {
			t.Fatalf("expected %q to be refused", path)
		}
	}
}

func TestBoundedReadClampsMaxBytes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "big.txt")
	if err := os.WriteFile(path, []byte(strings.Repeat("a", maxMCPReadBytes+100)), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	content, _, _, err := boundedRead(path, 0, maxMCPReadBytes*10)
	if err != nil {
		t.Fatalf("boundedRead: %v", err)
	}
	if len(content) != maxMCPReadBytes {
		t.Fatalf("content length = %d, want %d", len(content), maxMCPReadBytes)
	}
}

func TestBoundedReadLimitsAndTruncates(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "big.txt")
	body := strings.Repeat("a", 50000)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	content, truncated, total, err := boundedRead(path, 10, 100)
	if err != nil {
		t.Fatalf("boundedRead: %v", err)
	}
	if len(content) != 100 {
		t.Fatalf("content length = %d, want 100", len(content))
	}
	if !truncated {
		t.Fatal("expected truncated=true")
	}
	if total != 50000 {
		t.Fatalf("total = %d", total)
	}

	tail, tailTruncated, _, err := boundedRead(path, 49990, 100)
	if err != nil {
		t.Fatalf("boundedRead tail: %v", err)
	}
	if len(tail) != 10 || tailTruncated {
		t.Fatalf("tail length = %d truncated = %v", len(tail), tailTruncated)
	}
}

func TestMCPResultDetailRendering(t *testing.T) {
	result := mcpResult{
		OK:      true,
		Summary: map[string]any{"a": 1},
		Normal:  map[string]any{"b": 2},
		Full:    map[string]any{"c": 3},
	}
	summary := result.render(mcpDetailSummary)
	if _, ok := summary["b"]; ok {
		t.Fatal("summary leaked normal detail")
	}
	if _, ok := summary["c"]; ok {
		t.Fatal("summary leaked full detail")
	}
	normal := result.render(mcpDetailNormal)
	if normal["b"] != 2 {
		t.Fatal("normal missing detail")
	}
	if _, ok := normal["c"]; ok {
		t.Fatal("normal leaked full detail")
	}
	full := result.render(mcpDetailFull)
	if full["c"] != 3 {
		t.Fatal("full missing detail")
	}

	failed := mcpResult{OK: false, Summary: map[string]any{"a": 1}, Normal: map[string]any{"b": 2}, Full: map[string]any{"c": 3}}
	rendered := failed.render(mcpDetailSummary)
	if rendered["b"] != 2 || rendered["c"] != 3 {
		t.Fatal("failures must always render every section")
	}
}

func TestReadToolHonoursResponseDetail(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "note.txt"), []byte("body"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	server := newMCPServer(testMCPConfig(t, root))

	summary := toolPayload(t, server, "read_project_file", map[string]any{"path": "note.txt", "response_detail": "summary"})
	if _, ok := summary["root"]; ok {
		t.Fatal("summary response should not include root")
	}
	if summary["content"] != "body" {
		t.Fatalf("summary content = %v", summary["content"])
	}

	full := toolPayload(t, server, "read_project_file", map[string]any{"path": "note.txt", "response_detail": "full"})
	if full["root"] != root {
		t.Fatalf("full response root = %v", full["root"])
	}
	if full["total_size"] == nil {
		t.Fatal("full response missing total_size")
	}
}

func TestWriteConfigRefusesInvalidTOML(t *testing.T) {
	root := t.TempDir()
	server := newMCPServer(testMCPConfig(t, root))

	payload := toolPayload(t, server, "write_project_file", map[string]any{
		"path":    "config.toml",
		"content": "[openscadgen\nname = ",
	})
	if payload["ok"] != false {
		t.Fatalf("invalid config accepted: %v", payload)
	}
	if !strings.Contains(fmt.Sprint(payload["error"]), "validation failed") {
		t.Fatalf("error should mention validation: %v", payload["error"])
	}
	if _, err := os.Stat(filepath.Join(root, "config.toml")); !os.IsNotExist(err) {
		t.Fatal("invalid config must not be written")
	}
}

func TestWriteAndEditConfig(t *testing.T) {
	root := t.TempDir()
	server := newMCPServer(testMCPConfig(t, root))

	written := toolPayload(t, server, "write_project_file", map[string]any{
		"path":    "config.toml",
		"content": testValidConfig,
	})
	if written["ok"] != true {
		t.Fatalf("valid config refused: %v", written)
	}
	if written["created"] != true {
		t.Fatalf("created = %v", written["created"])
	}

	edited := toolPayload(t, server, "edit_project_file", map[string]any{
		"path":     "config.toml",
		"old_text": `name = "demo"`,
		"new_text": `name = "renamed"`,
	})
	if edited["ok"] != true {
		t.Fatalf("edit refused: %v", edited)
	}
	body, err := os.ReadFile(filepath.Join(root, "config.toml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(body), `name = "renamed"`) {
		t.Fatalf("edit not applied:\n%s", body)
	}
}

func TestProcessConfigReportsInvalidConfigWithoutRendering(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.toml"), []byte("not toml at all"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	server := newMCPServer(testMCPConfig(t, root))

	payload := toolPayload(t, server, "process_config", map[string]any{"path": "config.toml"})
	if payload["ok"] != false {
		t.Fatalf("expected failure payload: %v", payload)
	}
	if payload["config_valid"] != false {
		t.Fatalf("config_valid = %v", payload["config_valid"])
	}
}

func TestListProjectsDiscoversConfigs(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "cup")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project, "config.toml"), []byte(testValidConfig), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "notes.toml"), []byte("name = \"not a config\"\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	server := newMCPServer(testMCPConfig(t, root))

	payload := toolPayload(t, server, "list_projects", map[string]any{})
	if payload["count"] != float64(1) {
		t.Fatalf("count = %v (%v)", payload["count"], payload["projects"])
	}
}

func TestGitTools(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("config", "commit.gpgsign", "false")

	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("one"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "initial")

	server := newMCPServer(testMCPConfig(t, root))

	clean := toolPayload(t, server, "git_status", map[string]any{"response_detail": "summary"})
	if clean["ok"] != true {
		t.Fatalf("git_status failed: %v", clean)
	}
	if clean["changed_files"] != float64(0) {
		t.Fatalf("expected clean tree, got %v", clean)
	}

	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("two"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	dirty := toolPayload(t, server, "git_status", map[string]any{"response_detail": "summary"})
	if dirty["changed_files"] != float64(1) {
		t.Fatalf("changed_files = %v", dirty["changed_files"])
	}

	diff := toolPayload(t, server, "git_diff", map[string]any{"response_detail": "normal"})
	if diff["files_changed"] != float64(1) {
		t.Fatalf("files_changed = %v", diff["files_changed"])
	}
	if !strings.Contains(fmt.Sprint(diff["patch"]), "-one") {
		t.Fatalf("patch missing removal:\n%v", diff["patch"])
	}

	commit := toolPayload(t, server, "git_commit", map[string]any{"message": "update file"})
	if commit["ok"] != true {
		t.Fatalf("git_commit failed: %v", commit)
	}
	if strings.TrimSpace(fmt.Sprint(commit["commit"])) == "" {
		t.Fatal("commit hash missing")
	}

	after := toolPayload(t, server, "git_status", map[string]any{"response_detail": "summary"})
	if after["changed_files"] != float64(0) {
		t.Fatalf("tree not clean after commit: %v", after)
	}
}

func TestReadOnlyConfigHidesWriteTools(t *testing.T) {
	cfg := testMCPConfig(t, t.TempDir())
	cfg.ReadOnly = true
	for _, tool := range availableMCPTools(cfg) {
		if !tool.ReadOnly {
			t.Fatalf("read-only config exposed %q", tool.Name)
		}
	}
}

func TestMCPConfigDetailResolution(t *testing.T) {
	cfg := testMCPConfig(t, t.TempDir())
	cfg.Detail = mcpDetailSummary
	if got := cfg.detailLevel(""); got != mcpDetailSummary {
		t.Fatalf("default detail = %q", got)
	}
	if got := cfg.detailLevel("full"); got != mcpDetailFull {
		t.Fatalf("override detail = %q", got)
	}
	if got := cfg.detailLevel("nonsense"); got != mcpDetailSummary {
		t.Fatalf("invalid override should fall back to config: %q", got)
	}
}

func TestNewMCPConfigResolvesFlagsAndEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("OPENSCADGEN_MCP_READ_ONLY", "true")
	t.Setenv("OPENSCADGEN_MCP_DETAIL", "full")

	cfg := newMCPConfig(flagsForMCPTest(root), "")
	if !cfg.Enabled {
		t.Fatal("expected -ai to enable MCP")
	}
	if !cfg.ReadOnly {
		t.Fatal("expected env read-only")
	}
	if cfg.Detail != mcpDetailFull {
		t.Fatalf("detail = %q", cfg.Detail)
	}
	if len(cfg.Roots) != 1 || !strings.EqualFold(cfg.Roots[0], filepath.Clean(root)) {
		t.Fatalf("roots = %v", cfg.Roots)
	}
}

func TestNewMCPConfigRejectsBadDetail(t *testing.T) {
	flags := flagsForMCPTest(t.TempDir())
	flags.MCPDetail = "nonsense"
	cfg := newMCPConfig(flags, "")
	if cfg.Detail != mcpDetailNormal {
		t.Fatalf("detail = %q", cfg.Detail)
	}
	if len(cfg.Notices) == 0 {
		t.Fatal("expected a notice for the unknown detail level")
	}
}

func TestStdioCommandAndEndpoint(t *testing.T) {
	cfg := testMCPConfig(t, t.TempDir())
	cfg.AllowPush = true
	if url := cfg.mcpEndpointURL(7203); url != "http://127.0.0.1:7203/mcp" {
		t.Fatalf("endpoint = %q", url)
	}
	command := cfg.stdioCommand()
	for _, want := range []string{"-ai", "-mcp-stdio", "-mcp-allow-push"} {
		if !strings.Contains(command, want) {
			t.Fatalf("stdio command %q missing %q", command, want)
		}
	}
}

func TestBuildProcessPayloadSuccessCriteria(t *testing.T) {
	root := t.TempDir()
	artifact := filepath.Join(root, "part.stl")
	if err := os.WriteFile(artifact, []byte("solid part\nendsolid"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	failed := buildProcessPayload(root, models.ProcessResult{
		Instances: []models.InstanceConfig{
			{AutoName: "a", IsSuccessful: false, ConfigError: "renderer exploded"},
			{AutoName: "b", IsSuccessful: false, ConfigError: "renderer exploded"},
		},
	}, nil, "")
	if failed.OK {
		t.Fatal("a run where every instance failed must not be ok")
	}
	rendered := failed.render(mcpDetailSummary)
	if rendered["ok"] != false {
		t.Fatalf("rendered ok = %v", rendered["ok"])
	}
	if rendered["error"] != "all instances failed to render" {
		t.Fatalf("error = %v", rendered["error"])
	}

	success := buildProcessPayload(root, models.ProcessResult{
		Instances:  []models.InstanceConfig{{AutoName: "a", IsSuccessful: true, IsComplete: true}},
		STLResults: []models.GenerateSTLResult{{InstanceConfig: models.InstanceConfig{RunOutputPathV3: artifact}}},
	}, nil, "")
	if !success.OK {
		t.Fatalf("expected success, got %v", success.render(mcpDetailSummary))
	}
	successSummary := success.render(mcpDetailSummary)
	if successSummary["artifacts"] != 1 || successSummary["missing_artifacts"] != 0 {
		t.Fatalf("artifact accounting = %v / %v", successSummary["artifacts"], successSummary["missing_artifacts"])
	}

	missing := buildProcessPayload(root, models.ProcessResult{
		Instances:  []models.InstanceConfig{{AutoName: "a", IsSuccessful: true, IsComplete: true}},
		STLResults: []models.GenerateSTLResult{{InstanceConfig: models.InstanceConfig{RunOutputPathV3: filepath.Join(root, "gone.stl")}}},
	}, nil, "")
	if missing.OK {
		t.Fatal("expected missing artifact to fail the run")
	}

	// Relative run paths resolve against the config file's folder.
	if err := os.MkdirAll(filepath.Join(root, "export", "v0.1"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "export", "v0.1", "part.stl"), []byte("solid"), 0o644); err != nil {
		t.Fatalf("write relative artifact: %v", err)
	}
	relative := buildProcessPayload(root, models.ProcessResult{
		ConfigFile: filepath.Join(root, "config.toml"),
		Instances:  []models.InstanceConfig{{AutoName: "a", IsSuccessful: true, IsComplete: true}},
		STLResults: []models.GenerateSTLResult{{InstanceConfig: models.InstanceConfig{
			RunOutputPathV3: filepath.ToSlash(filepath.Join("export", "v0.1", "part.stl")),
		}}},
	}, nil, "")
	relRendered := relative.render(mcpDetailSummary)
	if relRendered["artifacts"] != 1 || relRendered["ok"] != true {
		t.Fatalf("relative artifact not resolved: %v", relRendered)
	}

	empty := buildProcessPayload(root, models.ProcessResult{}, nil, "")
	if empty.OK {
		t.Fatal("an empty run must not be ok")
	}

	errored := buildProcessPayload(root, models.ProcessResult{
		Instances:  []models.InstanceConfig{{AutoName: "a", IsSuccessful: true}},
		STLResults: []models.GenerateSTLResult{{OutputPath: artifact}},
	}, fmt.Errorf("image test failed"), "")
	if errored.OK {
		t.Fatal("a process error must fail the run")
	}
}

func TestListProjectFilesAcrossMultipleRoots(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()
	for path, body := range map[string]string{
		filepath.Join(rootA, "alpha.scad"): "a",
		filepath.Join(rootB, "beta.scad"):  "b",
	} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	server := newMCPServer(testMCPConfig(t, rootA, rootB))

	payload := toolPayload(t, server, "list_project_files", map[string]any{"response_detail": "summary"})
	entries, ok := payload["entries"].([]any)
	if !ok {
		t.Fatalf("entries = %T %v", payload["entries"], payload["entries"])
	}
	joined := make([]string, 0, len(entries))
	for _, entry := range entries {
		joined = append(joined, fmt.Sprint(entry))
	}
	combined := strings.Join(joined, ",")
	for _, want := range []string{filepath.Base(rootA) + "/alpha.scad", filepath.Base(rootB) + "/beta.scad"} {
		if !strings.Contains(combined, filepath.ToSlash(want)) {
			t.Fatalf("multi-root listing missing %q: %s", want, combined)
		}
	}
}

func TestEndpointUsesLoopbackForWildcardBind(t *testing.T) {
	cfg := testMCPConfig(t, t.TempDir())
	cfg.BindAddress = "0.0.0.0"
	if got := cfg.mcpEndpointURL(7203); got != "http://127.0.0.1:7203/mcp" {
		t.Fatalf("endpoint = %q", got)
	}
}

func TestToolPayloadsAreJSONEncodable(t *testing.T) {
	server := newMCPServer(testMCPConfig(t, t.TempDir()))
	response := callMCP(t, server, "tools/list", nil)
	if response.Error != nil {
		t.Fatalf("tools/list failed: %+v", response.Error)
	}
	if _, err := json.Marshal(response.Result); err != nil {
		t.Fatalf("result not encodable: %v", err)
	}
}

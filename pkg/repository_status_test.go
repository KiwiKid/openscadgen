package pkg

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRepositoryStatusForDirSummarizesWorktree(t *testing.T) {
	dir := t.TempDir()
	runGitTestCommand(t, dir, "init")
	runGitTestCommand(t, dir, "config", "user.email", "test@example.com")
	runGitTestCommand(t, dir, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("first\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitTestCommand(t, dir, "add", "tracked.txt")
	runGitTestCommand(t, dir, "commit", "-m", "initial project")
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "staged.txt"), []byte("staged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitTestCommand(t, dir, "add", "staged.txt")
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	status := RepositoryStatusForDir(dir)
	if status == nil {
		t.Fatal("RepositoryStatusForDir() = nil")
	}
	if status.Branch == "" || status.LastCommit != "initial project" {
		t.Fatalf("unexpected branch/commit: %#v", status)
	}
	if status.Staged != 1 || status.Modified != 1 || status.Untracked != 1 {
		t.Fatalf("unexpected change counts: %#v", status)
	}
}

func TestRepositoryStatusForDirReturnsNilOutsideGit(t *testing.T) {
	if status := RepositoryStatusForDir(t.TempDir()); status != nil {
		t.Fatalf("RepositoryStatusForDir() = %#v, want nil", status)
	}
}

func runGitTestCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

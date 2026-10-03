package pkg

import (
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kiwikid/openscadgen/pkg/models"
)

const repositoryStatusTimeout = 2 * time.Second

// RepositoryStatusForDir reports the Git state for dir without fetching,
// changing branches, or otherwise modifying the worktree. A nil result means
// dir is not inside a Git repository (or Git is unavailable).
func RepositoryStatusForDir(dir string) *models.RepositoryStatus {
	ctx, cancel := context.WithTimeout(context.Background(), repositoryStatusTimeout)
	defer cancel()

	statusOutput, err := runGit(ctx, dir, "status", "--porcelain=v2", "--branch")
	if err != nil {
		return nil
	}
	root, err := runGit(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil
	}

	status := &models.RepositoryStatus{Root: filepath.Clean(strings.TrimSpace(root))}
	for _, line := range strings.Split(statusOutput, "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.ab "):
			fields := strings.Fields(strings.TrimPrefix(line, "# branch.ab "))
			for _, field := range fields {
				if len(field) < 2 {
					continue
				}
				count, err := strconv.Atoi(field[1:])
				if err != nil {
					continue
				}
				switch field[0] {
				case '+':
					status.Ahead = count
				case '-':
					status.Behind = count
				}
			}
		case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 ") || strings.HasPrefix(line, "u "):
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			xy := fields[1]
			if len(xy) == 2 && xy[0] != '.' {
				status.Staged++
			}
			if len(xy) == 2 && xy[1] != '.' {
				status.Modified++
			}
		case strings.HasPrefix(line, "? "):
			status.Untracked++
		}
	}

	commitOutput, err := runGit(ctx, dir, "log", "-1", "--format=%ct%x00%s")
	if err == nil {
		parts := strings.SplitN(commitOutput, "\x00", 2)
		if len(parts) == 2 {
			if seconds, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				status.LastCommitAt = time.Unix(seconds, 0)
			}
			status.LastCommit = parts[1]
		}
	}
	return status
}

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", dir}, args...)
	command := exec.CommandContext(ctx, "git", commandArgs...)
	output, err := command.Output()
	return strings.TrimSpace(string(output)), err
}

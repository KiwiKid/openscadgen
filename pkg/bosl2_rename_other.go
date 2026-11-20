//go:build !windows

package pkg

// isRetryableRenameError is false outside Windows: POSIX platforms allow
// renaming a directory that is open in another process.
func isRetryableRenameError(error) bool {
	return false
}

//go:build windows

package pkg

import (
	"errors"
	"syscall"
)

// Windows error codes that indicate the directory is still open elsewhere.
const (
	errorSharingViolation = syscall.Errno(32) // ERROR_SHARING_VIOLATION
	errorLockViolation    = syscall.Errno(33) // ERROR_LOCK_VIOLATION
)

// isRetryableRenameError reports whether a directory rename failed because the
// directory is still open in another process. Windows reports this as a
// sharing violation; a directory that is pending deletion can instead surface
// as access denied or a lock violation.
func isRetryableRenameError(err error) bool {
	return errors.Is(err, errorSharingViolation) ||
		errors.Is(err, syscall.ERROR_ACCESS_DENIED) ||
		errors.Is(err, errorLockViolation)
}

//go:build windows

package update

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func replaceBinary(src string) error {
	dst, err := os.Executable()
	if err != nil {
		return err
	}
	// A running process on Windows usually has its own exe open. Move the
	// current exe aside, then move the new file into place.
	backup := dst + ".old"
	_ = os.Remove(backup)
	deadline := time.Now().Add(10 * time.Second)
	for {
		err := os.Rename(dst, backup)
		if err == nil {
			break
		}
		if !isLocked(err) || time.Now().After(deadline) {
			return fmt.Errorf("rename existing binary: %w", err)
		}
		time.Sleep(300 * time.Millisecond)
	}
	if err := os.Rename(src, dst); err != nil {
		_ = os.Rename(backup, dst)
		return fmt.Errorf("replace binary: %w", err)
	}
	_ = os.Remove(backup)
	return nil
}

// isLocked treats common Windows sharing errors as a transient lock.
func isLocked(err error) bool {
	if err == nil {
		return false
	}
	for _, s := range []string{"being used", "another process", "permission"} {
		if strings.Contains(strings.ToLower(err.Error()), s) {
			return true
		}
	}
	return false
}

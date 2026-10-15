//go:build !windows

package update

import (
	"fmt"
	"os"
)

func replaceBinary(src string) error {
	dst, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	return os.Chmod(dst, 0o755)
}

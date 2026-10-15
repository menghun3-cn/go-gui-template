//go:build !windows && !darwin

package platform

import (
	"os"
	"path/filepath"
)

func EnableAutoStart(name, command string) error {
	dir := autoStartDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, name+".desktop")
	content := "[Desktop Entry]\nType=Application\nName=" + name + "\nExec=" + command + "\n"
	return os.WriteFile(path, []byte(content), 0o755)
}

func DisableAutoStart(name string) error {
	path := filepath.Join(autoStartDir(), name+".desktop")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func AutoStartEnabled(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(autoStartDir(), name+".desktop"))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func autoStartDir() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return filepath.Join(v, "autostart")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "autostart")
}

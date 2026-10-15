package platform

import (
	"os"

	"github.com/menghun3-cn/go-gui-template/internal/server"
)

// AutoStart implements server.AutoStarter for the current platform.
type AutoStart struct {
	Name    string
	Command string
}

func (a *AutoStart) Enable() error {
	if a.Command == "" {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		a.Command = quote(exe)
	}
	return EnableAutoStart(a.Name, a.Command)
}

func (a *AutoStart) Disable() error {
	return DisableAutoStart(a.Name)
}

func (a *AutoStart) Enabled() (bool, error) {
	return AutoStartEnabled(a.Name)
}

var _ server.AutoStarter = (*AutoStart)(nil)

func quote(s string) string {
	return `"` + s + `"`
}

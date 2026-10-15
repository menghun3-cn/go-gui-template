// Package app wires the CLI entry point with the GUI and local HTTP server.
package app

import (
	"github.com/menghun3-cn/go-gui-template/internal/gui"
	"github.com/menghun3-cn/go-gui-template/internal/platform"
	"github.com/menghun3-cn/go-gui-template/internal/update"
)

type App struct {
	Version string
}

func New(version string) *App {
	if version == "" {
		version = "dev"
	}
	return &App{Version: version}
}

func (a *App) Run() error {
	return gui.Run(a.Options())
}

func (a *App) Options() gui.Options {
	return gui.Options{
		Version: a.Version,
		Updater: &update.Service{},
		AutoStart: &platform.AutoStart{
			Name: "go-gui-template",
		},
	}
}


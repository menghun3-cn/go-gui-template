//go:build nogui

package gui

import (
	"errors"

	"github.com/menghun3-cn/go-gui-template/internal/server"
)

func Run(opts Options) error {
	_ = opts
	return errors.New("gui disabled by -tags nogui")
}

type Options struct {
	Version   string
	Updater   server.Updater
	AutoStart server.AutoStarter
}

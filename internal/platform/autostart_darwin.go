//go:build darwin

package platform

import "errors"

func EnableAutoStart(name, command string) error {
	return errors.New("darwin autostart: wire to launchd or SMAppService")
}

func DisableAutoStart(name string) error {
	return errors.New("darwin autostart: wire to launchd or SMAppService")
}

func AutoStartEnabled(name string) (bool, error) {
	return false, nil
}

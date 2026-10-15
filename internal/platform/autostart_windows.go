//go:build windows

package platform

import (
	"golang.org/x/sys/windows/registry"
)

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func EnableAutoStart(name, command string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(name, command)
}

func DisableAutoStart(name string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.DeleteValue(name); err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}

func AutoStartEnabled(name string) (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer k.Close()
	_, _, err = k.GetStringValue(name)
	if err == registry.ErrNotExist {
		return false, nil
	}
	return err == nil, err
}

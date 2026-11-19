//go:build windows

package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	TaskName = "VibeVPN"
	RegKey   = `Software\Microsoft\Windows\CurrentVersion\Run`
	AppName  = "VibeVPN"
)

// Enable registers the application in Windows Task Scheduler (for elevated startup without UAC prompt)
// and in the HKCU Run registry key (for visibility in Windows Task Manager -> Startup).
func Enable() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exePath := filepath.Clean(exe)

	// 1. Task Scheduler: runs elevated (/RL HIGHEST) at user logon
	taskCmd := exec.Command("schtasks", "/Create", "/TN", TaskName, "/TR", fmt.Sprintf("\"%s\" --autostart", exePath), "/SC", "ONLOGON", "/RL", "HIGHEST", "/F")
	taskCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = taskCmd.Run()

	// 2. Registry HKCU Run key (visible in Windows Settings & Task Manager)
	k, err := registry.OpenKey(registry.CURRENT_USER, RegKey, registry.SET_VALUE)
	if err == nil {
		defer k.Close()
		_ = k.SetStringValue(AppName, fmt.Sprintf("\"%s\" --autostart", exePath))
	}

	return nil
}

// Disable unregisters the application from Task Scheduler and the registry.
func Disable() error {
	// 1. Delete scheduled task
	taskCmd := exec.Command("schtasks", "/Delete", "/TN", TaskName, "/F")
	taskCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = taskCmd.Run()

	// 2. Delete registry key
	k, err := registry.OpenKey(registry.CURRENT_USER, RegKey, registry.SET_VALUE)
	if err == nil {
		defer k.Close()
		_ = k.DeleteValue(AppName)
	}

	return nil
}

// IsEnabled checks if autostart is configured either in Task Scheduler or in the registry.
func IsEnabled() bool {
	// Check Task Scheduler
	taskCmd := exec.Command("schtasks", "/Query", "/TN", TaskName)
	taskCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := taskCmd.Run(); err == nil {
		return true
	}

	// Check registry
	k, err := registry.OpenKey(registry.CURRENT_USER, RegKey, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		if val, _, err := k.GetStringValue(AppName); err == nil && val != "" {
			return true
		}
	}

	return false
}

// Sync synchronizes the system state with the desired enabled status.
func Sync(enable bool) error {
	if enable {
		return Enable()
	}
	return Disable()
}

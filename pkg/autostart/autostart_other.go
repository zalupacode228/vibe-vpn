//go:build !windows

package autostart

func Enable() error {
	return nil
}

func Disable() error {
	return nil
}

func IsEnabled() bool {
	return false
}

func Sync(enable bool) error {
	return nil
}

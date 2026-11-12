//go:build !windows
// +build !windows

package pinger

import "time"

func PingICMP(host string, timeout time.Duration) int {
	return 0
}

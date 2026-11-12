//go:build windows
// +build windows

package pinger

import (
	"encoding/binary"
	"net"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modIphlpapi         = windows.NewLazySystemDLL("iphlpapi.dll")
	procIcmpCreateFile  = modIphlpapi.NewProc("IcmpCreateFile")
	procIcmpCloseHandle = modIphlpapi.NewProc("IcmpCloseHandle")
	procIcmpSendEcho    = modIphlpapi.NewProc("IcmpSendEcho")
)

type icmpEchoReply struct {
	Address       uint32
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       struct {
		Ttl         uint8
		Tos         uint8
		Flags       uint8
		OptionsSize uint8
		OptionsData uintptr
	}
}

// PingICMP measures ICMP round-trip latency to a host using Windows iphlpapi.dll (no admin required)
func PingICMP(host string, timeout time.Duration) int {
	if host == "" {
		return 0
	}

	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return 0
	}
	var ipv4 net.IP
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			ipv4 = v4
			break
		}
	}
	if ipv4 == nil {
		return 0
	}

	h, _, _ := procIcmpCreateFile.Call()
	if h == uintptr(windows.InvalidHandle) || h == 0 {
		return 0
	}
	defer procIcmpCloseHandle.Call(h)

	destIP := binary.LittleEndian.Uint32(ipv4)
	sendData := []byte("vibeping")
	replySize := unsafe.Sizeof(icmpEchoReply{}) + uintptr(len(sendData)) + 8
	replyBuffer := make([]byte, replySize)

	timeoutMs := uint32(timeout.Milliseconds())
	if timeoutMs == 0 {
		timeoutMs = 1500
	}

	ret, _, _ := procIcmpSendEcho.Call(
		h,
		uintptr(destIP),
		uintptr(unsafe.Pointer(&sendData[0])),
		uintptr(len(sendData)),
		0,
		uintptr(unsafe.Pointer(&replyBuffer[0])),
		uintptr(replySize),
		uintptr(timeoutMs),
	)

	if ret == 0 {
		return 0
	}

	reply := (*icmpEchoReply)(unsafe.Pointer(&replyBuffer[0]))
	if reply.Status != 0 {
		return 0
	}
	rtt := int(reply.RoundTripTime)
	if rtt == 0 {
		rtt = 1
	}
	return rtt
}

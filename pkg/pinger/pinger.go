package pinger

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"vibe-go/pkg/models"
)

// PingAddress performs a TCP handshake to determine socket reachability
func PingAddress(host string, port int, timeout time.Duration) int {
	if host == "" || port <= 0 {
		return 0
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return 0
	}
	defer conn.Close()
	return int(time.Since(start).Milliseconds())
}

// PingNode determines the latency to a node using TCP handshake for TCP protocols and ICMP for UDP protocols (Hysteria, WireGuard)
func PingNode(node models.Node, timeout time.Duration) int {
	if node.Server == "" {
		return 0
	}

	proto := strings.ToLower(node.Protocol)
	isUDP := proto == "hysteria2" || proto == "hysteria" || proto == "wireguard" || proto == "warp" || proto == "masque" || proto == "tuic"

	// 1. For UDP-only protocols (Hysteria, WireGuard, etc.), try ICMP ping first
	if isUDP {
		if icmpPing := PingICMP(node.Server, timeout); icmpPing > 0 {
			return icmpPing
		}
		if node.Port > 0 {
			if tcpPing := PingAddress(node.Server, node.Port, timeout); tcpPing > 0 {
				return tcpPing
			}
		}
		return 0
	}

	// 2. For TCP-based protocols (VLESS, VMess, Trojan, Shadowsocks, etc.), try TCP handshake to node port
	if node.Port > 0 {
		if tcpPing := PingAddress(node.Server, node.Port, timeout); tcpPing > 0 {
			return tcpPing
		}
	}

	// 3. Fallback to ICMP ping if TCP handshake failed or timed out
	return PingICMP(node.Server, timeout)
}

// PingNodeDeep is an alias to PingNode for backward compatibility
func PingNodeDeep(node models.Node, timeout time.Duration) int {
	return PingNode(node, timeout)
}

// PingAll pings all nodes concurrently with protocol-aware verification
func PingAll(nodes []models.Node) map[string]int {
	results := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 20) // Limit concurrency to 20 sockets

	for _, n := range nodes {
		wg.Add(1)
		go func(node models.Node) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			p := PingNode(node, 2000*time.Millisecond)
			mu.Lock()
			results[node.ID] = p
			mu.Unlock()
		}(n)
	}

	wg.Wait()
	return results
}

// PingWeb measures true end-to-end HTTP latency to verify active internet connectivity
func PingWeb(targetURL string, timeout time.Duration) (int, int, error) {
	if targetURL == "" {
		targetURL = "http://cp.cloudflare.com/generate_204"
	}

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	start := time.Now()
	resp, err := client.Get(targetURL)
	if err != nil {
		// Fallback to Google 204 endpoint if Cloudflare fails
		if !strings.Contains(targetURL, "gstatic.com") {
			return PingWeb("https://www.gstatic.com/generate_204", timeout)
		}
		return 0, 0, err
	}
	defer resp.Body.Close()

	latency := int(time.Since(start).Milliseconds())
	return latency, resp.StatusCode, nil
}

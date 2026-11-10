package pinger

import (
	"net"
	"strconv"
	"sync"
	"time"
	"vibe-go/pkg/models"
)

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

			p := PingAddress(node.Server, node.Port, 2500*time.Millisecond)
			mu.Lock()
			results[node.ID] = p
			mu.Unlock()
		}(n)
	}

	wg.Wait()
	return results
}

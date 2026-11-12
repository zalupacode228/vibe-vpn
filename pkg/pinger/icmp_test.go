package pinger

import (
	"testing"
	"vibe-go/pkg/models"
)

func TestPingAll(t *testing.T) {
	nodes := []models.Node{
		{
			ID:       "xhttp",
			Server:   "panel05.zamik.work.gd",
			Port:     443,
			Protocol: "vless",
			Security: "reality",
		},
		{
			ID:       "hysteria",
			Server:   "panel05.zamik.work.gd",
			Port:     8443,
			Protocol: "hysteria2",
		},
	}

	results := PingAll(nodes)
	t.Logf("Ping results: %+v", results)
	if results["xhttp"] == 0 || results["xhttp"] > 1000 {
		t.Errorf("xhttp ping unexpected: %d", results["xhttp"])
	}
	if results["hysteria"] == 0 {
		t.Errorf("hysteria ping should not be 0, got %d", results["hysteria"])
	}
}

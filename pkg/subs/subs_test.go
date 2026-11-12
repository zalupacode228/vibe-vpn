package subs

import (
	"testing"
)

func TestParseHysteria2WithObfs(t *testing.T) {
	raw := "hysteria2://d123a4de11fe4ffabbaa96ff1f6de13c@panel05.zamik.work.gd:51821?alpn=h3%2Ch2%2Chttp%2F1.1&fp=firefox&obfs=salamander&obfs-password=7gm3axb5evb05ze8&security=tls&sni=files.zamik.work.gd#hysteria"
	node := ParseNodeLink(raw)
	if node == nil {
		t.Fatalf("Failed to parse hysteria2 node")
	}

	if node.Protocol != "hysteria2" {
		t.Errorf("Expected protocol hysteria2, got %s", node.Protocol)
	}
	if node.Password != "d123a4de11fe4ffabbaa96ff1f6de13c" {
		t.Errorf("Unexpected password: %s", node.Password)
	}
	if node.Obfs != "salamander" {
		t.Errorf("Expected obfs salamander, got %s", node.Obfs)
	}
	if node.ObfsPassword != "7gm3axb5evb05ze8" {
		t.Errorf("Expected obfs-password 7gm3axb5evb05ze8, got %s", node.ObfsPassword)
	}
	if node.SNI != "files.zamik.work.gd" {
		t.Errorf("Expected sni files.zamik.work.gd, got %s", node.SNI)
	}
	if len(node.ALPN) != 3 || node.ALPN[0] != "h3" {
		t.Errorf("Unexpected ALPN: %v", node.ALPN)
	}
}

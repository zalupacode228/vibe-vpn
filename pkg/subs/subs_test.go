package subs

import (
	"testing"
)

func TestParseHysteria2WithObfs(t *testing.T) {
	raw := "hysteria2://dummy-user-password-test@hy2.example.com:51821?alpn=h3%2Ch2%2Chttp%2F1.1&fp=firefox&obfs=salamander&obfs-password=7gm3axb5evb05ze8&security=tls&sni=hy2.example.com#hysteria"
	node := ParseNodeLink(raw)
	if node == nil {
		t.Fatalf("Failed to parse hysteria2 node")
	}

	if node.Protocol != "hysteria2" {
		t.Errorf("Expected protocol hysteria2, got %s", node.Protocol)
	}
	if node.Password != "dummy-user-password-test" {
		t.Errorf("Unexpected password: %s", node.Password)
	}
	if node.Obfs != "salamander" {
		t.Errorf("Expected obfs salamander, got %s", node.Obfs)
	}
	if node.ObfsPassword != "7gm3axb5evb05ze8" {
		t.Errorf("Expected obfs-password 7gm3axb5evb05ze8, got %s", node.ObfsPassword)
	}
	if node.SNI != "hy2.example.com" {
		t.Errorf("Expected sni hy2.example.com, got %s", node.SNI)
	}
	if len(node.ALPN) != 3 || node.ALPN[0] != "h3" {
		t.Errorf("Unexpected ALPN: %v", node.ALPN)
	}
}

func TestParseTrojanVMessSS(t *testing.T) {
	// Trojan
	tr := ParseNodeLink("trojan://secretpass@tr.example.com:443?sni=tr.example.com&fp=chrome&type=ws&path=%2Fws&host=cdn.example.com#TrojanNode")
	if tr == nil || tr.Protocol != "trojan" || tr.Password != "secretpass" || tr.Network != "ws" || tr.Path != "/ws" || tr.Host != "cdn.example.com" {
		t.Fatalf("Failed to parse trojan link: %+v", tr)
	}

	// Shadowsocks SIP002
	// base64("aes-256-gcm:mysecret") = "YWVzLTI1Ni1nY206bXlzZWNyZXQ="
	ss := ParseNodeLink("ss://YWVzLTI1Ni1nY206bXlzZWNyZXQ=@ss.example.com:8388#SSNode")
	if ss == nil || ss.Protocol != "ss" || ss.Encryption != "aes-256-gcm" || ss.Password != "mysecret" || ss.Port != 8388 {
		t.Fatalf("Failed to parse ss SIP002 link: %+v", ss)
	}

	// VMess
	// {"ps":"VMessNode","add":"vm.example.com","port":"443","id":"11111111-1111-1111-1111-111111111111","scy":"auto","net":"ws","host":"vm.example.com","path":"/vm","tls":"tls","sni":"vm.example.com"}
	vm := ParseNodeLink("vmess://eyJwcyI6IlZNZXNzTm9kZSIsImFkZCI6InZtLmV4YW1wbGUuY29tIiwicG9ydCI6IjQ0MyIsImlkIjoiMTExMTExMTEtMTExMS0xMTExLTExMTEtMTExMTExMTExMTExIiwic2N5IjoiYXV0byIsIm5ldCI6IndzIiwiaG9zdCI6InZtLmV4YW1wbGUuY29tIiwicGF0aCI6Ii92bSIsInRscyI6InRscyIsInNuaSI6InZtLmV4YW1wbGUuY29tIn0=")
	if vm == nil || vm.Protocol != "vmess" || vm.Server != "vm.example.com" || vm.Port != 443 || vm.Network != "ws" || vm.Security != "tls" {
		t.Fatalf("Failed to parse vmess link: %+v", vm)
	}

	// Stable ID across fragment rename
	n1 := ParseNodeLink("vless://00000000-0000-0000-0000-000000000000@1.1.1.1:443?security=reality#Name1")
	n2 := ParseNodeLink("vless://00000000-0000-0000-0000-000000000000@1.1.1.1:443?security=reality#Name2")
	if n1 == nil || n2 == nil || n1.ID != n2.ID {
		t.Fatalf("Expected stable ID across rename, got %v vs %v", n1, n2)
	}
}

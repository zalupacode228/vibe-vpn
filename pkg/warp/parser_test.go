package warp

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"vibe-go/pkg/models"
	"vibe-go/pkg/singbox"
)

func TestParseWireguardConf(t *testing.T) {
	rawConf := `
[Interface]
PrivateKey = aGVsbG93b3JsZHByaXZhdGVrZXkxMjM0NTY3ODkwMTI=
Address = 172.16.0.2/32, 2606:4700:110:88bf:4e15:c3aa:69ee:ce72/128
DNS = 1.1.1.1, 1.0.0.1
MTU = 1280
Jc = 4
Jmin = 40
Jmax = 70
H1 = 1
H2 = 2
H3 = 3
H4 = 4

[Peer]
PublicKey = bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = 162.159.192.1:2408
Reserved = 163, 250, 45
`
	node, err := ParseWireguardConf(rawConf, "warp-germany.conf")
	if err != nil {
		t.Fatalf("ParseWireguardConf failed: %v", err)
	}

	if node.Name != "warp-germany" {
		t.Errorf("Expected node name 'warp-germany', got '%s'", node.Name)
	}
	if node.Protocol != "wireguard" {
		t.Errorf("Expected protocol 'wireguard', got '%s'", node.Protocol)
	}
	if node.ProtocolLabel != "Cloudflare WARP (AWG)" {
		t.Errorf("Expected protocol label 'Cloudflare WARP (AWG)', got '%s'", node.ProtocolLabel)
	}
	if node.Server != "162.159.192.1" || node.Port != 2408 {
		t.Errorf("Expected endpoint 162.159.192.1:2408, got %s:%d", node.Server, node.Port)
	}
	if len(node.Reserved) != 3 || node.Reserved[0] != 163 || node.Reserved[1] != 250 || node.Reserved[2] != 45 {
		t.Errorf("Expected reserved [163 250 45], got %v", node.Reserved)
	}
	if node.JC != 4 || node.H1 != 1 {
		t.Errorf("Expected JC=4, H1=1, got JC=%d, H1=%d", node.JC, node.H1)
	}

	// Verify sing-box generates valid JSON from this node!
	b := singbox.NewConfigBuilder()
	settings := models.Settings{
		SplitMode: "global",
		BypassLAN: true,
		MTU:       1280,
	}

	cfgBytes, err := b.GenerateConfig("global", nil, node, settings)
	if err != nil {
		t.Fatalf("GenerateConfig error: %v", err)
	}

	appData := os.Getenv("APPDATA")
	sbPath := filepath.Join(appData, "VPN", "core", "sing-box.exe")
	if _, err := os.Stat(sbPath); err == nil {
		tmpFile := filepath.Join(os.TempDir(), "test_warp_parsed.json")
		_ = os.WriteFile(tmpFile, cfgBytes, 0644)
		defer os.Remove(tmpFile)

		out, err := exec.Command(sbPath, "check", "-c", tmpFile).CombinedOutput()
		if err != nil {
			t.Fatalf("sing-box check failed: %v\nOutput: %s\nConfig:\n%s", err, string(out), string(cfgBytes))
		}
		t.Logf("sing-box check successfully passed for parsed WARP config!")
	}
}

func TestParseStandardWireguard(t *testing.T) {
	rawConf := `
# My WireGuard VPS
[Interface]
PrivateKey = aGVsbG93b3JsZHByaXZhdGVrZXkxMjM0NTY3ODkwMTI=
Address = 10.0.0.2/24

[Peer]
PublicKey = xBI5JNxODquOGUFVQx05G5kGz0e7A0/R3e5Pj9x8c1o=
Endpoint = 195.123.45.67:51820
AllowedIPs = 0.0.0.0/0
`
	node, err := ParseWireguardConf(rawConf, "my-vps.conf")
	if err != nil {
		t.Fatalf("ParseWireguardConf failed: %v", err)
	}

	if node.Name != "my-vps" {
		t.Errorf("Expected node name 'my-vps', got '%s'", node.Name)
	}
	if node.ProtocolLabel != "WireGuard" {
		t.Errorf("Expected protocol label 'WireGuard', got '%s'", node.ProtocolLabel)
	}
	if node.Server != "195.123.45.67" || node.Port != 51820 {
		t.Errorf("Expected 195.123.45.67:51820, got %s:%d", node.Server, node.Port)
	}
	if len(node.Reserved) > 0 {
		t.Errorf("Expected no reserved bytes for standard WG, got %v", node.Reserved)
	}
}

package singbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"vibe-go/pkg/models"
)

func TestGenerateConfigVLESS(t *testing.T) {
	b := NewConfigBuilder()
	node := &models.Node{
		ID:          "vless-test",
		Name:        "Test VLESS Reality",
		Protocol:    "vless",
		Server:      "1.1.1.1",
		Port:        443,
		UUID:        "00000000-0000-0000-0000-000000000000",
		Security:    "reality",
		SNI:         "cloudflare.com",
		Fingerprint: "chrome",
		PublicKey:   "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA",
		ShortID:     "0123456789abcdef",
	}

	settings := models.Settings{
		SplitMode:   "whitelist",
		BypassLAN:   true,
		BypassRU:    true,
		AdBlock:     true,
		StrictRoute: false,
		MTU:         1400,
	}

	rules := []models.AppRule{
		{
			Name:     "firefox.exe",
			Path:     `C:\Program Files\Mozilla Firefox\firefox.exe`,
			Enabled:  true,
			IsCustom: true,
		},
	}

	cfgBytes, err := b.GenerateConfig("whitelist", rules, node, settings)
	if err != nil {
		t.Fatalf("GenerateConfig error: %v", err)
	}

	tmpFile := filepath.Join(os.TempDir(), "test_vless_check.json")
	if err := os.WriteFile(tmpFile, cfgBytes, 0644); err != nil {
		t.Fatalf("Failed to write tmp file: %v", err)
	}
	defer os.Remove(tmpFile)

	appData := os.Getenv("APPDATA")
	sbPath := filepath.Join(appData, "VPN", "core", "sing-box.exe")
	if _, err := os.Stat(sbPath); err != nil {
		t.Skip("sing-box.exe not found at APPDATA/VPN/core")
	}

	out, err := exec.Command(sbPath, "check", "-c", tmpFile).CombinedOutput()
	if err != nil {
		t.Fatalf("sing-box check failed: %v\nOutput: %s\nConfig:\n%s", err, string(out), string(cfgBytes))
	}
}

func TestGenerateConfigAWG(t *testing.T) {
	b := NewConfigBuilder()
	node := &models.Node{
		ID:            "awg-test",
		Name:          "Test AWG WARP",
		Protocol:      "wireguard",
		Server:        "162.159.192.1",
		Port:          500,
		PrivateKey:    "aGVsbG93b3JsZGhlbGxvd29ybGRoZWxsb3dvcmxkMQ==",
		PeerPublicKey: "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
		LocalAddress:  "172.16.0.2/32",
		Reserved:      []int{1, 2, 3},
		AWG: map[string]interface{}{
			"jc":   120,
			"jmin": 23,
			"jmax": 911,
			"s1":   0,
			"s2":   0,
			"h1":   1,
			"h2":   2,
			"h3":   3,
			"h4":   4,
		},
	}

	settings := models.Settings{
		SplitMode: "global",
		BypassLAN: true,
		MTU:       1280,
	}

	cfgBytes, err := b.GenerateConfig("global", nil, node, settings)
	if err != nil {
		t.Fatalf("GenerateConfig error: %v", err)
	}

	tmpFile := filepath.Join(os.TempDir(), "test_awg_check.json")
	if err := os.WriteFile(tmpFile, cfgBytes, 0644); err != nil {
		t.Fatalf("Failed to write tmp file: %v", err)
	}
	defer os.Remove(tmpFile)

	appData := os.Getenv("APPDATA")
	sbPath := filepath.Join(appData, "VPN", "core", "sing-box.exe")
	if _, err := os.Stat(sbPath); err != nil {
		t.Skip("sing-box.exe not found at APPDATA/VPN/core")
	}

	out, err := exec.Command(sbPath, "check", "-c", tmpFile).CombinedOutput()
	if err != nil {
		t.Fatalf("sing-box check failed: %v\nOutput: %s\nConfig:\n%s", err, string(out), string(cfgBytes))
	}
}

func TestGenerateConfigWarpDetour(t *testing.T) {
	b := NewConfigBuilder()
	node := &models.Node{
		ID:          "vless-test",
		Name:        "Test VLESS Reality with WARP Detour",
		Protocol:    "vless",
		Server:      "1.1.1.1",
		Port:        443,
		UUID:        "00000000-0000-0000-0000-000000000000",
		Security:    "reality",
		SNI:         "cloudflare.com",
		Fingerprint: "chrome",
		PublicKey:   "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA",
		ShortID:     "0123456789abcdef",
	}

	settings := models.Settings{
		SplitMode:   "global",
		BypassLAN:   true,
		BypassRU:    true,
		AdBlock:     true,
		StrictRoute: false,
		KillSwitch:  true,
		WarpDetour:  true,
		MTU:         1400,
	}

	cfgBytes, err := b.GenerateConfig("global", nil, node, settings)
	if err != nil {
		t.Fatalf("GenerateConfig error: %v", err)
	}

	tmpFile := filepath.Join(os.TempDir(), "test_warp_detour_check.json")
	if err := os.WriteFile(tmpFile, cfgBytes, 0644); err != nil {
		t.Fatalf("Failed to write tmp file: %v", err)
	}
	defer os.Remove(tmpFile)

	appData := os.Getenv("APPDATA")
	sbPath := filepath.Join(appData, "VPN", "core", "sing-box.exe")
	if _, err := os.Stat(sbPath); err != nil {
		t.Skip("sing-box.exe not found at APPDATA/VPN/core")
	}

	out, err := exec.Command(sbPath, "check", "-c", tmpFile).CombinedOutput()
	if err != nil {
		t.Fatalf("sing-box check failed: %v\nOutput: %s\nConfig:\n%s", err, string(out), string(cfgBytes))
	}
}

func TestGenerateConfigDirectDNSLocal(t *testing.T) {
	b := NewConfigBuilder()
	node := &models.Node{
		ID:          "vless-test",
		Name:        "Test VLESS Reality with Direct DNS Local",
		Protocol:    "vless",
		Server:      "1.1.1.1",
		Port:        443,
		UUID:        "00000000-0000-0000-0000-000000000000",
		Security:    "reality",
		SNI:         "cloudflare.com",
		Fingerprint: "chrome",
		PublicKey:   "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA",
		ShortID:     "0123456789abcdef",
	}

	settings := models.Settings{
		SplitMode:   "whitelist",
		DirectDNS:   "local",
		BypassRU:    false,
		StrictRoute: false,
		MTU:         1400,
	}

	cfgBytes, err := b.GenerateConfig("whitelist", nil, node, settings)
	if err != nil {
		t.Fatalf("GenerateConfig error: %v", err)
	}

	tmpFile := filepath.Join(os.TempDir(), "test_direct_dns_local_check.json")
	if err := os.WriteFile(tmpFile, cfgBytes, 0644); err != nil {
		t.Fatalf("Failed to write tmp file: %v", err)
	}
	defer os.Remove(tmpFile)

	appData := os.Getenv("APPDATA")
	sbPath := filepath.Join(appData, "VPN", "core", "sing-box.exe")
	if _, err := os.Stat(sbPath); err != nil {
		t.Skip("sing-box.exe not found at APPDATA/VPN/core")
	}

	out, err := exec.Command(sbPath, "check", "-c", tmpFile).CombinedOutput()
	if err != nil {
		t.Fatalf("sing-box check failed with DirectDNS=local: %v\nOutput: %s\nConfig:\n%s", err, string(out), string(cfgBytes))
	}
}

func TestGenerateConfigExtraProtocols(t *testing.T) {
	b := NewConfigBuilder()
	settings := models.Settings{
		SplitMode: "global",
		BypassLAN: true,
		MTU:       1400,
	}

	nodes := []models.Node{
		{
			ID:         "ss-test",
			Name:       "Test SS",
			Protocol:   "ss",
			Server:     "1.1.1.1",
			Port:       8388,
			Encryption: "aes-256-gcm",
			Password:   "secret",
		},
		{
			ID:          "trojan-ws-test",
			Name:        "Test Trojan WS",
			Protocol:    "trojan",
			Server:      "1.1.1.1",
			Port:        443,
			Password:    "secret",
			SNI:         "example.com",
			Network:     "ws",
			Path:        "/ws",
			Host:        "cdn.example.com",
			Fingerprint: "chrome",
		},
		{
			ID:          "vmess-grpc-test",
			Name:        "Test VMess gRPC",
			Protocol:    "vmess",
			Server:      "1.1.1.1",
			Port:        443,
			UUID:        "00000000-0000-0000-0000-000000000000",
			Encryption:  "auto",
			Security:    "tls",
			SNI:         "example.com",
			Network:     "grpc",
			ServiceName: "grpc-svc",
		},
	}

	appData := os.Getenv("APPDATA")
	sbPath := filepath.Join(appData, "VPN", "core", "sing-box.exe")
	_, sbErr := os.Stat(sbPath)

	for _, n := range nodes {
		nodeCopy := n
		cfgBytes, err := b.GenerateConfig("global", nil, &nodeCopy, settings)
		if err != nil {
			t.Fatalf("GenerateConfig error for %s: %v", n.ID, err)
		}
		if sbErr == nil {
			tmpFile := filepath.Join(os.TempDir(), "test_"+n.ID+".json")
			if err := os.WriteFile(tmpFile, cfgBytes, 0644); err != nil {
				t.Fatalf("Failed to write tmp file: %v", err)
			}
			out, err := exec.Command(sbPath, "check", "-c", tmpFile).CombinedOutput()
			os.Remove(tmpFile)
			if err != nil {
				t.Fatalf("sing-box check failed for %s: %v\nOutput: %s\nConfig:\n%s", n.ID, err, string(out), string(cfgBytes))
			}
		}
	}
}

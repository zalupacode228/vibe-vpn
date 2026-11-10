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
		Server:      "panel05.zamik.work.gd",
		Port:        46787,
		UUID:        "6c8f967c-8f07-41d6-a3bf-293d6f5a8ae1",
		Security:    "reality",
		SNI:         "files.zamik.work.gd",
		Fingerprint: "firefox",
		PublicKey:   "ZrLg4VRO1xH-7Q5ZcMzvfP3AtVJwnUPAiUyh13DUiQQ",
		ShortID:     "200edd",
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

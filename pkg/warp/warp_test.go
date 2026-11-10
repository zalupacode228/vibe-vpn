package warp

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"vibe-go/pkg/models"
	"vibe-go/pkg/singbox"
)

func TestGenerateWarpAWGAndValidate(t *testing.T) {
	node, err := generateAWG()
	if err != nil {
		t.Fatalf("generateAWG failed: %v", err)
	}
	t.Logf("Generated AWG node: %+v", node)

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
	if _, err := os.Stat(sbPath); err != nil {
		t.Skip("sing-box.exe not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "test_warp_awg.json")
	if err := os.WriteFile(tmpFile, cfgBytes, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	defer os.Remove(tmpFile)

	out, err := exec.Command(sbPath, "check", "-c", tmpFile).CombinedOutput()
	if err != nil {
		t.Fatalf("sing-box check failed: %v\nOutput: %s\nConfig:\n%s", err, string(out), string(cfgBytes))
	}
	t.Logf("sing-box check passed for AWG!")
}

func TestGenerateWarpMasqueAndValidate(t *testing.T) {
	node, err := generateMasque()
	if err != nil {
		t.Fatalf("generateMasque failed: %v", err)
	}
	t.Logf("Generated MASQUE node: %+v", node)

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
	if _, err := os.Stat(sbPath); err != nil {
		t.Skip("sing-box.exe not found")
	}

	tmpFile := filepath.Join(os.TempDir(), "test_warp_masque.json")
	if err := os.WriteFile(tmpFile, cfgBytes, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	defer os.Remove(tmpFile)

	out, err := exec.Command(sbPath, "check", "-c", tmpFile).CombinedOutput()
	if err != nil {
		t.Fatalf("sing-box check failed: %v\nOutput: %s\nConfig:\n%s", err, string(out), string(cfgBytes))
	}
	t.Logf("sing-box check passed for MASQUE!")
}

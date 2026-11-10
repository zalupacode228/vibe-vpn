package procmon

import (
	"testing"
	"vibe-go/pkg/models"
)

func TestGetRunningApps(t *testing.T) {
	rules := []models.AppRule{
		{Name: "chrome.exe", Enabled: true},
	}
	apps := GetRunningApps(rules)
	if len(apps) == 0 {
		t.Logf("No user apps running currently, which is valid")
	} else {
		t.Logf("Found %d running user apps", len(apps))
		for i, a := range apps {
			if i < 5 {
				t.Logf("App: %s [%s] (%s)", a.Name, a.Category, a.Path)
			}
		}
	}
}

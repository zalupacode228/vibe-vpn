package storage

import (
	"testing"
)

func TestStorageLoad(t *testing.T) {
	s := NewStorage()

	settings := s.LoadSettings()
	if settings.SplitMode == "" {
		t.Errorf("Expected non-empty SplitMode")
	}

	nodes := s.LoadNodes()
	t.Logf("Loaded %d nodes from storage", len(nodes))

	subs := s.LoadSubscriptions()
	t.Logf("Loaded %d subscriptions from storage", len(subs))

	rules := s.LoadRules()
	t.Logf("Loaded %d rules from storage", len(rules))
}

func TestStorageSaveLoadWarpDetour(t *testing.T) {
	s := NewStorage()
	initial := s.LoadSettings()

	// Toggle to opposite
	target := !initial.WarpDetour
	initial.WarpDetour = target
	if err := s.SaveSettings(initial); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	reloaded := s.LoadSettings()
	if reloaded.WarpDetour != target {
		t.Fatalf("Expected WarpDetour to be %v, got %v", target, reloaded.WarpDetour)
	}

	// Toggle back
	initial.WarpDetour = !target
	if err := s.SaveSettings(initial); err != nil {
		t.Fatalf("SaveSettings revert failed: %v", err)
	}

	reloaded2 := s.LoadSettings()
	if reloaded2.WarpDetour != !target {
		t.Fatalf("Expected WarpDetour to revert to %v, got %v", !target, reloaded2.WarpDetour)
	}
}

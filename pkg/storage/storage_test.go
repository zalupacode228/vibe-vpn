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

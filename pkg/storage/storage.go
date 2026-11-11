package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"vibe-go/pkg/models"
)

type Storage struct {
	mu           sync.RWMutex
	appDir       string
	settingsPath string
	nodesPath    string
	subsPath     string
	rulesPath    string
	listsPath    string
}

func NewStorage() *Storage {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	dir := filepath.Join(appData, "VPN")
	_ = os.MkdirAll(dir, 0755)

	return &Storage{
		appDir:       dir,
		settingsPath: filepath.Join(dir, "settings.json"),
		nodesPath:    filepath.Join(dir, "nodes.json"),
		subsPath:     filepath.Join(dir, "subscriptions.json"),
		rulesPath:    filepath.Join(dir, "saved_rules.json"),
		listsPath:    filepath.Join(dir, "custom_lists.json"),
	}
}

func (s *Storage) GetAppDir() string {
	return s.appDir
}

func (s *Storage) LoadSettings() models.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	defaultSettings := models.Settings{
		Language:       "auto",
		SplitMode:      "whitelist",
		PrivacyBlur:    false,
		AutoStart:      false,
		MinimizeToTray: true,
		StrictRoute:    false,
		BypassLAN:      true,
		BypassRU:       false,
		MTU:            1400,
		TunMTU:         1400,
		RemoteDNS:      "https://1.1.1.1/dns-query",
		DirectDNS:      "local",
		UserAgent:      "v2rayNG/1.8.5",
		HWID:           GetMachineHWID(),
	}

	data, err := os.ReadFile(s.settingsPath)
	if err != nil {
		return defaultSettings
	}

	var loaded models.Settings
	if err := json.Unmarshal(data, &loaded); err != nil {
		return defaultSettings
	}
	if loaded.Language == "" {
		loaded.Language = "auto"
	}
	if loaded.SplitMode == "" {
		loaded.SplitMode = "whitelist"
	}
	if loaded.MTU == 0 {
		if loaded.TunMTU != 0 {
			loaded.MTU = loaded.TunMTU
		} else {
			loaded.MTU = 1400
		}
	}
	if loaded.UserAgent == "" {
		loaded.UserAgent = "v2rayNG/1.8.5"
	}
	if loaded.HWID == "" {
		loaded.HWID = GetMachineHWID()
	}
	return loaded
}

func (s *Storage) SaveSettings(settings models.Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read existing map to keep other properties untouched
	var existing map[string]interface{}
	if data, err := os.ReadFile(s.settingsPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}
	if existing == nil {
		existing = make(map[string]interface{})
	}

	existing["language"] = settings.Language
	existing["split_mode"] = settings.SplitMode
	existing["active_node_id"] = settings.ActiveNodeID
	existing["privacy_blur"] = settings.PrivacyBlur
	existing["auto_start"] = settings.AutoStart
	existing["minimize_to_tray"] = settings.MinimizeToTray
	existing["strict_route"] = settings.StrictRoute
	existing["bypass_lan"] = settings.BypassLAN
	existing["bypass_ru"] = settings.BypassRU
	existing["adblock"] = settings.AdBlock
	existing["mtu"] = settings.MTU
	if settings.RemoteDNS != "" {
		existing["remote_dns"] = settings.RemoteDNS
	}
	if settings.DirectDNS != "" {
		existing["direct_dns"] = settings.DirectDNS
	}
	existing["user_agent"] = settings.UserAgent
	existing["hwid"] = settings.HWID

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.settingsPath, data, 0644)
}

func (s *Storage) LoadNodes() []models.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.nodesPath)
	if err != nil {
		data, err = os.ReadFile(filepath.Join(s.appDir, "user_nodes.json"))
		if err != nil {
			return []models.Node{}
		}
	}
	var nodes []models.Node
	_ = json.Unmarshal(data, &nodes)
	if nodes == nil {
		return []models.Node{}
	}
	return nodes
}

func (s *Storage) SaveNodes(nodes []models.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if nodes == nil {
		nodes = []models.Node{}
	}

	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}
	_ = os.WriteFile(filepath.Join(s.appDir, "user_nodes.json"), data, 0644)
	return os.WriteFile(s.nodesPath, data, 0644)
}

func (s *Storage) LoadSubscriptions() []models.Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.subsPath)
	if err != nil {
		return []models.Subscription{}
	}
	var subs []models.Subscription
	_ = json.Unmarshal(data, &subs)
	if subs == nil {
		return []models.Subscription{}
	}
	return subs
}

func (s *Storage) SaveSubscriptions(subs []models.Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if subs == nil {
		subs = []models.Subscription{}
	}

	data, err := json.MarshalIndent(subs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.subsPath, data, 0644)
}

func (s *Storage) LoadRules() []models.AppRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.rulesPath)
	if err != nil {
		return []models.AppRule{}
	}
	var rules []models.AppRule
	_ = json.Unmarshal(data, &rules)
	for i := range rules {
		if rules[i].IsEnabled {
			rules[i].Enabled = true
		}
	}
	return rules
}

func (s *Storage) SaveRules(rules []models.AppRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range rules {
		rules[i].IsEnabled = rules[i].Enabled
	}
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.rulesPath, data, 0644)
}

func (s *Storage) LoadCustomLists() []models.CustomList {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.listsPath)
	if err != nil {
		return []models.CustomList{}
	}
	var lists []models.CustomList
	_ = json.Unmarshal(data, &lists)

	// Populate domains and IPs from cached data files if empty
	dataDir := filepath.Join(s.appDir, "custom_lists_data")
	for i := range lists {
		if len(lists[i].Domains) == 0 && len(lists[i].IPs) == 0 {
			cacheFile := filepath.Join(dataDir, lists[i].ID+".json")
			if cBytes, err := os.ReadFile(cacheFile); err == nil {
				var cached struct {
					Domains []string `json:"domains"`
					IPs     []string `json:"ips"`
				}
				if err := json.Unmarshal(cBytes, &cached); err == nil {
					lists[i].Domains = cached.Domains
					lists[i].IPs = cached.IPs
				}
			}
		}
	}

	return lists
}

func (s *Storage) SaveCustomLists(lists []models.CustomList) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dataDir := filepath.Join(s.appDir, "custom_lists_data")
	_ = os.MkdirAll(dataDir, 0755)

	// Save individual data cache files for each list
	for _, l := range lists {
		if len(l.Domains) > 0 || len(l.IPs) > 0 {
			cached := map[string]interface{}{
				"domains": l.Domains,
				"ips":     l.IPs,
			}
			cBytes, _ := json.Marshal(cached)
			_ = os.WriteFile(filepath.Join(dataDir, l.ID+".json"), cBytes, 0644)
		}
	}

	data, err := json.MarshalIndent(lists, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.listsPath, data, 0644)
}

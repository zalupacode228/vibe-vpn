package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vibe-go/pkg/models"
	"vibe-go/pkg/pinger"
	"vibe-go/pkg/procmon"
	"vibe-go/pkg/singbox"
	"vibe-go/pkg/storage"
	"vibe-go/pkg/subs"
	"vibe-go/pkg/tray"
	"vibe-go/pkg/warp"
)

//go:embed assets/core/sing-box.exe assets/core/wintun.dll
var embeddedCore embed.FS

type App struct {
	ctx          context.Context
	storage      *storage.Storage
	controller   *singbox.Controller
	builder      *singbox.ConfigBuilder
	tray         *tray.Tray
	activeNodeID string
	mu           sync.RWMutex
}

func NewApp() *App {
	s := storage.NewStorage()
	settings := s.LoadSettings()
	activeID := settings.ActiveNodeID

	nodes := s.LoadNodes()
	if activeID == "" && len(nodes) > 0 {
		activeID = nodes[0].ID
	}

	return &App{
		storage:      s,
		controller:   singbox.NewController(),
		builder:      singbox.NewConfigBuilder(),
		activeNodeID: activeID,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.unpackEmbeddedCore()
	t, err := tray.InitTray("VPN", func() {
		if a.ctx != nil {
			runtime.WindowShow(a.ctx)
			runtime.WindowUnminimise(a.ctx)
			runtime.WindowSetAlwaysOnTop(a.ctx, true)
			runtime.WindowSetAlwaysOnTop(a.ctx, false)
		}
	}, func() {
		if a.ctx != nil {
			runtime.Quit(a.ctx)
		}
	})
	if err == nil {
		a.tray = t
	}
}

func (a *App) unpackEmbeddedCore() {
	appData := a.storage.GetAppDir()
	coreDir := filepath.Join(appData, "core")
	_ = os.MkdirAll(coreDir, 0755)

	singboxDst := filepath.Join(coreDir, "sing-box.exe")
	if info, err := os.Stat(singboxDst); err != nil || info.Size() == 0 {
		if data, err := embeddedCore.ReadFile("assets/core/sing-box.exe"); err == nil {
			_ = os.WriteFile(singboxDst, data, 0755)
		}
	}

	wintunDst := filepath.Join(coreDir, "wintun.dll")
	if info, err := os.Stat(wintunDst); err != nil || info.Size() == 0 {
		if data, err := embeddedCore.ReadFile("assets/core/wintun.dll"); err == nil {
			_ = os.WriteFile(wintunDst, data, 0644)
		}
	}
}

func (a *App) domReady(ctx context.Context) {
	// Ready hook
}

func (a *App) shutdown(ctx context.Context) {
	if a.tray != nil {
		a.tray.Close()
	}
	if a.controller != nil {
		_ = a.controller.Stop()
	}
}

// Window Controls
func (a *App) WindowMin() {
	if a.ctx != nil {
		runtime.WindowMinimise(a.ctx)
	}
}

func (a *App) WindowToggleMax() {
	if a.ctx != nil {
		runtime.WindowToggleMaximise(a.ctx)
	}
}

func (a *App) WindowClose() {
	if a.ctx == nil {
		return
	}
	settings := a.storage.LoadSettings()
	if settings.MinimizeToTray {
		runtime.WindowHide(a.ctx)
	} else {
		runtime.Quit(a.ctx)
	}
}

func (a *App) SelectExeFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("context not initialized")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Выберите исполняемый файл приложения",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Исполняемые файлы (*.exe)",
				Pattern:     "*.exe",
			},
		},
	})
}

// Settings
func (a *App) GetSettings() models.Settings {
	s := a.storage.LoadSettings()
	a.mu.RLock()
	s.ActiveNodeID = a.activeNodeID
	a.mu.RUnlock()
	return s
}

func (a *App) SaveSettings(settings models.Settings) error {
	a.mu.Lock()
	if settings.ActiveNodeID != "" {
		a.activeNodeID = settings.ActiveNodeID
	} else {
		settings.ActiveNodeID = a.activeNodeID
	}
	a.mu.Unlock()
	return a.storage.SaveSettings(settings)
}

func (a *App) GetSystemLanguage() string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetUserDefaultLocaleName")
	buf := make([]uint16, 85)
	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r > 0 {
		locale := syscall.UTF16ToString(buf)
		lower := strings.ToLower(locale)
		if strings.HasPrefix(lower, "ru") || strings.HasPrefix(lower, "be") || strings.HasPrefix(lower, "uk") || strings.HasPrefix(lower, "kk") {
			return "ru"
		}
	}
	return "en"
}

// Nodes
func (a *App) GetNodes() []models.Node {
	return a.storage.LoadNodes()
}

func (a *App) SaveNodes(nodes []models.Node) error {
	return a.storage.SaveNodes(nodes)
}

func (a *App) GetActiveNodeID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.activeNodeID
}

func (a *App) SetActiveNodeID(id string) error {
	a.mu.Lock()
	a.activeNodeID = id
	a.mu.Unlock()

	s := a.storage.LoadSettings()
	s.ActiveNodeID = id
	if err := a.storage.SaveSettings(s); err != nil {
		return err
	}

	// If VPN is currently connected, hot-switch node on the fly!
	if a.controller.IsRunning() {
		_, _, err := a.ToggleVPN(true)
		return err
	}

	return nil
}

func (a *App) AddManualNode(link string) (*models.Node, error) {
	link = strings.TrimSpace(link)
	if link == "" {
		return nil, fmt.Errorf("пустая ссылка")
	}
	node := subs.ParseNodeLink(link)
	if node == nil {
		return nil, fmt.Errorf("не удалось разобрать ссылку узла")
	}

	nodes := a.storage.LoadNodes()
	// Check if already exists
	for _, n := range nodes {
		if n.ID == node.ID || (n.Server == node.Server && n.Port == node.Port) {
			return &n, nil
		}
	}
	nodes = append(nodes, *node)
	if err := a.storage.SaveNodes(nodes); err != nil {
		return nil, err
	}
	return node, nil
}

type FileResult struct {
	Content  string `json:"content"`
	FileName string `json:"file_name"`
}

func (a *App) GenerateWarp(mode string) (*models.Node, error) {
	node, err := warp.GenerateWarp(mode)
	if err != nil {
		return nil, err
	}

	nodes := a.storage.LoadNodes()
	// Place generated node at the top
	nodes = append([]models.Node{*node}, nodes...)
	if err := a.storage.SaveNodes(nodes); err != nil {
		return nil, err
	}

	_ = a.SetActiveNodeID(node.ID)
	return node, nil
}

func (a *App) SelectWarpConfigFile() (*FileResult, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("context not initialized")
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Выберите файл конфигурации WARP / WireGuard (.conf)",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "WireGuard & AmneziaWG Config (*.conf, *.wireguard, *.txt)",
				Pattern:     "*.conf;*.wireguard;*.txt",
			},
			{
				DisplayName: "Все файлы (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil || path == "" {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл: %w", err)
	}
	return &FileResult{
		Content:  string(data),
		FileName: filepath.Base(path),
	}, nil
}

func (a *App) ImportWarpConfig(content string, customName string) (*models.Node, error) {
	node, err := warp.ParseWireguardConf(content, customName)
	if err != nil {
		return nil, err
	}

	nodes := a.storage.LoadNodes()
	// Place imported node at the top
	nodes = append([]models.Node{*node}, nodes...)
	if err := a.storage.SaveNodes(nodes); err != nil {
		return nil, err
	}

	_ = a.SetActiveNodeID(node.ID)
	return node, nil
}

func (a *App) GetCoreLogs(maxLines int) string {
	return a.controller.GetLogs(maxLines)
}

func (a *App) ClearCoreLogs() error {
	return a.controller.ClearLogs()
}

func (a *App) DeleteNode(id string) error {
	nodes := a.storage.LoadNodes()
	updated := make([]models.Node, 0)
	for _, n := range nodes {
		if n.ID != id {
			updated = append(updated, n)
		}
	}

	a.mu.Lock()
	if a.activeNodeID == id {
		if len(updated) > 0 {
			a.activeNodeID = updated[0].ID
		} else {
			a.activeNodeID = ""
		}
		s := a.storage.LoadSettings()
		s.ActiveNodeID = a.activeNodeID
		_ = a.storage.SaveSettings(s)
	}
	a.mu.Unlock()

	if len(updated) == 0 && a.controller.IsRunning() {
		_ = a.controller.Stop()
	}

	return a.storage.SaveNodes(updated)
}

func (a *App) PingAllNodes() map[string]int {
	nodes := a.storage.LoadNodes()
	res := pinger.PingAll(nodes)

	// Update cached ping
	for i := range nodes {
		if p, ok := res[nodes[i].ID]; ok {
			nodes[i].Ping = p
		}
	}
	_ = a.storage.SaveNodes(nodes)
	return res
}

func (a *App) PingNode(id string) int {
	nodes := a.storage.LoadNodes()
	for i, n := range nodes {
		if n.ID == id {
			p := pinger.PingAddress(n.Server, n.Port, 3*time.Second)
			nodes[i].Ping = p
			_ = a.storage.SaveNodes(nodes)
			return p
		}
	}
	return 0
}

// Subscriptions
func (a *App) GetSubscriptions() []models.Subscription {
	return a.storage.LoadSubscriptions()
}

func (a *App) AddSubscription(name string, url string) ([]models.Node, error) {
	name = strings.TrimSpace(name)
	url = strings.TrimSpace(url)
	if name == "" || url == "" {
		return nil, fmt.Errorf("имя и ссылка не могут быть пустыми")
	}

	subID := fmt.Sprintf("sub_%d", time.Now().Unix())
	parsedNodes, err := subs.FetchAndParse(url, subID, name)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки подписки: %w", err)
	}

	subsList := a.storage.LoadSubscriptions()
	newSub := models.Subscription{
		ID:        subID,
		Name:      name,
		URL:       url,
		UpdatedAt: time.Now().Format("02.01 15:04"),
		Count:     len(parsedNodes),
	}
	subsList = append(subsList, newSub)
	_ = a.storage.SaveSubscriptions(subsList)

	// Merge into all nodes
	allNodes := a.storage.LoadNodes()
	allNodes = append(allNodes, parsedNodes...)
	_ = a.storage.SaveNodes(allNodes)

	return parsedNodes, nil
}

func (a *App) RefreshSubscription(id string) ([]models.Node, error) {
	subsList := a.storage.LoadSubscriptions()
	var targetSub *models.Subscription
	targetIdx := -1
	for i := range subsList {
		if subsList[i].ID == id {
			targetSub = &subsList[i]
			targetIdx = i
			break
		}
	}
	if targetSub == nil {
		return nil, fmt.Errorf("подписка не найдена")
	}

	newNodes, err := subs.FetchAndParse(targetSub.URL, targetSub.ID, targetSub.Name)
	if err != nil {
		return nil, err
	}

	subsList[targetIdx].UpdatedAt = time.Now().Format("02.01 15:04")
	subsList[targetIdx].Count = len(newNodes)
	_ = a.storage.SaveSubscriptions(subsList)

	// Replace nodes for this sub
	allNodes := a.storage.LoadNodes()
	var preserved []models.Node
	for _, n := range allNodes {
		if n.SubscriptionID != id {
			preserved = append(preserved, n)
		}
	}
	preserved = append(preserved, newNodes...)
	_ = a.storage.SaveNodes(preserved)

	return newNodes, nil
}

func (a *App) DeleteSubscription(id string) error {
	subsList := a.storage.LoadSubscriptions()
	updatedSubs := make([]models.Subscription, 0)
	for _, s := range subsList {
		if s.ID != id {
			updatedSubs = append(updatedSubs, s)
		}
	}
	_ = a.storage.SaveSubscriptions(updatedSubs)

	allNodes := a.storage.LoadNodes()
	preserved := make([]models.Node, 0)
	for _, n := range allNodes {
		if n.SubscriptionID != id {
			preserved = append(preserved, n)
		}
	}

	a.mu.Lock()
	foundActive := false
	for _, n := range preserved {
		if n.ID == a.activeNodeID {
			foundActive = true
			break
		}
	}
	if !foundActive {
		if len(preserved) > 0 {
			a.activeNodeID = preserved[0].ID
		} else {
			a.activeNodeID = ""
		}
		s := a.storage.LoadSettings()
		s.ActiveNodeID = a.activeNodeID
		_ = a.storage.SaveSettings(s)
	}
	a.mu.Unlock()

	if len(preserved) == 0 && a.controller.IsRunning() {
		_ = a.controller.Stop()
	}

	return a.storage.SaveNodes(preserved)
}

// App Rules & Processes
func (a *App) GetAppRules() []models.AppRule {
	return a.storage.LoadRules()
}

func (a *App) SaveAppRules(rules []models.AppRule) error {
	return a.storage.SaveRules(rules)
}

func (a *App) GetRunningApps() []models.ProcessInfo {
	rules := a.storage.LoadRules()
	return procmon.GetRunningApps(rules)
}

func (a *App) ToggleAppRule(path string, name string, enabled bool) error {
	rules := a.storage.LoadRules()
	found := false
	for i := range rules {
		if strings.EqualFold(rules[i].Path, path) || strings.EqualFold(rules[i].Name, name) {
			rules[i].Enabled = enabled
			rules[i].IsEnabled = enabled
			found = true
			break
		}
	}
	if !found {
		rules = append(rules, models.AppRule{
			Path:      path,
			Name:      name,
			Display:   strings.TrimSuffix(name, filepath.Ext(name)),
			Enabled:   enabled,
			IsEnabled: enabled,
		})
	}
	return a.storage.SaveRules(rules)
}

func (a *App) AddCustomApp(path string) (*models.AppRule, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("пустой путь к файлу")
	}
	name := filepath.Base(path)
	rule := models.AppRule{
		Path:      path,
		Name:      name,
		Display:   strings.TrimSuffix(name, filepath.Ext(name)),
		Category:  "Приложение",
		Enabled:   true,
		IsEnabled: true,
		IsCustom:  true,
	}

	rules := a.storage.LoadRules()
	for _, r := range rules {
		if strings.EqualFold(r.Path, path) {
			return &r, nil
		}
	}
	rules = append(rules, rule)
	if err := a.storage.SaveRules(rules); err != nil {
		return nil, err
	}
	return &rule, nil
}

// Admin rights
func (a *App) IsAdmin() bool {
	return isAdmin()
}

func (a *App) RestartAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	cwd, _ := syscall.UTF16PtrFromString(filepath.Dir(exe))

	releaseSingleInstanceMutex()

	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		0,
		uintptr(unsafe.Pointer(cwd)),
		1, // SW_SHOWNORMAL
	)
	if ret <= 32 {
		return fmt.Errorf("повышение прав отменено или ошибка (код %d)", ret)
	}

	if a.tray != nil {
		a.tray.Close()
	}
	if a.controller != nil {
		_ = a.controller.Stop()
	}

	os.Exit(0)
	return nil
}

// VPN conflict detection
func (a *App) CheckVPNConflicts() []string {
	return procmon.CheckVPNConflicts()
}

func (a *App) KillConflictingVPNs() error {
	return procmon.KillConflictingVPNs()
}

// Custom lists
func (a *App) GetCustomLists() []models.CustomList {
	return a.storage.LoadCustomLists()
}

func (a *App) ToggleSmartRule(key string, val bool) error {
	s := a.storage.LoadSettings()
	switch key {
	case "bypass_ru":
		s.BypassRU = val
	case "adblock":
		s.AdBlock = val
	case "privacy_blur":
		s.PrivacyBlur = val
	case "auto_start":
		s.AutoStart = val
	case "minimize_to_tray":
		s.MinimizeToTray = val
	case "strict_route":
		s.StrictRoute = val
	case "bypass_lan":
		s.BypassLAN = val
	default:
		// Fallback for any other boolean key
	}
	return a.storage.SaveSettings(s)
}

func (a *App) AddPresetList(presetKey string) (*models.CustomList, error) {
	lists := a.storage.LoadCustomLists()

	type presetDef struct {
		Name    string
		URL     string
		Action  string
		Domains []string
	}
	presets := map[string]presetDef{
		"youtube_discord": {
			Name:    "YouTube & Discord Fix",
			URL:     "https://raw.githubusercontent.com/runetfreedom/russia-v2ray-rules-dat/release/youtube.txt",
			Action:  "proxy",
			Domains: []string{"youtube.com", "googlevideo.com", "ytimg.com", "youtu.be", "ggpht.com", "discord.com", "discordapp.com", "discord.gg", "discord.media", "discordapp.net"},
		},
		"social_media": {
			Name:    "Соцсети (Instagram, X, Facebook)",
			URL:     "https://raw.githubusercontent.com/runetfreedom/russia-v2ray-rules-dat/release/meta.txt",
			Action:  "proxy",
			Domains: []string{"instagram.com", "cdninstagram.com", "facebook.com", "fbcdn.net", "twitter.com", "x.com", "t.co", "twimg.com", "threads.net"},
		},
		"ai_services": {
			Name:    "OpenAI & Claude",
			URL:     "https://raw.githubusercontent.com/v2fly/domain-list-community/release/openai.txt",
			Action:  "proxy",
			Domains: []string{"openai.com", "chatgpt.com", "oaistatic.com", "oaiusercontent.com", "claude.ai", "anthropic.com", "notion.so"},
		},
	}

	p, ok := presets[presetKey]
	if !ok {
		return nil, fmt.Errorf("пресет не найден")
	}

	for _, cl := range lists {
		if cl.Name == p.Name {
			return &cl, nil
		}
	}

	listID := fmt.Sprintf("preset-%s", presetKey)
	newItem := models.CustomList{
		ID:          listID,
		Name:        p.Name,
		URL:         p.URL,
		Action:      p.Action,
		Enabled:     true,
		Domains:     p.Domains,
		Count:       len(p.Domains),
		LastUpdated: "Встроенный набор",
	}

	lists = append(lists, newItem)
	if err := a.storage.SaveCustomLists(lists); err != nil {
		return nil, err
	}
	return &newItem, nil
}

func (a *App) AddCustomList(name string, urlStr string, action string) (*models.CustomList, error) {
	name = strings.TrimSpace(name)
	urlStr = strings.TrimSpace(urlStr)
	if name == "" || urlStr == "" {
		return nil, fmt.Errorf("имя и URL не могут быть пустыми")
	}
	if action == "" {
		action = "proxy"
	}

	var bodyBytes []byte
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(urlStr)
		if err != nil {
			return nil, fmt.Errorf("ошибка скачивания списка: %w", err)
		}
		defer resp.Body.Close()

		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
		}
	} else if data, err := os.ReadFile(urlStr); err == nil {
		bodyBytes = data
	} else {
		return nil, fmt.Errorf("укажите корректный URL (https://...) или локальный путь к файлу")
	}

	lines := strings.Split(string(bodyBytes), "\n")
	var domains []string
	var ips []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, "//") {
			continue
		}
		l = strings.TrimPrefix(l, "domain:")
		l = strings.TrimPrefix(l, "full:")
		l = strings.TrimPrefix(l, "regexp:")
		if l == "" {
			continue
		}

		if _, _, err := net.ParseCIDR(l); err == nil {
			ips = append(ips, l)
		} else if parsedIP := net.ParseIP(l); parsedIP != nil {
			if parsedIP.To4() != nil {
				ips = append(ips, l+"/32")
			} else {
				ips = append(ips, l+"/128")
			}
		} else {
			domains = append(domains, l)
		}
	}

	listID := fmt.Sprintf("list-%d", time.Now().UnixMilli())
	newItem := models.CustomList{
		ID:          listID,
		Name:        name,
		URL:         urlStr,
		Action:      action,
		Enabled:     true,
		Domains:     domains,
		IPs:         ips,
		Count:       len(domains) + len(ips),
		LastUpdated: time.Now().Format("02.01 15:04"),
	}

	lists := a.storage.LoadCustomLists()
	lists = append(lists, newItem)
	if err := a.storage.SaveCustomLists(lists); err != nil {
		return nil, err
	}
	return &newItem, nil
}

func (a *App) RefreshAllCustomLists() (int, error) {
	lists := a.storage.LoadCustomLists()
	client := &http.Client{Timeout: 15 * time.Second}
	updatedCount := 0

	for i := range lists {
		var bodyBytes []byte
		if strings.HasPrefix(lists[i].URL, "http://") || strings.HasPrefix(lists[i].URL, "https://") {
			resp, err := client.Get(lists[i].URL)
			if err != nil {
				continue
			}
			data, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}
			bodyBytes = data
		} else if data, err := os.ReadFile(lists[i].URL); err == nil {
			bodyBytes = data
		} else {
			continue
		}

		lines := strings.Split(string(bodyBytes), "\n")
		var domains []string
		var ips []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, "//") {
				continue
			}
			l = strings.TrimPrefix(l, "domain:")
			l = strings.TrimPrefix(l, "full:")
			l = strings.TrimPrefix(l, "regexp:")
			if l == "" {
				continue
			}

			if _, _, err := net.ParseCIDR(l); err == nil {
				ips = append(ips, l)
			} else if parsedIP := net.ParseIP(l); parsedIP != nil {
				if parsedIP.To4() != nil {
					ips = append(ips, l+"/32")
				} else {
					ips = append(ips, l+"/128")
				}
			} else {
				domains = append(domains, l)
			}
		}

		if len(domains) > 0 || len(ips) > 0 {
			lists[i].Domains = domains
			lists[i].IPs = ips
			lists[i].Count = len(domains) + len(ips)
			lists[i].LastUpdated = time.Now().Format("02.01 15:04")
			updatedCount++
		}
	}
	_ = a.storage.SaveCustomLists(lists)
	return updatedCount, nil
}

func (a *App) ToggleCustomList(id string, enabled bool) error {
	lists := a.storage.LoadCustomLists()
	for i := range lists {
		if lists[i].ID == id {
			lists[i].Enabled = enabled
			break
		}
	}
	return a.storage.SaveCustomLists(lists)
}

func (a *App) DeleteCustomList(id string) error {
	lists := a.storage.LoadCustomLists()
	var updated []models.CustomList
	for _, l := range lists {
		if l.ID != id {
			updated = append(updated, l)
		}
	}
	return a.storage.SaveCustomLists(updated)
}

// VPN Connection Control
func (a *App) ToggleVPN(enabled bool) (bool, string, error) {
	if !enabled {
		if err := a.controller.Stop(); err != nil {
			return false, err.Error(), err
		}
		return false, "Отключено", nil
	}

	// Starting VPN
	settings := a.storage.LoadSettings()
	nodes := a.storage.LoadNodes()
	if len(nodes) == 0 {
		return false, "Нет добавленных узлов", fmt.Errorf("no nodes available")
	}

	a.mu.RLock()
	currentID := a.activeNodeID
	a.mu.RUnlock()

	var activeNode *models.Node
	for i := range nodes {
		if nodes[i].ID == currentID {
			activeNode = &nodes[i]
			break
		}
	}
	if activeNode == nil {
		activeNode = &nodes[0]
		a.mu.Lock()
		a.activeNodeID = activeNode.ID
		a.mu.Unlock()
	}

	rules := a.storage.LoadRules()
	settings.CustomLists = a.storage.LoadCustomLists()
	configBytes, err := a.builder.GenerateConfig(settings.SplitMode, rules, activeNode, settings)
	if err != nil {
		return false, fmt.Sprintf("Ошибка конфигурации: %v", err), err
	}

	if err := a.controller.Start(configBytes); err != nil {
		return false, fmt.Sprintf("Ошибка запуска sing-box: %v", err), err
	}

	return true, "Подключено", nil
}

func (a *App) GetTrafficStats() models.TrafficStats {
	running, _, down, up, totalDown, totalUp := a.controller.GetStatus()
	settings := a.storage.LoadSettings()

	var nodeName string
	var serverIP string
	var ping int

	a.mu.RLock()
	curID := a.activeNodeID
	a.mu.RUnlock()

	nodes := a.storage.LoadNodes()
	for _, n := range nodes {
		if n.ID == curID {
			nodeName = n.Name
			serverIP = n.Server
			ping = n.Ping
			break
		}
	}

	modeStr := "Белый список"
	switch settings.SplitMode {
	case "blacklist":
		modeStr = "Черный список"
	case "global":
		modeStr = "Весь трафик (Global)"
	case "direct":
		modeStr = "Прямой (Bypass)"
	}

	return models.TrafficStats{
		Connected: running,
		DownSpeed: down,
		UpSpeed:   up,
		TotalDown: totalDown,
		TotalUp:   totalUp,
		Ping:      ping,
		IP:        serverIP,
		NodeName:  nodeName,
		Mode:      modeStr,
	}
}

func (a *App) GetConnections() []models.ConnectionGroup {
	conns, err := a.controller.GetConnections()
	if err != nil {
		return []models.ConnectionGroup{}
	}
	return conns
}

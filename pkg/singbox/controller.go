package singbox

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"vibe-go/pkg/models"
)

type Controller struct {
	mu          sync.Mutex
	corePath    string
	configPath  string
	cmd         *exec.Cmd
	logFile     *os.File
	isRunning   bool
	status      string
	downSpeed   int64
	upSpeed     int64
	totalDown   int64
	totalUp     int64
	stopTraffic chan struct{}
}

func NewController() *Controller {
	appData := os.Getenv("APPDATA")
	appDataCore := filepath.Join(appData, "VPN", "core", "sing-box.exe")
	configPath := filepath.Join(appData, "VPN", "config.json")

	// Check relative to executable first
	var corePath string
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidate := filepath.Join(exeDir, "assets", "core", "sing-box.exe")
		if _, err := os.Stat(candidate); err == nil {
			corePath = candidate
		}
	}

	if corePath == "" {
		if _, err := os.Stat(appDataCore); err == nil {
			corePath = appDataCore
		} else {
			corePath = "sing-box.exe"
		}
	}

	return &Controller{
		corePath:   corePath,
		configPath: configPath,
		status:     "Отключено",
	}
}

func (c *Controller) Start(configBytes []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		_ = c.stopInternal()
	}

	// Terminate any previous or orphaned sing-box instances
	killCmd := exec.Command("taskkill", "/F", "/IM", "sing-box.exe")
	killCmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	_ = killCmd.Run()
	time.Sleep(100 * time.Millisecond)

	// Ensure config directory exists and write config
	cfgDir := filepath.Dir(c.configPath)
	_ = os.MkdirAll(cfgDir, 0755)
	if err := os.WriteFile(c.configPath, configBytes, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Verify sing-box binary exists
	if _, err := os.Stat(c.corePath); err != nil {
		return fmt.Errorf("sing-box binary not found at %s", c.corePath)
	}

	coreDir := filepath.Dir(c.corePath)
	logPath := filepath.Join(cfgDir, "core.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logFile = nil
	}

	cmd := exec.Command(c.corePath, "run", "-c", c.configPath)
	cmd.Dir = coreDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		HideWindow:    true,
	}
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		return fmt.Errorf("failed to start sing-box: %w", err)
	}

	// Verification: wait 500ms to ensure process didn't exit immediately on error
	time.Sleep(500 * time.Millisecond)
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		if logFile != nil {
			logFile.Close()
		}
		errText := readLastLines(logPath, 15)
		return fmt.Errorf("sing-box crashed on startup: %s", errText)
	}

	c.cmd = cmd
	c.logFile = logFile
	c.isRunning = true
	c.status = "Подключено"

	c.stopTraffic = make(chan struct{})
	go c.pollTraffic(c.stopTraffic)

	return nil
}

func (c *Controller) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stopInternal()
}

func (c *Controller) stopInternal() error {
	if c.stopTraffic != nil {
		close(c.stopTraffic)
		c.stopTraffic = nil
	}
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
		c.cmd = nil
	}
	if c.logFile != nil {
		c.logFile.Close()
		c.logFile = nil
	}

	killCmd := exec.Command("taskkill", "/F", "/IM", "sing-box.exe")
	killCmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	_ = killCmd.Run()

	c.isRunning = false
	c.status = "Отключено"
	c.downSpeed = 0
	c.upSpeed = 0
	return nil
}

func (c *Controller) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isRunning
}

func (c *Controller) GetStatus() (bool, string, int64, int64, int64, int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isRunning, c.status, c.downSpeed, c.upSpeed, c.totalDown, c.totalUp
}

func (c *Controller) pollTraffic(stop chan struct{}) {
	client := &http.Client{Timeout: 3 * time.Second}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			resp, err := client.Get("http://127.0.0.1:9090/traffic")
			if err != nil {
				continue
			}
			var t struct {
				Up   int64 `json:"up"`
				Down int64 `json:"down"`
			}
			dec := json.NewDecoder(resp.Body)
			if err := dec.Decode(&t); err == nil {
				c.mu.Lock()
				c.downSpeed = t.Down
				c.upSpeed = t.Up
				c.totalDown += t.Down
				c.totalUp += t.Up
				c.mu.Unlock()
			}
			resp.Body.Close()
		}
	}
}

func formatBytesPretty(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	kb := float64(b) / 1024.0
	if kb < 1024.0 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024.0
	if mb < 1024.0 {
		return fmt.Sprintf("%.1f MB", mb)
	}
	gb := mb / 1024.0
	return fmt.Sprintf("%.2f GB", gb)
}

func (c *Controller) GetConnections() ([]models.ConnectionGroup, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:9090/connections")
	if err != nil {
		return []models.ConnectionGroup{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Connections []struct {
			ID       string `json:"id"`
			Metadata struct {
				Process         string `json:"process"`
				ProcessPath     string `json:"processPath"`
				Host            string `json:"host"`
				DestinationIP   string `json:"destinationIP"`
				DestinationPort string `json:"destinationPort"`
				SourceIP        string `json:"sourceIP"`
				Network         string `json:"network"`
				Type            string `json:"type"`
			} `json:"metadata"`
			Rule     string   `json:"rule"`
			Chains   []string `json:"chains"`
			Upload   int64    `json:"upload"`
			Download int64    `json:"download"`
		} `json:"connections"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return []models.ConnectionGroup{}, err
	}

	type groupAgg struct {
		process  string
		count    int
		down     int64
		up       int64
		children []models.ConnectionItem
	}

	groupsMap := make(map[string]*groupAgg)
	for _, conn := range data.Connections {
		proc := conn.Metadata.Process
		if proc == "" && conn.Metadata.ProcessPath != "" {
			cleanPath := strings.ReplaceAll(conn.Metadata.ProcessPath, "/", "\\")
			proc = filepath.Base(cleanPath)
		}
		if proc == "" && conn.Metadata.SourceIP != "" {
			proc = conn.Metadata.SourceIP
		}
		if proc == "" {
			proc = "Системный трафик"
		}

		group, exists := groupsMap[proc]
		if !exists {
			group = &groupAgg{
				process:  proc,
				count:    0,
				children: []models.ConnectionItem{},
			}
			groupsMap[proc] = group
		}
		group.count++
		group.down += conn.Download
		group.up += conn.Upload

		dest := conn.Metadata.Host
		if dest == "" {
			dest = conn.Metadata.DestinationIP
		}
		if dest == "" {
			dest = "direct"
		}
		if conn.Metadata.DestinationPort != "" {
			dest = fmt.Sprintf("%s:%s", dest, conn.Metadata.DestinationPort)
		}

		rule := conn.Rule
		if rule == "" && len(conn.Chains) > 0 {
			rule = conn.Chains[len(conn.Chains)-1]
		}
		if rule == "" {
			rule = "proxy"
		}

		group.children = append(group.children, models.ConnectionItem{
			ID:          conn.ID,
			Destination: dest,
			Rule:        rule,
			DownSpeed:   formatBytesPretty(conn.Download),
			UpSpeed:     formatBytesPretty(conn.Upload),
		})
	}

	var sortedGroups []*groupAgg
	for _, g := range groupsMap {
		sortedGroups = append(sortedGroups, g)
	}
	sort.Slice(sortedGroups, func(i, j int) bool {
		return (sortedGroups[i].down + sortedGroups[i].up) > (sortedGroups[j].down + sortedGroups[j].up)
	})

	var result []models.ConnectionGroup
	for _, g := range sortedGroups {
		result = append(result, models.ConnectionGroup{
			Process:   g.process,
			Count:     g.count,
			DownSpeed: formatBytesPretty(g.down),
			UpSpeed:   formatBytesPretty(g.up),
			Children:  g.children,
		})
	}
	return result, nil
}

func (c *Controller) GetLogs(maxLines int) string {
	if maxLines <= 0 {
		maxLines = 250
	}
	cfgDir := filepath.Dir(c.configPath)
	logPath := filepath.Join(cfgDir, "core.log")
	return readLastLines(logPath, maxLines)
}

func (c *Controller) ClearLogs() error {
	cfgDir := filepath.Dir(c.configPath)
	logPath := filepath.Join(cfgDir, "core.log")
	return os.WriteFile(logPath, []byte(""), 0644)
}

func readLastLines(filePath string, n int) string {
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}
	return sb.String()
}

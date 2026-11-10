package procmon

import (
	"sort"
	"strings"
	"syscall"
	"unsafe"
	"vibe-go/pkg/models"
)

var (
	modKernel32                    = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot   = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW            = modKernel32.NewProc("Process32FirstW")
	procProcess32NextW             = modKernel32.NewProc("Process32NextW")
	procCloseHandle                = modKernel32.NewProc("CloseHandle")
	procOpenProcess                = modKernel32.NewProc("OpenProcess")
	procQueryFullProcessImageNameW = modKernel32.NewProc("QueryFullProcessImageNameW")
	procTerminateProcess           = modKernel32.NewProc("TerminateProcess")
)

const (
	TH32CS_SNAPPROCESS                = 0x00000002
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	PROCESS_TERMINATE                 = 0x0001
)

type PROCESSENTRY32W struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]uint16
}

func GetRunningApps(activeRules []models.AppRule) []models.ProcessInfo {
	seen := make(map[string]bool)
	ruleMap := make(map[string]bool)
	var apps []models.ProcessInfo

	// 1. Add configured rules FIRST (so user-configured rules always persist and stay at top)
	for _, r := range activeRules {
		rName := r.Name
		if rName == "" && r.Path != "" {
			parts := strings.Split(r.Path, `\`)
			rName = parts[len(parts)-1]
		}
		if rName == "" {
			continue
		}
		lowerName := strings.ToLower(rName)
		lowerPath := strings.ToLower(r.Path)

		enabled := r.Enabled || r.IsEnabled
		ruleMap[lowerName] = enabled
		if lowerPath != "" {
			ruleMap[lowerPath] = enabled
		}

		if !seen[lowerName] {
			seen[lowerName] = true
			if lowerPath != "" {
				seen[lowerPath] = true
			}
			cat := r.Category
			if cat == "" {
				cat = categorizeApp(lowerName)
			}
			icon := GetAppIcon(r.Path)
			apps = append(apps, models.ProcessInfo{
				Name:     rName,
				Path:     r.Path,
				Category: cat,
				Enabled:  enabled,
				Icon:     icon,
			})
		}
	}

	// 2. Scan running processes
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot != uintptr(syscall.InvalidHandle) {
		defer procCloseHandle.Call(snapshot)

		var entry PROCESSENTRY32W
		entry.DwSize = uint32(unsafe.Sizeof(entry))

		ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))

		// System processes to ignore
		ignore := map[string]bool{
			"svchost.exe": true, "system": true, "registry": true, "smss.exe": true,
			"csrss.exe": true, "wininit.exe": true, "services.exe": true, "lsass.exe": true,
			"fontdrvhost.exe": true, "dwm.exe": true, "conhost.exe": true, "sihost.exe": true,
			"taskhostw.exe": true,
		}

		for ret != 0 {
			exeName := syscall.UTF16ToString(entry.SzExeFile[:])
			lowerExe := strings.ToLower(exeName)

			if !ignore[lowerExe] && strings.HasSuffix(lowerExe, ".exe") && !seen[lowerExe] {
				seen[lowerExe] = true

				fullPath := getProcessPath(entry.Th32ProcessID)
				if fullPath == "" {
					fullPath = exeName
				}

				cat := categorizeApp(lowerExe)
				enabled := ruleMap[strings.ToLower(fullPath)] || ruleMap[lowerExe]
				icon := GetAppIcon(fullPath)

				apps = append(apps, models.ProcessInfo{
					Name:     exeName,
					Path:     fullPath,
					Category: cat,
					Enabled:  enabled,
					Icon:     icon,
				})
			}

			ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		}
	}

	// 3. Sort: Enabled apps ALWAYS at the very top, then alphabetically by name!
	sort.SliceStable(apps, func(i, j int) bool {
		if apps[i].Enabled != apps[j].Enabled {
			return apps[i].Enabled // true comes before false
		}
		return strings.ToLower(apps[i].Name) < strings.ToLower(apps[j].Name)
	})

	return apps
}

func getProcessPath(pid uint32) string {
	h, _, _ := procOpenProcess.Call(PROCESS_QUERY_LIMITED_INFORMATION, 0, uintptr(pid))
	if h == 0 {
		return ""
	}
	defer procCloseHandle.Call(h)

	var buf [1024]uint16
	size := uint32(len(buf))
	ret, _, _ := procQueryFullProcessImageNameW.Call(h, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:size])
}

func categorizeApp(name string) string {
	browsers := []string{"chrome.exe", "firefox.exe", "msedge.exe", "opera.exe", "brave.exe", "yandex.exe", "vivaldi.exe"}
	for _, b := range browsers {
		if strings.Contains(name, b) {
			return "Браузер"
		}
	}
	messengers := []string{"telegram.exe", "discord.exe", "whatsapp.exe", "signal.exe", "slack.exe", "skype.exe"}
	for _, m := range messengers {
		if strings.Contains(name, m) {
			return "Мессенджер"
		}
	}
	games := []string{"steam.exe", "epicgameslauncher.exe", "riotclientservices.exe", "battle.net.exe", "gog.exe"}
	for _, g := range games {
		if strings.Contains(name, g) {
			return "Игры"
		}
	}
	media := []string{"spotify.exe", "vlc.exe", "mpc-hc.exe", "obs64.exe", "aimp.exe"}
	for _, md := range media {
		if strings.Contains(name, md) {
			return "Медиа"
		}
	}
	return "Приложение"
}

var knownVPNProcesses = map[string]string{
	"nekobox.exe":         "NekoBox",
	"nekoray.exe":         "NekoRay",
	"v2rayn.exe":          "v2rayN",
	"v2ray.exe":           "v2ray",
	"clash.exe":           "Clash",
	"clash-verge.exe":     "Clash Verge",
	"clash-meta.exe":      "Clash Meta",
	"mihomo.exe":          "Mihomo / Clash",
	"flclash.exe":         "FlClash",
	"xray.exe":            "Xray Core",
	"wireguard.exe":       "WireGuard",
	"openvpn.exe":         "OpenVPN",
	"amneziavpn.exe":      "Amnezia VPN",
	"amnezia.exe":         "Amnezia",
	"outline.exe":         "Outline VPN",
	"warp-svc.exe":        "Cloudflare WARP",
	"cloudflare warp.exe": "Cloudflare WARP",
	"hiddify.exe":         "Hiddify Next",
	"zapret.exe":          "Zapret",
	"goodbyedpi.exe":      "GoodbyeDPI",
}

// CheckVPNConflicts returns a list of names of running third-party VPN/proxy applications
func CheckVPNConflicts() []string {
	var conflicts []string
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == uintptr(syscall.InvalidHandle) {
		return conflicts
	}
	defer procCloseHandle.Call(snapshot)

	var entry PROCESSENTRY32W
	entry.DwSize = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))

	myPID := uint32(syscall.Getpid())
	seen := make(map[string]bool)

	for ret != 0 {
		if entry.Th32ProcessID != myPID {
			exeName := strings.ToLower(syscall.UTF16ToString(entry.SzExeFile[:]))
			if label, ok := knownVPNProcesses[exeName]; ok && !seen[label] {
				seen[label] = true
				conflicts = append(conflicts, label)
			}
		}
		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	}

	return conflicts
}

// KillConflictingVPNs terminates known conflicting VPN background processes
func KillConflictingVPNs() error {
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == uintptr(syscall.InvalidHandle) {
		return nil
	}
	defer procCloseHandle.Call(snapshot)

	var entry PROCESSENTRY32W
	entry.DwSize = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))

	myPID := uint32(syscall.Getpid())

	for ret != 0 {
		if entry.Th32ProcessID != myPID {
			exeName := strings.ToLower(syscall.UTF16ToString(entry.SzExeFile[:]))
			if _, ok := knownVPNProcesses[exeName]; ok {
				hProc, _, _ := procOpenProcess.Call(PROCESS_TERMINATE, 0, uintptr(entry.Th32ProcessID))
				if hProc != 0 {
					procTerminateProcess.Call(hProc, 1)
					procCloseHandle.Call(hProc)
				}
			}
		}
		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	}

	return nil
}

# VibeVPN — Fast & Simple Windows VPN Client

<div align="center">

<p align="center">
  <a href="README.md">🇷🇺 Русский</a> | <b>🇬🇧 English</b>
</p>

![Platform](https://img.shields.io/badge/Platform-Windows%2010%20%2F%2011%20x64-0078D6?logo=windows&logoColor=white)
![Version](https://img.shields.io/badge/Version-v1.0.8-brightgreen.svg)
![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v2-DF0000?logo=wails&logoColor=white)
![Sing-Box](https://img.shields.io/badge/Core-sing--box%201.14-FF6B00?logo=box&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![i18n](https://img.shields.io/badge/Language-RU%20%7C%20EN-blueviolet.svg)

**A fast, modern, and lightweight Windows VPN client featuring per-application Split Tunneling, Cloudflare WARP (AmneziaWG / MASQUE) generator and config import, WARP Detour chained proxying, smart domain & IP/CIDR geo-routing, real-time connection log viewer, and network noise filtering.**

[Download Release v1.0.8 (.exe)](https://github.com/zalupacode228/vibe-vpn/releases/tag/v1.0.8) • [Key Features](#-key-features) • [Screenshots](#-screenshots) • [Building from Source](#-building-from-source)

</div>

---

## 📸 Screenshots

<div align="center">

| **Dashboard (Connection & Telemetry)** | **Servers & Nodes (Subscriptions & Ping)** |
|:---:|:---:|
| <img src="assets/screenshots/dashboard_en.png?v=1.0.7" width="420" alt="Dashboard"/> | <img src="assets/screenshots/nodes_en.png?v=1.0.7" width="420" alt="Nodes"/> |

| **QR-Code Multi-Import** | **Client Settings & Preferences** |
|:---:|:---:|
| <img src="assets/screenshots/modal_qr_en.png?v=1.0.7" width="420" alt="QR Modal"/> | <img src="assets/screenshots/settings_en.png?v=1.0.7" width="420" alt="Settings"/> |

</div>

---

## ✨ Key Features

- **Split Tunneling** — whitelist / blacklist apps, global and direct modes.
- **WARP Detour (Clean IP)** — chain traffic through Cloudflare WARP on top of any proxy. ChatGPT, Claude, Netflix without captchas or datacenter blocks.
- **WARP Generator** — AmneziaWG (anti-DPI) and MASQUE/HTTP3 in one click; `.conf` file import.
- **Fastest Node** — auto-select the lowest-latency server.
- **Kill Switch** — block all traffic on unexpected tunnel drop.
- **Bypass RU** — route Russian services (banks, government, VK) directly, bypassing VPN.
- **AdBlock** — DNS-level ad and tracker blocking.
- **QR Import** — from screen, clipboard, file, or webcam; single nodes and subscriptions.
- **Core Logs** — real-time sing-box log viewer with Windows noise filtering.
- **Protocols** — VLESS (Reality, XHTTP, ML-KEM-768), VMess, Shadowsocks 2022, Trojan, Hysteria 2, TUIC v5, WireGuard, AmneziaWG.
- **Telemetry** — 60 fps speed graph, per-process connection tree, subscription widget with quotas.
- **Profile Backup** — export/import all settings to `.json`.
- **Privacy Blur** — hide IPs, domains, and keys for streaming and screenshots.
- **Material 3 Dark UI** — clean design, system tray with quick actions.

---

## 🚀 Download & Installation

1. Go to the [Releases](https://github.com/zalupacode228/vibe-vpn/releases/latest) page.
2. Download the latest `VibeVPN-v1.0.8-Windows-x64.exe` (or `VPN.exe`).
3. Run the executable.
   > **Note**: Requires Windows 10 or 11 (64-bit). Administrator privileges are required to create the TUN network adapter (`wintun.dll`).

---

## 🛠️ Building from Source

### Prerequisites
- **Go**: `1.22+` or `1.23+` ([golang.org](https://go.dev/dl/))
- **Node.js**: `18+` or `20+` ([nodejs.org](https://nodejs.org/))
- **Wails CLI v2**:
  ```powershell
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- **WebView2 Runtime** (installed by default on Windows 10/11)

### Development Mode
To run with live hot-reloading for frontend and Go backend:
```powershell
git clone https://github.com/zalupacode228/vibe-vpn.git
cd vibe-vpn

wails dev
```

### Production Build
To create a standalone, monolithic single-file binary:
```powershell
wails build
```
The compiled executable will be located at:
```
build/bin/VPN.exe
```

---

## 🏗️ Architecture

```
vibe-go/
├── app.go                 # Wails Go controller & Windows API bindings
├── main.go                # Application entrypoint, single-instance mutex & window config
├── assets/
│   ├── core/              # Embedded core (sing-box.exe with AWG + wintun.dll)
│   └── screenshots/       # Documentation screenshots
├── build/
│   ├── appicon.png        # Branded high-res vector shield icon (1024x1024)
│   └── windows/
│       └── icon.ico       # Multi-size Windows icon (16–256px) for taskbar & .exe
├── pkg/
│   ├── models/            # Core data structures (Nodes, Settings, Geo-rules, Subscriptions)
│   ├── singbox/           # Sing-box JSON generator, routing rules & process controller
│   ├── procmon/           # Windows active processes & exe icon extractor
│   ├── storage/           # State persistence in %APPDATA%\VPN
│   ├── subs/              # Subscription parser & Base64/Clash decoder
│   ├── pinger/            # Concurrent TCP/ICMP ping engine
│   ├── warp/              # Cloudflare WARP registration (AWG & MASQUE) & config parser
│   └── tray/              # Windows system tray integration
└── frontend/              # Lightweight frontend (Vanilla JS + Vite)
    ├── src/
    │   ├── main.js        # UI logic, log parsing & event listeners
    │   ├── i18n.js        # Localization dictionary & auto-detection
    │   └── style.css      # Dark Glassmorphism styles
    └── index.html         # Application layout
```

---

## 📄 License

Distributed under the **MIT License**. See [`LICENSE`](LICENSE) for details.

<div align="center">
Built with ❤️ using Go, Wails v2, and sing-box.
</div>

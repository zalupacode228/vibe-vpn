# VibeVPN — Fast & Simple Windows VPN Client

<div align="center">

<p align="center">
  <a href="README.md">🇷🇺 Русский</a> | <b>🇬🇧 English</b>
</p>

![Platform](https://img.shields.io/badge/Platform-Windows%2010%20%2F%2011%20x64-0078D6?logo=windows&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v2-DF0000?logo=wails&logoColor=white)
![Sing-Box](https://img.shields.io/badge/Core-sing--box%201.14-FF6B00?logo=box&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![i18n](https://img.shields.io/badge/Language-RU%20%7C%20EN-blueviolet.svg)

**A fast, modern, and lightweight Windows VPN client featuring per-application Split Tunneling, Cloudflare WARP (AmneziaWG / MASQUE) generator and config import, smart domain & IP/CIDR geo-routing, real-time connection log viewer, and network noise filtering.**

[Download Latest Release](https://github.com/zalupacode228/vibe-vpn/releases/latest) • [Key Features](#-key-features) • [Building from Source](#-building-from-source)

</div>

---

## 📸 Screenshots

<div align="center">

| **Dashboard (Dark Glassmorphism)** | **Servers & Nodes (Categorized)** |
|:---:|:---:|
| <img src="assets/screenshots/dashboard_en.png" width="420" alt="Dashboard"/> | <img src="assets/screenshots/nodes_en.png" width="420" alt="Nodes"/> |

| **QR-Code Multi-Import** | **Settings & Localization (RU / EN)** |
|:---:|:---:|
| <img src="assets/screenshots/modal_qr_en.png" width="420" alt="QR Modal"/> | <img src="assets/screenshots/settings_en.png" width="420" alt="Settings"/> |

</div>

---

## ✨ Key Features

### ⚡ Per-Application Split Tunneling
- Total control over which `.exe` applications route through VPN vs direct connection:
  - **Whitelist**: Route *only* selected apps through the VPN (browsers, messengers, games).
  - **Blacklist**: Route everything through VPN *except* selected apps.
  - **Global**: Route all computer traffic through the VPN tunnel.
  - **Direct (Bypass)**: All traffic goes straight without encryption.
- Automatically lists running Windows processes with their real-time system icons.
- Ability to pick and add any `.exe` application via the native Windows file dialog.

### 🛡️ Cloudflare WARP: Generator & Config Import (.conf)
- **Built-in Generator**:
  - **AmneziaWG (AWG)**: Anti-censorship obfuscation using randomized junk packets (`Jc`, `Jmin`, `Jmax`, `S1`–`S4`, `H1`–`H4`) to defeat DPI filters.
  - **MASQUE / HTTP/3**: Web traffic disguise over QUIC UDP port 443.
  - One-click Cloudflare client key registration and peer negotiation.
- **Config Import**:
  - Load existing `.conf`, `.wireguard`, `.txt` config files via native file picker.
  - Direct paste of raw config text.
  - Full support for AmneziaWG obfuscation parameters and `Reserved` / `ClientID` formats (comma-separated, array, or Base64).

### 📋 Connection Logs Viewer (Real-Time Core Logs)
- Integrated sing-box live log viewer directly accessible from the dashboard.
- **Smart Network Noise Filtering**: Automatically filters out malformed Windows DNS probes (NetBIOS, mDNS, NCSI network status checks) and high-frequency per-packet gaming traffic (CS2, etc.) — displaying only real connection events and actual errors.
- Filter pill toggles: *All*, *Errors*, *Warnings*.
- Auto-scroll, one-click copy to clipboard, and instant log clearing.

### 🌐 Smart Routing & Geo-Rules (Domains & IP/CIDR Subnets)
- **Bypass RU**: Automatically routes Russian domestic services (banks, government portals, social networks, domestic streaming) directly without VPN for maximum throughput.
- **IP & CIDR Subnet Support**: Add custom lists of IP addresses and CIDR subnets (e.g. Russian IP blocks like `178.249.0.0/16` or blocked ranges) with configurable action: **Proxy (In Tunnel)** or **Direct (Bypass)**.
- **Flexible List Sources**:
  - Online via direct URL / GitHub Raw `.txt` or `.list`.
  - Offline via local PC files using the **Browse...** button.
- **AdBlock**: DNS-level ad, tracker, and telemetry blocking.

### 📷 Universal QR-Code Importer
- **From Screen**: Instant screen capture with automatic QR code scanning.
- **From Clipboard (`Ctrl+V`)**: Instant detection of copied QR screenshot images or raw config text.
- **From File**: Drag-and-drop or file picker for PNG, JPG, WebP.
- **Live Webcam Feed**: Scan QR codes directly with your camera.
- Automatically recognizes single nodes, bulk node lists, and subscription links.

### 🚀 Modern Protocol Support
- VLESS (Reality, TLS, WebSocket, gRPC, **XHTTP / SplitHTTP, Post-Quantum ML-KEM-768 encryption**)
- VMess
- Shadowsocks (2022)
- Trojan
- Hysteria 2 / Hy2
- TUIC v5
- WireGuard & AmneziaWG (AWG)

### 📊 Telemetry, Subscriptions & Monitoring
- **Hybrid Subscription Widget**: Summary status strip on the main dashboard and detailed card on the Servers tab (supports quota limits, unlimited tiers, and expiration dates).
- **Active Node Auto-Pinning**: The currently connected server is always automatically pinned to the top of the node list.
- Real-time 60 FPS bandwidth speed graph (Download / Upload).
- High-precision ping latency test with visual color indicators (green / yellow / red).
- Interactive **Connection Tree** grouped by active Windows process names, displaying exact destination IPs, routing rules, and bandwidth per app.

### ⚙️ Customization & Privacy
- **Custom User-Agent & HWID**: Tailor client identification and hardware ID when updating subscriptions to ensure maximum server compatibility.
- **Privacy Blur Mode**: One-click blur of IP addresses, server domains, and private keys for safe livestreaming and screenshots.

### 🌍 Bilingual Localization & System Auto-Detection
- Automatically detects Windows system language at startup (Russian for RU/CIS systems, English for rest of world).
- Instant on-the-fly language switching in Settings without restarting the application.

### 🎨 Design & Ergonomics
- Polished Material 3 Dark theme (strict 16px grid rhythm, balanced typography contrast, clean visual accents).
- Minimization to system tray with quick-actions menu.

---

## 🚀 Download & Installation

1. Go to the [Releases](https://github.com/zalupacode228/vibe-vpn/releases/latest) page.
2. Download the latest `VibeVPN-v1.0.6-Windows-x64.exe` (or `VPN.exe`).
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

# VibeVPN — Fast & Simple Windows VPN Client

<div align="center">

<p align="center">
  <a href="README.md">🇷🇺 Русский</a> | <b>🇬🇧 English</b>
</p>

![Platform](https://img.shields.io/badge/Platform-Windows%2010%20%2F%2011%20x64-0078D6?logo=windows&logoColor=white)
![Version](https://img.shields.io/badge/Version-v1.0.7-brightgreen.svg)
![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v2-DF0000?logo=wails&logoColor=white)
![Sing-Box](https://img.shields.io/badge/Core-sing--box%201.14-FF6B00?logo=box&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![i18n](https://img.shields.io/badge/Language-RU%20%7C%20EN-blueviolet.svg)

**A fast, modern, and lightweight Windows VPN client featuring per-application Split Tunneling, Cloudflare WARP (AmneziaWG / MASQUE) generator and config import, WARP Detour chained proxying, smart domain & IP/CIDR geo-routing, real-time connection log viewer, and network noise filtering.**

[Download Release v1.0.7 (.exe)](https://github.com/zalupacode228/vibe-vpn/releases/tag/v1.0.7) • [What's New](#-whats-new-in-v107) • [Key Features](#-key-features) • [Building from Source](#-building-from-source)

</div>

---

## 🆕 What's New in v1.0.7

- 🛡️ **WARP Detour (Clean IP / Chained Proxy)**:
  - Route internet egress through Cloudflare WARP on top of any active upstream proxy (VLESS, Hysteria 2, Shadowsocks, Trojan). Hides the VPS IP address from destination sites, effortlessly bypassing datacenter blocks, Cloudflare turnstiles, and aggressive captchas.
  - One-click instant toggle button `WARP Clean IP` right inside the main header.
- ⚡ **Fastest Node Selection**:
  - Automatically benchmarks and switches to the lowest-latency server in one click from the active server dropdown or the Servers panel.
- 🌐 **Real Web Latency Test (Web Latency)**:
  - Measures true end-to-end HTTP RTT (code 204) through the active tunnel to verified endpoints instead of superficial TCP handshakes.
- ⛔ **Kill Switch (Leak Protection)**:
  - Immediately blocks unencrypted internet traffic if the VPN tunnel drops unexpectedly until the connection is safely restored.
- 🔄 **Background Subscription Auto-Update**:
  - Automatically synchronizes subscription node lists every 24 hours in the background.
- 💾 **Profile Backup & Restore**:
  - One-click export and import of all user settings, subscriptions, custom nodes, and routing rules to/from a portable JSON file.
- 📡 **Protocol-Aware Latency**:
  - Native ICMP Echo measurement for UDP protocols (Hysteria 2 / TUIC) and optimized TCP RTT for VLESS Reality / XHTTP without timeouts.

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

### ⚡ Per-Application Split Tunneling
- Total control over which `.exe` applications route through VPN vs direct connection:
  - **Whitelist**: Route *only* selected apps through the VPN (browsers, messengers, games).
  - **Blacklist**: Route everything through VPN *except* selected apps.
  - **Global**: Route all computer traffic through the VPN tunnel.
  - **Direct (Bypass)**: All traffic goes straight without encryption.
- Automatically lists running Windows processes with their real-time system icons.
- Ability to pick and add any `.exe` application via the native Windows file dialog.

### 🛡️ Cloudflare WARP & WARP Detour (Clean IP)
- **WARP Detour (Clean IP)**: Route outbound proxy traffic through Cloudflare's residential network. International services (ChatGPT, Claude, Netflix) see a clean IP without datacenter blocks or captchas.
- **Built-in Generator**:
  - **AmneziaWG (AWG)**: Anti-censorship obfuscation using randomized junk packets (`Jc`, `Jmin`, `Jmax`, `S1`–`S4`, `H1`–`H4`) to defeat DPI filters.
  - **MASQUE / HTTP/3**: Web traffic disguise over QUIC UDP port 443.
  - One-click Cloudflare client key registration and peer negotiation.
- **Config Import**:
  - Load existing `.conf`, `.wireguard`, `.txt` config files via native file picker or clipboard with full AmneziaWG parameter parsing.

### ⚡ Fastest Server Selection & Real Web Latency
- **⚡ Best Node**: Automatically selects and switches to the lowest-latency server in one click from the dropdown or nodes panel.
- **🌐 Real Web Latency Test (HTTP 204)**: Measures end-to-end data delivery across the active tunnel instead of superficial TCP handshakes.
- **Protocol-Aware Ping**: Measures true TCP RTT for Reality/TLS and system ICMP Echo for UDP protocols (Hysteria 2 / WireGuard).

### ⛔ Leak Protection & Backup Sync
- **Kill Switch**: Blocks all internet traffic outside the tunnel in case of unexpected VPN disconnection.
- **Background Subscription Auto-Update**: Silently updates subscription nodes every 24 hours and at startup.
- **One-Click Profile Backup**: Export and import settings, subscriptions, nodes, and custom rules as a single `.json` file.

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

### 🎨 Design & Ergonomics
- Polished Material 3 Dark theme (strict 16px grid rhythm, balanced typography contrast, clean visual accents).
- Minimization to system tray with quick-actions menu.

---

## 🚀 Download & Installation

1. Go to the [Releases](https://github.com/zalupacode228/vibe-vpn/releases/latest) page.
2. Download the latest `VibeVPN-v1.0.7-Windows-x64.exe` (or `VPN.exe`).
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

package singbox

import (
	"encoding/json"
	"net"
	"strings"
	"vibe-go/pkg/models"
)

var bypassRUDomains = []string{
	".ru", ".su", ".xn--p1ai", ".рф",
	"yandex.ru", "ya.ru", "dzen.ru", "mail.ru", "rambler.ru", "rutube.ru", "vk.com", "vk.ru", "ok.ru",
	"gosuslugi.ru", "mos.ru", "nalog.gov.ru", "cbr.ru", "kremlin.ru", "pravo.gov.ru", "customs.gov.ru",
	"sberbank.ru", "sber.ru", "tbank.ru", "tinkoff.ru", "vtb.ru", "alfa-bank.ru", "alfabank.ru",
	"gazprombank.ru", "raiffeisen.ru", "open.ru", "sovcombank.ru", "psbank.ru", "rshb.ru", "mkb.ru",
	"ozon.ru", "wildberries.ru", "wb.ru", "avito.ru", "market.yandex.ru", "megamarket.ru",
	"magnit.ru", "x5.ru", "perekrestok.ru", "sbermarket.ru", "kuper.ru", "samokat.ru", "cdek.ru", "pochta.ru",
	"kinopoisk.ru", "ivi.ru", "okko.tv", "kion.ru", "start.ru", "premier.one",
	"2gis.ru", "hh.ru", "cian.ru", "domclick.ru", "rbc.ru", "kommersant.ru", "vedomosti.ru", "tass.ru",
	"mts.ru", "megafon.ru", "beeline.ru", "t2.ru", "tele2.ru", "rostelecom.ru",
}

var adblockDomains = []string{
	"doubleclick.net", "adservice.google.com", "googleads.g.doubleclick.net",
	"pagead2.googlesyndication.com", "an.yandex.ru", "mc.yandex.ru", "ads.vk.com",
	"top-fwz1.mail.ru", "adguard.com", "telemetry.microsoft.com",
	"vortex.data.microsoft.com", "app-measurement.com", "analytics.google.com",
}

var appHelpers = map[string][]string{
	"firefox.exe":  {"crashhelper.exe", "pingsender.exe", "default-browser-agent.exe"},
	"chrome.exe":   {"elevation_service.exe", "notification_helper.exe"},
	"msedge.exe":   {"msedgewebview2.exe", "identity_helper.exe"},
	"telegram.exe": {"updater.exe"},
	"discord.exe":  {"update.exe"},
	"steam.exe":    {"steamservice.exe", "steamwebhelper.exe"},
	"spotify.exe":  {"spotifywebhelper.exe"},
}

type ConfigBuilder struct {
	warpNode *models.Node
}

func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

func (b *ConfigBuilder) SetWarpNode(wn *models.Node) {
	b.warpNode = wn
}

func (b *ConfigBuilder) GenerateConfig(
	mode string,
	rules []models.AppRule,
	node *models.Node,
	settings models.Settings,
) ([]byte, error) {
	serverHost := ""
	serverSNI := ""
	if node != nil {
		serverHost = strings.TrimSpace(node.Server)
		serverSNI = strings.TrimSpace(node.SNI)
	}

	proto := ""
	if node != nil {
		proto = strings.ToLower(strings.TrimSpace(node.Protocol))
	}
	isMasque := proto == "masque"
	isWireguard := !isMasque && (proto == "wireguard" || proto == "warp" || (node != nil && node.PrivateKey != ""))

	var wgEndpoints []map[string]interface{}
	var proxyOut map[string]interface{}

	targetOutbound := "proxy-out"
	warpDetourActive := settings.WarpDetour && node != nil && !isWireguard && !isMasque
	if warpDetourActive {
		targetOutbound = "warp-detour"
		warpPrivKey := "aGVsbG93b3JsZGhlbGxvd29ybGRoZWxsb3dvcmxkMQ=="
		var warpReserved []int
		if b.warpNode != nil && b.warpNode.PrivateKey != "" {
			warpPrivKey = b.warpNode.PrivateKey
			warpReserved = b.warpNode.Reserved
		}

		peerObj := map[string]interface{}{
			"address":     "162.159.192.1",
			"port":        500,
			"public_key":  "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
			"allowed_ips": []string{"0.0.0.0/0", "::/0"},
		}
		if len(warpReserved) >= 3 {
			peerObj["reserved"] = warpReserved
		}

		warpEndpoint := map[string]interface{}{
			"type":        "wireguard",
			"tag":         "warp-detour",
			"detour":      "proxy-out",
			"address":     []string{"172.16.0.2/32"},
			"private_key": warpPrivKey,
			"peers":       []interface{}{peerObj},
			"mtu":         1280,
		}
		wgEndpoints = append(wgEndpoints, warpEndpoint)
	}

	if isWireguard && node != nil {
		serverPort := node.Port
		if serverPort <= 0 {
			serverPort = 500
		}
		if serverHost == "" {
			serverHost = "162.159.192.1"
		}
		peerPub := node.PeerPublicKey
		if peerPub == "" {
			peerPub = "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo="
		}
		localAddrs := []string{"172.16.0.2/32"}
		if node.LocalAddress != "" {
			parts := strings.Split(node.LocalAddress, ",")
			localAddrs = []string{}
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					localAddrs = append(localAddrs, p)
				}
			}
		}

		peerObj := map[string]interface{}{
			"address":                       serverHost,
			"port":                          serverPort,
			"public_key":                    peerPub,
			"allowed_ips":                   []string{"0.0.0.0/0", "::/0"},
			"persistent_keepalive_interval": 25,
		}
		if len(node.Reserved) >= 3 {
			peerObj["reserved"] = node.Reserved
		}

		endpointObj := map[string]interface{}{
			"type":        "wireguard",
			"tag":         "proxy-out",
			"address":     localAddrs,
			"private_key": node.PrivateKey,
			"peers":       []interface{}{peerObj},
			"mtu":         1280,
		}

		// AmneziaWG obfuscation parameters supported by sing-box-lx
		awgMap := node.AWG
		awgFields := []string{"jc", "jmin", "jmax", "s1", "s2", "s3", "s4", "h1", "h2", "h3", "h4", "i1", "i2", "i3", "i4", "i5"}
		for _, f := range awgFields {
			var val interface{}
			if awgMap != nil {
				val = awgMap[f]
			}
			if val == nil {
				switch f {
				case "jc":
					if node.JC != 0 {
						val = node.JC
					}
				case "jmin":
					if node.Jmin != 0 {
						val = node.Jmin
					}
				case "jmax":
					if node.Jmax != 0 {
						val = node.Jmax
					}
				case "s1":
					if node.S1 != 0 {
						val = node.S1
					}
				case "s2":
					if node.S2 != 0 {
						val = node.S2
					}
				case "s3":
					if node.S3 != 0 {
						val = node.S3
					}
				case "s4":
					if node.S4 != 0 {
						val = node.S4
					}
				case "h1":
					if node.H1 != 0 {
						val = node.H1
					}
				case "h2":
					if node.H2 != 0 {
						val = node.H2
					}
				case "h3":
					if node.H3 != 0 {
						val = node.H3
					}
				case "h4":
					if node.H4 != 0 {
						val = node.H4
					}
				case "i1":
					if node.I1 != "" {
						val = node.I1
					}
				}
			}
			if val != nil {
				endpointObj[f] = val
			}
		}

		wgEndpoints = []map[string]interface{}{endpointObj}

	} else if isMasque && node != nil {
		serverPort := node.Port
		if serverPort <= 0 {
			serverPort = 443
		}
		if serverHost == "" {
			serverHost = "162.159.198.2"
		}
		if serverSNI == "" {
			serverSNI = "www.microsoft.com"
		}
		pubKey := node.PublicKey
		if pubKey == "" {
			pubKey = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEIaU7MToJm9NKp8YfGxR6r+/h4mcG7SxI8tsW8OR1A5tv/zCzVbCRRh2t87/kxnP6lAy0lkr7qYwu+ox+k3dr6w=="
		}
		v4Addr := node.LocalAddress
		if v4Addr == "" {
			v4Addr = "172.16.0.2/32"
		} else if !strings.Contains(v4Addr, "/") {
			v4Addr += "/32"
		}

		proxyOut = map[string]interface{}{
			"type":        "masque",
			"tag":         "proxy-out",
			"server":      serverHost,
			"server_port": serverPort,
			"profile":     "cloudflare",
			"vhttp":       "auto",
			"tls": map[string]interface{}{
				"server_name": serverSNI,
			},
			"private_key": node.PrivateKey,
			"public_key":  pubKey,
			"ip":          v4Addr,
			"mtu":         1280,
		}

	} else if node != nil && node.RawOutbound != nil && len(node.RawOutbound) > 0 {
		proxyOut = make(map[string]interface{})
		for k, v := range node.RawOutbound {
			proxyOut[k] = v
		}
		proxyOut["tag"] = "proxy-out"
		delete(proxyOut, "name") // sing-box forbids unexpected 'name' in outbound schema

		if s, ok := proxyOut["server"].(string); ok && s != "" {
			serverHost = s
		}
		if tlsMap, ok := proxyOut["tls"].(map[string]interface{}); ok {
			if sName, ok := tlsMap["server_name"].(string); ok && sName != "" {
				serverSNI = sName
			}
		}
		if t, ok := proxyOut["type"].(string); ok && t == "vless" {
			proxyOut["packet_encoding"] = "xudp"
		}

	} else if node != nil {
		nodeType := proto
		if nodeType == "" {
			nodeType = "vless"
		}
		serverPort := node.Port
		if serverPort <= 0 {
			serverPort = 443
		}

		proxyOut = map[string]interface{}{
			"tag":         "proxy-out",
			"type":        nodeType,
			"server":      serverHost,
			"server_port": serverPort,
		}

		if nodeType == "vless" {
			proxyOut["packet_encoding"] = "xudp"
			if node.UUID != "" {
				proxyOut["uuid"] = node.UUID
			}
			if node.Flow != "" {
				proxyOut["flow"] = node.Flow
			}
			if node.Encryption != "" {
				proxyOut["encryption"] = node.Encryption
			}
			tlsMap := map[string]interface{}{
				"enabled": true,
			}
			if serverSNI != "" {
				tlsMap["server_name"] = serverSNI
			}
			if node.Fingerprint != "" {
				tlsMap["utls"] = map[string]interface{}{
					"enabled":     true,
					"fingerprint": node.Fingerprint,
				}
			}
			if node.Security == "reality" {
				tlsMap["reality"] = map[string]interface{}{
					"enabled":    true,
					"public_key": node.PublicKey,
					"short_id":   node.ShortID,
				}
			}
			proxyOut["tls"] = tlsMap

			if node.Network != "" && node.Network != "tcp" {
				transMap := map[string]interface{}{
					"type": node.Network,
					"path": node.Path,
					"host": node.Host,
				}
				if node.Mode != "" {
					transMap["mode"] = node.Mode
				}
				proxyOut["transport"] = transMap
			}
		} else if nodeType == "hysteria2" || nodeType == "hy2" {
			proxyOut["type"] = "hysteria2"
			if node.Password != "" {
				proxyOut["password"] = node.Password
			}
			tlsMap := map[string]interface{}{
				"enabled": true,
			}
			if serverSNI != "" {
				tlsMap["server_name"] = serverSNI
			}
			proxyOut["tls"] = tlsMap
		}
	}

	// Collect server host and SNI to prevent routing loop & ensure local resolution
	var bypassDomains []string
	var bypassIPs []string

	candidates := []string{serverHost, serverSNI}
	for _, cand := range candidates {
		cand = strings.TrimSpace(cand)
		if cand == "" {
			continue
		}
		if ip := net.ParseIP(cand); ip != nil {
			if !contains(bypassIPs, cand) {
				bypassIPs = append(bypassIPs, cand)
			}
		} else {
			if !contains(bypassDomains, cand) {
				bypassDomains = append(bypassDomains, cand)
			}
			// Attempt to resolve IP to exclude from tunnel
			if ips, err := net.LookupIP(cand); err == nil {
				for _, ip := range ips {
					if ip4 := ip.To4(); ip4 != nil {
						ipStr := ip4.String()
						if !contains(bypassIPs, ipStr) {
							bypassIPs = append(bypassIPs, ipStr)
						}
					}
				}
			}
		}
	}

	// 1. Base route rules: Sniffing, DNS Hijacking, QUIC Block
	routeRules := []map[string]interface{}{
		{"action": "sniff"},
		{"action": "hijack-dns", "protocol": []string{"dns"}},
		{"action": "hijack-dns", "port": []int{53}},
		{"action": "reject", "port": []int{443}, "network": []string{"udp"}},
	}

	// 2. Exclude VPN server IP and domain from tunnel to prevent routing loop
	if len(bypassDomains) > 0 {
		routeRules = append(routeRules, map[string]interface{}{
			"action":   "route",
			"domain":   bypassDomains,
			"outbound": "direct",
		})
	}
	if len(bypassIPs) > 0 {
		var cidrs []string
		for _, ip := range bypassIPs {
			cidrs = append(cidrs, ip+"/32")
		}
		routeRules = append(routeRules, map[string]interface{}{
			"action":   "route",
			"ip_cidr":  cidrs,
			"outbound": "direct",
		})
	}

	// 3. Bypass LAN
	if settings.BypassLAN {
		routeRules = append(routeRules, map[string]interface{}{
			"action":        "route",
			"ip_is_private": true,
			"outbound":      "direct",
		})
	}

	// 4. Smart Routing: AdBlock
	if settings.AdBlock {
		routeRules = append(routeRules, map[string]interface{}{
			"action":        "reject",
			"domain_suffix": adblockDomains,
		})
	}

	// 5. Smart Routing: Bypass RU
	if settings.BypassRU {
		routeRules = append(routeRules, map[string]interface{}{
			"action":        "route",
			"domain_suffix": bypassRUDomains,
			"outbound":      "direct",
		})
	}

	// 6. Custom Lists
	for _, cl := range settings.CustomLists {
		if cl.Enabled {
			targetOut := targetOutbound
			if cl.Action == "direct" {
				targetOut = "direct"
			}
			if len(cl.Domains) > 0 {
				routeRules = append(routeRules, map[string]interface{}{
					"action":        "route",
					"domain_suffix": cl.Domains,
					"outbound":      targetOut,
				})
			}
			if len(cl.IPs) > 0 {
				routeRules = append(routeRules, map[string]interface{}{
					"action":   "route",
					"ip_cidr":  cl.IPs,
					"outbound": targetOut,
				})
			}
		}
	}

	// 7. Process-based split routing
	var rawNames []string
	var processPaths []string
	for _, r := range rules {
		if r.Enabled || r.IsEnabled {
			if r.Name != "" {
				rawNames = append(rawNames, r.Name)
			}
			if r.Path != "" && r.IsCustom {
				processPaths = append(processPaths, r.Path)
			}
		}
	}

	// Expand helper processes
	allProcessNames := make(map[string]struct{})
	for _, pn := range rawNames {
		allProcessNames[pn] = struct{}{}
		if helpers, ok := appHelpers[strings.ToLower(pn)]; ok {
			for _, h := range helpers {
				allProcessNames[h] = struct{}{}
			}
		}
	}
	var processNames []string
	for name := range allProcessNames {
		processNames = append(processNames, name)
	}

	switch mode {
	case "whitelist":
		if len(processNames) > 0 || len(processPaths) > 0 {
			entry := map[string]interface{}{
				"action":   "route",
				"outbound": targetOutbound,
			}
			if len(processNames) > 0 {
				entry["process_name"] = processNames
			}
			if len(processPaths) > 0 {
				entry["process_path"] = processPaths
			}
			routeRules = append(routeRules, entry)
		}
		routeRules = append(routeRules, map[string]interface{}{
			"action":   "route",
			"outbound": "direct",
		})

	case "blacklist":
		if len(processNames) > 0 || len(processPaths) > 0 {
			entry := map[string]interface{}{
				"action":   "route",
				"outbound": "direct",
			}
			if len(processNames) > 0 {
				entry["process_name"] = processNames
			}
			if len(processPaths) > 0 {
				entry["process_path"] = processPaths
			}
			routeRules = append(routeRules, entry)
		}
		routeRules = append(routeRules, map[string]interface{}{
			"action":   "route",
			"outbound": targetOutbound,
		})

	case "global":
		routeRules = append(routeRules, map[string]interface{}{
			"action":   "route",
			"outbound": targetOutbound,
		})

	default: // "direct"
		routeRules = append(routeRules, map[string]interface{}{
			"action":   "route",
			"outbound": "direct",
		})
	}

	// DNS Rules
	var dnsRules []map[string]interface{}
	if len(bypassDomains) > 0 {
		dnsRules = append(dnsRules, map[string]interface{}{
			"domain": bypassDomains,
			"server": "direct-dns",
		})
	}
	if settings.BypassRU {
		dnsRules = append(dnsRules, map[string]interface{}{
			"domain_suffix": bypassRUDomains,
			"server":        "direct-dns",
		})
	}
	for _, cl := range settings.CustomLists {
		if cl.Enabled && cl.Action == "direct" && len(cl.Domains) > 0 {
			dnsRules = append(dnsRules, map[string]interface{}{
				"domain_suffix": cl.Domains,
				"server":        "direct-dns",
			})
		}
	}
	dnsRules = append(dnsRules, map[string]interface{}{
		"domain_suffix": []string{".local", ".lan", ".home", ".internal", ".arpa"},
		"server":        "local-dns",
	})

	finalDNS := "remote-dns"
	if mode == "direct" {
		finalDNS = "direct-dns"
	}

	remoteDNSServer := map[string]interface{}{
		"tag":    "remote-dns",
		"type":   "https",
		"server": "8.8.8.8",
		"path":   "/dns-query",
	}
	if isWireguard || isMasque {
		remoteDNSServer = map[string]interface{}{
			"tag":    "remote-dns",
			"type":   "udp",
			"server": "1.1.1.1",
			"detour": "proxy-out",
		}
	} else if proxyOut != nil {
		remoteDNSServer["detour"] = targetOutbound
	}

	dnsServers := []map[string]interface{}{
		remoteDNSServer,
		{
			"tag":    "direct-dns",
			"type":   "udp",
			"server": "77.88.8.8",
		},
		{
			"tag":  "local-dns",
			"type": "local",
		},
	}

	tunMTU := settings.MTU
	if tunMTU <= 0 {
		tunMTU = settings.TunMTU
	}
	if tunMTU <= 0 || tunMTU > 1500 {
		tunMTU = 1400
	}

	outbounds := []map[string]interface{}{
		{"tag": "direct", "type": "direct"},
		{"tag": "block", "type": "block"},
	}
	if proxyOut != nil {
		outbounds = append([]map[string]interface{}{proxyOut}, outbounds...)
	}

	finalRoute := "direct"
	if settings.KillSwitch && mode == "global" {
		finalRoute = "block"
	}

	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"timestamp": true,
		},
		"dns": map[string]interface{}{
			"servers":  dnsServers,
			"rules":    dnsRules,
			"final":    finalDNS,
			"strategy": "prefer_ipv4",
		},
		"inbounds": []map[string]interface{}{
			{
				"type":                     "tun",
				"tag":                      "tun-in",
				"interface_name":           "vpn-tun",
				"address":                  []string{"172.19.0.1/30"},
				"mtu":                      tunMTU,
				"auto_route":               true,
				"strict_route":             settings.StrictRoute || settings.KillSwitch,
				"endpoint_independent_nat": true,
				"stack":                    "system",
			},
		},
		"outbounds": outbounds,
		"route": map[string]interface{}{
			"default_domain_resolver": "direct-dns",
			"rules":                   routeRules,
			"auto_detect_interface":   true,
			"final":                   finalRoute,
		},
		"experimental": map[string]interface{}{
			"clash_api": map[string]interface{}{
				"external_controller": "127.0.0.1:9090",
				"external_ui":         "",
				"secret":              "",
				"default_mode":        "rule",
			},
		},
	}

	if len(wgEndpoints) > 0 {
		config["endpoints"] = wgEndpoints
	}

	return json.MarshalIndent(config, "", "  ")
}

func contains(arr []string, val string) bool {
	for _, item := range arr {
		if strings.EqualFold(item, val) {
			return true
		}
	}
	return false
}

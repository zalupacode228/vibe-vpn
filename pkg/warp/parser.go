package warp

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"vibe-go/pkg/models"
)

// ParseWireguardConf parses a standard WireGuard or AmneziaWG (AWG) configuration file (.conf).
func ParseWireguardConf(content string, fileName string) (*models.Node, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var section string

	var privateKey string
	var addresses []string
	var dnsServers []string
	var mtu int

	// AWG parameters
	var jc, jmin, jmax, s1, s2, s3, s4, h1, h2, h3, h4 int
	var i1 string
	hasAWG := false

	// Peer parameters
	var peerPublicKey string
	var endpoint string
	var allowedIPs []string
	var reserved []int
	var hasReserved bool

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		rawLine := scanner.Text()
		line := strings.TrimSpace(rawLine)

		// Check for comment-based Reserved/ClientID before stripping
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			commentContent := strings.TrimSpace(strings.TrimLeft(line, "#; "))
			lowerComment := strings.ToLower(commentContent)
			if strings.HasPrefix(lowerComment, "reserved") || strings.HasPrefix(lowerComment, "clientid") || strings.HasPrefix(lowerComment, "client_id") {
				parts := strings.SplitN(commentContent, "=", 2)
				if len(parts) == 2 && !hasReserved {
					if res := parseReservedField(parts[1]); len(res) >= 3 {
						reserved = res
						hasReserved = true
					}
				}
			}
			continue
		}

		if line == "" {
			continue
		}

		// Section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.Trim(line, "[] \t"))
			continue
		}

		// Key-Value parsing
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		// Strip trailing inline comments if present
		if commentIdx := strings.Index(val, " #"); commentIdx != -1 {
			val = strings.TrimSpace(val[:commentIdx])
		}

		switch section {
		case "interface":
			switch key {
			case "privatekey", "private_key":
				privateKey = val
			case "address":
				for _, addr := range strings.Split(val, ",") {
					addr = strings.TrimSpace(addr)
					if addr != "" {
						addresses = append(addresses, addr)
					}
				}
			case "dns":
				for _, d := range strings.Split(val, ",") {
					d = strings.TrimSpace(d)
					if d != "" {
						dnsServers = append(dnsServers, d)
					}
				}
			case "mtu":
				if n, err := strconv.Atoi(val); err == nil {
					mtu = n
				}
			// AmneziaWG obfuscation parameters
			case "jc":
				if n, err := strconv.Atoi(val); err == nil {
					jc = n
					hasAWG = true
				}
			case "jmin":
				if n, err := strconv.Atoi(val); err == nil {
					jmin = n
					hasAWG = true
				}
			case "jmax":
				if n, err := strconv.Atoi(val); err == nil {
					jmax = n
					hasAWG = true
				}
			case "s1":
				if n, err := strconv.Atoi(val); err == nil {
					s1 = n
					hasAWG = true
				}
			case "s2":
				if n, err := strconv.Atoi(val); err == nil {
					s2 = n
					hasAWG = true
				}
			case "s3":
				if n, err := strconv.Atoi(val); err == nil {
					s3 = n
					hasAWG = true
				}
			case "s4":
				if n, err := strconv.Atoi(val); err == nil {
					s4 = n
					hasAWG = true
				}
			case "h1":
				if n, err := strconv.Atoi(val); err == nil {
					h1 = n
					hasAWG = true
				}
			case "h2":
				if n, err := strconv.Atoi(val); err == nil {
					h2 = n
					hasAWG = true
				}
			case "h3":
				if n, err := strconv.Atoi(val); err == nil {
					h3 = n
					hasAWG = true
				}
			case "h4":
				if n, err := strconv.Atoi(val); err == nil {
					h4 = n
					hasAWG = true
				}
			case "i1":
				i1 = val
				hasAWG = true
			}

		case "peer":
			switch key {
			case "publickey", "public_key":
				if peerPublicKey == "" {
					peerPublicKey = val
				}
			case "endpoint":
				if endpoint == "" {
					endpoint = val
				}
			case "allowedips", "allowed_ips":
				for _, ip := range strings.Split(val, ",") {
					ip = strings.TrimSpace(ip)
					if ip != "" {
						allowedIPs = append(allowedIPs, ip)
					}
				}
			case "reserved", "clientid", "client_id":
				if !hasReserved {
					if res := parseReservedField(val); len(res) >= 3 {
						reserved = res
						hasReserved = true
					}
				}
			}
		}
	}

	if privateKey == "" {
		return nil, fmt.Errorf("в конфигурации не найден PrivateKey в секции [Interface]")
	}
	if peerPublicKey == "" {
		return nil, fmt.Errorf("в конфигурации не найден PublicKey в секции [Peer]")
	}

	// Parse Endpoint into host and port
	serverHost := "162.159.192.1"
	serverPort := 2408
	if hasAWG {
		serverPort = 500
	}

	if endpoint != "" {
		host, portStr, err := net.SplitHostPort(endpoint)
		if err == nil {
			serverHost = host
			if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
				serverPort = p
			}
		} else {
			// Might be missing port
			serverHost = strings.Trim(endpoint, "[]")
		}
	}

	localAddrStr := "172.16.0.2/32"
	if len(addresses) > 0 {
		localAddrStr = strings.Join(addresses, ", ")
	}

	isCloudflare := strings.Contains(strings.ToLower(serverHost), "cloudflare") ||
		strings.HasPrefix(serverHost, "162.159.") ||
		strings.HasPrefix(serverHost, "188.114.") ||
		peerPublicKey == "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=" ||
		hasReserved

	protoLabel := "WireGuard"
	if isCloudflare {
		if hasAWG {
			protoLabel = "Cloudflare WARP (AWG)"
		} else {
			protoLabel = "Cloudflare WARP"
		}
	} else if hasAWG {
		protoLabel = "AmneziaWG"
	}

	// Determine node name
	nodeName := strings.TrimSpace(fileName)
	if nodeName != "" {
		nodeName = strings.TrimSuffix(nodeName, ".conf")
		nodeName = strings.TrimSuffix(nodeName, ".wireguard")
		nodeName = strings.TrimSuffix(nodeName, ".txt")
	}
	if nodeName == "" {
		nodeName = protoLabel
	}

	var awgMap map[string]interface{}
	if hasAWG {
		awgMap = map[string]interface{}{}
		if jc != 0 {
			awgMap["jc"] = jc
		}
		if jmin != 0 {
			awgMap["jmin"] = jmin
		}
		if jmax != 0 {
			awgMap["jmax"] = jmax
		}
		if s1 != 0 {
			awgMap["s1"] = s1
		}
		if s2 != 0 {
			awgMap["s2"] = s2
		}
		if s3 != 0 {
			awgMap["s3"] = s3
		}
		if s4 != 0 {
			awgMap["s4"] = s4
		}
		if h1 != 0 {
			awgMap["h1"] = h1
		}
		if h2 != 0 {
			awgMap["h2"] = h2
		}
		if h3 != 0 {
			awgMap["h3"] = h3
		}
		if h4 != 0 {
			awgMap["h4"] = h4
		}
		if i1 != "" {
			awgMap["i1"] = i1
		}
	}

	node := &models.Node{
		ID:               fmt.Sprintf("node_wg_%d", time.Now().UnixMilli()),
		Name:             nodeName,
		Protocol:         "wireguard",
		ProtocolLabel:    protoLabel,
		Server:           serverHost,
		Port:             serverPort,
		PrivateKey:       privateKey,
		PeerPublicKey:    peerPublicKey,
		LocalAddress:     localAddrStr,
		SubscriptionName: "Custom WARP / WireGuard",
		Ping:             0,
		AWG:              awgMap,
		JC:               jc,
		Jmin:             jmin,
		Jmax:             jmax,
		S1:               s1,
		S2:               s2,
		S3:               s3,
		S4:               s4,
		H1:               h1,
		H2:               h2,
		H3:               h3,
		H4:               h4,
		I1:               i1,
	}

	if hasReserved {
		node.Reserved = reserved
	}
	if mtu > 0 {
		// can be used if needed
	}
	_ = dnsServers

	return node, nil
}

// parseReservedField parses various Reserved/ClientID formats:
// 1) "1, 2, 3" or "1,2,3"
// 2) "[1, 2, 3]"
// 3) Base64 encoded string e.g. "AQID"
func parseReservedField(val string) []int {
	val = strings.TrimSpace(val)
	val = strings.Trim(val, "[]\"'")

	// Check if comma-separated integers
	if strings.Contains(val, ",") {
		parts := strings.Split(val, ",")
		var res []int
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if n, err := strconv.Atoi(p); err == nil {
				res = append(res, n)
			}
		}
		if len(res) >= 3 {
			return res[:3]
		}
	}

	// Try base64 decoding
	if dec, err := base64.StdEncoding.DecodeString(val); err == nil && len(dec) >= 3 {
		return []int{int(dec[0]), int(dec[1]), int(dec[2])}
	}

	// Try 3 space-separated numbers
	fields := strings.Fields(val)
	if len(fields) >= 3 {
		var res []int
		for _, f := range fields {
			if n, err := strconv.Atoi(f); err == nil {
				res = append(res, n)
			}
		}
		if len(res) >= 3 {
			return res[:3]
		}
	}

	return nil
}

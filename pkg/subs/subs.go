package subs

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"vibe-go/pkg/models"
)

type FetchResult struct {
	Nodes       []models.Node
	Upload      int64
	Download    int64
	Total       int64
	Expire      int64
	HasUserInfo bool
}

func ParseUserInfo(headerVal string) (upload, download, total, expire int64) {
	if headerVal == "" {
		return
	}
	parts := strings.Split(headerVal, ";")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 {
			k := strings.ToLower(strings.TrimSpace(kv[0]))
			v, _ := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
			switch k {
			case "upload":
				upload = v
			case "download":
				download = v
			case "total":
				total = v
			case "expire":
				expire = v
			}
		}
	}
	return
}

func FetchAndParse(subURL string, subID string, subName string, userAgent string, hwid string) (*FetchResult, error) {
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		ua = "v2rayNG/1.8.5"
	}
	h := strings.TrimSpace(hwid)

	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			req.Header.Set("User-Agent", ua)
			if h != "" {
				req.Header.Set("HWID", h)
				req.Header.Set("X-HWID", h)
				req.Header.Set("X-Device-Id", h)
				req.Header.Set("Device-Id", h)
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", subURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ua)
	if h != "" {
		req.Header.Set("HWID", h)
		req.Header.Set("X-HWID", h)
		req.Header.Set("X-Device-Id", h)
		req.Header.Set("Device-Id", h)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	userinfoHdr := resp.Header.Get("Subscription-Userinfo")
	if userinfoHdr == "" {
		userinfoHdr = resp.Header.Get("subscription-userinfo")
	}
	hasUserInfo := strings.TrimSpace(userinfoHdr) != ""
	u, d, tot, exp := ParseUserInfo(userinfoHdr)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(bodyBytes)
	// Try base64 decode
	trimmed := strings.TrimSpace(content)
	if decoded, err := base64.StdEncoding.DecodeString(trimmed); err == nil {
		content = string(decoded)
	} else if decoded, err := base64.RawStdEncoding.DecodeString(trimmed); err == nil {
		content = string(decoded)
	} else if decoded, err := base64.URLEncoding.DecodeString(trimmed); err == nil {
		content = string(decoded)
	}

	lines := strings.Split(content, "\n")
	var nodes []models.Node

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		node := ParseNodeLink(line)
		if node != nil {
			node.SubscriptionID = subID
			node.SubscriptionName = subName
			nodes = append(nodes, *node)
		}
	}

	return &FetchResult{
		Nodes:       nodes,
		Upload:      u,
		Download:    d,
		Total:       tot,
		Expire:      exp,
		HasUserInfo: hasUserInfo,
	}, nil
}

func decodeBase64Any(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawStdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.URLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.RawURLEncoding.DecodeString(s)
}

func parseVMessLink(raw string) *models.Node {
	encoded := strings.TrimPrefix(raw, "vmess://")
	if idx := strings.Index(encoded, "#"); idx != -1 {
		encoded = encoded[:idx]
	}
	decoded, err := decodeBase64Any(encoded)
	if err != nil {
		return nil
	}

	var v struct {
		PS   string      `json:"ps"`
		Add  string      `json:"add"`
		Port interface{} `json:"port"`
		ID   string      `json:"id"`
		Scy  string      `json:"scy"`
		Net  string      `json:"net"`
		Type string      `json:"type"`
		Host string      `json:"host"`
		Path string      `json:"path"`
		TLS  string      `json:"tls"`
		SNI  string      `json:"sni"`
		ALPN string      `json:"alpn"`
		FP   string      `json:"fp"`
	}
	if err := json.Unmarshal(decoded, &v); err != nil || v.Add == "" || v.ID == "" {
		return nil
	}

	port := 443
	switch p := v.Port.(type) {
	case float64:
		port = int(p)
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && parsed > 0 {
			port = parsed
		}
	}
	if port <= 0 {
		port = 443
	}

	name := strings.TrimSpace(v.PS)
	if name == "" {
		name = fmt.Sprintf("%s:%d", v.Add, port)
	}

	hashInput := fmt.Sprintf("vmess:%s:%d:%s:%s:%s", v.Add, port, v.ID, v.Net, v.Path)
	h := sha256.Sum256([]byte(hashInput))
	id := fmt.Sprintf("vmess_%s_%d_%s", v.Add, port, hex.EncodeToString(h[:4]))

	sec := strings.TrimSpace(v.Scy)
	if sec == "" {
		sec = "auto"
	}

	node := &models.Node{
		ID:            id,
		Name:          name,
		Protocol:      "vmess",
		ProtocolLabel: "VMess",
		Server:        strings.TrimSpace(v.Add),
		Port:          port,
		UUID:          strings.TrimSpace(v.ID),
		Encryption:    sec,
		Network:       strings.TrimSpace(v.Net),
		Host:          strings.TrimSpace(v.Host),
		Path:          strings.TrimSpace(v.Path),
		Security:      strings.TrimSpace(v.TLS),
		SNI:           strings.TrimSpace(v.SNI),
		Fingerprint:   strings.TrimSpace(v.FP),
	}
	if v.ALPN != "" {
		for _, a := range strings.Split(v.ALPN, ",") {
			if a = strings.TrimSpace(a); a != "" {
				node.ALPN = append(node.ALPN, a)
			}
		}
	}
	return node
}

func ParseNodeLink(raw string) *models.Node {
	if strings.HasPrefix(strings.ToLower(raw), "vmess://") {
		return parseVMessLink(raw)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil
	}

	scheme := strings.ToLower(u.Scheme)
	name, _ := url.QueryUnescape(u.Fragment)
	if name == "" {
		name = u.Host
	}

	port, _ := strconv.Atoi(u.Port())
	if port == 0 {
		port = 443
	}

	rawWithoutFrag := raw
	if idx := strings.Index(rawWithoutFrag, "#"); idx != -1 {
		rawWithoutFrag = rawWithoutFrag[:idx]
	}
	h := sha256.Sum256([]byte(rawWithoutFrag))
	id := fmt.Sprintf("%s_%s_%d_%s", scheme, u.Hostname(), port, hex.EncodeToString(h[:4]))

	node := &models.Node{
		ID:       id,
		Name:     name,
		Protocol: scheme,
		Server:   u.Hostname(),
		Port:     port,
	}

	q := u.Query()

	switch scheme {
	case "vless":
		if u.User == nil {
			return nil
		}
		node.UUID = u.User.Username()
		node.Flow = q.Get("flow")
		node.Security = q.Get("security")
		node.SNI = q.Get("sni")
		node.Fingerprint = q.Get("fp")
		node.PublicKey = q.Get("pbk")
		node.ShortID = q.Get("sid")
		node.SpiderX = q.Get("spx")
		node.Network = q.Get("type")
		node.Path = q.Get("path")
		node.Host = q.Get("host")
		node.ServiceName = q.Get("serviceName")
		node.Mode = q.Get("mode")
		node.Encryption = q.Get("encryption")

		if node.Security == "reality" {
			node.ProtocolLabel = "VLESS Reality"
		} else {
			node.ProtocolLabel = "VLESS"
		}

	case "trojan":
		if u.User == nil {
			return nil
		}
		node.Password = u.User.Username()
		node.Security = q.Get("security")
		if node.Security == "" {
			node.Security = "tls"
		}
		node.SNI = q.Get("sni")
		node.Fingerprint = q.Get("fp")
		node.Network = q.Get("type")
		node.Path = q.Get("path")
		node.Host = q.Get("host")
		node.ServiceName = q.Get("serviceName")
		node.Insecure = q.Get("insecure") == "1" || q.Get("allowInsecure") == "1"
		if alpnStr := q.Get("alpn"); alpnStr != "" {
			for _, p := range strings.Split(alpnStr, ",") {
				if p = strings.TrimSpace(p); p != "" {
					node.ALPN = append(node.ALPN, p)
				}
			}
		}
		node.ProtocolLabel = "Trojan"

	case "hysteria2", "hy2":
		if u.User != nil {
			node.Password = u.User.Username()
		}
		node.SNI = q.Get("sni")
		node.Obfs = q.Get("obfs")
		node.ObfsPassword = q.Get("obfs-password")
		if node.ObfsPassword == "" {
			node.ObfsPassword = q.Get("obfs_password")
		}
		node.Insecure = q.Get("insecure") == "1" || q.Get("allowInsecure") == "1"
		if alpnStr := q.Get("alpn"); alpnStr != "" {
			parts := strings.Split(alpnStr, ",")
			var alpnList []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					alpnList = append(alpnList, p)
				}
			}
			if len(alpnList) > 0 {
				node.ALPN = alpnList
			}
		}
		node.ProtocolLabel = "Hysteria 2"

	case "ss":
		node.ProtocolLabel = "Shadowsocks"
		if u.User != nil {
			if pass, hasPass := u.User.Password(); hasPass {
				node.Encryption = u.User.Username()
				node.Password = pass
			} else {
				// SIP002: userinfo is base64(method:password)
				if decoded, err := decodeBase64Any(u.User.Username()); err == nil {
					if parts := strings.SplitN(string(decoded), ":", 2); len(parts) == 2 {
						node.Encryption = parts[0]
						node.Password = parts[1]
					}
				}
			}
		} else if u.Host != "" {
			// Legacy format: ss://BASE64(method:password@host:port)#name
			if decoded, err := decodeBase64Any(u.Host); err == nil {
				if lu, err := url.Parse("ss://" + string(decoded)); err == nil && lu.Hostname() != "" {
					node.Server = lu.Hostname()
					if p, err := strconv.Atoi(lu.Port()); err == nil && p > 0 {
						node.Port = p
					}
					if lu.User != nil {
						node.Encryption = lu.User.Username()
						node.Password, _ = lu.User.Password()
					}
				}
			}
		}

	case "warp", "wireguard":
		node.ProtocolLabel = "WARP"

	default:
		return nil
	}

	return node
}

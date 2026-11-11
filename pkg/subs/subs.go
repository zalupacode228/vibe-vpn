package subs

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"vibe-go/pkg/models"
)

func FetchAndParse(subURL string, subID string, subName string, userAgent string, hwid string) ([]models.Node, error) {
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

	return nodes, nil
}

func ParseNodeLink(raw string) *models.Node {
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

	id := fmt.Sprintf("%s_%s_%d", scheme, u.Hostname(), port)

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

		if node.Security == "reality" {
			node.ProtocolLabel = "VLESS Reality"
		} else {
			node.ProtocolLabel = "VLESS"
		}

	case "hysteria2", "hy2":
		node.Password = u.User.Username()
		node.SNI = q.Get("sni")
		node.ProtocolLabel = "Hysteria 2"

	case "ss":
		node.ProtocolLabel = "Shadowsocks"
		if u.User != nil {
			node.Password, _ = u.User.Password()
		}

	case "warp", "wireguard":
		node.ProtocolLabel = "WARP"

	default:
		return nil
	}

	return node
}

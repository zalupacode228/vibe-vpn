package warp

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"vibe-go/pkg/models"
)

func GenerateWarp(mode string) (*models.Node, error) {
	if mode == "masque" {
		return generateMasque()
	}
	return generateAWG()
}

func generateAWG() (*models.Node, error) {
	// 1. Generate X25519 keypair
	privKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate x25519 key: %w", err)
	}
	privBytes := privKey.Bytes()
	pubBytes := privKey.PublicKey().Bytes()
	privB64 := base64.StdEncoding.EncodeToString(privBytes)
	pubB64 := base64.StdEncoding.EncodeToString(pubBytes)

	now := time.Now().UTC().Format(time.RFC3339)
	body := map[string]interface{}{
		"key":        pubB64,
		"install_id": "",
		"fcm_token":  "",
		"tos":        now,
		"type":       "ios",
		"locale":     "en_US",
	}
	bodyBytes, _ := json.Marshal(body)

	baseAPIs := []string{
		"https://api.cloudflareclient.com/v0i1909051800",
		"https://api.cloudflareclient.com/v0a2158",
		"https://api.cloudflareclient.com/v0a3160",
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var regData map[string]interface{}
	var usedAPI string
	var lastErr error

	for _, api := range baseAPIs {
		req, err := http.NewRequest("POST", api+"/reg", bytes.NewReader(bodyBytes))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", "okhttp/3.12.1")
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err == nil {
			regData = parsed
			usedAPI = api
			break
		}
	}

	if regData == nil || usedAPI == "" {
		return nil, fmt.Errorf("Cloudflare WARP registration failed: %v", lastErr)
	}

	res, ok := regData["result"].(map[string]interface{})
	if !ok {
		res = regData
	}

	accID, _ := res["id"].(string)
	token, _ := res["token"].(string)
	if accID == "" || token == "" {
		return nil, fmt.Errorf("failed to get account id or token from Cloudflare")
	}

	// 2. Patch warp_enabled: true
	patchBody := map[string]interface{}{"warp_enabled": true}
	patchBytes, _ := json.Marshal(patchBody)

	patchReq, err := http.NewRequest("PATCH", fmt.Sprintf("%s/reg/%s", usedAPI, accID), bytes.NewReader(patchBytes))
	if err == nil {
		patchReq.Header.Set("User-Agent", "okhttp/3.12.1")
		patchReq.Header.Set("Content-Type", "application/json")
		patchReq.Header.Set("Authorization", "Bearer "+token)

		if patchResp, err := client.Do(patchReq); err == nil {
			var patchData map[string]interface{}
			_ = json.NewDecoder(patchResp.Body).Decode(&patchData)
			patchResp.Body.Close()
			if pRes, ok := patchData["result"].(map[string]interface{}); ok {
				res = pRes
			}
		}
	}

	cfg, _ := res["config"].(map[string]interface{})
	if cfg == nil {
		cfg = make(map[string]interface{})
	}

	peers, _ := cfg["peers"].([]interface{})
	peerPub := "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo="
	if len(peers) > 0 {
		if pMap, ok := peers[0].(map[string]interface{}); ok {
			if pub, ok := pMap["public_key"].(string); ok && pub != "" {
				peerPub = pub
			}
		}
	}

	cidStr, _ := cfg["client_id"].(string)
	reserved := []int{0, 0, 0}
	if cidStr != "" {
		if dec, err := base64.StdEncoding.DecodeString(cidStr); err == nil && len(dec) >= 3 {
			reserved = []int{int(dec[0]), int(dec[1]), int(dec[2])}
		}
	}

	v4 := "172.16.0.2"
	v6 := ""
	if iface, ok := cfg["interface"].(map[string]interface{}); ok {
		if addrs, ok := iface["addresses"].(map[string]interface{}); ok {
			if v, ok := addrs["v4"].(string); ok && v != "" {
				v4 = v
			}
			if v, ok := addrs["v6"].(string); ok && v != "" {
				v6 = v
			}
		}
	}

	localAddresses := []string{v4 + "/32"}
	if v6 != "" {
		localAddresses = append(localAddresses, v6+"/128")
	}

	selectedIP := "162.159.192.1"
	selectedPort := 500

	awgMap := map[string]interface{}{
		"jc":   120,
		"jmin": 23,
		"jmax": 911,
		"s1":   0,
		"s2":   0,
		"h1":   1,
		"h2":   2,
		"h3":   3,
		"h4":   4,
		"i1":   WarpAwgI1,
	}

	rawOut := map[string]interface{}{
		"type":            "wireguard",
		"tag":             "proxy-out",
		"server":          selectedIP,
		"server_port":     selectedPort,
		"local_address":   localAddresses,
		"private_key":     privB64,
		"peer_public_key": peerPub,
		"reserved":        reserved,
		"mtu":             1280,
		"awg":             awgMap,
	}

	node := &models.Node{
		ID:               fmt.Sprintf("node_warp_awg_%d", time.Now().UnixMilli()),
		Name:             "Cloudflare WARP (AWG)",
		Protocol:         "wireguard",
		ProtocolLabel:    "Cloudflare WARP (AWG)",
		Server:           selectedIP,
		Port:             selectedPort,
		PrivateKey:       privB64,
		PeerPublicKey:    peerPub,
		LocalAddress:     v4 + "/32",
		Reserved:         reserved,
		SubscriptionName: "Cloudflare WARP",
		Ping:             0,
		AWG:              awgMap,
		JC:               120,
		Jmin:             23,
		Jmax:             911,
		S1:               0,
		S2:               0,
		H1:               1,
		H2:               2,
		H3:               3,
		H4:               4,
		I1:               WarpAwgI1,
		RawOutbound:      rawOut,
	}

	return node, nil
}

func generateMasque() (*models.Node, error) {
	client := &http.Client{Timeout: 9 * time.Second}
	req, err := http.NewRequest("GET", "https://masque.configwireguard1.workers.dev/api/list", nil)
	if err == nil {
		req.Header.Set("User-Agent", "Mozilla/5.0")
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			var poolData struct {
				Keys []struct {
					Address       string `json:"address"`
					PrivateKeyRaw string `json:"private_key_raw"`
				} `json:"keys"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&poolData); err == nil && len(poolData.Keys) > 0 {
				idx := int(time.Now().UnixNano() % int64(len(poolData.Keys)))
				if idx < 0 {
					idx = -idx
				}
				k := poolData.Keys[idx]
				v4 := "172.16.0.2"

				return &models.Node{
					ID:               fmt.Sprintf("node_warp_masque_%d", time.Now().UnixMilli()),
					Name:             "Cloudflare WARP (MASQUE)",
					Protocol:         "masque",
					ProtocolLabel:    "Cloudflare WARP (MASQUE)",
					Server:           "162.159.198.2",
					Port:             443,
					PrivateKey:       k.PrivateKeyRaw,
					SNI:              "www.microsoft.com",
					LocalAddress:     v4 + "/32",
					SubscriptionName: "Cloudflare WARP",
					Ping:             0,
				}, nil
			}
		}
	}

	// Fallback to AWG if pool unavailable
	return generateAWG()
}

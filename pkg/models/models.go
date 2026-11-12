package models

type Node struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Protocol         string                 `json:"protocol"`
	ProtocolLabel    string                 `json:"protocol_label,omitempty"`
	Server           string                 `json:"server"`
	Port             int                    `json:"port"`
	UUID             string                 `json:"uuid,omitempty"`
	Password         string                 `json:"password,omitempty"`
	Flow             string                 `json:"flow,omitempty"`
	Security         string                 `json:"security,omitempty"`
	SNI              string                 `json:"sni,omitempty"`
	Fingerprint      string                 `json:"fp,omitempty"`
	PublicKey        string                 `json:"pbk,omitempty"`
	ShortID          string                 `json:"sid,omitempty"`
	SpiderX          string                 `json:"spx,omitempty"`
	ALPN             []string               `json:"alpn,omitempty"`
	Path             string                 `json:"path,omitempty"`
	Host             string                 `json:"host,omitempty"`
	ServiceName      string                 `json:"service_name,omitempty"`
	Network          string                 `json:"network,omitempty"`
	Mode             string                 `json:"mode,omitempty"`
	Encryption       string                 `json:"encryption,omitempty"`
	SubscriptionID   string                 `json:"subscription_id,omitempty"`
	SubscriptionName string                 `json:"subscription_name,omitempty"`
	Ping             int                    `json:"ping"`
	Country          string                 `json:"country,omitempty"`
	PrivateKey       string                 `json:"private_key,omitempty"`
	PeerPublicKey    string                 `json:"peer_public_key,omitempty"`
	LocalAddress     string                 `json:"local_address,omitempty"`
	Reserved         []int                  `json:"reserved,omitempty"`
	RawOutbound      map[string]interface{} `json:"raw_outbound,omitempty"`
	AWG              map[string]interface{} `json:"awg,omitempty"`
	JC               int                    `json:"jc,omitempty"`
	Jmin             int                    `json:"jmin,omitempty"`
	Jmax             int                    `json:"jmax,omitempty"`
	S1               int                    `json:"s1,omitempty"`
	S2               int                    `json:"s2,omitempty"`
	S3               int                    `json:"s3,omitempty"`
	S4               int                    `json:"s4,omitempty"`
	H1               int                    `json:"h1,omitempty"`
	H2               int                    `json:"h2,omitempty"`
	H3               int                    `json:"h3,omitempty"`
	H4               int                    `json:"h4,omitempty"`
	I1               string                 `json:"i1,omitempty"`
}

type Subscription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	UpdatedAt string `json:"updated_at"`
	Count     int    `json:"count"`
	Upload      int64  `json:"upload"`
	Download    int64  `json:"download"`
	Total       int64  `json:"total"`
	Expire      int64  `json:"expire"`
	HasUserInfo bool   `json:"has_userinfo"`
}

type CustomList struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url,omitempty"`
	Action      string   `json:"action,omitempty"`
	RuleType    string   `json:"rule_type,omitempty"` // geosite, domain, ip
	Outbound    string   `json:"outbound,omitempty"`  // proxy, direct, block
	Enabled     bool     `json:"enabled"`
	Domains     []string `json:"domains,omitempty"`
	IPs         []string `json:"ips,omitempty"`
	Count       int      `json:"count"`
	LastUpdated string   `json:"last_updated,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

type AppRule struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Display   string `json:"display,omitempty"`
	Category  string `json:"category,omitempty"`
	Enabled   bool   `json:"enabled"`
	IsEnabled bool   `json:"is_enabled,omitempty"`
	IsCustom  bool   `json:"is_custom,omitempty"`
}

type Settings struct {
	Language       string       `json:"language,omitempty"`
	SplitMode      string       `json:"split_mode"`
	ActiveNodeID   string       `json:"active_node_id,omitempty"`
	PrivacyBlur    bool         `json:"privacy_blur"`
	AutoStart      bool         `json:"auto_start"`
	MinimizeToTray bool         `json:"minimize_to_tray"`
	StrictRoute    bool         `json:"strict_route"`
	BypassLAN      bool         `json:"bypass_lan"`
	BypassRU       bool         `json:"bypass_ru"`
	AdBlock        bool         `json:"adblock"`
	CustomLists    []CustomList `json:"custom_lists,omitempty"`
	MTU            int          `json:"mtu,omitempty"`
	TunMTU         int          `json:"tun_mtu,omitempty"`
	RemoteDNS      string       `json:"remote_dns,omitempty"`
	DirectDNS      string       `json:"direct_dns,omitempty"`
	UserAgent              string       `json:"user_agent,omitempty"`
	HWID                   string       `json:"hwid,omitempty"`
	AutoUpdateSubs         bool         `json:"auto_update_subs"`
	AutoUpdateIntervalHours int         `json:"auto_update_interval_hours,omitempty"`
	KillSwitch             bool         `json:"kill_switch"`
	WarpDetour             bool         `json:"warp_detour"`
}

type BackupProfile struct {
	Version       string         `json:"version"`
	ExportedAt    string         `json:"exported_at"`
	Settings      Settings       `json:"settings"`
	Subscriptions []Subscription `json:"subscriptions"`
	Nodes         []Node         `json:"nodes"`
	AppRules      []AppRule      `json:"app_rules"`
	CustomLists   []CustomList   `json:"custom_lists"`
}

type ProcessInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Category string `json:"category"`
	Enabled  bool   `json:"enabled"`
	Icon     string `json:"icon,omitempty"`
}

type TrafficStats struct {
	Connected bool   `json:"connected"`
	DownSpeed int64  `json:"down_speed"`
	UpSpeed   int64  `json:"up_speed"`
	TotalDown int64  `json:"total_down"`
	TotalUp   int64  `json:"total_up"`
	Ping      int    `json:"ping"`
	IP        string `json:"ip"`
	NodeName  string `json:"node_name"`
	Mode      string `json:"mode"`
}

type ConnectionGroup struct {
	Process   string           `json:"process"`
	Count     int              `json:"count"`
	DownSpeed string           `json:"down_speed"`
	UpSpeed   string           `json:"up_speed"`
	Children  []ConnectionItem `json:"children"`
}

type ConnectionItem struct {
	ID          string `json:"id"`
	Destination string `json:"destination"`
	Rule        string `json:"rule"`
	DownSpeed   string `json:"down_speed"`
	UpSpeed     string `json:"up_speed"`
}

type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	ReleaseNotes   string `json:"release_notes"`
	DownloadURL    string `json:"download_url"`
	AssetSize      int64  `json:"asset_size"`
	PublishedAt    string `json:"published_at"`
}

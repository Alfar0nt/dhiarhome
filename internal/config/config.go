package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxmox       ProxmoxConfig        `yaml:"proxmox"`
	Docker        DockerConfig         `yaml:"docker"`
	Services      []ServiceConfig      `yaml:"services"`
	Appearance    AppearanceConfig     `yaml:"appearance"`
	Widgets       WidgetsConfig        `yaml:"widgets"`
	Network       NetworkConfig        `yaml:"network"`
	Todos         TodoConfig           `yaml:"todos"`
	MediaServices []MediaServiceConfig `yaml:"media_services"`
	Bookmarks     []BookmarkGroup      `yaml:"bookmarks"`
	Notifications NotificationsConfig  `yaml:"notifications"`
}

type MediaServiceConfig struct {
	Name   string `yaml:"name"`    // "Sonarr", "Radarr", "Overseerr"
	URL    string `yaml:"url"`     // API base URL (e.g. http://192.168.1.100:8989)
	APIKey string `yaml:"api_key"` // API key
	WebUI  string `yaml:"webui"`   // Web UI URL (e.g. http://192.168.1.100:8989)
}

type AppearanceConfig struct {
	BackgroundImage   string  `yaml:"background_image"`
	BackgroundURL     string  `yaml:"background_url"`
	BackgroundOpacity float64 `yaml:"background_opacity"`
	BackgroundBlur    int     `yaml:"background_blur"`
	Logo              string  `yaml:"logo"`
	Theme             string  `yaml:"theme"`
	CardOpacity       float64 `yaml:"card_opacity"`
	CardBlur          int     `yaml:"card_blur"`
	AccentColor       string  `yaml:"accent_color"`
}

type ProxmoxConfig struct {
	URL         string            `yaml:"url"`
	NodeName    string            `yaml:"node_name"`
	TokenID     string            `yaml:"token_id"`
	TokenSecret string            `yaml:"token_secret"`
	Mock        bool              `yaml:"mock"`
	ExtraDisks  []ExtraDiskConfig `yaml:"extra_disks"`
}

type ExtraDiskConfig struct {
	Mountpoint string `yaml:"mountpoint"`  // required: e.g. "/mnt/data"
	Label      string `yaml:"label"`       // optional: friendly name (defaults to mountpoint)
	Total      string `yaml:"total"`       // optional: manual override, e.g. "500GB", "1TB"
	Used       string `yaml:"used"`        // optional: manual override, e.g. "200GB"
	AutoDetect bool   `yaml:"auto_detect"` // if true (default), read from filesystem via statfs
}

type DockerConfig struct {
	Socket            string   `yaml:"socket"`
	MonitorContainers []string `yaml:"monitor_containers"`
	// Remote Docker TLS options
	SkipTLS   bool   `yaml:"skip_tls"`
	TLSCACert string `yaml:"tls_ca_cert"`
	TLSCert   string `yaml:"tls_cert"`
	TLSKey    string `yaml:"tls_key"`
	// Portainer API integration
	PortainerURL   string `yaml:"portainer_url"`
	PortainerKey   string `yaml:"portainer_api_key"`
	PortainerEnvID int    `yaml:"portainer_env_id"`
}

type ServiceConfig struct {
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	SkipTLS bool   `yaml:"skip_tls"`
}

type WidgetsConfig struct {
	Weather    WeatherWidgetConfig    `yaml:"weather"`
	DateTime   DateTimeWidgetConfig   `yaml:"datetime"`
	SystemInfo SystemInfoWidgetConfig `yaml:"system_info"`
	CustomText CustomTextWidgetConfig `yaml:"custom_text"`
}

type WeatherWidgetConfig struct {
	Enabled      bool    `yaml:"enabled"`
	Latitude     float64 `yaml:"latitude"`
	Longitude    float64 `yaml:"longitude"`
	Units        string  `yaml:"units"` // "celsius" or "fahrenheit"
	CacheMinutes int     `yaml:"cache_minutes"`
	Mock         bool    `yaml:"mock"`
}

type DateTimeWidgetConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Timezone  string `yaml:"timezone"` // e.g., "America/New_York"
	Format24h bool   `yaml:"format_24h"`
}

type SystemInfoWidgetConfig struct {
	Enabled bool `yaml:"enabled"`
}

type CustomTextWidgetConfig struct {
	Enabled bool   `yaml:"enabled"`
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

type NetworkConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Interfaces     []NetIfConfig `yaml:"interfaces"`
	ShowSpeed      bool          `yaml:"show_speed"`
	ShowTotal      bool          `yaml:"show_total_transfer"`
	UpdateInterval int           `yaml:"update_interval"`
	Mock           bool          `yaml:"mock"`
}

type NetIfConfig struct {
	Name  string `yaml:"name"`
	Label string `yaml:"label"`
}

type TodoConfig struct {
	Enabled  bool   `yaml:"enabled"`
	FilePath string `yaml:"file_path"` // JSON file for persistence
	Title    string `yaml:"title"`
}

type BookmarkGroup struct {
	Group string         `yaml:"group"`
	Links []BookmarkLink `yaml:"links"`
}

type BookmarkLink struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Icon        string `yaml:"icon"` // Lucide icon name, image path, or "favicon"
	Description string `yaml:"description"`
	NewTab      bool   `yaml:"new_tab"` // Open in new tab (default true)
}

type NotificationsConfig struct {
	Telegram TelegramConfig `yaml:"telegram"`
}

type TelegramConfig struct {
	Enabled         bool   `yaml:"enabled"`
	BotToken        string `yaml:"bot_token"`
	ChatID          string `yaml:"chat_id"`
	MessageThreadID int    `yaml:"message_thread_id"` // optional: topic/thread ID for forum groups
	NotifyUp        bool   `yaml:"notify_up"`
	NotifyDown      bool   `yaml:"notify_down"`
	Cooldown        int    `yaml:"cooldown"`    // minutes between repeat alerts (default: 5)
	SilentHours     []int  `yaml:"silent_hours"` // optional: hours to suppress (e.g., [23,0,1])
	Mock            bool   `yaml:"mock"`        // dry-run: log to stdout instead of sending
}

// parseSizeRegex matches a number (with optional decimal) followed by a unit suffix.
var parseSizeRegex = regexp.MustCompile(`(?i)^\s*([\d.]+)\s*(B|KB|MB|GB|TB|KIB|MIB|GIB|TIB)?\s*$`)

// ParseSize converts a human-readable size string (e.g. "500GB", "1.5TB", "200MB") to bytes.
func ParseSize(s string) (int64, error) {
	m := parseSizeRegex.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("invalid size format: %q", s)
	}
	val, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number in size: %q", m[1])
	}
	unit := strings.ToUpper(m[2])
	if unit == "" {
		unit = "B"
	}
	multipliers := map[string]float64{
		"B":   1,
		"KB":  1000,
		"MB":  1000 * 1000,
		"GB":  1000 * 1000 * 1000,
		"TB":  1000 * 1000 * 1000 * 1000,
		"KIB": 1024,
		"MIB": 1024 * 1024,
		"GIB": 1024 * 1024 * 1024,
		"TIB": 1024 * 1024 * 1024 * 1024,
	}
	mult, ok := multipliers[unit]
	if !ok {
		return 0, fmt.Errorf("unknown size unit: %q", unit)
	}
	return int64(val * mult), nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.setDefaults()
	cfg.validate()
	return &cfg, nil
}

func (c *Config) setDefaults() {
	if c.Appearance.Theme == "" {
		c.Appearance.Theme = "dark"
	}
	if c.Appearance.BackgroundOpacity == 0 {
		c.Appearance.BackgroundOpacity = 0.3
	}
	if c.Appearance.BackgroundBlur == 0 {
		c.Appearance.BackgroundBlur = 5
	}
	if c.Appearance.CardOpacity == 0 {
		c.Appearance.CardOpacity = 0.6
	}
	if c.Appearance.CardBlur == 0 {
		c.Appearance.CardBlur = 12
	}
	if c.Appearance.AccentColor == "" {
		c.Appearance.AccentColor = "#3b82f6"
	}
	// Widget defaults
	if c.Widgets.Weather.CacheMinutes == 0 {
		c.Widgets.Weather.CacheMinutes = 15
	}
	if c.Widgets.Weather.Units == "" {
		c.Widgets.Weather.Units = "celsius"
	}
	if c.Widgets.DateTime.Timezone == "" {
		c.Widgets.DateTime.Timezone = "Local"
	}
	if c.Widgets.CustomText.Title == "" {
		c.Widgets.CustomText.Title = "Note"
	}
	// Network defaults
	if c.Network.UpdateInterval == 0 {
		c.Network.UpdateInterval = 3
	}
	// Todo defaults
	if c.Todos.FilePath == "" {
		c.Todos.FilePath = "data/todos.json"
	}
	if c.Todos.Title == "" {
		c.Todos.Title = "To-Do"
	}
	// Telegram defaults
	if c.Notifications.Telegram.Cooldown == 0 {
		c.Notifications.Telegram.Cooldown = 5
	}
	if c.Notifications.Telegram.Enabled {
		if c.Notifications.Telegram.SilentHours == nil {
			c.Notifications.Telegram.SilentHours = []int{}
		}
	}
}

// validate checks config for common issues and logs warnings.
// Features with bad config are gracefully disabled (not crashed).
func (c *Config) validate() {
	// Appearance validation
	if c.Appearance.BackgroundOpacity < 0 || c.Appearance.BackgroundOpacity > 1 {
		log.Printf("[WARN] appearance.background_opacity %.2f out of range (0-1), clamping to 0.3", c.Appearance.BackgroundOpacity)
		c.Appearance.BackgroundOpacity = 0.3
	}
	if c.Appearance.CardOpacity < 0 || c.Appearance.CardOpacity > 1 {
		log.Printf("[WARN] appearance.card_opacity %.2f out of range (0-1), clamping to 0.6", c.Appearance.CardOpacity)
		c.Appearance.CardOpacity = 0.6
	}
	if c.Appearance.BackgroundBlur < 0 || c.Appearance.BackgroundBlur > 30 {
		log.Printf("[WARN] appearance.background_blur %d out of range (0-30), clamping to 5", c.Appearance.BackgroundBlur)
		c.Appearance.BackgroundBlur = 5
	}
	if c.Appearance.CardBlur < 0 || c.Appearance.CardBlur > 50 {
		log.Printf("[WARN] appearance.card_blur %d out of range (0-50), clamping to 12", c.Appearance.CardBlur)
		c.Appearance.CardBlur = 12
	}

	// Validate background URL if provided
	if c.Appearance.BackgroundURL != "" {
		if _, err := url.ParseRequestURI(c.Appearance.BackgroundURL); err != nil {
			log.Printf("[WARN] appearance.background_url is invalid (%v), ignoring", err)
			c.Appearance.BackgroundURL = ""
		}
	}

	// Proxmox validation (only if not mock)
	if !c.Proxmox.Mock && c.Proxmox.URL != "" {
		if _, err := url.ParseRequestURI(c.Proxmox.URL); err != nil {
			log.Printf("[WARN] proxmox.url is invalid (%v), falling back to mock mode", err)
			c.Proxmox.Mock = true
		}
		if c.Proxmox.TokenID == "" || c.Proxmox.TokenSecret == "" {
			log.Println("[WARN] proxmox.token_id or proxmox.token_secret is empty, falling back to mock mode")
			c.Proxmox.Mock = true
		}
	}

	// Weather validation
	if c.Widgets.Weather.Enabled && !c.Widgets.Weather.Mock {
		if c.Widgets.Weather.Latitude == 0 && c.Widgets.Weather.Longitude == 0 {
			log.Println("[WARN] widgets.weather enabled but latitude/longitude not set, disabling")
			c.Widgets.Weather.Enabled = false
		}
		if c.Widgets.Weather.Units != "celsius" && c.Widgets.Weather.Units != "fahrenheit" {
			log.Printf("[WARN] widgets.weather.units '%s' invalid, defaulting to celsius", c.Widgets.Weather.Units)
			c.Widgets.Weather.Units = "celsius"
		}
		if c.Widgets.Weather.CacheMinutes < 1 {
			c.Widgets.Weather.CacheMinutes = 15
		}
	}

	// DateTime validation
	if c.Widgets.DateTime.Enabled {
		if _, err := time.LoadLocation(c.Widgets.DateTime.Timezone); err != nil {
			log.Printf("[WARN] widgets.datetime.timezone '%s' invalid, using Local", c.Widgets.DateTime.Timezone)
			c.Widgets.DateTime.Timezone = "Local"
		}
	}

	// Network validation
	if c.Network.Enabled {
		if c.Network.UpdateInterval < 1 {
			log.Println("[WARN] network.update_interval < 1s, clamping to 1")
			c.Network.UpdateInterval = 1
		}
		if c.Network.UpdateInterval > 60 {
			log.Println("[WARN] network.update_interval > 60s, clamping to 60")
			c.Network.UpdateInterval = 60
		}
	}

	// Service URL validation
	for i, svc := range c.Services {
		if svc.URL != "" {
			if _, err := url.ParseRequestURI(svc.URL); err != nil {
				log.Printf("[WARN] services[%d].url '%s' is invalid (%v), skipping", i, svc.Name, err)
			}
		}
	}

	// Media services validation
	for i, ms := range c.MediaServices {
		if ms.URL != "" {
			if _, err := url.ParseRequestURI(ms.URL); err != nil {
				log.Printf("[WARN] media_services[%d].url '%s' is invalid (%v), skipping", i, ms.Name, err)
			}
		}
	}

	// Bookmark URL validation
	for gi, group := range c.Bookmarks {
		for li, link := range group.Links {
			if link.URL != "" {
				if _, err := url.ParseRequestURI(link.URL); err != nil {
					log.Printf("[WARN] bookmarks[%d].links[%d].url '%s' is invalid (%v)", gi, li, link.Name, err)
				}
			}
		}
	}

	// Extra disks validation
	for i, ed := range c.Proxmox.ExtraDisks {
		if ed.Mountpoint == "" {
			log.Printf("[WARN] proxmox.extra_disks[%d] missing mountpoint, skipping", i)
			continue
		}
		if ed.Total != "" {
			if _, err := ParseSize(ed.Total); err != nil {
				log.Printf("[WARN] proxmox.extra_disks[%d].total '%s' invalid (%v), skipping", i, ed.Total, err)
				c.Proxmox.ExtraDisks[i].Total = ""
			}
		}
		if ed.Used != "" {
			if _, err := ParseSize(ed.Used); err != nil {
				log.Printf("[WARN] proxmox.extra_disks[%d].used '%s' invalid (%v), skipping", i, ed.Used, err)
				c.Proxmox.ExtraDisks[i].Used = ""
			}
		}
	}

	// Telegram validation
	if c.Notifications.Telegram.Enabled && !c.Notifications.Telegram.Mock {
		if c.Notifications.Telegram.BotToken == "" {
			log.Println("[WARN] notifications.telegram.enabled but bot_token is empty, disabling")
			c.Notifications.Telegram.Enabled = false
		}
		if c.Notifications.Telegram.ChatID == "" {
			log.Println("[WARN] notifications.telegram.enabled but chat_id is empty, disabling")
			c.Notifications.Telegram.Enabled = false
		}
		if c.Notifications.Telegram.Cooldown < 1 {
			c.Notifications.Telegram.Cooldown = 1
		}
	}

	// Print feature summary
	features := []string{}
	if !c.Proxmox.Mock && c.Proxmox.URL != "" {
		features = append(features, "Proxmox")
	} else if c.Proxmox.Mock {
		features = append(features, "Proxmox (mock)")
	}
	if c.Widgets.Weather.Enabled {
		features = append(features, fmt.Sprintf("Weather (%s)", c.Widgets.Weather.Units))
	}
	if c.Widgets.DateTime.Enabled {
		features = append(features, "DateTime")
	}
	if c.Widgets.SystemInfo.Enabled {
		features = append(features, "SystemInfo")
	}
	if c.Network.Enabled {
		features = append(features, fmt.Sprintf("Network (%d interfaces)", len(c.Network.Interfaces)))
	}
	if len(c.Services) > 0 {
		features = append(features, fmt.Sprintf("Services (%d)", len(c.Services)))
	}
	if len(c.MediaServices) > 0 {
		features = append(features, fmt.Sprintf("MediaServices (%d)", len(c.MediaServices)))
	}
	if c.Todos.Enabled {
		features = append(features, "Todos")
	}
	if len(c.Bookmarks) > 0 {
		totalLinks := 0
		for _, g := range c.Bookmarks {
			totalLinks += len(g.Links)
		}
		features = append(features, fmt.Sprintf("Bookmarks (%d links)", totalLinks))
	}
	if len(c.Proxmox.ExtraDisks) > 0 {
		features = append(features, fmt.Sprintf("ExtraDisks (%d)", len(c.Proxmox.ExtraDisks)))
	}
	if c.Notifications.Telegram.Enabled {
		if c.Notifications.Telegram.Mock {
			features = append(features, "Telegram (mock)")
		} else {
			features = append(features, "Telegram")
		}
	}
	if len(features) > 0 {
		log.Printf("Active features: %s", strings.Join(features, ", "))
	}
}

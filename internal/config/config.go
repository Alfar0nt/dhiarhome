package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxmox    ProxmoxConfig    `yaml:"proxmox"`
	Docker     DockerConfig     `yaml:"docker"`
	Services   []ServiceConfig  `yaml:"services"`
	Appearance AppearanceConfig `yaml:"appearance"`
	Widgets    WidgetsConfig    `yaml:"widgets"`
	Network    NetworkConfig    `yaml:"network"`
	Todos      TodoConfig       `yaml:"todos"`
}

type AppearanceConfig struct {
	BackgroundImage   string  `yaml:"background_image"`
	BackgroundURL     string  `yaml:"background_url"`
	BackgroundOpacity float64 `yaml:"background_opacity"`
	BackgroundBlur    int     `yaml:"background_blur"`
	Theme             string  `yaml:"theme"`
	CardOpacity       float64 `yaml:"card_opacity"`
	CardBlur          int     `yaml:"card_blur"`
	AccentColor       string  `yaml:"accent_color"`
}

type ProxmoxConfig struct {
	URL         string `yaml:"url"`
	NodeName    string `yaml:"node_name"`
	TokenID     string `yaml:"token_id"`
	TokenSecret string `yaml:"token_secret"`
	Mock        bool   `yaml:"mock"`
}

type DockerConfig struct {
	Socket            string   `yaml:"socket"`
	MonitorContainers []string `yaml:"monitor_containers"`
}

type ServiceConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
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
}

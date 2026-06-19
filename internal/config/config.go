package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

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
	Bookmarks  []BookmarkGroup  `yaml:"bookmarks"`
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
	NodeName string `yaml:"node_name"`
}

type DockerConfig struct {
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
	Enabled bool   `yaml:"enabled"`
	Units   string `yaml:"units"`
}

type DateTimeWidgetConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Timezone  string `yaml:"timezone"`
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
}

type NetIfConfig struct {
	Name  string `yaml:"name"`
	Label string `yaml:"label"`
}

type TodoConfig struct {
	Enabled  bool   `yaml:"enabled"`
	FilePath string `yaml:"file_path"`
	Title    string `yaml:"title"`
}

type BookmarkGroup struct {
	Group string         `yaml:"group"`
	Links []BookmarkLink `yaml:"links"`
}

type BookmarkLink struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Icon        string `yaml:"icon"`
	Description string `yaml:"description"`
	NewTab      bool   `yaml:"new_tab"`
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
	if c.Widgets.Weather.Units == "" {
		c.Widgets.Weather.Units = "celsius"
	}
	if c.Widgets.DateTime.Timezone == "" {
		c.Widgets.DateTime.Timezone = "Local"
	}
	if c.Widgets.CustomText.Title == "" {
		c.Widgets.CustomText.Title = "Note"
	}
	if c.Network.UpdateInterval == 0 {
		c.Network.UpdateInterval = 3
	}
	if c.Todos.FilePath == "" {
		c.Todos.FilePath = "data/todos.json"
	}
	if c.Todos.Title == "" {
		c.Todos.Title = "To-Do"
	}
}

func (c *Config) validate() {
	// Appearance validation
	if c.Appearance.BackgroundOpacity < 0 || c.Appearance.BackgroundOpacity > 1 {
		c.Appearance.BackgroundOpacity = 0.3
	}
	if c.Appearance.CardOpacity < 0 || c.Appearance.CardOpacity > 1 {
		c.Appearance.CardOpacity = 0.6
	}
	if c.Appearance.BackgroundBlur < 0 || c.Appearance.BackgroundBlur > 30 {
		c.Appearance.BackgroundBlur = 5
	}
	if c.Appearance.CardBlur < 0 || c.Appearance.CardBlur > 50 {
		c.Appearance.CardBlur = 12
	}
	if c.Appearance.BackgroundURL != "" {
		if _, err := url.ParseRequestURI(c.Appearance.BackgroundURL); err != nil {
			log.Printf("[WARN] appearance.background_url is invalid, ignoring")
			c.Appearance.BackgroundURL = ""
		}
	}

	// Weather validation
	if c.Widgets.Weather.Units != "celsius" && c.Widgets.Weather.Units != "fahrenheit" {
		c.Widgets.Weather.Units = "celsius"
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
			c.Network.UpdateInterval = 1
		}
		if c.Network.UpdateInterval > 60 {
			c.Network.UpdateInterval = 60
		}
	}

	// Service URL validation
	for i, svc := range c.Services {
		if svc.URL != "" {
			if _, err := url.ParseRequestURI(svc.URL); err != nil {
				log.Printf("[WARN] services[%d].url '%s' is invalid, skipping", i, svc.Name)
			}
		}
	}

	// Bookmark URL validation
	for gi, group := range c.Bookmarks {
		for li, link := range group.Links {
			if link.URL != "" {
				if _, err := url.ParseRequestURI(link.URL); err != nil {
					log.Printf("[WARN] bookmarks[%d].links[%d].url '%s' is invalid", gi, li, link.Name)
				}
			}
		}
	}

	// Print feature summary
	features := []string{}
	features = append(features, "Proxmox (demo)")
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
	if len(features) > 0 {
		log.Printf("Active features: %s", strings.Join(features, ", "))
	}
}

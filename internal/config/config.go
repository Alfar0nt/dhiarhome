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
}

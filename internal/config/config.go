package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxmox  ProxmoxConfig  `yaml:"proxmox"`
	Docker   DockerConfig   `yaml:"docker"`
	Services []ServiceConfig `yaml:"services"`
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
	return &cfg, nil
}

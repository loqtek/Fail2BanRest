package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Auth     AuthConfig     `yaml:"auth"`
	Fail2ban Fail2banConfig `yaml:"fail2ban"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	Host string    `yaml:"host"`
	Port int       `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type AuthConfig struct {
	JWTSecret   string `yaml:"jwt_secret"`
	TokenExpiry string `yaml:"token_expiry"`
}

type Fail2banConfig struct {
	ClientPath string `yaml:"client_path"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

var defaultConfig = Config{
	Server: ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
		TLS: TLSConfig{
			Enabled: false,
		},
	},
	Auth: AuthConfig{
		JWTSecret:   "change-this-secret",
		TokenExpiry: "24h",
	},
	Fail2ban: Fail2banConfig{
		ClientPath: "/usr/bin/fail2ban-client",
	},
	Logging: LoggingConfig{
		Level: "info",
	},
}

func LoadConfig(path string) (*Config, error) {
	config := defaultConfig

	if path == "" {
		// Try common config locations
		paths := []string{"config.yaml", "config.yml", "/etc/fail2rest/config.yaml"}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Validate
	if config.Auth.JWTSecret == "" || config.Auth.JWTSecret == "change-this-secret" {
		return nil, fmt.Errorf("jwt_secret must be set in config")
	}

	return &config, nil
}

func (c *Config) GetTokenExpiry() (time.Duration, error) {
	return time.ParseDuration(c.Auth.TokenExpiry)
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}


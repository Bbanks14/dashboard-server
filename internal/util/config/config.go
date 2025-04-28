// File: internal/util/config/config.go
package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config contains all application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
}

// ServerConfig contains server-specific configuration
type ServerConfig struct {
	Port         string `yaml:"port"`
	Environment  string `yaml:"environment"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// DatabaseConfig contains database-specific configuration
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

// AuthConfig contains authentication-specific configuration
type AuthConfig struct {
	JWTSecret        string `yaml:"jwt_secret"`
	TokenExpiration  int    `yaml:"token_expiration"` // in hours
	RefreshTokenExpiration int `yaml:"refresh_token_expiration"` // in days
}

// LoadConfig loads configuration from the specified file
func LoadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// File: configs/app.yaml
server:
  port: "8080"
  environment: "development"
  read_timeout: 10
  write_timeout: 10

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "password"
  name: "dashboard_db"

auth:
  jwt_secret: "your_jwt_secret_key_change_in_production"
  token_expiration: 24
  refresh_token_expiration: 7

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	Log    LogConfig    `yaml:"log"`
}

type ServerConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	DebugPort int    `yaml:"debug_port"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	overrideFromEnv(&cfg)

	return &cfg, nil
}

func overrideFromEnv(cfg *Config) {
	if host := os.Getenv("DB_HOST"); host != "" {
		fmt.Printf("[CONFIG] Overriding DB_HOST from env: %s (was: %s)\n", host, cfg.DB.Host)
		cfg.DB.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		oldPort := cfg.DB.Port
		fmt.Sscanf(port, "%d", &cfg.DB.Port)
		fmt.Printf("[CONFIG] Overriding DB_PORT from env: %d (was: %d)\n", cfg.DB.Port, oldPort)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		fmt.Printf("[CONFIG] Overriding DB_USER from env: %s (was: %s)\n", user, cfg.DB.User)
		cfg.DB.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		fmt.Printf("[CONFIG] Overriding DB_PASSWORD from env: ***\n")
		cfg.DB.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		fmt.Printf("[CONFIG] Overriding DB_NAME from env: %s (was: %s)\n", dbname, cfg.DB.DBName)
		cfg.DB.DBName = dbname
	}
	if sslmode := os.Getenv("DB_SSLMODE"); sslmode != "" {
		fmt.Printf("[CONFIG] Overriding DB_SSLMODE from env: %s (was: %s)\n", sslmode, cfg.DB.SSLMode)
		cfg.DB.SSLMode = sslmode
	}
}

package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server  ServerConfig  `toml:"server"`
	DB      DBConfig      `toml:"db"`
	Twitter TwitterConfig `toml:"twitter"`
}

type ServerConfig struct {
	Port          int    `toml:"port"`
	SessionSecret string `toml:"session_secret"`
}

type DBConfig struct {
	Path string `toml:"path"`
}

type TwitterConfig struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	RedirectURL  string `toml:"redirect_url"`
}

func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.DB.Path == "" {
		cfg.DB.Path = "db/pcgw.db"
	}
	return &cfg, nil
}

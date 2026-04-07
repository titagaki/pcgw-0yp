package config

import (
	"os"

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
	Host   string `toml:"host"`
	Port   int    `toml:"port"`
	User   string `toml:"user"`
	Passwd string `toml:"passwd"`
	DBName string `toml:"dbname"`
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
	if cfg.DB.Host == "" {
		cfg.DB.Host = "127.0.0.1"
	}
	if cfg.DB.Port == 0 {
		cfg.DB.Port = 3306
	}
	if cfg.DB.User == "" {
		cfg.DB.User = "pcgw"
	}
	if cfg.DB.DBName == "" {
		cfg.DB.DBName = "pcgw"
	}
	// 環境変数があればTOMLの値を上書き
	if v := os.Getenv("SESSION_SECRET"); v != "" {
		cfg.Server.SessionSecret = v
	}
	if v := os.Getenv("DB_PASSWD"); v != "" {
		cfg.DB.Passwd = v
	}
	if v := os.Getenv("TWITTER_CLIENT_ID"); v != "" {
		cfg.Twitter.ClientID = v
	}
	if v := os.Getenv("TWITTER_CLIENT_SECRET"); v != "" {
		cfg.Twitter.ClientSecret = v
	}
	return &cfg, nil
}

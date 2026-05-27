package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App     App     `mapstructure:"app"`
	Server  Server  `mapstructure:"server"`
	Storage Storage `mapstructure:"storage"`
}

type App struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	LogPath string `mapstructure:"logPath"`
	Env     string `mapstructure:"env"`
}

type Server struct {
	Host             string        `mapstructure:"host"`
	Port             string        `mapstructure:"port"`
	Timeout          time.Duration `mapstructure:"timeout"`
	IdleTimeout      time.Duration `mapstructure:"idleTimeout"`
	KeepAliveTimeout time.Duration `mapstructure:"keepAliveTimeout"`
}

type Storage struct {
	User           string        `mapstructure:"user"`
	Password       string        `mapstructure:"password"`
	Database       string        `mapstructure:"database"`
	Timezone       string        `mapstructure:"timezone"`
	Host           string        `mapstructure:"host"`
	Port           string        `mapstructure:"port"`
	SSLMode        string        `mapstructure:"sslMode"`
	ConnectTimeout time.Duration `mapstructure:"connectTimeout"`
	PingTimeout    time.Duration `mapstructure:"pingTimeout"`
	AmountOfConns  int32         `mapstructure:"amountOfConns"`
}

func MustInitConfig(path string) *Config {
	if path == "" {
		panic("config path is empty")
	}

	filename := filepath.Join(path, "config.yaml")
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Errorf("failed to read config file: %w", err))
	}
	data = []byte(os.ExpandEnv(string(data)))
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBuffer(data)); err != nil {
		panic(fmt.Errorf("failed to read config: %w", err))
	}
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %w", err))
	}
	return cfg
}

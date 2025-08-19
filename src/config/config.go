package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env          string `yaml:"env" env-default:"local"`
	StoragePath  string `yaml:"storage_path" env-required:"true"`
	DatabaseMess string `yaml:"database_url" env-default:"postgres://user:postgres@localhost:5432/mess?sslmode=disable"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"192.168.0.22:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := "/home/nexus/Desktop/my-project/pet/messenger/messenger/src/config/config.yaml"

	if configPath == "" {
		log.Fatalf("Invalid path %s", configPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file is not exist %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %s", err)
	}

	return &cfg
}

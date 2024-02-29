package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `env:"ENV" env_default:"local"`
	Port     string `env:"DB_PORT" env_default:"5432"`
	Host     string `env:"DB_HOST" env_default:"localhost"`
	Name     string `env:"DB_NAME" env_default:"postgres"`
	User     string `env:"DB_USER" env_default:"user"`
	Password string `env:"DB_PASSWORD" env_default:"password"`
}

type HTTPServer struct {
	Address     string        `env:"HTTP_SERVER_ADDRESS" env_default:"0.0.0.0:8082"`
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT" env_default: "4s"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env_default: "30s"`
}

func MustLoad() (*Config, *HTTPServer) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH env variable is not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error opening config file: %s", err)
	}

	var cfg Config
	var server HTTPServer
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error reading config file: %s", err)
	}
	err = cleanenv.ReadConfig(configPath, &server)
	if err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	return &cfg, &server
}

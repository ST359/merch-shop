package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DBPort      string `env:"DATABASE_PORT" env-default:"5432"`
	DBUser      string `env:"DATABASE_USER" env-default:"postgres"`
	DBPass      string `env:"DATABASE_PASSWORD" env-default:"password"`
	DBName      string `env:"DATABASE_NAME" env-default:"shop"`
	DBHost      string `env:"DATABASE_HOST" env-default:"db"`
	ServicePort string `env:"SERVER_PORT" env-default:"8080"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}
	return &cfg
}

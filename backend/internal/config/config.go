package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `env:"APP_ENV" env-default:"local"`
	FrontendURL string `env:"FRONTEND_URL" env-default:"http://localhost:5173"`
	Server      Server `yaml:"server"`
	DB          DB     `yaml:"db"`
	JWT         JWT    `yaml:"jwt"`
}

type Server struct {
	Port         int           `env:"SERVER_PORT" env-default:"8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" env-default:"10s"`
}

type DB struct {
	Host     string `env:"DB_HOST" env-required:"true"`
	Port     int    `env:"DB_PORT" env-default:"5432"`
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	DBName   string `env:"DB_NAME" env-required:"true"`
}

type JWT struct {
	Secret    string        `env:"JWT_SECRET" env-required:"true"`
	AccessTTL time.Duration `env:"JWT_ACCESS_TTL" env-default:"15m"`
}

func MustLoad() *Config {
	var cfg Config

	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var path string
	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}

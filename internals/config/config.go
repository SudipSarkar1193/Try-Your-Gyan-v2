package config

import (
	"flag"
	
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Addr string `yaml:"address" env:"HTTP_SERVER_ADDRESS" env-default:"localhost:5000"`
}

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"dev"`
	PsqlInfo   string `yaml:"postgresqlInfo" env:"PSQL_INFO"`
	APIKey     string `env:"API_KEY"`
	JWTSecret  string `env:"JWT_SECRET_KEY"`
	HTTPServer `yaml:"http_server"`
}

// Load configuration from environment variables or a YAML file
func MustLoad() *Config {
	// Attempt to load environment variables from a .env file (for local development)
	if err := LoadEnvFile(".env"); err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}

	// Attempt to load configuration from environment variables first
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err == nil && cfg.PsqlInfo != "" {
		log.Println("Loaded configuration from environment variables")
		return &cfg
	}

	// If not fully set, fall back to reading from the YAML file
	var configPath string
	configPath = os.Getenv("CONFIG_PATH")
	if configPath == "" {
		flags := flag.String("config", "config/config.yaml", "path to the config file")
		flag.Parse()
		configPath = *flags
	}

	// Check if Configuration File Exists
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		log.Fatalf("Config file doesn't exist: %s", configPath)
	}

	// Load configuration from the YAML file
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Can't read config file: %s", err.Error())
	}

	return &cfg
}

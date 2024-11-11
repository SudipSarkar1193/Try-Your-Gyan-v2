package config

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Addr string `yaml:"address" env:"HTTP_SERVER_ADDRESS" env-required:"true"`
}

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-required:"true" env-default:"production"`
	PsqlInfo   string `yaml:"postgresqlInfo" env-required:"true"`
	HTTPServer `yaml:"http_server"`
}

func MustLoad() *Config {

	if err := LoadEnvFile(".env"); err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}

	var configPath string

	configPath = os.Getenv("CONFIG_PATH")
	fmt.Println("configPath", configPath)

	if configPath == "" {
		flags := flag.String("config", "", "path to the config file")
		flag.Parse()

		configPath = *flags

		if configPath == "" {
			log.Fatal("Config path is not set")
		}
	}

	//Check if Configuration File Exists
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		log.Fatalf("config file doesn't exist : %s", configPath)
	}

	//Read and Parse the Configuration File
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Can't read config file: %s", err.Error())
	}

	return &cfg
}

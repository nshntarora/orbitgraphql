package config

import (
	"io"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Origin       string `toml:"origin"`
	Port         int    `toml:"port"`
	CacheBackend string `toml:"cache_backend"`
	Handlers     struct {
		GraphQLPath     string `toml:"graphql_path"`
		FlushAllPath    string `toml:"flush_all_path"`
		FlushByTypePath string `toml:"flush_by_type_path"`
		DebugPath       string `toml:"debug_path"`
		HealthPath      string `toml:"health_path"`
	} `toml:"handlers"`
}

const CONFIG_FILE = "./config.toml"

func NewConfig() *Config {

	// look for the config.toml file in the current directory
	// if it doesn't exist, use the default configuration
	// if it does exist, use the configuration from the file
	_, err := os.Stat(CONFIG_FILE)
	if os.IsNotExist(err) {
		log.Panic("config.toml file not found")
	}

	configFile, err := os.Open(CONFIG_FILE)
	if err != nil {
		log.Panic("error opening config.toml file")
	}
	defer configFile.Close()

	fileContent, err := io.ReadAll(configFile)
	if err != nil {
		log.Panic("error reading config.toml file")
	}

	var cfg Config
	_, err = toml.Decode(string(fileContent), &cfg)
	if err != nil {
		log.Panic(err)
	}

	if cfg.Handlers.GraphQLPath == "" {
		cfg.Handlers.GraphQLPath = "/graphql"
	}

	if cfg.Handlers.FlushAllPath == "" {
		cfg.Handlers.FlushAllPath = "/flush"
	}

	if cfg.Handlers.FlushByTypePath == "" {
		cfg.Handlers.FlushByTypePath = "/flush.type"
	}

	if cfg.Handlers.DebugPath == "" {
		cfg.Handlers.DebugPath = "/debug"
	}

	if cfg.Handlers.HealthPath == "" {
		cfg.Handlers.HealthPath = "/health"
	}

	return &Config{
		Origin:       cfg.Origin,
		Port:         cfg.Port,
		CacheBackend: cfg.CacheBackend,
		Handlers:     cfg.Handlers,
	}
}

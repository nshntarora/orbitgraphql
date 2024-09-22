package config

import (
	"io"
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Origin          string   `toml:"origin"`
	Port            int      `toml:"port"`
	CacheBackend    string   `toml:"cache_backend"`
	CacheHeaderName string   `toml:"cache_header_name"`
	ScopeHeaders    []string `toml:"scope_headers"`
	Handlers        struct {
		GraphQLPath     string `toml:"graphql_path"`
		FlushAllPath    string `toml:"flush_all_path"`
		FlushByTypePath string `toml:"flush_by_type_path"`
		DebugPath       string `toml:"debug_path"`
		HealthPath      string `toml:"health_path"`
	} `toml:"handlers"`
	Redis struct {
		Host string `toml:"host"`
		Port int    `toml:"port"`
	} `toml:"redis"`
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
	err = toml.Unmarshal(fileContent, &cfg)
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

	if cfg.CacheHeaderName == "" {
		cfg.CacheHeaderName = "x-orbit-cache"
	}

	return &cfg
}

package config

import (
	"io"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Origin          string   `toml:"origin" envconfig:"ORBIT_ORIGIN"`
	Port            int      `toml:"port" envconfig:"ORBIT_PORT"`
	CacheBackend    string   `toml:"cache_backend" envconfig:"ORBIT_CACHE_BACKEND"`
	CacheHeaderName string   `toml:"cache_header_name" envconfig:"ORBIT_CACHE_HEADER_NAME"`
	CacheTTL        int      `toml:"cache_ttl" envconfig:"ORBIT_CACHE_TTL"`
	ScopeHeaders    []string `toml:"scope_headers" envconfig:"ORBIT_SCOPE_HEADERS"`
	PrimaryKeyField string   `toml:"primary_key_field" envconfig:"ORBIT_PRIMARY_KEY_FIELD"`

	// Handlers configuration
	HandlersGraphQLPath     string `toml:"handlers_graphql_path" envconfig:"ORBIT_HANDLERS_GRAPHQL_PATH"`
	HandlersFlushAllPath    string `toml:"handlers_flush_all_path" envconfig:"ORBIT_HANDLERS_FLUSH_ALL_PATH"`
	HandlersFlushByTypePath string `toml:"handlers_flush_by_type_path" envconfig:"ORBIT_HANDLERS_FLUSH_BY_TYPE_PATH"`
	HandlersDebugPath       string `toml:"handlers_debug_path" envconfig:"ORBIT_HANDLERS_DEBUG_PATH"`
	HandlersHealthPath      string `toml:"handlers_health_path" envconfig:"ORBIT_HANDLERS_HEALTH_PATH"`

	// Redis configuration
	RedisHost string `toml:"redis_host" envconfig:"ORBIT_REDIS_HOST"`
	RedisPort int    `toml:"redis_port" envconfig:"ORBIT_REDIS_PORT"`

	// Logging configuration
	LogLevel  string `toml:"log_level" envconfig:"ORBIT_LOG_LEVEL"`
	LogFormat string `toml:"log_format" envconfig:"ORBIT_LOG_FORMAT"`
}

var CONFIG_FILE = "./config.toml"

func NewConfig() *Config {

	var cfg Config

	// first parse the configuration from the config.toml file
	ParseAndUpdateConfigFromTOML(&cfg)

	// then override the configuration from environment variables
	OverrideConfigFromEnv(&cfg)

	if cfg.HandlersGraphQLPath == "" {
		cfg.HandlersGraphQLPath = "/graphql"
	}

	if cfg.HandlersFlushAllPath == "" {
		cfg.HandlersFlushAllPath = "/flush"
	}

	if cfg.HandlersFlushByTypePath == "" {
		cfg.HandlersFlushByTypePath = "/flush.type"
	}

	if cfg.HandlersDebugPath == "" {
		cfg.HandlersDebugPath = "/debug"
	}

	if cfg.HandlersHealthPath == "" {
		cfg.HandlersHealthPath = "/health"
	}

	if cfg.CacheHeaderName == "" {
		cfg.CacheHeaderName = "x-orbit-cache"
	}

	if cfg.CacheTTL == 0 {
		// default cache TTL is 1 hour
		cfg.CacheTTL = 3600
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	if cfg.LogFormat == "" {
		cfg.LogFormat = "text"
	}

	return &cfg
}

func ParseAndUpdateConfigFromTOML(cfg *Config) {
	// look for the config.toml file in the current directory
	// if it doesn't exist, use the default configuration
	// if it does exist, use the configuration from the file
	_, err := os.Stat(CONFIG_FILE)
	if os.IsNotExist(err) {
		log.Print("config.toml file not found")
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
	err = toml.Unmarshal(fileContent, &cfg)
	if err != nil {
		log.Panic(err)
	}
}

func OverrideConfigFromEnv(cfg *Config) {
	err := envconfig.Process("ORBIT", cfg)
	if err != nil {
		log.Panic(err)
	}
}

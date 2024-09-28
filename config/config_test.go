package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testConfigFile = "./test_config.toml"

func createTestConfigFile(content string) {
	os.WriteFile(testConfigFile, []byte(content), 0644)
}

func removeTestConfigFile() {
	os.Remove(testConfigFile)
}

func TestNewConfigFileNotFound(t *testing.T) {
	// Ensure the test config file does not exist
	removeTestConfigFile()

	// Temporarily change the CONFIG_FILE constant
	originalConfigFile := CONFIG_FILE
	CONFIG_FILE = testConfigFile
	defer func() { CONFIG_FILE = originalConfigFile }()

	assert.Panics(t, func() {
		NewConfig()
	}, "Expected NewConfig to panic when config file is not found")
}

func TestNewConfigDefaultValues(t *testing.T) {
	configContent := `
        origin = "http://localhost"
        port = 8080
        cache_backend = "memory"
        primary_key_field = "id"
        redis_host = "localhost"
        redis_port = 6379
    `
	createTestConfigFile(configContent)
	defer removeTestConfigFile()

	// Temporarily change the CONFIG_FILE constant
	originalConfigFile := CONFIG_FILE
	CONFIG_FILE = testConfigFile
	defer func() { CONFIG_FILE = originalConfigFile }()

	cfg := NewConfig()

	assert.Equal(t, "/graphql", cfg.HandlersGraphQLPath)
	assert.Equal(t, "/flush", cfg.HandlersFlushAllPath)
	assert.Equal(t, "/flush.type", cfg.HandlersFlushByTypePath)
	assert.Equal(t, "/debug", cfg.HandlersDebugPath)
	assert.Equal(t, "/health", cfg.HandlersHealthPath)
	assert.Equal(t, "x-orbit-cache", cfg.CacheHeaderName)
}

func TestNewConfigValidFile(t *testing.T) {
	configContent := `
        origin = "http://localhost"
        port = 8080
        cache_backend = "memory"
        cache_header_name = "x-custom-cache"
        primary_key_field = "id"
        handlers_graphql_path = "/custom_graphql"
        handlers_flush_all_path = "/custom_flush"
        handlers_flush_by_type_path = "/custom_flush.type"
        handlers_debug_path = "/custom_debug"
        handlers_health_path = "/custom_health"
        redis_host = "localhost"
        redis_port = 6379
    `
	createTestConfigFile(configContent)
	defer removeTestConfigFile()

	// Temporarily change the CONFIG_FILE constant
	originalConfigFile := CONFIG_FILE
	CONFIG_FILE = testConfigFile
	defer func() { CONFIG_FILE = originalConfigFile }()

	cfg := NewConfig()

	assert.Equal(t, "http://localhost", cfg.Origin)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "memory", cfg.CacheBackend)
	assert.Equal(t, "x-custom-cache", cfg.CacheHeaderName)
	assert.Equal(t, "id", cfg.PrimaryKeyField)
	assert.Equal(t, "/custom_graphql", cfg.HandlersGraphQLPath)
	assert.Equal(t, "/custom_flush", cfg.HandlersFlushAllPath)
	assert.Equal(t, "/custom_flush.type", cfg.HandlersFlushByTypePath)
	assert.Equal(t, "/custom_debug", cfg.HandlersDebugPath)
	assert.Equal(t, "/custom_health", cfg.HandlersHealthPath)
	assert.Equal(t, "localhost", cfg.RedisHost)
	assert.Equal(t, 6379, cfg.RedisPort)
}

func TestNewConfigCacheBackendFromEnv(t *testing.T) {
	configContent := `
        origin = "http://localhost"
        port = 8080
        cache_backend = "memory"
        primary_key_field = "id"
        redis_host = "localhost"
        redis_port = 6379
    `
	createTestConfigFile(configContent)
	defer removeTestConfigFile()

	// Temporarily change the CONFIG_FILE constant
	originalConfigFile := CONFIG_FILE
	CONFIG_FILE = testConfigFile
	defer func() { CONFIG_FILE = originalConfigFile }()

	// Set environment variable for CacheBackend
	os.Setenv("ORBIT_CACHE_BACKEND", "redis")
	defer os.Unsetenv("ORBIT_CACHE_BACKEND")

	cfg := NewConfig()

	assert.Equal(t, "redis", cfg.CacheBackend, "Expected CacheBackend to be 'redis' from environment variable")
}

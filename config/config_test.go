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
        [redis]
        host = "localhost"
        port = 6379
    `
	createTestConfigFile(configContent)
	defer removeTestConfigFile()

	// Temporarily change the CONFIG_FILE constant
	originalConfigFile := CONFIG_FILE
	CONFIG_FILE = testConfigFile
	defer func() { CONFIG_FILE = originalConfigFile }()

	cfg := NewConfig()

	assert.Equal(t, "/graphql", cfg.Handlers.GraphQLPath)
	assert.Equal(t, "/flush", cfg.Handlers.FlushAllPath)
	assert.Equal(t, "/flush.type", cfg.Handlers.FlushByTypePath)
	assert.Equal(t, "/debug", cfg.Handlers.DebugPath)
	assert.Equal(t, "/health", cfg.Handlers.HealthPath)
	assert.Equal(t, "x-orbit-cache", cfg.CacheHeaderName)
}

func TestNewConfigValidFile(t *testing.T) {
	configContent := `
        origin = "http://localhost"
        port = 8080
        cache_backend = "memory"
        cache_header_name = "x-custom-cache"
        primary_key_field = "id"
        [handlers]
        graphql_path = "/custom_graphql"
        flush_all_path = "/custom_flush"
        flush_by_type_path = "/custom_flush.type"
        debug_path = "/custom_debug"
        health_path = "/custom_health"
        [redis]
        host = "localhost"
        port = 6379
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
	assert.Equal(t, "/custom_graphql", cfg.Handlers.GraphQLPath)
	assert.Equal(t, "/custom_flush", cfg.Handlers.FlushAllPath)
	assert.Equal(t, "/custom_flush.type", cfg.Handlers.FlushByTypePath)
	assert.Equal(t, "/custom_debug", cfg.Handlers.DebugPath)
	assert.Equal(t, "/custom_health", cfg.Handlers.HealthPath)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
}

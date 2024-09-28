package handlers

import (
	"orbitgraphql/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCacheOptions(t *testing.T) {
	// Mock configuration
	cfg := &config.Config{
		CacheBackend:    "inmemory",
		PrimaryKeyField: "id",
	}

	// Test values
	values := []interface{}{"value1", "value2"}

	// Call the function
	options := GetCacheOptions(cfg, values)

	// Assertions
	assert.NotNil(t, options)
	assert.NotNil(t, options.QueryStore)
	assert.NotNil(t, options.ObjectStore)
	assert.Equal(t, "id", options.IDField)

	// Check if QueryStore and ObjectStore are initialized
	assert.NotNil(t, QueryStore)
	assert.NotNil(t, ObjectStore)

	// Check if the Prefix is correctly generated
	expectedPrefix := "dmFsdWUxOjp2YWx1ZTI="
	assert.Equal(t, expectedPrefix, options.Prefix)
}

func TestGetCacheOptionsWithRedis(t *testing.T) {
	// Mock configuration for Redis
	cfg := &config.Config{
		CacheBackend:    "redis",
		PrimaryKeyField: "id",
		RedisHost:       "localhost",
		RedisPort:       6379,
	}

	// Test values
	values := []interface{}{"value1", "value2"}

	// Call the function
	options := GetCacheOptions(cfg, values)

	// Assertions
	assert.NotNil(t, options)
	assert.NotNil(t, options.QueryStore)
	assert.NotNil(t, options.ObjectStore)
	assert.Equal(t, "id", options.IDField)

	// Check if QueryStore and ObjectStore are initialized
	assert.NotNil(t, QueryStore)
	assert.NotNil(t, ObjectStore)

	// Check if the Prefix is correctly generated
	expectedPrefix := "dmFsdWUxOjp2YWx1ZTI="
	assert.Equal(t, expectedPrefix, options.Prefix)
}

func TestGetCacheOptionsWithEmptyValues(t *testing.T) {
	// Mock configuration
	cfg := &config.Config{
		CacheBackend:    "inmemory",
		PrimaryKeyField: "id",
	}

	// Test empty values
	values := []interface{}{}

	// Call the function
	options := GetCacheOptions(cfg, values)

	// Assertions
	assert.NotNil(t, options)
	assert.NotNil(t, options.QueryStore)
	assert.NotNil(t, options.ObjectStore)
	assert.Equal(t, "id", options.IDField)

	// Check if QueryStore and ObjectStore are initialized
	assert.NotNil(t, QueryStore)
	assert.NotNil(t, ObjectStore)

	// Check if the Prefix is correctly generated
	expectedPrefix := ""
	assert.Equal(t, expectedPrefix, options.Prefix)
}

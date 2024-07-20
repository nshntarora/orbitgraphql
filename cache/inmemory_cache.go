package cache

import (
	"encoding/json"
	"graphql_cache/utils/file_utils"
)

type InMemoryCache struct {
	cache map[string]interface{}
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		cache: make(map[string]interface{}),
	}
}

func (c *InMemoryCache) Set(key string, value interface{}) error {
	c.cache[key] = value
	return nil
}

func (c *InMemoryCache) Get(key string) (interface{}, error) {
	return c.cache[key], nil
}

func (c *InMemoryCache) Del(key string) error {
	delete(c.cache, key)
	return nil
}

func (c *InMemoryCache) Exists(key string) (bool, error) {
	_, exists := c.cache[key]
	return exists, nil
}

func (c *InMemoryCache) Map() (map[string]interface{}, error) {
	return c.cache, nil
}

func (c *InMemoryCache) JSON() ([]byte, error) {
	return json.Marshal(c.cache)
}

func (c *InMemoryCache) Debug(identifier string) error {
	f := file_utils.NewFile(identifier + ".cache.json")
	defer f.Close()
	jsonContent, _ := c.JSON()
	f.Write(string(jsonContent))
	return nil
}

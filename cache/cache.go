package cache

import "encoding/json"

type Cache interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Del(key string) error
	Exists(key string) (bool, error)
	Map() (map[string]interface{}, error)
	JSON() ([]byte, error)
}

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

package cache

import (
	"encoding/json"
	"errors"
	"graphql_cache/utils/file_utils"
	"regexp"
	"strings"
	"sync"
	"time"
)

type InMemoryCache struct {
	data       map[string]interface{}
	expiration map[string]*time.Time
	ttl        int
	mu         sync.Mutex
}

func NewInMemoryCache(ttl int) *InMemoryCache {
	cache := &InMemoryCache{
		mu:         sync.Mutex{},
		data:       make(map[string]interface{}),
		expiration: make(map[string]*time.Time),
		ttl:        ttl,
	}
	go cache.cleanup()
	return cache
}

func (c *InMemoryCache) Key(key string) string {
	return key
}

func (c *InMemoryCache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[c.Key(key)] = deepCopy(value)
	if value == nil {
		c.expiration[c.Key(key)] = nil
	} else {
		t := time.Now().Add(time.Duration(c.ttl) * time.Second)
		c.expiration[c.Key(key)] = &t
	}
	return nil
}

func (c *InMemoryCache) Get(key string) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expiration, exists := c.expiration[c.Key(key)]
	if !exists || expiration == nil || time.Now().After(*expiration) {
		return nil, errors.New("key not found")
	}
	value, exists := c.data[c.Key(key)]
	if !exists {
		return nil, errors.New("key not found")
	}
	return deepCopy(value), nil
}

func (c *InMemoryCache) Del(key string) error {
	c.Set(c.Key(key), nil)
	return nil
}

func (c *InMemoryCache) Exists(key string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expiration, exists := c.expiration[c.Key(key)]
	if !exists || expiration == nil || time.Now().After(*expiration) {
		return false, nil
	}
	return true, nil
}

func (c *InMemoryCache) Map() (map[string]interface{}, error) {
	copy := make(map[string]interface{})
	now := time.Now()
	for k, v := range c.data {
		if expiration, exists := c.expiration[k]; exists && expiration != nil && now.Before(*expiration) {
			copy[k] = v
		}
	}
	return copy, nil
}

func (c *InMemoryCache) JSON() ([]byte, error) {
	copy, err := c.Map()
	if err != nil {
		return nil, err
	}
	return json.Marshal(copy)
}

func (c *InMemoryCache) Debug(identifier string) error {
	f := file_utils.NewFile("../" + identifier + ".cache.json")
	defer f.Close()
	jsonContent, _ := c.JSON()
	f.Write(string(jsonContent))
	return nil
}

func (c *InMemoryCache) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]interface{})
	c.expiration = make(map[string]*time.Time)
	return nil
}

func (c *InMemoryCache) DeleteByPrefix(prefix string) error {
	var re = regexp.MustCompile(`(?m)` + strings.ReplaceAll(c.Key(prefix), "*", ".*"))
	for k := range c.data {
		if re.Match([]byte(k)) {
			c.Del(k)
		}
	}
	return nil
}

func (c *InMemoryCache) cleanup() {
	for {
		time.Sleep(time.Duration(c.ttl) * time.Second)
		c.mu.Lock()
		now := time.Now()
		for key, expiration := range c.expiration {
			if expiration != nil && now.After(*expiration) {
				c.Del(key)
			}
		}
		c.mu.Unlock()
	}
}

func deepCopy(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for k, v := range val {
			newMap[k] = deepCopy(v)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(val))
		for i, v := range val {
			newSlice[i] = deepCopy(v)
		}
		return newSlice
	default:
		return v
	}
}

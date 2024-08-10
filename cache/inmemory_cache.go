package cache

import (
	"encoding/json"
	"errors"
	"graphql_cache/utils/file_utils"
	"regexp"
	"strings"
)

type InMemoryCache struct {
	data map[string]interface{}
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		data: make(map[string]interface{}),
	}
}

func (c *InMemoryCache) Set(key string, value interface{}) error {
	c.data[key] = deepCopy(value)
	return nil
}

func (c *InMemoryCache) Get(key string) (interface{}, error) {
	value, exists := c.data[key]
	if !exists {
		return nil, errors.New("key not found")
	}
	return deepCopy(value), nil
}

func (c *InMemoryCache) Del(key string) error {
	c.Set(key, nil)
	return nil
}

func (c *InMemoryCache) Exists(key string) (bool, error) {
	_, exists := c.data[key]
	return exists, nil
}

func (c *InMemoryCache) Map() (map[string]interface{}, error) {
	copy := make(map[string]interface{})
	for k, v := range c.data {
		copy[k] = v
	}
	return copy, nil
}

func (c *InMemoryCache) JSON() ([]byte, error) {
	return json.Marshal(c.data)
}

func (c *InMemoryCache) Debug(identifier string) error {
	f := file_utils.NewFile("../" + identifier + ".cache.json")
	defer f.Close()
	jsonContent, _ := c.JSON()
	f.Write(string(jsonContent))
	return nil
}

func (c *InMemoryCache) Flush() error {
	c.data = make(map[string]interface{})
	return nil
}

func (c *InMemoryCache) DeleteByPrefix(prefix string) error {
	var re = regexp.MustCompile(`(?m)` + strings.ReplaceAll(prefix, "*", ".*"))

	for k := range c.data {
		// regex match the prefix to the key
		// if the key is gql:* then delete all keys which start with gql
		if re.Match([]byte(k)) {
			c.Del(k)
		}
	}
	return nil
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

package cache

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// RedisCache implements the Cache interface and uses Redis as the cache store
type RedisCache struct {
	cache *redis.Client
}

func NewRedisCache() Cache {
	c := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &RedisCache{
		cache: c,
	}
}

func (c *RedisCache) Set(key string, value interface{}) error {
	valueType := reflect.TypeOf(value)
	switch valueType.Kind() {
	case reflect.Map:
		br, _ := json.Marshal(value)
		c.cache.Set(ctx, key, string(br), 0)
		c.cache.Set(ctx, key+"_type", "reflect.Map", 0)
	case reflect.Slice:
		br, _ := json.Marshal(value)
		c.cache.Set(ctx, key, string(br), 0)
		c.cache.Set(ctx, key+"_type", "reflect.Slice", 0)
	default:
		c.cache.Set(ctx, key, value, 0)
	}
	return nil
}

func (c *RedisCache) Get(key string) (interface{}, error) {
	typeValue, _ := c.cache.Get(ctx, key+"_type").Result()
	val, err := c.cache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	switch typeValue {
	case "reflect.Map":
		var m map[string]interface{}
		json.Unmarshal([]byte(val), &m)
		return m, nil
	case "reflect.Slice":
		var s []interface{}
		json.Unmarshal([]byte(val), &s)
		return s, nil
	}

	return val, nil
}

func (c *RedisCache) Del(key string) error {
	c.cache.Del(ctx, key)
	return nil
}

func (c *RedisCache) Exists(key string) (bool, error) {
	val, err := c.cache.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val == 1, nil
}

func (c *RedisCache) Map() (map[string]interface{}, error) {
	return nil, nil
}

func (c *RedisCache) JSON() ([]byte, error) {
	return nil, nil
}

func (c *RedisCache) Debug(identifier string) error {
	return nil
}

func (c *RedisCache) Flush() error {
	c.cache.FlushAll(ctx)
	return nil
}

func (c *RedisCache) DeleteByPrefix(prefix string) error {
	allKeys := c.cache.Keys(ctx, prefix+"*")
	if allKeys == nil {
		return nil
	}

	for _, key := range allKeys.Val() {
		c.cache.Del(ctx, key)
	}

	return nil
}

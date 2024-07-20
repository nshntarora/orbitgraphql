package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisCache struct {
	cache *redis.Client
}

// var client *redis.Client

// func init() {
// 	client = redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "", // no password set
// 		DB:       1,  // use default DB
// 	})
// }

func NewRedisCache() *RedisCache {
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
	c.cache.Set(ctx, key, value, 0)
	return nil
}

func (c *RedisCache) Get(key string) (interface{}, error) {
	val, err := c.cache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
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

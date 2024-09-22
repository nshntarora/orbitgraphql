package handlers

import (
	"encoding/base64"
	"fmt"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"net/http"
	"strings"
)

func GetHandlers(cfg *config.Config) *http.ServeMux {
	api := http.NewServeMux()
	api.Handle(cfg.Handlers.DebugPath, GetDebugHandler(cfg))
	api.Handle(cfg.Handlers.FlushAllPath, GetFlushCacheHandler(cfg))
	api.Handle(cfg.Handlers.FlushByTypePath, GetFlushCacheByTypeHandler(cfg))
	api.Handle(cfg.Handlers.GraphQLPath, GetCacheHandler(cfg))
	return api
}

func GetCacheOptions(cfg *config.Config, values []interface{}) *graphcache.GraphCacheOptions {
	valueStr := make([]string, 0)
	for _, val := range values {
		valueStr = append(valueStr, fmt.Sprintf("%v", val))
	}
	valueHash := base64.StdEncoding.EncodeToString([]byte(strings.Join(valueStr, "::")))

	return &graphcache.GraphCacheOptions{
		Backend:   graphcache.CacheBackend(cfg.CacheBackend),
		RedisHost: cfg.Redis.Host,
		RedisPort: cfg.Redis.Port,
		Prefix:    valueHash,
	}
}

func GetScopeValues(cfg *config.Config, r *http.Request) []interface{} {
	values := make([]interface{}, 0)
	for _, header := range cfg.ScopeHeaders {
		if header != "" {
			values = append(values, r.Header.Get(header))
		}
	}
	return values
}

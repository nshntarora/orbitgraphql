package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"graphql_cache/cache"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var QueryStore *cache.Cache
var ObjectStore *cache.Cache

func GetHandlers(cfg *config.Config) *http.ServeMux {
	api := http.NewServeMux()
	api.Handle(cfg.HandlersDebugPath, GetDebugHandler(cfg))
	api.Handle(cfg.HandlersFlushAllPath, GetFlushCacheHandler(cfg))
	api.Handle(cfg.HandlersFlushByTypePath, GetFlushCacheByTypeHandler(cfg))
	api.Handle(cfg.HandlersGraphQLPath, GetCacheHandler(cfg))
	return api
}

func GetNewCacheStore(cfg *config.Config) cache.Cache {
	if cfg.CacheBackend == "redis" {
		cache.NewRedisCache(cfg.RedisHost, strconv.Itoa(cfg.RedisPort), cfg.CacheTTL)
	}
	return cache.NewInMemoryCache(cfg.CacheTTL)
}

func GetCacheOptions(cfg *config.Config, values []interface{}) *graphcache.GraphCacheOptions {
	if QueryStore == nil {
		qs := GetNewCacheStore(cfg)
		QueryStore = &qs
	}
	if ObjectStore == nil {
		os := GetNewCacheStore(cfg)
		ObjectStore = &os
	}

	valueStr := make([]string, 0)
	for _, val := range values {
		valueStr = append(valueStr, fmt.Sprintf("%v", val))
	}
	valueHash := base64.StdEncoding.EncodeToString([]byte(strings.Join(valueStr, "::")))

	return &graphcache.GraphCacheOptions{
		QueryStore:  *QueryStore,
		ObjectStore: *ObjectStore,
		Prefix:      valueHash,
		IDField:     cfg.PrimaryKeyField,
	}
}

func GetScopeValues(cfg *config.Config, r *http.Request) []interface{} {
	values := make([]interface{}, 0)
	splittedHeaderNames := strings.Split(cfg.ScopeHeaders, ",")
	headerNames := make([]string, 0)
	for _, header := range splittedHeaderNames {
		headerNames = append(headerNames, strings.TrimSpace(header))
	}
	for _, header := range headerNames {
		if header != "" {
			values = append(values, r.Header.Get(header))
		}
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		return values
	}
	r.Body = io.NopCloser(bytes.NewReader(requestBody))

	request := graphcache.GraphQLRequest{}
	request.FromBytes(requestBody)

	values = append(values, request.Query, request.OperationName, request.Variables)
	return values
}

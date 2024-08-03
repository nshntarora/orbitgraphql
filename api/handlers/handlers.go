package handlers

import (
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"net/http"
)

func GetHandlers(cache *graphcache.GraphCache, cfg *config.Config) *http.ServeMux {
	api := http.NewServeMux()
	api.Handle(cfg.Handlers.DebugPath, GetDebugHandler(cache, cfg))
	api.Handle(cfg.Handlers.FlushAllPath, GetFlushCacheHandler(cache, cfg))
	api.Handle(cfg.Handlers.FlushByTypePath, GetFlushCacheByTypeHandler(cache, cfg))
	api.Handle(cfg.Handlers.GraphQLPath, GetCacheHandler(cache, cfg))
	return api
}

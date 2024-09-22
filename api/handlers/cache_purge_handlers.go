package handlers

import (
	"encoding/json"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"net/http"
)

type FlushCacheByTypeRequest struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func GetFlushCacheHandler(cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cache := graphcache.NewGraphCacheWithOptions(GetCacheOptions(cfg, GetScopeValues(cfg, r)))
		cache.Flush()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

func GetFlushCacheByTypeHandler(cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cache := graphcache.NewGraphCacheWithOptions(GetCacheOptions(cfg, GetScopeValues(cfg, r)))
		flushByTypeRequest := FlushCacheByTypeRequest{}
		err := json.NewDecoder(r.Body).Decode(&flushByTypeRequest)
		if err != nil {
			http.Error(w, "error decoding request", http.StatusBadRequest)
		}
		cache.FlushByType(flushByTypeRequest.Type, flushByTypeRequest.ID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

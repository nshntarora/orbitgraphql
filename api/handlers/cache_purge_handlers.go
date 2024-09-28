package handlers

import (
	"context"
	"encoding/json"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"io"
	"net/http"
)

type FlushCacheByTypeRequest struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func GetFlushCacheHandler(cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		cache := graphcache.NewGraphCacheWithOptions(ctx, GetCacheOptions(cfg, GetScopeValues(cfg, r)))
		cache.Flush()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

func GetFlushCacheByTypeHandler(cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		cache := graphcache.NewGraphCacheWithOptions(ctx, GetCacheOptions(cfg, GetScopeValues(cfg, r)))
		flushByTypeRequest := FlushCacheByTypeRequest{}
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading request", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(bytes, &flushByTypeRequest)
		if err != nil {
			http.Error(w, "error decoding request", http.StatusBadRequest)
			return
		}

		cache.FlushByType(flushByTypeRequest.Type, flushByTypeRequest.ID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

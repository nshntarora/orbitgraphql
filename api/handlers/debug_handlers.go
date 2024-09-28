package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"orbitgraphql/config"
	"orbitgraphql/graphcache"
)

func GetDebugHandler(cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		cache := graphcache.NewGraphCacheWithOptions(ctx, GetCacheOptions(cfg, GetScopeValues(cfg, r)))
		resp := cache.Look()
		br, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "error marshalling response", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(br)
	})
}

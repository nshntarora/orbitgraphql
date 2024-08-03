package handlers

import (
	"encoding/json"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"net/http"
)

func GetFlushCacheHandler(Cache *graphcache.GraphCache, cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Cache.Flush()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

func GetFlushCacheByTypeHandler(Cache *graphcache.GraphCache, cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flushByTypeRequest := struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&flushByTypeRequest)
		if err != nil {
			http.Error(w, "error decoding request", http.StatusBadRequest)
		}
		Cache.FlushByType(flushByTypeRequest.Type, flushByTypeRequest.ID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

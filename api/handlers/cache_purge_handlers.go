package handlers

import (
	"encoding/json"
	"fmt"
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
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading request", http.StatusBadRequest)
			return
		}
		fmt.Println("flush by type body - ", string(bytes))
		err = json.Unmarshal(bytes, &flushByTypeRequest)
		if err != nil {
			http.Error(w, "error decoding request", http.StatusBadRequest)
			return
		}

		// err := json.NewDecoder(r.Body).Decode(&flushByTypeRequest)
		// if err != nil {
		// 	http.Error(w, "error decoding request", http.StatusBadRequest)
		// 	return
		// }
		cache.FlushByType(flushByTypeRequest.Type, flushByTypeRequest.ID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

package handlers

import (
	"encoding/json"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"net/http"
)

func GetDebugHandler(Cache *graphcache.GraphCache, cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Cache.Look()
		br, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "error marshalling response", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(br)
	})
}

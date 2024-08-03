package handlers

import (
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"net/http"
)

func GetDebugHandler(Cache *graphcache.GraphCache, cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Cache.Debug()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

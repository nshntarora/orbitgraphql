package api

import (
	"graphql_cache/api/handlers"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"net/http"
	"strconv"
)

type Server struct {
	httpServer *http.Server
	cache      *graphcache.GraphCache
	cfg        *config.Config
}

func NewGraphCache(opts *graphcache.GraphCacheOptions) *graphcache.GraphCache {
	return graphcache.NewGraphCacheWithOptions(opts)
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Port),
			Handler: handlers.GetHandlers(cfg),
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

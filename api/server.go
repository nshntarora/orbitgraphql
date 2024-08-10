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

func NewGraphCache(cfg *config.Config) *graphcache.GraphCache {
	return graphcache.NewGraphCache(cfg)
}

func NewServer(cfg *config.Config) *Server {
	cache := NewGraphCache(cfg)
	return &Server{
		cache: cache,
		cfg:   cfg,
		httpServer: &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Port),
			Handler: handlers.GetHandlers(cache, cfg),
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

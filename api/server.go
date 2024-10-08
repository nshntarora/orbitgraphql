package api

import (
	"net/http"
	"orbitgraphql/api/handlers"
	"orbitgraphql/config"
	"strconv"
)

type Server struct {
	httpServer *http.Server
	cfg        *config.Config
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

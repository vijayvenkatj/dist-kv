package api

import (
	"errors"
	"log"
	"net/http"
	"time"
)

type Server struct {
	HttpServer *http.Server
}

type Config struct {
	Address     string
	GrpcAddress string
	ElectionT   time.Duration

	NodeID    uint32
	ShardList map[uint32][]Location

	Path string
}

func NewServer(config Config) *Server {

	shardManager := NewShardManager(config.ShardList, config.Address, config)
	handler := NewHandler(shardManager, config)
	router := NewRouter(handler)

	return &Server{
		HttpServer: &http.Server{
			Addr:    config.Address,
			Handler: corsMiddleware(router),
		},
	}
}

func (server *Server) ListenAndServe() error {
	if err := server.HttpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %s\n", err)
		return err
	}
	return nil
}

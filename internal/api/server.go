package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/vijayvenkatj/kv-store/internal/server/store"
)

type Server struct {
	HttpServer *http.Server
}

func NewServer(addr, path string) *Server {

	storeInstance := store.New(path)
	handler := NewHandler(storeInstance)
	router := NewRouter(handler)

	return &Server{
		HttpServer: &http.Server{
			Addr:    addr,
			Handler: corsMiddleware(router),
		},
	}
}

func (server *Server) ListenAndServe() {
	if err := server.HttpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %s\n", err)
	}
}

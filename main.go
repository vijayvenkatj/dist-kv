package main

import (
	"github.com/vijayvenkatj/kv-store/internal/api"
)

func main() {

	server := api.NewServer(":8080", "tmp")
	server.ListenAndServe()

	return
}

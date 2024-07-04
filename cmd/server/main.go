package main

import (
	"github.com/Tarunshrma/prolog/internal/server"
)

func main() {
	// This is the main entry point for the application
	// It should start the HTTP server and listen for requests

	srv := server.NewHttpServer(":8080")
	srv.ListenAndServe()
}

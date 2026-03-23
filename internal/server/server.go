package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// StartServer запускает HTTP сервер на указанном порту с заданным handler
func StartServer(handler http.Handler) *http.Server {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9065"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: handler,
	}

	go func() {
		log.Printf("Starting service on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	return srv
}
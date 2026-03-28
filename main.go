package main

import (
	"context"
	"time"

	"github.com/vitalii-q/selena-users-service/internal/bootstrap"
	"github.com/vitalii-q/selena-users-service/internal/router"
	"github.com/vitalii-q/selena-users-service/internal/server"
	//"github.com/vitalii-q/selena-users-service/internal/logger"
)

func main() {
	// --- Context with cancel for graceful shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Logger setup ---
	//logger.Setup()

	// --- Bootstrap all dependencies ---
	deps := bootstrap.NewBootstrap(ctx)
	defer deps.DB.Close()

	// --- Router setup ---
	r := router.SetupRouter(deps.DB, deps.UserHandler, deps.AuthHandler, deps.UserHotelsHandler, deps.LocationsHandler)

	// --- HTTP server ---
	srv := server.StartServer(r)

	// graceful shutdown
	server.GracefulShutdown(srv, cancel, 5*time.Second)
}
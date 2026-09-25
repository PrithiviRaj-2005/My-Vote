package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pulsevote/config"
	"pulsevote/controllers"
	"pulsevote/repository"
	"pulsevote/routes"
	"pulsevote/services"
	ws "pulsevote/websocket"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables from .env file if available
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found, using system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("==================================================")
	log.Println("   PULSEVOTE – Live Polling Tool Backend Server   ")
	log.Println("==================================================")

	// 2. Connect to MongoDB
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatalf("[MongoDB Error] Could not connect to MongoDB: %v\nPlease ensure MongoDB is running or check your MONGO_URI in .env", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.Client.Disconnect(ctx)
		log.Println("[MongoDB] Disconnected cleanly.")
	}()

	// 3. Connect to Redis
	rdb, err := config.ConnectRedis()
	if err != nil {
		log.Fatalf("[Redis Error] Could not connect to Redis: %v\nPlease ensure Redis is running or check your REDIS_URL in .env", err)
	}
	defer func() {
		_ = rdb.Close()
		log.Println("[Redis] Disconnected cleanly.")
	}()

	// 4. Initialize Gorilla WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()
	log.Println("[WebSocket] Gorilla WebSocket Hub started.")

	// 5. Initialize Realtime Service (Redis Live Counts + Pub/Sub + WebSocket)
	realtimeService := services.NewRealtimeService(rdb, hub)

	// 6. Start Redis Pub/Sub Background Subscriber
	subCtx, cancelSubscriber := context.WithCancel(context.Background())
	defer cancelSubscriber()
	realtimeService.StartSubscriber(subCtx)

	// 7. Initialize Repositories
	userRepo := repository.NewUserRepository(db.Users)
	pollRepo := repository.NewPollRepository(db.Polls, db.Votes)

	// 8. Initialize Services
	authService := services.NewAuthService(userRepo)
	pollService := services.NewPollService(pollRepo, realtimeService)

	// 9. Initialize Controllers
	authCtrl := controllers.NewAuthController(authService)
	pollCtrl := controllers.NewPollController(pollService)
	voteCtrl := controllers.NewVoteController(pollService)

	// 10. Setup Gin Router
	router := routes.SetupRouter(authCtrl, pollCtrl, voteCtrl, hub)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 11. Start HTTP Server in a goroutine
	go func() {
		log.Printf("[Server] PulseVote backend is running on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Server Error] Failed to listen and serve: %v", err)
		}
	}()

	// 12. Graceful Shutdown listener
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Server] Shutting down PulseVote server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[Server] Forced shutdown error: %v", err)
	}

	fmt.Println("[Server] PulseVote server exited cleanly.")
}

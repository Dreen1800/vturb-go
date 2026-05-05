package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"vturb-go/internal/config"
	"vturb-go/internal/platform"
	"vturb-go/internal/repository"
	"vturb-go/internal/service"
	"vturb-go/internal/worker"
)

func main() {
	cfg := config.Load()

	fmt.Printf("🔴 Redis Addr: %s\n", cfg.Redis.Addr)
	fmt.Printf("📦 Database: %s\n", cfg.Database.URL)

	// Connect to database
	db, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// Initialize services
	videoRepo := repository.NewVideoRepository(db)
	videoSvc := service.NewVideoService(videoRepo)
	cfClient := platform.NewCloudflareClient(cfg.Cloudflare.AccountID, cfg.Cloudflare.APIToken)

	// Create Asynq server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	// Register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeSyncVideoStatus, worker.HandleSyncVideoStatusTask(videoSvc, cfClient))

	fmt.Println("🔄 Worker started - Waiting for jobs...")
	fmt.Println("   Press Ctrl+C to stop")

	// Start server in a goroutine
	go func() {
		if err := srv.Run(mux); err != nil {
			log.Printf("Worker error: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down worker...")
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	srv.Stop()
	
	<-shutdownCtx.Done()
	fmt.Println("👋 Worker stopped")
}

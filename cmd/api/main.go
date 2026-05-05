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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"vturb-go/internal/config"
	"vturb-go/internal/handler"
	"vturb-go/internal/platform"
	"vturb-go/internal/repository"
	"vturb-go/internal/service"
)

func main() {
	cfg := config.Load()

	// Connect to database
	db, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		log.Printf("⚠️  Failed to connect to database: %v", err)
		log.Println("🔄 Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
		db, err = pgxpool.New(context.Background(), cfg.Database.URL)
		if err != nil {
			log.Fatalf("❌ Failed to connect to database after retry: %v", err)
		}
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(context.Background()); err != nil {
		log.Printf("⚠️  Failed to ping database: %v", err)
		log.Println("🔄 Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
		if err := db.Ping(context.Background()); err != nil {
			log.Fatalf("❌ Failed to ping database after retry: %v", err)
		}
	}
	log.Println("✅ Connected to PostgreSQL")

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Printf("⚠️  Failed to run migrations: %v", err)
	} else {
		log.Println("✅ Migrations completed")
	}

	// Initialize repositories and services
	videoRepo := repository.NewVideoRepository(db)
	videoSvc := service.NewVideoService(videoRepo)
	cfClient := platform.NewCloudflareClient(cfg.Cloudflare.AccountID, cfg.Cloudflare.APIToken)

	// Asynq client for background jobs
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Redis.Addr})
	defer asynqClient.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	


	// Healthcheck
	r.Get("/health", handler.HealthCheck(cfg))
	r.Get("/healthz", handler.HealthCheck(cfg))

	// API routes
	r.Route("/api", func(api chi.Router) {
		api.Route("/videos", func(v chi.Router) {
			v.Post("/", handler.CreateVideoHandler(videoSvc))
			v.Get("/", handler.ListVideosHandler(videoSvc))
			v.Options("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.WriteHeader(http.StatusOK)
			})
			v.Route("/{id}", func(vid chi.Router) {
				vid.Get("/", handler.GetVideoHandler(videoSvc))
				vid.Post("/upload", handler.GenerateUploadURLHandler(videoSvc, cfClient))
				vid.Post("/proxy-upload", handler.NewProxyUploadHandler(videoSvc, cfClient).HandleUpload)
				vid.Patch("/finalize", handler.FinalizeUploadHandler(videoSvc))
				vid.Options("/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					w.WriteHeader(http.StatusOK)
				})
			})
		})
		api.Post("/webhooks/cloudflare", handler.CloudflareWebhookHandler(videoSvc, asynqClient))
		api.Post("/embed/resolve", handler.ResolveEmbedHandler(videoSvc))
		api.Options("/embed/resolve", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
		})
	})

	r.Get("/embed.js", handler.ServeEmbedJS(cfg))

	// Serve static files (frontend)
	fs := http.FileServer(http.Dir("./web"))
	r.Handle("/web/*", http.StripPrefix("/web/", fs))

	// Serve specific files at root
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})
	r.Get("/embed-test.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/embed-test.html")
	})

	// Wrap router with CORS handler to intercept OPTIONS before Chi
	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("🌐 CORS Handler: %s %s", req.Method, req.URL.Path)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Filename")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
		if req.Method == "OPTIONS" {
			log.Printf("✅ Intercepted OPTIONS request")
			w.WriteHeader(http.StatusOK)
			return
		}
		r.ServeHTTP(w, req)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: corsHandler,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	fmt.Printf("🚀 Server running on http://localhost:%s\n", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("👋 Server stopped gracefully")
}

func runMigrations(db *pgxpool.Pool) error {
	migrationSQL := `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    filename VARCHAR(255),
    size_bytes BIGINT,
    status VARCHAR(50) DEFAULT 'pending',
    embed_token VARCHAR(255) UNIQUE NOT NULL,
    stream_uid VARCHAR(255),
    stream_status VARCHAR(50),
    manifest_url TEXT,
    thumbnail_url TEXT,
    duration_sec INT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
CREATE INDEX IF NOT EXISTS idx_videos_embed_token ON videos(embed_token);
CREATE INDEX IF NOT EXISTS idx_videos_stream_uid ON videos(stream_uid);
`
	_, err := db.Exec(context.Background(), migrationSQL)
	return err
}

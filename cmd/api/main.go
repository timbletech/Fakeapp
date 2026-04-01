package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"device_only/internal/config"
	"device_only/internal/deepfake"
	"device_only/internal/handlers"
	"device_only/internal/middleware"
	"device_only/internal/orchestration"
	"device_only/internal/repository"
	"device_only/internal/service"

	simapi "device_only/internal/sim/api"
	simconfig "device_only/internal/sim/config"
	simmiddleware "device_only/internal/sim/middleware"
	simprovider "device_only/internal/sim/provider"
	simservice "device_only/internal/sim/service"
	simstore "device_only/internal/sim/store"
)

func main() {
	cfg := config.LoadConfig()

	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	// Wait for DB to be ready
	for i := 0; i < 5; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Printf("Waiting for DB connection...")
		time.Sleep(2 * time.Second)
	}

	log.Println("Applying database migrations...")
	migrationBytes, err := os.ReadFile("migrations/001_initial_schema.sql")
	if err != nil {
		log.Fatalf("Failed to read migrations file 001: %v", err)
	}
	if _, err := db.Exec(string(migrationBytes)); err != nil {
		log.Fatalf("Failed to execute migrations 001: %v", err)
	}

	migrationBytes2, err := os.ReadFile("migrations/002_user_ref_unique.sql")
	if err != nil {
		log.Fatalf("Failed to read migrations file 002: %v", err)
	}
	if _, err := db.Exec(string(migrationBytes2)); err != nil {
		log.Printf("Warning: Migration 002 failed (might already be applied): %v", err)
	}

	migrationBytes3, err := os.ReadFile("migrations/003_add_client_id.sql")
	if err != nil {
		log.Fatalf("Failed to read migrations file 003: %v", err)
	}
	if _, err := db.Exec(string(migrationBytes3)); err != nil {
		log.Printf("Warning: Migration 003 failed (might already be applied): %v", err)
	}

	log.Println("Database migrations applied successfully.")

	repo := repository.NewPostgresRepo(db)

	deviceSvc := service.NewDeviceService(repo)
	authSvc := service.NewAuthService(repo, cfg)

	// --- Initialize Sim/Timble API ---
	simCfg, err := simconfig.Load()
	if err != nil {
		log.Fatalf("Sim configuration error: %v", err)
	}

	sessionStore := simstore.NewSessionStore()
	sekuraProvider := simprovider.NewSekuraProvider(
		simCfg.SekuraBaseURL,
		simCfg.SekuraClientKey,
		simCfg.SekuraClientSecret,
		simCfg.SekuraRefreshToken,
	)

	// Construct the dynamic base URL from configuration for the SIM redirect logic
	baseURL := simCfg.ServerDomain

	simAuthService := simservice.NewAuthService(
		sessionStore,
		sekuraProvider,
		baseURL,
		simCfg.SessionTTLSeconds,
		simCfg.SessionMaxAttempts,
	)
	simAuthHandler := simapi.NewAuthHandler(simAuthService)

	orchestrationStore := orchestration.NewStore()
	api := handlers.NewAPI(deviceSvc, authSvc, simAuthService, cfg, orchestrationStore)

	deepfakeClient := deepfake.NewClient(cfg.DeepfakeServiceURL)
	faceClient := deepfake.NewClient(cfg.FaceServiceURL)
	deepfakeHandler := handlers.NewDeepfakeHandler(deepfakeClient)
	faceHandler := handlers.NewFaceHandler(faceClient)

	voiceClient := deepfake.NewVoiceClient(cfg.VoiceServiceURL)
	voiceHandler := handlers.NewVoiceHandler(voiceClient)

	mux := http.NewServeMux()
	api.SetupRoutes(mux)
	deepfakeHandler.SetupRoutes(mux)
	faceHandler.SetupRoutes(mux)
	voiceHandler.SetupRoutes(mux)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "API is running",
		})
	})

	mux.Handle("/v1/sim/start", simmiddleware.APIKeyAuth(simCfg.TimbleAPIKey, http.HandlerFunc(simAuthHandler.Start)))
	mux.Handle("/v1/sim/complete", simmiddleware.APIKeyAuth(simCfg.TimbleAPIKey, http.HandlerFunc(simAuthHandler.Complete)))
	// No auth required for redirect so device can seamlessly navigate to it
	mux.Handle("/v1/sim/redirect/", http.HandlerFunc(simAuthHandler.Redirect))
	// No auth required for polling proxy so browser/mobile clients can call it directly
	mux.Handle("/v1/sim/poll/", http.HandlerFunc(simAuthHandler.Poll))

	// Serve static files for demo
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/demo/", http.StripPrefix("/demo/", fs))
	mux.Handle("/", http.StripPrefix("/", fs))
	// ----------------------------------

	// Apply CORS middleware
	handler := middleware.CORSMiddleware(mux)

	// In an actual production scenario, HTTPS configuration and CORS would go here.
	server := &http.Server{
		Addr:    cfg.ServerHost + ":" + cfg.ServerPort,
		Handler: handler,
	}

	go func() {
		log.Printf("Server starting on %s:%s...", cfg.ServerHost, cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

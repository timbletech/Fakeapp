package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	ServerPort          string
	ServerDomain        string // e.g. https://api.mydomain.com
	TimbleAPIKey        string
	SekuraBaseURL       string
	SekuraClientKey     string
	SekuraClientSecret  string
	SekuraRefreshToken  string
	SessionTTLSeconds   int
	SessionMaxAttempts  int
	PollingMaxRetries   int
	PollingRetryDelayMs int
}

// Load reads configuration from a .env file (if present) and environment variables.
// Missing required variables cause a descriptive error so the server never starts
// in an invalid state.
func Load() (*Config, error) {
	// Not fatal if .env is absent — system env vars take precedence.
	_ = godotenv.Load()

	cfg := &Config{}
	var missing []string

	cfg.ServerPort = envOr("SERVER_PORT", "8080")
	cfg.ServerDomain = envOr("SERVER_DOMAIN", "http://localhost:"+cfg.ServerPort)

	cfg.TimbleAPIKey = os.Getenv("TIMBLE_API_KEY")
	if cfg.TimbleAPIKey == "" {
		missing = append(missing, "TIMBLE_API_KEY")
	}

	cfg.SekuraBaseURL = os.Getenv("SEKURA_BASE_URL")
	if cfg.SekuraBaseURL == "" {
		missing = append(missing, "SEKURA_BASE_URL")
	}

	cfg.SekuraClientKey = os.Getenv("SEKURA_CLIENT_KEY")
	if cfg.SekuraClientKey == "" {
		missing = append(missing, "SEKURA_CLIENT_KEY")
	}

	cfg.SekuraClientSecret = os.Getenv("SEKURA_CLIENT_SECRET")
	if cfg.SekuraClientSecret == "" {
		missing = append(missing, "SEKURA_CLIENT_SECRET")
	}

	cfg.SekuraRefreshToken = os.Getenv("SEKURA_REFRESH_TOKEN")
	if cfg.SekuraRefreshToken == "" {
		missing = append(missing, "SEKURA_REFRESH_TOKEN")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	var err error

	cfg.SessionTTLSeconds, err = envInt("SESSION_TTL_SECONDS", 300)
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_TTL_SECONDS: %w", err)
	}

	cfg.SessionMaxAttempts, err = envInt("SESSION_MAX_ATTEMPTS", 3)
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_MAX_ATTEMPTS: %w", err)
	}

	cfg.PollingMaxRetries, err = envInt("POLLING_MAX_RETRIES", 3)
	if err != nil {
		return nil, fmt.Errorf("invalid POLLING_MAX_RETRIES: %w", err)
	}

	cfg.PollingRetryDelayMs, err = envInt("POLLING_RETRY_DELAY_MS", 2000)
	if err != nil {
		return nil, fmt.Errorf("invalid POLLING_RETRY_DELAY_MS: %w", err)
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	return strconv.Atoi(v)
}

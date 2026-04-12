package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN                  string
	ServerHost             string
	ServerPort             string
	DevMode                bool
	DevPrivateKey          string
	ChallengeExpirySeconds int
	AuthTokenExpirySeconds int
	DeviceApprovalExpirySeconds int
	DeepfakeServiceURL         string
	FaceServiceURL             string
	VoiceServiceURL            string
}

func LoadConfig() *Config {
	_ = godotenv.Load() // Ignore error if .env is missing, could be supplied by environment

	devMode, _ := strconv.ParseBool(getEnv("DEV_MODE", "false"))
	challengeExpiry, err := strconv.Atoi(getEnv("CHALLENGE_EXPIRY_SECONDS", "120"))
	if err != nil {
		challengeExpiry = 120
	}

	authTokenExpiry, err := strconv.Atoi(getEnv("AUTH_TOKEN_EXPIRY_SECONDS", "300"))
	if err != nil {
		authTokenExpiry = 300
	}

	deviceApprovalExpiry, err := strconv.Atoi(getEnv("DEVICE_APPROVAL_EXPIRY_SECONDS", "300"))
	if err != nil {
		deviceApprovalExpiry = 300
	}

	cfg := &Config{
		DBDSN:                  getEnv("DB_DSN", "postgres://user:password@localhost:5432/timble?sslmode=disable"),
		ServerHost:             getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:             getEnv("SERVER_PORT", "8080"),
		DevMode:                devMode,
		DevPrivateKey:          getEnv("DEV_PRIVATE_KEY", ""),
		ChallengeExpirySeconds: challengeExpiry,
		AuthTokenExpirySeconds:     authTokenExpiry,
		DeviceApprovalExpirySeconds: deviceApprovalExpiry,
		DeepfakeServiceURL:         getEnv("DEEPFAKE_SERVICE_URL", "http://localhost:8000"),
		FaceServiceURL:         getEnv("FACE_SERVICE_URL", "http://localhost:8001"),
		VoiceServiceURL:        getEnv("VOICE_SERVICE_URL", "http://localhost:8002"),
	}

	if cfg.DevMode {
		log.Println("WARNING: Server started in DEV_MODE=true. Simulating device signatures.")
		if cfg.DevPrivateKey == "" {
			log.Println("WARNING: DEV_PRIVATE_KEY is empty. DEV_MODE relies on it for test signatures.")
		}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

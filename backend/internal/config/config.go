package config

import "os"

type Config struct {
	Port            string
	DatabaseURL     string
	SpotifyClientID string
	FrontendURL     string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres:///metaloreian?host=/tmp&sslmode=disable"),
		SpotifyClientID: getEnv("SPOTIFY_CLIENT_ID", ""),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

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
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://metaloreian:metaloreian_dev@localhost:5432/metaloreian?sslmode=disable"),
		SpotifyClientID: getEnv("SPOTIFY_CLIENT_ID", "37a4b40e4fa24e5caa7f219f32899689"),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

package config

import "os"

type Config struct {
	Port               string
	Debug              bool
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	ClaudeAPIKey       string
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
	RedirectBaseURL    string
	FrontendURL        string
}

func Load() (*Config, error) {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		Debug:              getEnv("DEBUG", "") != "",
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/cospec?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		ClaudeAPIKey:       getEnv("CLAUDE_API_KEY", ""),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectBaseURL:    getEnv("REDIRECT_BASE_URL", "http://localhost:8080"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	RateLimitPerMin int
	CleanupInterval time.Duration
	UserTTL         time.Duration
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{
		Port:            port,
		RateLimitPerMin: 5,
		CleanupInterval: 10 * time.Minute,
		UserTTL:         5 * time.Minute,
	}
}

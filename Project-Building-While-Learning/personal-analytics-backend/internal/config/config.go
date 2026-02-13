package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server setting
	Port string

	// Database
	DBPath string

	// Redis
	RedisAddr string

	// Auth
	JWTSecret string

	// Log Level
	LogLevel string // can we can debug, error or verbose

	// ShutdownTimeout
	ShutdownTimeout time.Duration // Go's standard for time 5*time.Second

	// RateLimit
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// WorkerPoolSize
	WorkerPoolSize int
}

func Load() (*Config, error) {
	// Config{} actually builds the house from the blueprint.
	// By adding the curly braces, you are initializing it. It creates a real object in the computer's memory where all the strings are empty "" and the numbers are 0.
	cfg := &Config{}

	// Load the PORT
	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// Load DBPath (default : "./data.db")
	cfg.DBPath = os.Getenv("DB_PATH")
	if cfg.DBPath == "" {
		cfg.DBPath = "./data.db"
	}

	// Load RedisAddr
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")

	if host == "" {
		host = "localhost"
	}

	if port == "" {
		port = "6379"
	}
	cfg.RedisAddr = host + ":" + port

	// Load JWT_SECRET
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required but not set")
	}

	// Load LogLevel
	cfg.LogLevel = os.Getenv("LOG_LEVEL")
	if cfg.LogLevel == "" {
		cfg.LogLevel = "DEBUG" // Setting default one
	}

	// Load ShutdownTimeout
	timeout, err := strconv.Atoi(os.Getenv("SHUTDOWN_TIMEOUT"))
	if err != nil || timeout == 0 {
		timeout = 5
	}

	cfg.ShutdownTimeout = time.Duration(timeout) * time.Second

	// Load RateLimitRequests
	rateLimitrequests, err := strconv.Atoi(os.Getenv("RATE_LIMIT_REQUESTS"))
	if err != nil || rateLimitrequests == 0 {
		rateLimitrequests = 100
	}
	cfg.RateLimitRequests = rateLimitrequests

	// Load RateLimitWindow
	rateLimitWindow, err := strconv.Atoi(os.Getenv("RATE_LIMIT_WINDOW"))
	if err != nil || rateLimitWindow == 0 {
		rateLimitWindow = 60
	}
	cfg.RateLimitWindow = time.Duration(rateLimitWindow) * time.Second

	// Load WorkerPoolSize
	workerPoolSize, err := strconv.Atoi(os.Getenv("WORKERPOOL_SIZE"))
	if err != nil || workerPoolSize == 0 {
		workerPoolSize = 3

	}
	cfg.WorkerPoolSize = workerPoolSize
	return cfg, nil
}

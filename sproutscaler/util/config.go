package util

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds configuration settings for polling and EMA calculations
type Config struct {
	MaxEntries          int           // Maximum entries to store in StatsStorage
	PollingInterval     time.Duration // Interval between polls (in seconds)
	EMADepth            int           // Depth for EMA calculation (used to calculate alpha)
	BaseSensitivityUp   float64       // Base sensitivity for scaling up
	BaseSensitivityDown float64       // Base sensitivity for scaling down
}

// NewConfig initializes and returns a Config struct with environment variables or default values
func NewConfig() *Config {
	maxEntries := getEnvAsInt("MAX_ENTRIES", 10)
	pollingInterval := time.Duration(getEnvAsInt("POLLING_INTERVAL_SECONDS", 1)) * time.Second
	emaDepth := getEnvAsInt("EMA_DEPTH", 10)
	baseSensitivityUp := getEnvAsFloat("BASE_SENSITIVITY_UP", 1.0)   // Default sensitivity for scaling up
	baseSensitivityDown := getEnvAsFloat("BASE_SENSITIVITY_DOWN", 5) // Default sensitivity for scaling down

	return &Config{
		MaxEntries:          maxEntries,
		PollingInterval:     pollingInterval,
		EMADepth:            emaDepth,
		BaseSensitivityUp:   baseSensitivityUp,
		BaseSensitivityDown: baseSensitivityDown,
	}
}

// Helper functions for fetching environment variables
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatalf("Invalid integer value for %s: %v", key, err)
	}
	return value
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Fatalf("Invalid float value for %s: %v", key, err)
	}
	return value
}

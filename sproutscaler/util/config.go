package util

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds configuration settings for polling and EMA calculations
type Config struct {
	MaxEntries      int           // Maximum entries to store in StatsStorage
	PollingInterval time.Duration // Interval between polls (in seconds)
	EMADepth        int           // Depth for EMA calculation (used to calculate alpha)
	BaseSensitivity float64       // Base sensitivity for scaling adjustments
}

// NewConfig initializes and returns a Config struct with environment variables or default values
func NewConfig() *Config {
	maxEntries := getEnvAsInt("MAX_ENTRIES", 5)
	pollingInterval := time.Duration(getEnvAsInt("POLLING_INTERVAL_SECONDS", 2)) * time.Second
	emaDepth := getEnvAsInt("EMA_DEPTH", 5)
	baseSensitivity := getEnvAsFloat("BASE_SENSITIVITY", 0.05)

	return &Config{
		MaxEntries:      maxEntries,
		PollingInterval: pollingInterval,
		EMADepth:        emaDepth,
		BaseSensitivity: baseSensitivity,
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

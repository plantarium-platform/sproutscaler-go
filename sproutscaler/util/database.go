package util

import (
	"log"
	"math"
)

// BackendStats holds the relevant simplified stats from HAProxy, focusing only on response time and its EMA
type BackendStats struct {
	Rtime    int     `json:"rtime"`     // Response time
	RtimeEMA float64 `json:"rtime_ema"` // EMA for response time
}

// StatsStorage will hold the stats over time and the last EMA value for response time
type StatsStorage struct {
	Stats        []BackendStats
	MaxEntries   int
	LastRtimeEMA float64 // Last calculated EMA for response time
}

// NewStatsStorage initializes a new StatsStorage with a specified capacity
func NewStatsStorage(maxEntries int) *StatsStorage {
	return &StatsStorage{
		Stats:      make([]BackendStats, 0, maxEntries),
		MaxEntries: maxEntries,
	}
}

// AddStatWithEMA adds a stat entry, calculating and updating the EMA for response time
func (s *StatsStorage) AddStatWithEMA(stats BackendStats, alpha float64) {
	if len(s.Stats) == 0 {
		// Initialize EMA with the first observation value
		s.LastRtimeEMA = float64(stats.Rtime)
		stats.RtimeEMA = float64(stats.Rtime)
	} else {
		// Calculate EMA based on the last EMA value
		s.LastRtimeEMA = alpha*float64(stats.Rtime) + (1-alpha)*s.LastRtimeEMA
		stats.RtimeEMA = s.LastRtimeEMA
	}

	// Log the current response time and updated EMA
	log.Printf("New stat added - Rtime: %d, RtimeEMA: %.2f", stats.Rtime, stats.RtimeEMA)

	// Manage the Stats slice to keep within the max entries limit
	if len(s.Stats) >= s.MaxEntries {
		s.Stats = s.Stats[1:] // Remove oldest if at capacity
	}
	s.Stats = append(s.Stats, stats)
}

// CalculateInstanceAdjustment calculates the number of instances to add/remove based on Delta Percent of response time EMA
func (s *StatsStorage) CalculateInstanceAdjustment(baseSensitivity float64, instanceCount int) int {
	if len(s.Stats) < s.MaxEntries {
		log.Println("Not enough data to calculate Delta Percent")
		return 0 // Not enough data to calculate Delta Percent
	}

	// Calculate Delta Percent comparing current EMA with the EMA from `N` steps ago
	deltaPercent := (s.LastRtimeEMA - s.Stats[0].RtimeEMA) / s.Stats[0].RtimeEMA * 100
	adjustedSensitivity := baseSensitivity * math.Exp(1.0/float64(instanceCount+1))

	// Log calculation steps
	log.Printf("Calculating instance adjustment:")
	log.Printf(" - Last RtimeEMA: %.2f", s.LastRtimeEMA)
	log.Printf(" - RtimeEMA %d steps ago: %.2f", s.MaxEntries, s.Stats[0].RtimeEMA)
	log.Printf(" - Delta Percent: %.2f%%", deltaPercent)
	log.Printf(" - Base Sensitivity: %.4f", baseSensitivity)
	log.Printf(" - Adjusted Sensitivity: %.4f", adjustedSensitivity)

	// Calculate and log the number of instances to add/remove
	instanceAdjustment := int(math.Round(deltaPercent * adjustedSensitivity))
	log.Printf(" - Calculated Instance Adjustment: %d", instanceAdjustment)

	return instanceAdjustment
}

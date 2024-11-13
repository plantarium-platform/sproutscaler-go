package util

import (
	"log"
)

// BackendStats holds the relevant stats from HAProxy, focusing on response time and instance count
type BackendStats struct {
	Rtime         int     `json:"rtime"`          // Response time
	RtimeEMA      float64 `json:"rtime_ema"`      // EMA for response time
	InstanceCount int     `json:"instance_count"` // Instance count at this tick
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

// AddStatWithEMA adds a stat entry, calculating and updating the EMA for response time and instance count
func (s *StatsStorage) AddStatWithEMA(stats BackendStats, alpha float64) {
	if stats.Rtime == 0 {
		// Set EMA to 0 if response time is 0
		stats.RtimeEMA = 0
		log.Println("Rtime is 0, setting RtimeEMA to 0.")
	} else if len(s.Stats) == 0 || s.LastRtimeEMA == 0 {
		// Initialize EMA with the first non-zero observation value
		s.LastRtimeEMA = float64(stats.Rtime)
		stats.RtimeEMA = s.LastRtimeEMA
		log.Printf("Initializing EMA with first non-zero Rtime: %d", stats.Rtime)
	} else {
		// Calculate EMA based on the last EMA value
		s.LastRtimeEMA = alpha*float64(stats.Rtime) + (1-alpha)*s.LastRtimeEMA
		stats.RtimeEMA = s.LastRtimeEMA
	}

	// Log the current response time, instance count, and updated EMA
	log.Printf("New stat added - Rtime: %d, RtimeEMA: %.2f, InstanceCount: %d", stats.Rtime, stats.RtimeEMA, stats.InstanceCount)

	// Manage the Stats slice to keep within the max entries limit
	if len(s.Stats) >= s.MaxEntries {
		s.Stats = s.Stats[1:] // Remove oldest if at capacity
	}
	s.Stats = append(s.Stats, stats)
}

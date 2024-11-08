package sproutscaler

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

// BackendStats holds the relevant simplified stats from HAProxy, including EMA
type BackendStats struct {
	Rate     float64 `json:"rate"`      // Request rate
	Hrsp5xx  int     `json:"hrsp_5xx"`  // 5xx responses
	Qcur     float64 `json:"qcur"`      // Current queue length
	Rtime    int     `json:"rtime"`     // Response time
	RateEMA  float64 `json:"rate_ema"`  // EMA for rate
	QcurEMA  float64 `json:"qcur_ema"`  // EMA for current queue length
	RtimeEMA float64 `json:"rtime_ema"` // EMA for response time
}

// StatsStorage will hold the stats over time and the last EMA values
type StatsStorage struct {
	Stats        []BackendStats
	MaxEntries   int
	LastRateEMA  float64 // Last calculated EMA for Rate
	LastQcurEMA  float64 // Last calculated EMA for Qcur
	LastRtimeEMA float64 // Last calculated EMA for Rtime
}

func (s *StatsStorage) AddStatWithEMA(stats BackendStats, alpha float64) {
	// If this is the first entry, initialize the EMAs with the current values
	if len(s.Stats) == 0 {
		s.LastRateEMA = stats.Rate
		s.LastQcurEMA = stats.Qcur
		s.LastRtimeEMA = float64(stats.Rtime)

		// Set the initial EMA values for the current stat
		stats.RateEMA = stats.Rate
		stats.QcurEMA = stats.Qcur
		stats.RtimeEMA = float64(stats.Rtime)
	} else {
		// Calculate the new EMA for each parameter
		s.LastRateEMA = alpha*stats.Rate + (1-alpha)*s.LastRateEMA
		s.LastQcurEMA = alpha*stats.Qcur + (1-alpha)*s.LastQcurEMA
		s.LastRtimeEMA = alpha*float64(stats.Rtime) + (1-alpha)*s.LastRtimeEMA

		// Update the current stat with the new EMA values
		stats.RateEMA = s.LastRateEMA
		stats.QcurEMA = s.LastQcurEMA
		stats.RtimeEMA = s.LastRtimeEMA
	}

	// If we have already stored enough entries, remove the oldest
	if len(s.Stats) >= s.MaxEntries {
		s.Stats = s.Stats[1:]
	}

	// Add the new stat with the updated EMAs
	s.Stats = append(s.Stats, stats)
}

// PrintStats is just for debugging purposes to print out stored stats
func (s *StatsStorage) PrintStats() {
	for _, entry := range s.Stats {
		fmt.Printf("Rate: %f | RateEMA: %f | Hrsp5xx: %d | Qcur: %f | QcurEMA: %f | Rtime: %d | RtimeEMA: %f\n",
			entry.Rate, entry.RateEMA, entry.Hrsp5xx, entry.Qcur, entry.QcurEMA, entry.Rtime, entry.RtimeEMA)
	}
}

// PollHAProxyStats fetches stats from HAProxy every `interval` seconds using resty and calculates EMA
func PollHAProxyStats(backendName string, storage *StatsStorage, interval time.Duration, EMADepth int) {
	// Hardcoded base URL
	const baseURL = "http://localhost:5555/v3/services/haproxy/stats/native"

	// Create a new resty client
	client := resty.New()
	client.SetDisableWarn(true)
	client.SetBaseURL(baseURL)
	client.SetBasicAuth("admin", "mypassword") // Replace with your actual username and password
	client.SetHeader("Content-Type", "application/json")

	alpha := 2.0 / float64(EMADepth+1)

	for {
		// Send GET request to the HAProxy stats URL
		resp, err := client.R().Get("")
		if err != nil {
			log.Printf("Failed to fetch stats from HAProxy: %v", err)
			time.Sleep(interval)
			continue
		}

		// Parse the response into the desired format
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &data); err != nil {
			log.Printf("Failed to parse stats: %v", err)
			time.Sleep(interval)
			continue
		}

		// Loop through stats and find backend stats
		for _, item := range data["stats"].([]interface{}) {
			statMap := item.(map[string]interface{})
			if statMap["type"] == "backend" && statMap["name"] == backendName {
				// Extract backend stats
				backendStats := statMap["stats"].(map[string]interface{})

				// Prepare backend stats
				stats := BackendStats{
					Rate:    backendStats["rate"].(float64),
					Hrsp5xx: int(backendStats["hrsp_5xx"].(float64)),
					Qcur:    backendStats["qcur"].(float64),
					Rtime:   int(backendStats["rtime"].(float64)),
				}

				// Store the stat with EMA calculation
				storage.AddStatWithEMA(stats, alpha)

				// Log the parsed stat with EMAs
				log.Printf("Stored stat with EMA for backend '%s': %+v", backendName, stats)
			}
		}

		// Sleep before making the next request
		time.Sleep(interval)
	}
}

package scaler

import (
	"encoding/json"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/plantarium-platform/sproutscaler-go/sproutscaler/util"
)

// Poller fetches stats from HAProxy, updates StatsStorage, and triggers scaling adjustments
type Poller struct {
	Storage         *util.StatsStorage // Stats storage to hold historical data and calculate EMA
	Scaler          *SproutScaler      // Scaler to handle instance adjustments
	BaseSensitivity float64            // Base sensitivity for scaling
	Interval        time.Duration      // Polling interval
	BackendName     string             // Backend to monitor
}

// NewPoller initializes a Poller with StatsStorage, scaler, and polling configurations
func NewPoller(storage *util.StatsStorage, scaler *SproutScaler, baseSensitivity float64, interval time.Duration, backendName string) *Poller {
	return &Poller{
		Storage:         storage,
		Scaler:          scaler,
		BaseSensitivity: baseSensitivity,
		Interval:        interval,
		BackendName:     backendName,
	}
}

// StartPolling begins polling HAProxy stats, updating StatsStorage, and triggering scaling as needed
func (p *Poller) StartPolling() {
	client := resty.New()
	client.SetDisableWarn(true)
	client.SetBaseURL("http://localhost:5555/v3/services/haproxy/stats/native")
	client.SetBasicAuth("admin", "mypassword") // Replace with actual credentials
	client.SetHeader("Content-Type", "application/json")

	// Calculate the EMA smoothing factor
	alpha := 2.0 / float64(p.Storage.MaxEntries+1)

	for {
		resp, err := client.R().Get("")
		if err != nil {
			log.Printf("Failed to fetch stats from HAProxy: %v", err)
			time.Sleep(p.Interval)
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &data); err != nil {
			log.Printf("Failed to parse stats: %v", err)
			time.Sleep(p.Interval)
			continue
		}

		// Loop through stats and find backend stats
		for _, item := range data["stats"].([]interface{}) {
			statMap := item.(map[string]interface{})
			if statMap["type"] == "backend" && statMap["name"] == p.BackendName {
				// Extract backend stats
				backendStats := statMap["stats"].(map[string]interface{})

				// Prepare backend stats focusing only on response time
				rtime := int(backendStats["rtime"].(float64))
				if rtime == 0 {
					log.Printf("Skipping stats with Rtime=0 for backend '%s'", p.BackendName)
					continue
				}

				stats := util.BackendStats{Rtime: rtime}

				// Store the stat with EMA calculation
				p.Storage.AddStatWithEMA(stats, alpha)

				// Log the parsed stat with EMA for response time
				log.Printf("Stored stat with EMA for backend '%s': %+v", p.BackendName, stats)

				// Calculate the instances to add/remove based on response time EMA
				instanceAdjustment := p.Storage.CalculateInstanceAdjustment(p.BaseSensitivity, p.Scaler.GetInstanceCount())
				log.Printf("Calculated instances to add/remove based on response time EMA: %d", instanceAdjustment)

				// Scale instances as needed
				p.Scaler.Scale(instanceAdjustment)
			}
		}

		// Sleep before making the next request
		time.Sleep(p.Interval)
	}
}

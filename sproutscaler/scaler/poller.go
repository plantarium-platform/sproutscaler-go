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
	Storage     *util.StatsStorage
	Scaler      *SproutScaler
	Algorithm   *ScalingAlgorithm
	Interval    time.Duration
	BackendName string
}

// NewPoller initializes a Poller with StatsStorage, Scaler, Algorithm, and polling configurations
func NewPoller(storage *util.StatsStorage, scaler *SproutScaler, algorithm *ScalingAlgorithm, interval time.Duration, backendName string) *Poller {
	return &Poller{
		Storage:     storage,
		Scaler:      scaler,
		Algorithm:   algorithm,
		Interval:    interval,
		BackendName: backendName,
	}
}

// StartPolling initiates polling, retrieving HAProxy stats, updating StatsStorage, and adjusting scaling as needed
func (p *Poller) StartPolling() {
	client := p.initializeClient()
	alpha := 2.0 / float64(p.Storage.MaxEntries+1) // EMA smoothing factor

	for {
		data, err := p.fetchHAProxyStats(client)
		if err != nil {
			log.Printf("Error fetching HAProxy stats: %v", err)
			time.Sleep(p.Interval)
			continue
		}

		if backendStats := p.extractBackendStats(data); backendStats != nil {
			p.processStats(*backendStats, alpha)
		}

		time.Sleep(p.Interval) // Pause before the next polling cycle
	}
}

// initializeClient sets up the HTTP client for polling HAProxy stats
func (p *Poller) initializeClient() *resty.Client {
	client := resty.New()
	client.SetDisableWarn(true)
	client.SetBaseURL("http://localhost:5555/v3/services/haproxy/stats/native")
	client.SetBasicAuth("admin", "mypassword") // Replace with actual credentials
	client.SetHeader("Content-Type", "application/json")
	return client
}

// fetchHAProxyStats retrieves the HAProxy stats data
func (p *Poller) fetchHAProxyStats(client *resty.Client) (map[string]interface{}, error) {
	resp, err := client.R().Get("")
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, err
	}

	return data, nil
}

// extractBackendStats extracts backend stats for the specified backend name and includes the current instance count
func (p *Poller) extractBackendStats(data map[string]interface{}) *util.BackendStats {
	for _, item := range data["stats"].([]interface{}) {
		statMap := item.(map[string]interface{})
		if statMap["type"] == "backend" && statMap["name"] == p.BackendName {
			backendStats := statMap["stats"].(map[string]interface{})
			rtime := int(backendStats["rtime"].(float64))

			// Include the current instance count in the BackendStats
			instanceCount := p.Scaler.GetInstanceCount()

			// Return the extracted stats, allowing zero values
			return &util.BackendStats{Rtime: rtime, InstanceCount: instanceCount}
		}
	}
	return nil
}

// processStats updates the storage with new stats and triggers scaling if necessary
func (p *Poller) processStats(stats util.BackendStats, alpha float64) {
	p.Storage.AddStatWithEMA(stats, alpha)

	instanceAdjustment := p.Algorithm.CalculateInstanceAdjustment(p.Scaler.GetInstanceCount())
	if instanceAdjustment != 0 {
		p.Scaler.Scale(instanceAdjustment)
	}
}

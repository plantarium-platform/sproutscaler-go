package scaler

import (
	"github.com/plantarium-platform/sproutscaler-go/sproutscaler/util"
	"log"
	"math"
)

// ScalingAlgorithm holds the logic to calculate instance adjustments based on stored stats
type ScalingAlgorithm struct {
	BaseSensitivityUp   float64
	BaseSensitivityDown float64
	Storage             *util.StatsStorage
}

// NewScalingAlgorithm initializes a new ScalingAlgorithm with a base sensitivity and reference to StatsStorage
func NewScalingAlgorithm(baseSensitivityUp float64, baseSensitivityDown float64, storage *util.StatsStorage) *ScalingAlgorithm {
	return &ScalingAlgorithm{
		BaseSensitivityUp:   baseSensitivityUp,
		BaseSensitivityDown: baseSensitivityDown,
		Storage:             storage,
	}
}

// CalculateInstanceAdjustment calculates the number of instances to add/remove based on Delta Percent of response time EMA and instance delta
func (a *ScalingAlgorithm) CalculateInstanceAdjustment(currentInstanceCount int) int {
	if len(a.Storage.Stats) < a.Storage.MaxEntries {
		log.Println("Not enough data to calculate Delta Percent")
		return 0 // Not enough data to calculate Delta Percent
	}

	// Check if response time has been zero for the last `N` intervals
	allZero := true
	for _, stat := range a.Storage.Stats {
		if stat.RtimeEMA > 0 {
			allZero = false
			break
		}
	}
	if allZero {
		log.Println("Response time EMA has been zero across all recent intervals; recommending removal of all instances > 1.")
		return -currentInstanceCount + 1
	}

	// Calculate Delta Percent by comparing current EMA with the EMA from `N` steps ago
	var deltaPercent float64
	previousEMA := a.Storage.Stats[0].RtimeEMA

	if previousEMA == 0 {
		log.Println("Previous EMA is 0; cannot calculate Delta Percent.")
		return 0
	} else {
		deltaPercent = (a.Storage.LastRtimeEMA - previousEMA) / previousEMA
	}

	// Calculate the absolute difference in instance count between now and `N` steps ago
	instancePreviousDelta := int(math.Abs(float64(currentInstanceCount - a.Storage.Stats[0].InstanceCount)))

	// If instances were recently added or removed, log and prevent further adjustments until cooldown ends
	if instancePreviousDelta != 0 {
		log.Println("Instance count has changed recently; entering cooldown period.")
		return 0 // Prevent scaling adjustments during cooldown
	}

	// Adjust sensitivity based on the current instance count
	// Using separate base sensitivities for adding and removing instances
	var adjustedSensitivity float64
	if deltaPercent > 0 {
		// Use base sensitivity for adding instances
		adjustedSensitivity = a.BaseSensitivityUp * math.Exp(6/float64(currentInstanceCount+1))
	} else {
		// Use base sensitivity for removing instances
		adjustedSensitivity = a.BaseSensitivityDown * math.Exp(4.83/float64(currentInstanceCount+1))
	}

	// Log calculation steps
	log.Printf("Calculating instance adjustment:")
	log.Printf(" - Last RtimeEMA: %.2f", a.Storage.LastRtimeEMA)
	log.Printf(" - RtimeEMA %d steps ago: %.2f", a.Storage.MaxEntries, previousEMA)
	log.Printf(" - Delta Percent: %.4f", deltaPercent*100)
	log.Printf(" - Instance Count %d steps ago: %d", a.Storage.MaxEntries, a.Storage.Stats[0].InstanceCount)
	log.Printf(" - Base Sensitivity (Up): %.4f", a.BaseSensitivityUp)
	log.Printf(" - Base Sensitivity (Down): %.4f", a.BaseSensitivityDown)
	log.Printf(" - Adjusted Sensitivity: %.4f", adjustedSensitivity)
	log.Printf(" - Instance Previous Delta (absolute): %d", instancePreviousDelta)

	// Calculate and log the number of instances to add/remove
	instanceAdjustment := int(math.Round(deltaPercent * adjustedSensitivity))
	log.Printf(" - Calculated Instance Adjustment: %d", instanceAdjustment)

	return instanceAdjustment
}

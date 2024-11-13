package scaler

import (
	"github.com/plantarium-platform/sproutscaler-go/sproutscaler/util"
	"log"
	"math"
)

// ScalingAlgorithm holds the logic to calculate instance adjustments based on stored stats
type ScalingAlgorithm struct {
	BaseSensitivity float64
	Storage         *util.StatsStorage
}

// NewScalingAlgorithm initializes a new ScalingAlgorithm with a base sensitivity and reference to StatsStorage
func NewScalingAlgorithm(baseSensitivity float64, storage *util.StatsStorage) *ScalingAlgorithm {
	return &ScalingAlgorithm{
		BaseSensitivity: baseSensitivity,
		Storage:         storage,
	}
}

// CalculateInstanceAdjustment calculates the number of instances to add/remove based on Delta Percent of response time EMA and instance delta
func (a *ScalingAlgorithm) CalculateInstanceAdjustment(currentInstanceCount int) int {
	if len(a.Storage.Stats) < a.Storage.MaxEntries {
		log.Println("Not enough data to calculate Delta Percent")
		return 0 // Not enough data to calculate Delta Percent
	}

	// Calculate Delta Percent, avoiding division by zero
	var deltaPercent float64
	previousEMA := a.Storage.Stats[0].RtimeEMA

	if previousEMA == 0 {
		log.Println("Previous EMA is 0; cannot calculate Delta Percent.")
		deltaPercent = 0
	} else {
		deltaPercent = (a.Storage.LastRtimeEMA - previousEMA) / previousEMA * 100
	}

	// Calculate the difference in instance count between now and `N` steps ago
	instancePreviousDelta := currentInstanceCount - a.Storage.Stats[0].InstanceCount

	// If instances were recently added, log the adjustment message
	if instancePreviousDelta != 0 {
		log.Println("We detected a change in the number of instances, so we will adjust delta by this value.")
	}

	// Adjust sensitivity based on the current instance count
	adjustedSensitivity := a.BaseSensitivity * math.Exp(1.0/float64(currentInstanceCount+1))

	// Log calculation steps
	log.Printf("Calculating instance adjustment:")
	log.Printf(" - Last RtimeEMA: %.2f", a.Storage.LastRtimeEMA)
	log.Printf(" - RtimeEMA %d steps ago: %.2f", a.Storage.MaxEntries, previousEMA)
	log.Printf(" - Delta Percent: %.2f%%", deltaPercent)
	log.Printf(" - Instance Count %d steps ago: %d", a.Storage.MaxEntries, a.Storage.Stats[0].InstanceCount)
	log.Printf(" - Base Sensitivity: %.4f", a.BaseSensitivity)
	log.Printf(" - Adjusted Sensitivity: %.4f", adjustedSensitivity)
	log.Printf(" - Instance Previous Delta: %d", instancePreviousDelta)

	// Calculate and log the number of instances to add/remove
	instanceAdjustment := int(math.Round(deltaPercent*adjustedSensitivity)) - instancePreviousDelta
	log.Printf(" - Calculated Instance Adjustment (with instancePreviousDelta): %d", instanceAdjustment)

	return instanceAdjustment
}

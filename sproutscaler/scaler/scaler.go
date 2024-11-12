package scaler

import (
	"fmt"
	"github.com/plantarium-platform/graftnode-go/services/haproxy"
	"log"
)

// SproutScaler manages scaling decisions, handling instance addition and removal
type SproutScaler struct {
	haproxyClient *haproxy.HAProxyClient
	backendName   string
	maxInstances  int
	instances     []int
}

// NewSproutScaler initializes and returns a new SproutScaler
func NewSproutScaler(haproxyClient *haproxy.HAProxyClient, backendName string, maxInstances int) *SproutScaler {
	return &SproutScaler{
		haproxyClient: haproxyClient,
		backendName:   backendName,
		maxInstances:  maxInstances,
		instances:     make([]int, 0, maxInstances),
	}
}

// AddInstance adds a new instance to HAProxy and tracks it locally
func (s *SproutScaler) AddInstance() error {
	if len(s.instances) >= s.maxInstances {
		return fmt.Errorf("cannot add more instances, maximum of %d reached", s.maxInstances)
	}

	nextInstanceID := len(s.instances) + 1
	serverName := fmt.Sprintf("java-service-%d", nextInstanceID)

	err := s.haproxyClient.BindService(s.backendName, serverName, "localhost", 8080+nextInstanceID-1)
	if err != nil {
		return fmt.Errorf("failed to add server to HAProxy: %v", err)
	}

	s.instances = append(s.instances, nextInstanceID)
	log.Printf("Added server %s to the backend", serverName)
	return nil
}

// RemoveInstance removes the most recently added instance from HAProxy and local tracking
func (s *SproutScaler) RemoveInstance() error {
	if len(s.instances) == 0 {
		log.Println("No servers to remove from the backend")
		return nil
	}

	lastInstanceID := s.instances[len(s.instances)-1]
	serverName := fmt.Sprintf("java-service-%d", lastInstanceID)

	err := s.haproxyClient.DeleteServer(s.backendName, serverName)
	if err != nil {
		return fmt.Errorf("failed to remove server %s from HAProxy: %v", serverName, err)
	}

	s.instances = s.instances[:len(s.instances)-1]
	log.Printf("Removed server %s from the backend", serverName)
	return nil
}

// GetInstanceCount returns the current number of instances
func (s *SproutScaler) GetInstanceCount() int {
	return len(s.instances)
}

// Scale adjusts the number of instances based on the calculated adjustment value.
// Positive values of instanceAdjustment add instances; negative values remove instances.
func (s *SproutScaler) Scale(instanceAdjustment int) {
	if instanceAdjustment > 0 {
		for i := 0; i < instanceAdjustment; i++ {
			if err := s.AddInstance(); err != nil {
				log.Printf("Failed to add instance: %v", err)
				break
			}
		}
	} else if instanceAdjustment < 0 {
		for i := 0; i < -instanceAdjustment; i++ {
			if err := s.RemoveInstance(); err != nil {
				log.Printf("Failed to remove instance: %v", err)
				break
			}
		}
	} else {
		log.Println("No scaling adjustment required")
	}
}

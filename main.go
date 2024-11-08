package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/plantarium-platform/graftnode-go/services/haproxy"
	"github.com/plantarium-platform/sproutscaler-go/sproutscaler"
	"log"
)

// getEnv fetches the value of an environment variable or returns a default if not set
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func startJavaApp(id int, port int) {
	cmd := exec.Command("java", "-jar", "resources/java-service-example-0.1-all.jar", "--instance-id="+strconv.Itoa(id), "--request-delay=1000")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "MICRONAUT_SERVER_PORT="+strconv.Itoa(port))
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting Java app with ID %d: %v\n", id, err)
		return
	}
	fmt.Printf("Java app with ID %d started on port %d with a delay of 200ms\n", id, port)
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Java app with ID %d exited with error: %v\n", id, err)
	} else {
		fmt.Printf("Java app with ID %d exited successfully\n", id)
	}
}

func main() {
	// Fetch max entries and polling interval from environment variables with defaults
	maxEntriesStr := getEnv("MAX_ENTRIES", "20")
	pollingIntervalStr := getEnv("POLLING_INTERVAL_SECONDS", "2")
	EMADepthStr := getEnv("EMA_DEPTH", "5")

	// Convert fetched values to integers
	maxEntries, err := strconv.Atoi(maxEntriesStr)
	if err != nil {
		log.Fatalf("Invalid MAX_ENTRIES value: %v", err)
	}
	pollingInterval, err := strconv.Atoi(pollingIntervalStr)
	if err != nil {
		log.Fatalf("Invalid POLLING_INTERVAL_SECONDS value: %v", err)
	}

	EMADepth, err := strconv.Atoi(EMADepthStr)
	if err != nil {
		log.Fatalf("Invalid POLLING_INTERVAL_SECONDS value: %v", err)
	}

	// Number of Java app instances
	instances := 5
	startPort := 8080

	// Start Java applications
	for i := 1; i <= instances; i++ {
		go startJavaApp(i, startPort+i-1)
	}

	fmt.Println("All Java instances started. Waiting for requests...")

	// Create a new HAProxyClient
	haproxyClient := haproxy.NewHAProxyClient()

	// Create a new SproutScaler with a maximum of 5 instances
	scaler := sproutscaler.NewSproutScaler(haproxyClient, "service-backend", 5)

	// Delete all existing servers from the backend
	err = haproxyClient.DeleteAllServersFromBackend("service-backend")
	if err != nil {
		fmt.Printf("Error deleting servers from backend: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Deleted all servers from the 'service-backend'")

	// Add the first instance to the backend
	err = scaler.AddInstance()
	if err != nil {
		fmt.Printf("Error adding instance to HAProxy: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Added the first instance to HAProxy")

	// Create a StatsStorage object using the configuration value
	statsStorage := sproutscaler.StatsStorage{
		Stats:      make([]sproutscaler.BackendStats, 0),
		MaxEntries: maxEntries, // Parameterized value from environment variable or default
	}

	// Start polling HAProxy stats in a separate goroutine
	go sproutscaler.PollHAProxyStats("service-backend", &statsStorage, time.Duration(pollingInterval)*time.Second, EMADepth)

	// Keep the main goroutine running
	select {}
}

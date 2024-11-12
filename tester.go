package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

// SendRequests continuously sends HTTP GET requests to the specified URL with random delays
func SendRequests(url string, baseDelay time.Duration, id int) {
	client := &http.Client{
		Timeout: 10 * time.Second, // Add timeout to prevent hanging requests
	}

	for {
		// Record the start time
		startTime := time.Now()

		// Send the HTTP GET request
		resp, err := client.Get(url)
		duration := time.Since(startTime)

		// Log response and execution time
		if err != nil {
			log.Printf("[Thread %d] Failed to send request: %v", id, err)
		} else {
			log.Printf("[Thread %d] Response: %s | Time taken: %v", id, resp.Status, duration)
			resp.Body.Close()
		}

		// Calculate and apply delay with random jitter
		jitter := time.Duration(rand.Float64() * float64(baseDelay))
		totalDelay := baseDelay + jitter
		time.Sleep(totalDelay)
	}
}

func main() {
	// Configurable parameters
	const url = "http://localhost/hello"
	numThreads := 50         // Number of concurrent threads (goroutines)
	baseDelay := time.Second // Base delay between requests

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Start the specified number of goroutines
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			SendRequests(url, baseDelay, threadID)
		}(i)
	}

	// Wait for interrupt signal
	<-sigChan
	log.Println("Received interrupt signal, shutting down...")
}

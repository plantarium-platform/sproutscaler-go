package main

import (
	"fmt"
	"github.com/plantarium-platform/graftnode-go/services/haproxy"
	"github.com/plantarium-platform/sproutscaler-go/sproutscaler"
	"os"
	"os/exec"
	"strconv"
)

func startJavaApp(id int, port int) {
	cmd := exec.Command("java", "-jar", "resources/java-service-example-0.1-all.jar", "--instance-id="+strconv.Itoa(id), "--request-delay=200")
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
	instances := 5
	startPort := 8080

	for i := 1; i <= instances; i++ {
		go startJavaApp(i, startPort+i-1)
	}

	fmt.Println("All Java instances started. Waiting for requests...")

	// Create a new HAProxyClient
	haproxyClient := haproxy.NewHAProxyClient()

	// Create a new SproutScaler with a maximum of 5 instances
	scaler := sproutscaler.NewSproutScaler(haproxyClient, "service-backend", 5)

	// Delete all existing servers from the backend
	err := haproxyClient.DeleteAllServersFromBackend("service-backend")
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

	// Keep the main goroutine running
	select {}
}

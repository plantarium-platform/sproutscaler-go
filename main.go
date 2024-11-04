package main

import (
	"fmt"
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

	// Keep the main goroutine running
	select {}
}

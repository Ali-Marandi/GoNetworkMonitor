package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ali-Marandi/gonetworkmonitor/pkg/monitor"
)

func main() {
	ifaces, err := monitor.ListInterfaces()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}

	if len(ifaces) == 0 {
		log.Fatal("No network interfaces found")
	}

	// Use the first available interface for monitoring
	targetIface := ifaces[0]
	m := monitor.NewMonitor(targetIface)

	if err := m.Start(); err != nil {
		log.Fatalf("Failed to start monitor: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Monitoring... Press Ctrl+C to stop.")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.Report()
		case <-sigChan:
			fmt.Println("\nShutting down...")
			m.Report()
			return
		}
	}
}

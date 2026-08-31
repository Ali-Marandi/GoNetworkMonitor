package main

import (
	"flag"
	"log"

	"github.com/Ali-Marandi/GoNetworkMonitor/pkg/api"
	"github.com/Ali-Marandi/GoNetworkMonitor/pkg/config"
)

// Version is set at build time by the release workflow.
var Version = "dev"

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	interfaceName := flag.String("interface", "", "Network interface to monitor (default: config/auto)")
	port := flag.Int("port", 0, "Override listen port")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		log.Printf("gonetmon %s", Version)
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	if *interfaceName != "" {
		cfg.SetInterface(*interfaceName)
	}
	if *port > 0 {
		cfg.SetPort(*port)
	}

	server := api.NewServer(cfg)
	log.Printf("GoNetworkMonitor %s starting", Version)
	if err := server.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

# GoNetworkMonitor

A lightweight, high-performance network monitoring tool written in Go. Designed for real-time traffic analysis and statistics gathering with minimal resource overhead.

## Features

- **Real-time Monitoring:** Tracks packet counts and data throughput across network interfaces.
- **Interface Auto-discovery:** Automatically identifies available network interfaces for monitoring.
- **Concurrent Design:** Leverages Go's goroutines for efficient packet processing without blocking.
- **Graceful Shutdown:** Ensures accurate final reporting upon termination.

## Installation

Ensure you have Go (1.21 or later) installed.

```bash
git clone https://github.com/Ali-Marandi/gonetworkmonitor.git
cd gonetworkmonitor
go build -o network-monitor cmd/main.go
```

## Usage

Run the monitor with administrative privileges (required for packet capture):

```bash
sudo ./network-monitor
```

## Architecture

The project is structured following Go best practices:
- `cmd/`: Contains the main application entry point.
- `pkg/`: Contains reusable logic for network monitoring and interface management.

## Future Enhancements

- Integration with `gopacket` for deep packet inspection (DPI).
- Support for filtering by protocol (TCP, UDP, ICMP).
- Exporting metrics to Prometheus or InfluxDB.

## License

MIT License

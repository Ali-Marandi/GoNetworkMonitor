# GoNetworkMonitor

<div align="center">

![GoNetworkMonitor Banner](https://img.shields.io/badge/GoNetworkMonitor-v2.0.0-6366f1?style=for-the-badge&logo=go&logoColor=white)

[![CI](https://github.com/Ali-Marandi/GoNetworkMonitor/actions/workflows/ci.yml/badge.svg)](https://github.com/Ali-Marandi/GoNetworkMonitor/actions/workflows/ci.yml)
[![Release](https://github.com/Ali-Marandi/GoNetworkMonitor/actions/workflows/release.yml/badge.svg)](https://github.com/Ali-Marandi/GoNetworkMonitor/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/Ali-Marandi/GoNetworkMonitor)](https://github.com/Ali-Marandi/GoNetworkMonitor/releases)
[![GitHub stars](https://img.shields.io/github/stars/Ali-Marandi/GoNetworkMonitor?style=social)](https://github.com/Ali-Marandi/GoNetworkMonitor/stargazers)

**A commercial-grade, high-performance network monitoring tool written in Go.**
Real-time packet capture, deep protocol analysis, and a beautiful web dashboard — all in a single binary.

[Download Latest Release](https://github.com/Ali-Marandi/GoNetworkMonitor/releases/latest) · [Report Bug](https://github.com/Ali-Marandi/GoNetworkMonitor/issues) · [Request Feature](https://github.com/Ali-Marandi/GoNetworkMonitor/issues)

</div>

---

## Overview

GoNetworkMonitor is a production-ready network analysis platform built for network engineers, security professionals, and system administrators. It leverages Go's concurrency model and the battle-tested `gopacket` library to deliver deep packet inspection with minimal resource overhead, served through a modern real-time web dashboard accessible from any browser.

## Features

| Category | Capabilities |
|---|---|
| **Packet Capture** | Real-time capture via `gopacket`/libpcap, BPF filter support, promiscuous mode |
| **Protocol Analysis** | TCP, UDP, ICMP, ICMPv6, DNS, HTTP, ARP, IPv4, IPv6 |
| **Dashboard** | Real-time traffic charts, protocol distribution, top talkers, connection table |
| **API** | REST API + WebSocket streaming for live data integration |
| **Alerting** | Configurable bandwidth and PPS thresholds with alert history |
| **Connection Tracking** | Active connection table with packet/byte counters |
| **Configuration** | JSON config file + live API-based updates |
| **Deployment** | Single binary, Docker support, multi-platform |

## Quick Start

### Prerequisites

- Linux (recommended) or macOS
- `libpcap` installed (`sudo apt-get install libpcap-dev` or `brew install libpcap`)
- Root/Administrator privileges (required for raw packet capture)

### Download Pre-built Binary

```bash
# Linux (amd64)
curl -LO https://github.com/Ali-Marandi/GoNetworkMonitor/releases/latest/download/gonetmon-v2.0.0-linux-amd64.tar.gz
tar -xzf gonetmon-v2.0.0-linux-amd64.tar.gz
sudo ./gonetmon-linux-amd64
```

Then open your browser at **http://localhost:8080**

### Build from Source

```bash
# Clone the repository
git clone https://github.com/Ali-Marandi/GoNetworkMonitor.git
cd GoNetworkMonitor

# Install dependencies
sudo apt-get install -y libpcap-dev gcc

# Build
make build

# Run
sudo ./dist/gonetmon
```

### Docker

```bash
docker run --rm \
  --cap-add=NET_ADMIN \
  --cap-add=NET_RAW \
  --network=host \
  -p 8080:8080 \
  ghcr.io/ali-marandi/gonetworkmonitor:latest
```

## Usage

```
Usage of gonetmon:
  -config string     Path to configuration file (default "config.json")
  -interface string  Network interface to monitor (default: auto-detect)
  -port int          Override listen port
  -version           Print version and exit
```

### Examples

```bash
# Auto-detect interface, default port 8080
sudo ./gonetmon

# Monitor specific interface
sudo ./gonetmon -interface eth0

# Custom port and config
sudo ./gonetmon -interface eth0 -port 9090 -config /etc/gonetmon/config.json
```

## Configuration

GoNetworkMonitor uses a JSON configuration file (`config.json`). All settings can also be updated live via the web dashboard.

```json
{
  "listen_addr": "0.0.0.0",
  "port": 8080,
  "interface": "auto",
  "snap_len": 65535,
  "promiscuous": true,
  "bpf_filter": "",
  "alerts": {
    "enabled": true,
    "bandwidth_mbps_threshold": 100,
    "pps_threshold": 10000,
    "new_conn_per_second_threshold": 500
  },
  "data_dir": "./data",
  "max_history_seconds": 3600
}
```

## REST API Reference

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/interfaces` | List available network interfaces |
| `POST` | `/api/capture/start` | Start packet capture |
| `POST` | `/api/capture/stop` | Stop packet capture |
| `GET` | `/api/capture/status` | Get capture status |
| `GET` | `/api/stats/snapshot` | Current statistics snapshot |
| `GET` | `/api/stats/timeseries` | Historical time-series data |
| `GET` | `/api/stats/connections` | Active connection table |
| `GET` | `/api/alerts` | Recent alerts |
| `GET/POST` | `/api/config` | Get or update configuration |
| `WS` | `/api/ws` | WebSocket stream for real-time stats |

## Architecture

```
GoNetworkMonitor/
├── cmd/gonetmon/          # Application entry point
├── pkg/
│   ├── capture/           # Packet capture engine (gopacket/libpcap)
│   ├── api/               # REST API + WebSocket server
│   ├── stats/             # Time-series and connection tracking
│   ├── alert/             # Alert management and threshold checking
│   └── config/            # Configuration management
├── web/
│   ├── index.html         # Dashboard SPA
│   └── static/
│       ├── css/style.css  # Modern dark theme
│       └── js/app.js      # Dashboard logic + Chart.js
├── .github/workflows/     # CI/CD pipelines
├── Dockerfile
└── Makefile
```

## Performance

GoNetworkMonitor is designed for high-throughput environments. It uses non-blocking packet processing via buffered Go channels, lock-free atomic counters for hot-path statistics, a WebSocket push model that eliminates polling overhead, and minimal memory allocations in the capture hot path.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](.github/CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

<div align="center">
Made with Go by <a href="https://github.com/Ali-Marandi">Ali Marandi</a>
</div>

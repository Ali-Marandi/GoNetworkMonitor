# GoNetworkMonitor

<div align="center">

![GoNetworkMonitor Banner](https://img.shields.io/badge/GoNetworkMonitor-v3.1.0-6366f1?style=for-the-badge&logo=go&logoColor=white)

[![CI](https://github.com/Ali-Marandi/GoNetworkMonitor/actions/workflows/ci.yml/badge.svg)](https://github.com/Ali-Marandi/GoNetworkMonitor/actions/workflows/ci.yml)
[![Release](https://github.com/Ali-Marandi/GoNetworkMonitor/actions/workflows/release.yml/badge.svg)](https://github.com/Ali-Marandi/GoNetworkMonitor/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/Ali-Marandi/GoNetworkMonitor)](https://github.com/Ali-Marandi/GoNetworkMonitor/releases)

**A high-performance network monitoring tool written in Go.**
Real-time packet capture, protocol analysis, alerting, persistent metrics, Prometheus export, and a web dashboard in a single binary.

[Download Latest Release](https://github.com/Ali-Marandi/GoNetworkMonitor/releases/latest) · [Report Bug](https://github.com/Ali-Marandi/GoNetworkMonitor/issues)

</div>

## Current status

The `master` branch tracks the v3.1.0 codebase. CI and release tooling require Go 1.25, and the application entry point is `cmd/gonetmon/main.go`.

By default the HTTP dashboard binds to `127.0.0.1:8080`. Set `listen_addr` explicitly in `config.json` if remote access is required.

## Features

| Category | Capabilities |
|---|---|
| **Packet Capture** | Real-time capture via `gopacket`/libpcap, BPF filter support, promiscuous mode |
| **Protocol Analysis** | TCP, UDP, ICMP, ICMPv6, DNS, HTTP, ARP, IPv4, IPv6 |
| **Dashboard** | Real-time traffic charts, protocol distribution, top talkers, connection table |
| **API** | REST API + WebSocket streaming for live data integration |
| **Alerting** | Configurable bandwidth and PPS thresholds with alert history |
| **Intelligence** | Statistical anomaly detection for traffic spikes |
| **Storage** | SQLite-backed time-series and alert persistence |
| **Observability** | Prometheus metrics at `/metrics` |
| **Deployment** | Single binary, Docker support, multi-platform release builds |

## Quick Start

### Prerequisites

- Go 1.25+ when building from source
- Linux or macOS (Windows requires Npcap)
- `libpcap` installed (`sudo apt-get install libpcap-dev` or `brew install libpcap`)
- Root/Administrator privileges for packet capture

### Build from Source

```bash
git clone https://github.com/Ali-Marandi/GoNetworkMonitor.git
cd GoNetworkMonitor
sudo apt-get install -y libpcap-dev gcc
make build
sudo ./dist/gonetmon
```

Open **http://127.0.0.1:8080**.

### Usage

```text
Usage of gonetmon:
  -config string     Path to configuration file (default "config.json")
  -interface string  Network interface to monitor (default: config/auto)
  -port int          Override listen port
  -version           Print version and exit
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

## Configuration

GoNetworkMonitor uses `config.json`; settings can also be updated through the API.

```json
{
  "listen_addr": "127.0.0.1",
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

For remote dashboard access, change `listen_addr` deliberately and protect the service with network controls or a reverse proxy with authentication. The built-in API currently does not provide user authentication.

## REST API

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/interfaces` | List available network interfaces |
| `POST` | `/api/capture/start` | Start packet capture |
| `POST` | `/api/capture/stop` | Stop packet capture |
| `GET` | `/api/capture/status` | Get capture status |
| `GET` | `/api/stats/snapshot` | Current statistics snapshot |
| `GET` | `/api/stats/timeseries` | Recent time-series data |
| `GET` | `/api/stats/connections` | Active connection table |
| `GET` | `/api/alerts` | Recent alerts |
| `GET/POST` | `/api/config` | Get or update configuration |
| `WS` | `/api/ws` | WebSocket stream for real-time stats |
| `GET` | `/metrics` | Prometheus metrics |

## Architecture

```text
GoNetworkMonitor/
├── cmd/gonetmon/          # Application entry point
├── pkg/
│   ├── capture/           # Packet capture engine (gopacket/libpcap)
│   ├── api/               # REST API + WebSocket server and embedded UI
│   ├── stats/              # Time-series, connection tracking, Prometheus
│   ├── alert/              # Threshold and anomaly alerting
│   ├── config/             # Configuration management
│   └── storage/            # SQLite persistence
├── .github/workflows/      # CI/CD pipelines
├── Dockerfile
└── Makefile
```

## Development

```bash
make fmt
make vet
make test
make build
```

Contributions are welcome. See `.github/CONTRIBUTING.md` for contribution guidelines.

## License

MIT — see `LICENSE`.

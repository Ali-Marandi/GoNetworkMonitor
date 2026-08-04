# GoNetworkMonitor Commercial Architecture

## Overview
GoNetworkMonitor is being upgraded from a simple CLI tool to a commercial-grade network monitoring solution. It features a high-performance Go backend using `gopacket` for real-time packet inspection, a REST API, and a modern web-based dashboard for visualization and configuration.

## Core Components

### 1. Packet Capture Engine (Backend - Go)
- **Library**: `github.com/google/gopacket`
- **Features**:
  - Promiscuous mode capture on selected interfaces
  - BPF (Berkeley Packet Filter) support
  - Multi-threaded packet processing using Go routines
  - Protocol parsing (Ethernet, IPv4/IPv6, TCP, UDP, ICMP, DNS, HTTP)

### 2. Analytics & Storage Engine (Backend - Go)
- **Metrics Storage**: In-memory time-series aggregation with periodic flush to disk (SQLite or JSON logs)
- **Features**:
  - Bandwidth calculation (bps, pps)
  - Top talkers (IPs, Ports, Protocols)
  - Connection tracking
  - Alerting based on thresholds

### 3. API Server (Backend - Go)
- **Framework**: `net/http` or `gin-gonic/gin`
- **Endpoints**:
  - `/api/interfaces`: List available network interfaces
  - `/api/capture/start`: Start capture on an interface
  - `/api/capture/stop`: Stop capture
  - `/api/stats/realtime`: WebSocket or SSE for real-time metrics
  - `/api/stats/history`: Historical data
  - `/api/config`: Manage settings and filters

### 4. Web Dashboard (Frontend - HTML/JS/CSS)
- **Tech Stack**: Embedded static files (React/Vue or Vanilla JS with TailwindCSS and Chart.js/ECharts)
- **Features**:
  - Real-time traffic graphs
  - Protocol distribution pie charts
  - Active connections table
  - Interface selection and filter configuration

## Directory Structure
```
GoNetworkMonitor/
├── cmd/
│   └── gonetmon/
│       └── main.go           # Application entry point
├── pkg/
│   ├── capture/              # Packet capture and parsing logic
│   ├── api/                  # REST API and WebSocket handlers
│   ├── stats/                # Metrics aggregation and storage
│   └── config/               # Configuration management
├── web/
│   ├── index.html            # Dashboard UI
│   ├── css/
│   └── js/
├── go.mod
├── go.sum
└── README.md
```

## Development Plan
1. **Phase 1: Backend Core Enhancement**
   - Implement real `gopacket` capture logic
   - Add protocol parsers and statistics aggregator
2. **Phase 2: API and Web Server**
   - Implement REST API
   - Implement WebSocket for real-time updates
   - Embed web assets into the Go binary
3. **Phase 3: Frontend Dashboard**
   - Build a responsive, modern UI
   - Integrate charts and tables
4. **Phase 4: Commercial Features**
   - Add alerting, BPF filtering, and data export
5. **Phase 5: Packaging and Release**
   - Cross-compile binaries for Windows, Linux, macOS
   - Create GitHub Release with assets

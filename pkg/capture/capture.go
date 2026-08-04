package capture

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Ali-Marandi/GoNetworkMonitor/pkg/stats"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// PacketInfo holds parsed information about a single captured packet.
type PacketInfo struct {
	Timestamp   time.Time
	SrcIP       string
	DstIP       string
	SrcPort     uint16
	DstPort     uint16
	Protocol    string
	Length      int
	TCPFlags    string
	DNSQuery    string
	HTTPMethod  string
	HTTPHost    string
}

// Stats holds aggregated statistics for a capture session.
type Stats struct {
	mu sync.RWMutex

	TotalPackets  int64
	TotalBytes    int64
	StartTime     time.Time
	LastUpdated   time.Time

	// Protocol distribution
	ProtocolCounts map[string]int64

	// Top talkers
	SrcIPCounts map[string]int64
	DstIPCounts map[string]int64

	// Port statistics
	SrcPortCounts map[uint16]int64
	DstPortCounts map[uint16]int64

	// Rate tracking (last second)
	PacketsPerSec float64
	BytesPerSec   float64
	lastPackets   int64
	lastBytes     int64
	lastRateTime  time.Time
}

// NewStats creates a new Stats object.
func NewStats() *Stats {
	return &Stats{
		StartTime:      time.Now(),
		LastUpdated:    time.Now(),
		ProtocolCounts: make(map[string]int64),
		SrcIPCounts:    make(map[string]int64),
		DstIPCounts:    make(map[string]int64),
		SrcPortCounts:  make(map[uint16]int64),
		DstPortCounts:  make(map[uint16]int64),
		lastRateTime:   time.Now(),
	}
}

// Update incorporates a new packet into the statistics.
func (s *Stats) Update(pkt PacketInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	atomic.AddInt64(&s.TotalPackets, 1)
	atomic.AddInt64(&s.TotalBytes, int64(pkt.Length))
	s.LastUpdated = pkt.Timestamp

	s.ProtocolCounts[pkt.Protocol]++
	if pkt.SrcIP != "" {
		s.SrcIPCounts[pkt.SrcIP]++
	}
	if pkt.DstIP != "" {
		s.DstIPCounts[pkt.DstIP]++
	}
	if pkt.SrcPort != 0 {
		s.SrcPortCounts[pkt.SrcPort]++
	}
	if pkt.DstPort != 0 {
		s.DstPortCounts[pkt.DstPort]++
	}
}

// UpdateRates calculates per-second rates.
func (s *Stats) UpdateRates() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(s.lastRateTime).Seconds()
	if elapsed > 0 {
		currentPackets := atomic.LoadInt64(&s.TotalPackets)
		currentBytes := atomic.LoadInt64(&s.TotalBytes)
		s.PacketsPerSec = float64(currentPackets-s.lastPackets) / elapsed
		s.BytesPerSec = float64(currentBytes-s.lastBytes) / elapsed
		s.lastPackets = currentPackets
		s.lastBytes = currentBytes
		s.lastRateTime = now
	}
}

// Snapshot returns a thread-safe copy of the current statistics.
func (s *Stats) Snapshot() StatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap := StatsSnapshot{
		TotalPackets:  atomic.LoadInt64(&s.TotalPackets),
		TotalBytes:    atomic.LoadInt64(&s.TotalBytes),
		StartTime:     s.StartTime,
		LastUpdated:   s.LastUpdated,
		PacketsPerSec: s.PacketsPerSec,
		BytesPerSec:   s.BytesPerSec,
		ProtocolCounts: make(map[string]int64),
		TopSrcIPs:     make(map[string]int64),
		TopDstIPs:     make(map[string]int64),
	}

	for k, v := range s.ProtocolCounts {
		snap.ProtocolCounts[k] = v
	}

	// Copy top 20 source IPs
	for k, v := range s.SrcIPCounts {
		snap.TopSrcIPs[k] = v
	}
	for k, v := range s.DstIPCounts {
		snap.TopDstIPs[k] = v
	}

	duration := time.Since(s.StartTime).Seconds()
	if duration > 0 {
		snap.AvgPacketsPerSec = float64(snap.TotalPackets) / duration
		snap.AvgBytesPerSec = float64(snap.TotalBytes) / duration
	}

	return snap
}

// StatsSnapshot is a point-in-time copy of statistics.
type StatsSnapshot struct {
	TotalPackets     int64              `json:"total_packets"`
	TotalBytes       int64              `json:"total_bytes"`
	StartTime        time.Time          `json:"start_time"`
	LastUpdated      time.Time          `json:"last_updated"`
	PacketsPerSec    float64            `json:"packets_per_sec"`
	BytesPerSec      float64            `json:"bytes_per_sec"`
	AvgPacketsPerSec float64            `json:"avg_packets_per_sec"`
	AvgBytesPerSec   float64            `json:"avg_bytes_per_sec"`
	ProtocolCounts   map[string]int64   `json:"protocol_counts"`
	TopSrcIPs        map[string]int64   `json:"top_src_ips"`
	TopDstIPs        map[string]int64   `json:"top_dst_ips"`
}

// Engine manages the packet capture lifecycle.
type Engine struct {
	mu        sync.Mutex
	iface     string
	snapLen   int32
	promisc   bool
	bpfFilter string
	handle    *pcap.Handle
	stats     *Stats
	packets   chan PacketInfo
	stopCh    chan struct{}
	running   bool
	Listeners []chan PacketInfo
}

// NewEngine creates a new capture engine.
func NewEngine(iface string, snapLen int32, promisc bool, bpfFilter string) *Engine {
	return &Engine{
		iface:     iface,
		snapLen:   snapLen,
		promisc:   promisc,
		bpfFilter: bpfFilter,
		stats:     NewStats(),
		packets:   make(chan PacketInfo, 1000),
		Listeners: make([]chan PacketInfo, 0),
	}
}

// Stats returns the statistics object.
func (e *Engine) Stats() *Stats {
	return e.stats
}

// IsRunning returns true if capture is active.
func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

// Start begins packet capture on the configured interface.
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("capture already running")
	}

	handle, err := pcap.OpenLive(e.iface, e.snapLen, e.promisc, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("failed to open interface %s: %w", e.iface, err)
	}

	if e.bpfFilter != "" {
		if err := handle.SetBPFFilter(e.bpfFilter); err != nil {
			handle.Close()
			return fmt.Errorf("failed to set BPF filter: %w", err)
		}
	}

	e.handle = handle
	e.stopCh = make(chan struct{})
	e.running = true
	e.stats = NewStats()

	go e.captureLoop()
	go e.processLoop()
	go e.rateLoop()

	log.Printf("[capture] Started on interface %s (promisc=%v, filter=%q)", e.iface, e.promisc, e.bpfFilter)
	return nil
}

// Stop halts packet capture gracefully.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return
	}
	close(e.stopCh)
	if e.handle != nil {
		e.handle.Close()
	}
	e.running = false
	log.Printf("[capture] Stopped on interface %s", e.iface)
}

// captureLoop reads raw packets from the handle.
func (e *Engine) captureLoop() {
	src := gopacket.NewPacketSource(e.handle, e.handle.LinkType())
	src.NoCopy = true
	for {
		select {
		case <-e.stopCh:
			return
		case pkt, ok := <-src.Packets():
			if !ok {
				return
			}
			info := parsePacket(pkt)
			select {
			case e.packets <- info:
			default:
				// drop if buffer full
			}
		}
	}
}

// processLoop processes packets from the channel.
func (e *Engine) processLoop() {
	for {
		select {
		case <-e.stopCh:
			return
		case pkt := <-e.packets:
			e.stats.Update(pkt)

			// Update Prometheus counters
			stats.RecordPacket(e.iface, pkt.Protocol, pkt.Length)

			// Notify listeners
			for _, ch := range e.Listeners {
				select {
				case ch <- pkt:
				default:
				}
			}
		}
	}
}

// rateLoop periodically updates rate statistics.
func (e *Engine) rateLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.stats.UpdateRates()
		}
	}
}

// parsePacket extracts structured information from a raw gopacket.
func parsePacket(pkt gopacket.Packet) PacketInfo {
	info := PacketInfo{
		Timestamp: pkt.Metadata().Timestamp,
		Length:    pkt.Metadata().Length,
		Protocol:  "Other",
	}

	// Network layer
	if netLayer := pkt.NetworkLayer(); netLayer != nil {
		switch v := netLayer.(type) {
		case *layers.IPv4:
			info.SrcIP = v.SrcIP.String()
			info.DstIP = v.DstIP.String()
		case *layers.IPv6:
			info.SrcIP = v.SrcIP.String()
			info.DstIP = v.DstIP.String()
		}
	}

	// Transport layer
	if transLayer := pkt.TransportLayer(); transLayer != nil {
		switch v := transLayer.(type) {
		case *layers.TCP:
			info.Protocol = "TCP"
			info.SrcPort = uint16(v.SrcPort)
			info.DstPort = uint16(v.DstPort)
			flags := ""
			if v.SYN {
				flags += "SYN "
			}
			if v.ACK {
				flags += "ACK "
			}
			if v.FIN {
				flags += "FIN "
			}
			if v.RST {
				flags += "RST "
			}
			if v.PSH {
				flags += "PSH "
			}
			info.TCPFlags = flags
		case *layers.UDP:
			info.Protocol = "UDP"
			info.SrcPort = uint16(v.SrcPort)
			info.DstPort = uint16(v.DstPort)
		}
	}

	// Application layer
	if appLayer := pkt.ApplicationLayer(); appLayer != nil {
		payload := appLayer.Payload()
		if len(payload) > 4 {
			// Simple HTTP detection
			method := string(payload[:4])
			switch method {
			case "GET ", "POST", "PUT ", "DELE", "HEAD", "OPTI":
				info.Protocol = "HTTP"
				info.HTTPMethod = method
			}
		}
	}

	// DNS layer
	dnsLayer := pkt.Layer(layers.LayerTypeDNS)
	if dnsLayer != nil {
		dns, _ := dnsLayer.(*layers.DNS)
		info.Protocol = "DNS"
		if len(dns.Questions) > 0 {
			info.DNSQuery = string(dns.Questions[0].Name)
		}
	}

	// ICMP
	if icmpLayer := pkt.Layer(layers.LayerTypeICMPv4); icmpLayer != nil {
		info.Protocol = "ICMP"
	}
	if icmpLayer := pkt.Layer(layers.LayerTypeICMPv6); icmpLayer != nil {
		info.Protocol = "ICMPv6"
	}

	// ARP
	if arpLayer := pkt.Layer(layers.LayerTypeARP); arpLayer != nil {
		info.Protocol = "ARP"
	}

	return info
}

// ListInterfaces returns all available network interfaces.
func ListInterfaces() ([]InterfaceInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []InterfaceInfo
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		var addrStrs []string
		for _, a := range addrs {
			addrStrs = append(addrStrs, a.String())
		}
		result = append(result, InterfaceInfo{
			Name:    iface.Name,
			Flags:   iface.Flags.String(),
			Addrs:   addrStrs,
			MTU:     iface.MTU,
			HWAddr:  iface.HardwareAddr.String(),
		})
	}
	return result, nil
}

// InterfaceInfo holds information about a network interface.
type InterfaceInfo struct {
	Name   string   `json:"name"`
	Flags  string   `json:"flags"`
	Addrs  []string `json:"addrs"`
	MTU    int      `json:"mtu"`
	HWAddr string   `json:"hw_addr"`
}

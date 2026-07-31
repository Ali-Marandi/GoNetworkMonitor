package monitor

import (
	"fmt"
	"net"
	"time"
)

type PacketStats struct {
	TotalPackets int64
	TotalBytes   int64
	StartTime    time.Time
}

type Monitor struct {
	Interface string
	Stats     PacketStats
}

func NewMonitor(iface string) *Monitor {
	return &Monitor{
		Interface: iface,
		Stats: PacketStats{
			StartTime: time.Now(),
		},
	}
}

func (m *Monitor) Start() error {
	fmt.Printf("Starting network monitor on interface: %s\n", m.Interface)
	
	// In a real scenario, we would use gopacket here.
	// For this demonstration, we'll simulate packet capturing.
	
	go func() {
		for {
			m.Stats.TotalPackets++
			m.Stats.TotalBytes += 64 // Simulating 64-byte packets
			time.Sleep(100 * time.Millisecond)
		}
	}()
	
	return nil
}

func (m *Monitor) Report() {
	duration := time.Since(m.Stats.StartTime).Seconds()
	fmt.Printf("\n--- Network Report for %s ---\n", m.Interface)
	fmt.Printf("Duration: %.2fs\n", duration)
	fmt.Printf("Total Packets: %d\n", m.Stats.TotalPackets)
	fmt.Printf("Total Bytes: %d\n", m.Stats.TotalBytes)
	if duration > 0 {
		fmt.Printf("Avg Rate: %.2f packets/s\n", float64(m.Stats.TotalPackets)/duration)
	}
}

func ListInterfaces() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	
	var names []string
	for _, i := range ifaces {
		names = append(names, i.Name)
	}
	return names, nil
}

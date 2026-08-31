package alert

import (
	"fmt"
	"sync"
	"time"
)

// Severity represents the severity level of an alert.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Alert represents a triggered alert event.
type Alert struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Severity  Severity  `json:"severity"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Resolved  bool      `json:"resolved"`
}

// Manager manages alert generation and history.
type Manager struct {
	mu        sync.RWMutex
	alerts    []Alert
	maxSize   int
	counter   int
	Listeners []chan Alert
}

// NewManager creates a new alert manager.
func NewManager(maxSize int) *Manager {
	return &Manager{
		alerts:    make([]Alert, 0, maxSize),
		maxSize:   maxSize,
		Listeners: make([]chan Alert, 0),
	}
}

// Trigger creates and stores a new alert.
func (m *Manager) Trigger(severity Severity, title, message string, value, threshold float64) Alert {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	a := Alert{
		ID:        fmt.Sprintf("alert-%d", m.counter),
		Timestamp: time.Now(),
		Severity:  severity,
		Title:     title,
		Message:   message,
		Value:     value,
		Threshold: threshold,
		Resolved:  false,
	}

	if len(m.alerts) >= m.maxSize {
		m.alerts = m.alerts[1:]
	}
	m.alerts = append(m.alerts, a)

	// Notify listeners
	for _, ch := range m.Listeners {
		select {
		case ch <- a:
		default:
		}
	}

	return a
}

// GetAll returns all alerts.
func (m *Manager) GetAll() []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Alert, len(m.alerts))
	copy(result, m.alerts)
	return result
}

// GetRecent returns the last n alerts.
func (m *Manager) GetRecent(n int) []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if n <= 0 {
		return []Alert{}
	}
	if n >= len(m.alerts) {
		result := make([]Alert, len(m.alerts))
		copy(result, m.alerts)
		return result
	}
	result := make([]Alert, n)
	copy(result, m.alerts[len(m.alerts)-n:])
	return result
}

// Checker holds state for threshold checking.
type Checker struct {
	manager            *Manager
	lastBandwidthAlert time.Time
	lastPPSAlert       time.Time
	cooldown            time.Duration
}

// NewChecker creates a new threshold checker.
func NewChecker(manager *Manager) *Checker {
	return &Checker{
		manager:  manager,
		cooldown: 30 * time.Second,
	}
}

// CheckBandwidth checks if bandwidth exceeds the threshold.
func (c *Checker) CheckBandwidth(bytesPerSec, thresholdMbps float64) {
	if time.Since(c.lastBandwidthAlert) < c.cooldown {
		return
	}
	currentMbps := bytesPerSec * 8 / 1_000_000
	if currentMbps > thresholdMbps {
		c.manager.Trigger(
			SeverityWarning,
			"High Bandwidth Usage",
			fmt.Sprintf("Current bandwidth %.2f Mbps exceeds threshold %.2f Mbps", currentMbps, thresholdMbps),
			currentMbps,
			thresholdMbps,
		)
		c.lastBandwidthAlert = time.Now()
	}
}

// CheckPPS checks if packets-per-second exceeds the threshold.
func (c *Checker) CheckPPS(pps, thresholdPPS float64) {
	if time.Since(c.lastPPSAlert) < c.cooldown {
		return
	}
	if pps > thresholdPPS {
		c.manager.Trigger(
			SeverityCritical,
			"High Packet Rate",
			fmt.Sprintf("Current PPS %.0f exceeds threshold %.0f — possible DDoS", pps, thresholdPPS),
			pps,
			thresholdPPS,
		)
		c.lastPPSAlert = time.Now()
	}
}

package stats

import (
	"sync"
	"time"
)

// DataPoint represents a single time-series data point.
type DataPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	PacketsPerSec float64   `json:"packets_per_sec"`
	BytesPerSec   float64   `json:"bytes_per_sec"`
	MbpsIn        float64   `json:"mbps_in"`
	MbpsOut       float64   `json:"mbps_out"`
}

// TimeSeries stores a rolling window of data points.
type TimeSeries struct {
	mu      sync.RWMutex
	points  []DataPoint
	maxSize int
}

// NewTimeSeries creates a new time series with a maximum number of points.
func NewTimeSeries(maxSize int) *TimeSeries {
	return &TimeSeries{
		points:  make([]DataPoint, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a new data point, evicting the oldest if at capacity.
func (ts *TimeSeries) Add(dp DataPoint) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if len(ts.points) >= ts.maxSize {
		ts.points = ts.points[1:]
	}
	ts.points = append(ts.points, dp)
}

// GetAll returns a copy of all data points.
func (ts *TimeSeries) GetAll() []DataPoint {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	result := make([]DataPoint, len(ts.points))
	copy(result, ts.points)
	return result
}

// GetLast returns the last n data points.
func (ts *TimeSeries) GetLast(n int) []DataPoint {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if n <= 0 {
		return []DataPoint{}
	}
	if n >= len(ts.points) {
		result := make([]DataPoint, len(ts.points))
		copy(result, ts.points)
		return result
	}
	result := make([]DataPoint, n)
	copy(result, ts.points[len(ts.points)-n:])
	return result
}

// ConnectionEntry tracks a single observed connection.
type ConnectionEntry struct {
	SrcIP     string    `json:"src_ip"`
	DstIP     string    `json:"dst_ip"`
	SrcPort   uint16    `json:"src_port"`
	DstPort   uint16    `json:"dst_port"`
	Protocol  string    `json:"protocol"`
	Packets   int64     `json:"packets"`
	Bytes     int64     `json:"bytes"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// ConnectionKey uniquely identifies a connection.
type ConnectionKey struct {
	SrcIP   string
	DstIP   string
	SrcPort uint16
	DstPort uint16
	Proto   string
}

// ConnectionTable tracks active connections.
type ConnectionTable struct {
	mu      sync.RWMutex
	entries map[ConnectionKey]*ConnectionEntry
	maxSize int
}

// NewConnectionTable creates a new connection table.
func NewConnectionTable(maxSize int) *ConnectionTable {
	return &ConnectionTable{
		entries: make(map[ConnectionKey]*ConnectionEntry),
		maxSize: maxSize,
	}
}

// Update adds or updates a connection entry.
func (ct *ConnectionTable) Update(key ConnectionKey, bytes int64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if entry, ok := ct.entries[key]; ok {
		entry.Packets++
		entry.Bytes += bytes
		entry.LastSeen = time.Now()
	} else {
		if len(ct.entries) >= ct.maxSize {
			// Evict oldest entry
			var oldest *ConnectionKey
			var oldestTime time.Time
			for k, v := range ct.entries {
				if oldest == nil || v.LastSeen.Before(oldestTime) {
					kCopy := k
					oldest = &kCopy
					oldestTime = v.LastSeen
				}
			}
			if oldest != nil {
				delete(ct.entries, *oldest)
			}
		}
		ct.entries[key] = &ConnectionEntry{
			SrcIP:     key.SrcIP,
			DstIP:     key.DstIP,
			SrcPort:   key.SrcPort,
			DstPort:   key.DstPort,
			Protocol:  key.Proto,
			Packets:   1,
			Bytes:     bytes,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		}
	}
}

// GetAll returns a snapshot of all connections.
func (ct *ConnectionTable) GetAll() []ConnectionEntry {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	result := make([]ConnectionEntry, 0, len(ct.entries))
	for _, v := range ct.entries {
		result = append(result, *v)
	}
	return result
}

// Count returns the number of tracked connections.
func (ct *ConnectionTable) Count() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return len(ct.entries)
}

package alert

import (
	"fmt"
	"math"
	"sync"
)

// AnomalyDetector uses moving averages to detect traffic spikes.
type AnomalyDetector struct {
	mu            sync.Mutex
	manager       *Manager
	history       []float64
	windowSize    int
	threshold     float64 // Standard deviations from mean
}

// NewAnomalyDetector creates a new detector.
func NewAnomalyDetector(mgr *Manager, windowSize int, threshold float64) *AnomalyDetector {
	return &AnomalyDetector{
		manager:    mgr,
		windowSize: windowSize,
		threshold:  threshold,
		history:    make([]float64, 0, windowSize),
	}
}

// Observe adds a new data point and checks for anomalies.
func (ad *AnomalyDetector) Observe(value float64) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	if len(ad.history) < 10 {
		ad.history = append(ad.history, value)
		return
	}

	mean, stdDev := ad.stats()
	if value > mean+(ad.threshold*stdDev) && value > 100 { // Only alert for significant values
		ad.manager.Trigger(
			SeverityCritical,
			"Anomaly Detected",
			fmt.Sprintf("Traffic spike detected: %.2f is significantly above normal (mean: %.2f, stdDev: %.2f)", value, mean, stdDev),
			value,
			mean+(ad.threshold*stdDev),
		)
	}

	if len(ad.history) >= ad.windowSize {
		ad.history = ad.history[1:]
	}
	ad.history = append(ad.history, value)
}

func (ad *AnomalyDetector) stats() (mean, stdDev float64) {
	sum := 0.0
	for _, v := range ad.history {
		sum += v
	}
	mean = sum / float64(len(ad.history))

	sqDiffSum := 0.0
	for _, v := range ad.history {
		sqDiffSum += math.Pow(v-mean, 2)
	}
	stdDev = math.Sqrt(sqDiffSum / float64(len(ad.history)))
	return
}

package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Config holds the application configuration.
type Config struct {
	mu *sync.RWMutex

	// Server settings
	ListenAddr string `json:"listen_addr"`
	Port       int    `json:"port"`

	// Capture settings
	Interface   string `json:"interface"`
	SnapLen     int32  `json:"snap_len"`
	Promiscuous bool   `json:"promiscuous"`
	BPFFilter   string `json:"bpf_filter"`

	// Alert thresholds
	Alerts AlertConfig `json:"alerts"`

	// Storage
	DataDir    string `json:"data_dir"`
	MaxHistory int    `json:"max_history_seconds"`
}

// AlertConfig holds threshold-based alert settings.
type AlertConfig struct {
	Enabled          bool    `json:"enabled"`
	BandwidthMbps    float64 `json:"bandwidth_mbps_threshold"`
	PacketsPerSecond float64 `json:"pps_threshold"`
	NewConnPerSecond float64 `json:"new_conn_per_second_threshold"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		mu:          &sync.RWMutex{},
		ListenAddr:  "127.0.0.1",
		Port:        8080,
		Interface:   "auto",
		SnapLen:     65535,
		Promiscuous: true,
		BPFFilter:   "",
		Alerts: AlertConfig{
			Enabled:          true,
			BandwidthMbps:    100,
			PacketsPerSecond: 10000,
			NewConnPerSecond: 500,
		},
		DataDir:    "./data",
		MaxHistory: 3600,
	}
}

func (c *Config) mutex() *sync.RWMutex {
	if c.mu == nil {
		c.mu = &sync.RWMutex{}
	}
	return c.mu
}

// Load reads configuration from a JSON file.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save writes the current configuration to a JSON file.
func (c *Config) Save(path string) error {
	mu := c.mutex()
	mu.RLock()
	defer mu.RUnlock()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(c)
}

// Get returns a copy of the current config safely.
func (c *Config) Get() Config {
	mu := c.mutex()
	mu.RLock()
	defer mu.RUnlock()
	return *c
}

// Update applies a new configuration atomically while preserving the lock.
func (c *Config) Update(newCfg Config) {
	mu := c.mutex()
	mu.Lock()
	defer mu.Unlock()
	newCfg.mu = c.mu
	*c = newCfg
}

// SetInterface updates the capture interface without copying the lock.
func (c *Config) SetInterface(name string) {
	mu := c.mutex()
	mu.Lock()
	defer mu.Unlock()
	c.Interface = name
}

// SetPort updates the HTTP port without copying the lock.
func (c *Config) SetPort(port int) {
	mu := c.mutex()
	mu.Lock()
	defer mu.Unlock()
	c.Port = port
}

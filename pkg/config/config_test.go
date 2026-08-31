package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ListenAddr != "127.0.0.1" {
		t.Fatalf("ListenAddr = %q, want 127.0.0.1", cfg.ListenAddr)
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want 8080", cfg.Port)
	}
}

func TestConfigUpdateAndSave(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SetInterface("eth0")
	cfg.SetPort(9090)

	got := cfg.Get()
	if got.Interface != "eth0" || got.Port != 9090 {
		t.Fatalf("unexpected config after setters: %+v", got)
	}

	path := filepath.Join(t.TempDir(), "config.json")
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Interface != "eth0" || loaded.Port != 9090 {
		t.Fatalf("unexpected loaded config: %+v", loaded.Get())
	}
}

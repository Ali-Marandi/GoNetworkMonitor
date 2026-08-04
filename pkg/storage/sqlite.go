package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

// DB manages the persistent SQLite storage.
type DB struct {
	conn *sql.DB
}

// NewDB initializes a new SQLite database.
func NewDB(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dataDir, "network_monitor.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

func (db *DB) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS stats_history (
			id INTEGER PRIMARY KEY,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			packets_per_sec REAL,
			bytes_per_sec REAL,
			mbps REAL
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			timestamp DATETIME,
			severity TEXT,
			title TEXT,
			message TEXT,
			value REAL,
			threshold REAL
		)`,
		`CREATE TABLE IF NOT EXISTS top_ips (
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			ip TEXT,
			direction TEXT, -- 'src' or 'dst'
			packet_count INTEGER
		)`,
	}

	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}
	return nil
}

// SaveStats persists a time-series data point.
func (db *DB) SaveStats(pps, bps, mbps float64) error {
	_, err := db.conn.Exec(
		"INSERT INTO stats_history (packets_per_sec, bytes_per_sec, mbps) VALUES (?, ?, ?)",
		pps, bps, mbps,
	)
	return err
}

// SaveAlert persists an alert event.
func (db *DB) SaveAlert(id, severity, title, message string, timestamp time.Time, value, threshold float64) error {
	_, err := db.conn.Exec(
		"INSERT INTO alerts (id, timestamp, severity, title, message, value, threshold) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, timestamp, severity, title, message, value, threshold,
	)
	return err
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetHistory returns historical stats for a given duration.
func (db *DB) GetHistory(duration time.Duration) ([]map[string]interface{}, error) {
	rows, err := db.conn.Query(
		"SELECT timestamp, packets_per_sec, bytes_per_sec, mbps FROM stats_history WHERE timestamp > ?",
		time.Now().Add(-duration),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var ts time.Time
		var pps, bps, mbps float64
		if err := rows.Scan(&ts, &pps, &bps, &mbps); err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"timestamp":       ts,
			"packets_per_sec": pps,
			"bytes_per_sec":   bps,
			"mbps":            mbps,
		})
	}
	return result, nil
}

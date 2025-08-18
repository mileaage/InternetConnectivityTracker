package db

import (
	"database/sql"
	"fmt"
	"time"

	"WifiTracker/internals/monitor"

	_ "github.com/mattn/go-sqlite3"
)

// storage interface
type Storage interface {
	LogConnectivityCheck(deviceID string, success bool, responseTime time.Duration, timestamp time.Time, err error) error
	LogStatusChange(deviceID string, from, to monitor.ConnectionStatus, timestamp time.Time) error
	LogOutageStart(deviceID string, timestamp time.Time) error
	LogOutageEnd(deviceID string, duration time.Duration, timestamp time.Time) error
	Close() error
}

type DatabaseStorage struct {
	db    *sql.DB
	stmts map[string]*sql.Stmt
}

func NewDatabaseStorage(dbPath string) (*DatabaseStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &DatabaseStorage{
		db:    db,
		stmts: make(map[string]*sql.Stmt),
	}

	// databse schema
	if err := storage.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// map statements
	if err := storage.prepareStatements(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return storage, nil
}

func (d *DatabaseStorage) runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS connectivity_checks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id TEXT NOT NULL,
			success BOOLEAN NOT NULL,
			response_time INTEGER, -- milliseconds
			timestamp DATETIME NOT NULL,
			error TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS status_changes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id TEXT NOT NULL,
			from_status TEXT NOT NULL,
			to_status TEXT NOT NULL,
			timestamp DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS outages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			duration INTEGER -- milliseconds
		)`,
		`CREATE INDEX IF NOT EXISTS idx_connectivity_device_time ON connectivity_checks(device_id, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_status_device_time ON status_changes(device_id, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_outages_device_time ON outages(device_id, start_time)`,
	}

	for _, migration := range migrations {
		if _, err := d.db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}
	return nil
}

func (d *DatabaseStorage) prepareStatements() error {
	statements := map[string]string{
		"insertConnectivityCheck": `INSERT INTO connectivity_checks (device_id, success, response_time, timestamp, error) VALUES (?, ?, ?, ?, ?)`,
		"insertStatusChange":      `INSERT INTO status_changes (device_id, from_status, to_status, timestamp) VALUES (?, ?, ?, ?)`,
		"insertOutageStart":       `INSERT INTO outages (device_id, start_time) VALUES (?, ?)`,
		"updateOutageEnd":         `UPDATE outages SET end_time = ?, duration = ? WHERE device_id = ? AND end_time IS NULL`,
	}

	for name, query := range statements {
		stmt, err := d.db.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement %s: %w", name, err)
		}
		d.stmts[name] = stmt
	}
	return nil
}

func (d *DatabaseStorage) LogConnectivityCheck(deviceID string, success bool, responseTime time.Duration, timestamp time.Time, err error) error {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	responseTimeMs := responseTime.Milliseconds()

	_, execErr := d.stmts["insertConnectivityCheck"].Exec(
		deviceID,
		success,
		responseTimeMs,
		timestamp,
		errStr,
	)

	return execErr
}

func (d *DatabaseStorage) LogStatusChange(deviceID string, from, to monitor.ConnectionStatus, timestamp time.Time) error {
	_, err := d.stmts["insertStatusChange"].Exec(
		deviceID,
		from.String(),
		to.String(),
		timestamp,
	)
	return err
}

func (d *DatabaseStorage) LogOutageStart(deviceID string, timestamp time.Time) error {
	_, err := d.stmts["insertOutageStart"].Exec(deviceID, timestamp)
	return err
}

func (d *DatabaseStorage) LogOutageEnd(deviceID string, duration time.Duration, timestamp time.Time) error {
	durationMs := duration.Milliseconds()
	_, err := d.stmts["updateOutageEnd"].Exec(timestamp, durationMs, deviceID)
	return err
}

func (d *DatabaseStorage) Close() error {
	for _, stmt := range d.stmts {
		stmt.Close()
	}

	return d.db.Close()
}

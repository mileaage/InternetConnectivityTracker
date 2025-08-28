package db

import (
	"database/sql"
	"time"
)

type DowntimeEvent struct {
	ID        int
	DeviceID  string
	StartTime time.Time
	EndTime   sql.NullTime
	Duration  sql.NullInt64
}

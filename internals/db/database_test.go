package db

import (
	"WifiTracker/util"
	"os"
	"testing"
	"time"
)

func TestDowntimeSpan(t *testing.T) {
	tempFile, err := os.CreateTemp("", "tempdatabase-*.db")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}

	defer os.Remove(tempFile.Name()) // clean up

	dbStorage, err := NewDatabaseStorage(tempFile.Name())
	if err != nil {
		t.Fatalf("Error creating database: %v", err)
	}

	dbStorage.LogOutageStart("bbed78db-4aa8-46bc-930e-e689aabf5eb0", time.Now())
	dbStorage.LogOutageEnd("bbed78db-4aa8-46bc-930e-e689aabf5eb0", time.Minute, time.Now())

	_, err = dbStorage.GetDowntimes(util.OneDayAgo())
	if err != nil {
		t.Fatalf("Error fetching dowtimes: %v", err)
	}

}

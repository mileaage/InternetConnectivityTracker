/*
	The primary use of this file to fetch the downtimes so that we
	can put it on the website
*/

package db

import (
	"time"
)

func GetDowntimes(timespan time.Time) ([]DowntimeEvent, error) {
	dbStorage, err := NewDatabaseStorage(`downtimedata.db`)
	if err != nil {
		return nil, err
	}

	defer dbStorage.Close()

	results, err := dbStorage.GetDowntimes(timespan)
	if err != nil {
		return nil, err
	}

	return results, nil
}

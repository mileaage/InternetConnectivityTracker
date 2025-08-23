package util

import "time"

func OneDayAgo() time.Time {
	return time.Now().Add(-24 * time.Hour)
}

func OneWeekAgo() time.Time {
	return time.Now().Add(-7 * 24 * time.Hour)
}

func OneMonthAgo() time.Time {
	return time.Now().Add(-31 * 24 * time.Hour)
}

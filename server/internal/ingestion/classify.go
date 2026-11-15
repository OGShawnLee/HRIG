package ingestion

import "hrig/server/internal/db"

// ClassifyReading classifies a reading value against metric_threshold limits.
// Critical bounds (inclusive) are evaluated before warning bounds (exclusive).
// For "manual" mode, it always returns normal.
func ClassifyReading(value float64, mode string, warnMin, warnMax, critMin, critMax *float64) db.ReadingStatus {
	if mode == "manual" {
		return db.ReadingStatusNormal
	}

	if critMin != nil && value <= *critMin {
		return db.ReadingStatusCritical
	}
	if critMax != nil && value >= *critMax {
		return db.ReadingStatusCritical
	}

	if warnMin != nil && value < *warnMin {
		return db.ReadingStatusAbnormal
	}
	if warnMax != nil && value > *warnMax {
		return db.ReadingStatusAbnormal
	}

	return db.ReadingStatusNormal
}

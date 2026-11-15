package realtime

import (
	"fmt"
	"time"
)

// FormatMET formats Mission Elapsed Time relative to epochUTC as "T+000d HH:MM:SS"
// (or "T-000d HH:MM:SS" if recordedAtUTC precedes epochUTC).
func FormatMET(recordedAtUTC, epochUTC time.Time) string {
	diff := recordedAtUTC.UTC().Sub(epochUTC.UTC())
	sign := "+"
	if diff < 0 {
		sign = "-"
		diff = -diff
	}

	totalSeconds := int64(diff / time.Second)
	days := totalSeconds / 86400
	rem := totalSeconds % 86400
	hours := rem / 3600
	rem %= 3600
	minutes := rem / 60
	seconds := rem % 60

	return fmt.Sprintf("T%s%03dd %02d:%02d:%02d", sign, days, hours, minutes, seconds)
}

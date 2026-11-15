package ingestion

import (
	"testing"

	"hrig/server/internal/db"
)

func floatPtr(v float64) *float64 {
	return &v
}

func TestClassifyReading_HeartRateBoundaries(t *testing.T) {
	warnMin := floatPtr(50)
	warnMax := floatPtr(100)
	critMin := floatPtr(40)
	critMax := floatPtr(200)

	cases := []struct {
		name  string
		value float64
		want  db.ReadingStatus
	}{
		{name: "critical_min_inclusive_40", value: 40, want: db.ReadingStatusCritical},
		{name: "below_critical_min_39", value: 39, want: db.ReadingStatusCritical},
		{name: "abnormal_low_41", value: 41, want: db.ReadingStatusAbnormal},
		{name: "abnormal_low_49_9", value: 49.9, want: db.ReadingStatusAbnormal},
		{name: "warning_min_exclusive_50", value: 50, want: db.ReadingStatusNormal},
		{name: "normal_mid_60", value: 60, want: db.ReadingStatusNormal},
		{name: "warning_max_exclusive_100", value: 100, want: db.ReadingStatusNormal},
		{name: "abnormal_high_101", value: 101, want: db.ReadingStatusAbnormal},
		{name: "abnormal_high_199", value: 199, want: db.ReadingStatusAbnormal},
		{name: "critical_max_inclusive_200", value: 200, want: db.ReadingStatusCritical},
		{name: "above_critical_max_220", value: 220, want: db.ReadingStatusCritical},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyReading(tc.value, "numeric", warnMin, warnMax, critMin, critMax)
			if got != tc.want {
				t.Fatalf("ClassifyReading(%v) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestClassifyReading_ManualMode(t *testing.T) {
	warnMin := floatPtr(50)
	warnMax := floatPtr(100)
	critMin := floatPtr(40)
	critMax := floatPtr(200)

	if got := ClassifyReading(250, "manual", warnMin, warnMax, critMin, critMax); got != db.ReadingStatusNormal {
		t.Fatalf("manual mode with bounds returned %q, want %q", got, db.ReadingStatusNormal)
	}
	if got := ClassifyReading(0, "manual", nil, nil, nil, nil); got != db.ReadingStatusNormal {
		t.Fatalf("manual mode without bounds returned %q, want %q", got, db.ReadingStatusNormal)
	}
}

func TestClassifyReading_OneSidedAndScaleBounds(t *testing.T) {
	// spo2: warning_min=94, critical_min=90, no max
	if got := ClassifyReading(90, "numeric", floatPtr(94), nil, floatPtr(90), nil); got != db.ReadingStatusCritical {
		t.Fatalf("spo2=90 got %q, want critical", got)
	}
	if got := ClassifyReading(93, "numeric", floatPtr(94), nil, floatPtr(90), nil); got != db.ReadingStatusAbnormal {
		t.Fatalf("spo2=93 got %q, want abnormal", got)
	}
	if got := ClassifyReading(98, "numeric", floatPtr(94), nil, floatPtr(90), nil); got != db.ReadingStatusNormal {
		t.Fatalf("spo2=98 got %q, want normal", got)
	}

	// mood: warning 2..4, critical 1..5
	if got := ClassifyReading(1, "numeric", floatPtr(2), floatPtr(4), floatPtr(1), floatPtr(5)); got != db.ReadingStatusCritical {
		t.Fatalf("mood=1 got %q, want critical", got)
	}
	if got := ClassifyReading(4, "numeric", floatPtr(2), floatPtr(4), floatPtr(1), floatPtr(5)); got != db.ReadingStatusNormal {
		t.Fatalf("mood=4 got %q, want normal", got)
	}
	if got := ClassifyReading(5, "numeric", floatPtr(2), floatPtr(4), floatPtr(1), floatPtr(5)); got != db.ReadingStatusCritical {
		t.Fatalf("mood=5 got %q, want critical", got)
	}
}

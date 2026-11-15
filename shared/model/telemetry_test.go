package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEnumValidation(t *testing.T) {
	validStatuses := []ReadingStatus{StatusNormal, StatusAbnormal, StatusCritical, StatusStale}
	for _, s := range validStatuses {
		if !s.IsValid() {
			t.Fatalf("expected ReadingStatus %q to be valid", s)
		}
	}

	invalidStatuses := []ReadingStatus{"", "nominal", "caution", "NORMAL"}
	for _, s := range invalidStatuses {
		if s.IsValid() {
			t.Fatalf("expected ReadingStatus %q to be invalid", s)
		}
	}

	if !StatusNormal.IsPersistable() || !StatusAbnormal.IsPersistable() || !StatusCritical.IsPersistable() {
		t.Fatal("expected normal/abnormal/critical to be persistable")
	}
	if StatusStale.IsPersistable() {
		t.Fatal("expected stale to NOT be persistable")
	}

	if !EnvIVA.IsValid() || !EnvEVA.IsValid() {
		t.Fatal("expected iva and eva to be valid EnvironmentMode")
	}
	if EnvironmentMode("moon").IsValid() || EnvironmentMode("").IsValid() {
		t.Fatal("expected unknown EnvironmentMode to be invalid")
	}

	if !TierPrimary.IsValid() || !TierSecondary.IsValid() {
		t.Fatal("expected primary and secondary to be valid MetricTier")
	}
	if !ValueNumeric.IsValid() || !ValueScale.IsValid() || !ValueText.IsValid() || !ValueImage.IsValid() {
		t.Fatal("expected numeric/scale/text/image to be valid ValueType")
	}
	if !DeviceWearable.IsValid() || !DeviceEnvironmentSensor.IsValid() || !DeviceManual.IsValid() || !DeviceLabAssay.IsValid() || !DeviceImaging.IsValid() {
		t.Fatal("expected all DeviceKind constants to be valid")
	}
	if !MediaUltrasound.IsValid() || !MediaOcular.IsValid() {
		t.Fatal("expected ultrasound and ocular to be valid MediaKind")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 11, 15, 18, 0, 0, 0, time.UTC)
	astroID := "11111111-1111-1111-1111-111111111111"
	unitCode := "/min"
	val := 72.5
	note := "resting"
	ackBy := "CMO-1"
	label := "United States"

	astronaut := Astronaut{
		ID:               astroID,
		FullName:         "Elena Vance",
		BirthDate:        time.Date(1990, 5, 10, 0, 0, 0, 0, time.UTC),
		Age:              36,
		NationalityCode:  "US",
		NationalityLabel: &label,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	assertRoundTrip(t, astronaut)

	wMin := 50.0
	wMax := 100.0
	cMin := 40.0
	cMax := 200.0
	metric := Metric{
		ID:           "22222222-2222-2222-2222-222222222222",
		Code:         "heart_rate",
		Name:         "Heart rate",
		Category:     "cardiac",
		Tier:         TierPrimary,
		MetricType:   1,
		ValueType:    ValueNumeric,
		BaseUnitCode: &unitCode,
		DisplayOrder: 11,
		IsActive:     true,
		Threshold: &MetricThreshold{
			ID:                 "33333333-3333-3333-3333-333333333333",
			MetricID:           "22222222-2222-2222-2222-222222222222",
			MetricCode:         "heart_rate",
			WarningMin:         &wMin,
			WarningMax:         &wMax,
			CriticalMin:        &cMin,
			CriticalMax:        &cMax,
			ClassificationMode: "numeric",
			Source:             "test",
			UpdatedAt:          now,
		},
	}
	assertRoundTrip(t, metric)

	frame := TelemetryFrame{
		Source:        "wearable",
		DeviceCode:    "biomonitor_01",
		AstronautID:   &astroID,
		Environment:   EnvIVA,
		MetricType:    1,
		MetricCode:    "heart_rate",
		UnitCode:      &unitCode,
		ValueNumeric:  &val,
		Status:        StatusNormal,
		RecordedAtUTC: now,
		MET:           "T+001d 02:00:00",
		Note:          &note,
	}
	latest := LatestReading{
		TelemetryFrame: frame,
		LastSeenUTC:    now,
		IsStale:        false,
	}
	assertRoundTrip(t, latest)

	alert := AlertEvent{
		ID:                "alert-1",
		AstronautID:       &astroID,
		Source:            "wearable",
		DeviceCode:        "biomonitor_01",
		Environment:       EnvIVA,
		MetricCode:        "heart_rate",
		ValueNumeric:      &val,
		UnitCode:          &unitCode,
		Status:            StatusCritical,
		Message:           "Critical heart_rate",
		TriggeredAtUTC:    now,
		MET:               "T+001d 02:00:00",
		Acknowledged:      true,
		AcknowledgedBy:    &ackBy,
		AcknowledgedAtUTC: &now,
	}
	env := RealtimeEnvelope{
		Type:      "snapshot",
		Readings:  []LatestReading{latest},
		Alerts:    []AlertEvent{alert},
		Reading:   &latest,
		Alert:     &alert,
		SentAtUTC: now,
	}
	assertRoundTrip(t, env)

	checkin := CheckInRequest{
		AstronautID:   &astroID,
		Environment:   EnvIVA,
		MetricCode:    "ultrasound",
		ValueText:     &note,
		Status:        &frame.Status,
		RecordedAtUTC: &now,
		Media: []CheckInMedia{
			{
				Kind:          MediaUltrasound,
				FilePath:      "media/ultrasound/test.png",
				MimeType:      "image/png",
				Findings:      &note,
				CapturedAtUTC: &now,
			},
		},
	}
	assertRoundTrip(t, checkin)

	health := HealthStatus{
		Status:          "ok",
		Store:           "memory",
		TimeUTC:         now,
		ActiveAlerts:    1,
		RealtimeClients: 2,
		IngestClients:   1,
	}
	assertRoundTrip(t, health)
}

func assertRoundTrip[T any](t *testing.T, v T) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %T: %v", v, err)
	}
	var out T
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal %T: %v", v, err)
	}
	b2, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("re-marshal %T: %v", v, err)
	}
	if string(b) != string(b2) {
		t.Fatalf("round-trip mismatch for %T:\ngot:  %s\nwant: %s", v, string(b2), string(b))
	}
}

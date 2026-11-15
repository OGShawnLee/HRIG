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
	if !EscalationSelfCare.IsValid() || !EscalationCMO.IsValid() || !EscalationFlightSurgeon.IsValid() {
		t.Fatal("expected self_care, cmo, and flight_surgeon to be valid EscalationLevel")
	}
	if EscalationLevel("").IsValid() || EscalationLevel("physician").IsValid() {
		t.Fatal("expected unknown EscalationLevel to be invalid")
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
	unitSymbol := "bpm"
	unitDim := "rate"
	var unitDecimals int16 = 0
	displaySym := "bpm"
	baseUnitID := "44444444-4444-4444-4444-444444444444"
	metric := Metric{
		ID:                "22222222-2222-2222-2222-222222222222",
		Code:              "heart_rate",
		Name:              "Heart rate",
		Category:          "cardiac",
		Tier:              TierPrimary,
		MetricType:        1,
		ValueType:         ValueNumeric,
		BaseUnitID:        &baseUnitID,
		BaseUnitCode:      &unitCode,
		BaseUnitSymbol:    &unitSymbol,
		BaseUnitDimension: &unitDim,
		BaseUnitDecimals:  &unitDecimals,
		DisplayUnitSymbol: &displaySym,
		DisplayOrder:      11,
		IsActive:          true,
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

	trend := TrendPoint{
		ID:            "55555555-5555-5555-5555-555555555555",
		MeasurementID: "66666666-6666-6666-6666-666666666666",
		MetricID:      metric.ID,
		MetricCode:    "heart_rate",
		AstronautID:   &astroID,
		DeviceID:      "77777777-7777-7777-7777-777777777777",
		Environment:   EnvIVA,
		UnitCode:      &unitCode,
		UnitSymbol:    &unitSymbol,
		ValueNumeric:  &val,
		Status:        StatusNormal,
		RecordedAtUTC: now,
		MET:           "T+001d 02:00:00",
	}
	assertRoundTrip(t, trend)

	astroName := "Elena Vance"
	measNote := "Daily resting health check"
	valNote := "Within baseline"
	mediaID := "88888888-8888-8888-8888-888888888888"
	mediaKind := MediaUltrasound
	mediaPath := "media/ultrasound/test.png"
	mediaMime := "image/png"
	mediaFindings := "No anomalies"
	var scaleVal int16 = 4
	textVal := "clear scan"

	historyNumeric := HistoryEntry{
		MeasurementID:     "66666666-6666-6666-6666-666666666666",
		ValueID:           "55555555-5555-5555-5555-555555555555",
		AstronautID:       &astroID,
		AstronautName:     &astroName,
		DeviceCode:        "biomonitor_01",
		DeviceName:        "IVA biomonitor",
		Environment:       EnvIVA,
		MetricCode:        "heart_rate",
		MetricName:        "Heart rate",
		ValueType:         ValueNumeric,
		UnitCode:          &unitCode,
		UnitSymbol:        &unitSymbol,
		DisplayUnitSymbol: &displaySym,
		ValueNumeric:      &val,
		Status:            StatusNormal,
		RecordedAtUTC:     now,
		MET:               "T+001d 02:00:00",
		MeasurementNote:   &measNote,
		ValueNote:         &valNote,
	}
	assertRoundTrip(t, historyNumeric)

	historyMedia := HistoryEntry{
		MeasurementID:      "99999999-9999-9999-9999-999999999999",
		ValueID:            "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		AstronautID:        &astroID,
		AstronautName:      &astroName,
		DeviceCode:         "ultrasound_01",
		DeviceName:         "Diagnostic ultrasound",
		Environment:        EnvIVA,
		MetricCode:         "ultrasound",
		MetricName:         "Diagnostic ultrasound",
		ValueType:          ValueImage,
		ValueScale:         &scaleVal,
		ValueText:          &textVal,
		Status:             StatusNormal,
		RecordedAtUTC:      now,
		MET:                "T+001d 02:00:00",
		MediaAssetID:       &mediaID,
		MediaKind:          &mediaKind,
		MediaFilePath:      &mediaPath,
		MediaMimeType:      &mediaMime,
		MediaFindings:      &mediaFindings,
		MediaCapturedAtUTC: &now,
	}
	assertRoundTrip(t, historyMedia)

	historyQuery := HistoryQuery{
		AstronautID: astroID,
		Limit:       50,
		Offset:      25,
	}
	assertRoundTrip(t, historyQuery)

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

	hazard := SpaceHazard{
		ID:                 "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		Code:               "space_radiation",
		Name:               "Space Radiation",
		Summary:            "Ionizing radiation beyond Earth's magnetosphere.",
		ClinicalImpact:     "DNA damage and immune system changes.",
		RelatedMetricCodes: []string{"radiation_rate", "radiation_dose_cum"},
		Source:             "NASA HRP Five Hazards",
		DisplayOrder:       1,
	}
	assertRoundTrip(t, hazard)

	envIVA := EnvIVA
	proc := MedicalProcedure{
		ID:                 "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
		Code:               "hypoxia_cabin_o2",
		Title:              "Hypoxia & Low Cabin O2 Response",
		Category:           "respiratory_cabin",
		BodySystem:         "respiratory",
		Summary:            "Operational checklist for low cabin O2.",
		WhenToUse:          "When cabin O2 drops or SpO2 trends downward.",
		Escalation:         EscalationCMO,
		EstimatedMinutes:   10,
		Environment:        &envIVA,
		Equipment:          []string{"Pulse oximeter"},
		Steps:              []string{"Check cabin O2 sensor", "Verify SpO2"},
		Warnings:           []string{"Escalate to Flight Surgeon if cyanosis appears"},
		RelatedMetricCodes: []string{"cabin_o2", "spo2"},
		HazardCodes:        []string{"hostile_closed_env"},
		Keywords:           []string{"hypoxia", "shortness of breath"},
		Source:             "NASA HRP / ISS Operational Checklist (demo - clinical review required)",
		DisplayOrder:       1,
	}
	assertRoundTrip(t, proc)

	searchRes := SearchResults{
		Query:      "oxygen",
		Tab:        "all",
		Astronauts: []Astronaut{astronaut},
		Procedures: []MedicalProcedure{proc},
		Hazards:    []SpaceHazard{hazard},
		Metrics:    []Metric{metric},
	}
	assertRoundTrip(t, searchRes)
}

func TestMedicalKnowledgeNormalization(t *testing.T) {
	var res SearchResults
	res.Query = "spo2"
	res.Tab = "all"
	res.Procedures = []MedicalProcedure{{Code: "hypoxia_cabin_o2"}}
	res.Hazards = []SpaceHazard{{Code: "hostile_closed_env"}}
	res.Normalize()

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal normalized SearchResults: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal SearchResults map: %v", err)
	}
	for _, key := range []string{"astronauts", "procedures", "hazards", "metrics"} {
		arr, ok := raw[key].([]any)
		if !ok || arr == nil {
			t.Fatalf("expected non-nil JSON array for %q, got %#v", key, raw[key])
		}
	}
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

func TestIngestDTOJSONContract(t *testing.T) {
	now := time.Date(2026, 11, 15, 8, 20, 0, 0, time.UTC)
	astro := "Ana Torres"
	numVal := 101.3
	scaleVal := 4
	textVal := "simulated normal ultrasound finding"
	kpaUnit := "kPa"
	scoreUnit := "{score}"

	fullReq := IngestRequest{
		DeviceCode:    "biomonitor_01",
		Environment:   EnvIVA,
		RecordedAtUTC: now.Format(time.RFC3339Nano),
		AstronautName: &astro,
		Readings: []IngestReading{
			{
				MetricCode:   "cabin_pressure",
				ValueNumeric: &numVal,
				UnitCode:     &kpaUnit,
			},
			{
				MetricCode: "mood",
				ValueScale: &scaleVal,
				UnitCode:   &scoreUnit,
			},
			{
				MetricCode: "ultrasound",
				ValueText:  &textVal,
			},
		},
	}
	assertRoundTrip(t, fullReq)

	rawReq, err := json.Marshal(fullReq)
	if err != nil {
		t.Fatalf("marshal full IngestRequest: %v", err)
	}
	var reqMap map[string]any
	if err := json.Unmarshal(rawReq, &reqMap); err != nil {
		t.Fatalf("unmarshal IngestRequest map: %v", err)
	}
	for _, key := range []string{"device_code", "environment", "recorded_at_utc", "astronaut_name", "readings"} {
		if _, ok := reqMap[key]; !ok {
			t.Fatalf("expected wire key %q in IngestRequest JSON: %s", key, string(rawReq))
		}
	}

	readingsSlice, ok := reqMap["readings"].([]any)
	if !ok || len(readingsSlice) != 3 {
		t.Fatalf("expected 3 readings in wire JSON, got %T", reqMap["readings"])
	}
	r0 := readingsSlice[0].(map[string]any)
	for _, key := range []string{"metric_code", "value_numeric", "unit_code"} {
		if _, ok := r0[key]; !ok {
			t.Fatalf("expected wire key %q in numeric reading JSON", key)
		}
	}
	r1 := readingsSlice[1].(map[string]any)
	for _, key := range []string{"metric_code", "value_scale", "unit_code"} {
		if _, ok := r1[key]; !ok {
			t.Fatalf("expected wire key %q in scale reading JSON", key)
		}
	}
	r2 := readingsSlice[2].(map[string]any)
	for _, key := range []string{"metric_code", "value_text"} {
		if _, ok := r2[key]; !ok {
			t.Fatalf("expected wire key %q in text/image reading JSON", key)
		}
	}
	if _, exists := r2["unit_code"]; exists {
		t.Fatalf("expected nil unit_code to be omitted from wire JSON, got %s", string(rawReq))
	}

	envReq := IngestRequest{
		DeviceCode:    "cabin_sensor_01",
		Environment:   EnvEVA,
		RecordedAtUTC: now.Format(time.RFC3339Nano),
		Readings: []IngestReading{
			{
				MetricCode:   "cabin_pressure",
				ValueNumeric: &numVal,
				UnitCode:     &kpaUnit,
			},
		},
	}
	assertRoundTrip(t, envReq)

	rawEnvReq, err := json.Marshal(envReq)
	if err != nil {
		t.Fatalf("marshal env IngestRequest: %v", err)
	}
	var envMap map[string]any
	if err := json.Unmarshal(rawEnvReq, &envMap); err != nil {
		t.Fatalf("unmarshal env IngestRequest map: %v", err)
	}
	if _, exists := envMap["astronaut_name"]; exists {
		t.Fatalf("expected nil astronaut_name to be omitted from wire JSON, got %s", string(rawEnvReq))
	}

	res := IngestResult{
		MeasurementID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		RecordedAtUTC: now,
		ValuesCount:   3,
	}
	assertRoundTrip(t, res)

	rawRes, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal IngestResult: %v", err)
	}
	var resMap map[string]any
	if err := json.Unmarshal(rawRes, &resMap); err != nil {
		t.Fatalf("unmarshal IngestResult map: %v", err)
	}
	for _, key := range []string{"measurement_id", "recorded_at_utc", "values_count"} {
		if _, ok := resMap[key]; !ok {
			t.Fatalf("expected wire key %q in IngestResult JSON: %s", key, string(rawRes))
		}
	}
}


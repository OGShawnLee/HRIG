package realtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"hrig/server/internal/db"
	"hrig/shared/model"
)

func TestFormatMET(t *testing.T) {
	epoch := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		ts   time.Time
		want string
	}{
		{"zero", epoch, "T+000d 00:00:00"},
		{"positive_multi_day", epoch.Add(2*24*time.Hour + 3*time.Hour + 4*time.Minute + 5*time.Second), "T+002d 03:04:05"},
		{"negative", epoch.Add(-(1*time.Hour + 2*time.Minute + 3*time.Second)), "T-000d 01:02:03"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatMET(tc.ts, epoch); got != tc.want {
				t.Fatalf("FormatMET = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestHubPublishFrames_ReadingAndCriticalAlert(t *testing.T) {
	epoch := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	now := epoch.Add(90 * time.Second)

	hub := NewHub(15*time.Second, epoch)
	hub.nowFn = func() time.Time { return now }

	client := &Client{hub: hub, send: make(chan model.RealtimeEnvelope, sendBufferSize)}
	if !hub.register(client) {
		t.Fatal("failed to register client")
	}
	// Drain initial snapshot
	snapEnv := <-client.send
	if snapEnv.Type != "snapshot" {
		t.Fatalf("expected initial snapshot envelope, got %q", snapEnv.Type)
	}

	hrVal := 72.0
	spo2Val := 85.0
	unitMin := "/min"
	unitPct := "%"
	astroID := "11111111-1111-1111-1111-111111111111"

	hub.PublishFrames([]model.TelemetryFrame{
		{
			Source:        "wearable",
			DeviceCode:    "biomonitor_01",
			AstronautID:   &astroID,
			Environment:   model.EnvIVA,
			MetricType:    1,
			MetricCode:    "heart_rate",
			UnitCode:      &unitMin,
			ValueNumeric:  &hrVal,
			Status:        model.StatusNormal,
			RecordedAtUTC: now,
		},
		{
			Source:        "wearable",
			DeviceCode:    "biomonitor_01",
			AstronautID:   &astroID,
			Environment:   model.EnvIVA,
			MetricType:    1,
			MetricCode:    "spo2",
			UnitCode:      &unitPct,
			ValueNumeric:  &spo2Val,
			Status:        model.StatusCritical,
			RecordedAtUTC: now,
		},
	})

	if got := hub.ActiveAlertsCount(); got != 1 {
		t.Fatalf("ActiveAlertsCount = %d, want 1", got)
	}

	env1 := <-client.send
	if env1.Type != "reading" || env1.Reading == nil || env1.Reading.MetricCode != "heart_rate" {
		t.Fatalf("unexpected env1: %+v", env1)
	}
	if env1.Reading.MET != "T+000d 00:01:30" {
		t.Fatalf("unexpected MET %q", env1.Reading.MET)
	}

	env2 := <-client.send
	if env2.Type != "reading" || env2.Reading == nil || env2.Reading.MetricCode != "spo2" {
		t.Fatalf("unexpected env2: %+v", env2)
	}

	env3 := <-client.send
	if env3.Type != "alert" || env3.Alert == nil {
		t.Fatalf("expected alert envelope, got %+v", env3)
	}
	if env3.Alert.Status != model.StatusCritical || env3.Alert.MetricCode != "spo2" || env3.Alert.Acknowledged {
		t.Fatalf("unexpected alert payload: %+v", env3.Alert)
	}
}

func TestHubStaleMonitorAndAckAlert(t *testing.T) {
	epoch := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	now := epoch

	hub := NewHub(50*time.Millisecond, epoch)
	hub.nowFn = func() time.Time { return now }

	client := &Client{hub: hub, send: make(chan model.RealtimeEnvelope, sendBufferSize)}
	if !hub.register(client) {
		t.Fatal("failed to register client")
	}
	<-client.send // snapshot

	val := 88.0
	unitPct := "%"
	hub.PublishFrames([]model.TelemetryFrame{
		{
			Source:        "wearable",
			DeviceCode:    "biomonitor_01",
			Environment:   model.EnvIVA,
			MetricType:    1,
			MetricCode:    "spo2",
			UnitCode:      &unitPct,
			ValueNumeric:  &val,
			Status:        model.StatusCritical,
			RecordedAtUTC: now,
		},
	})
	<-client.send // reading
	alertEnv := <-client.send
	if alertEnv.Type != "alert" || alertEnv.Alert == nil {
		t.Fatalf("expected alert envelope, got %+v", alertEnv)
	}

	// Ack unknown alert returns ErrAlertNotFound
	if _, err := hub.AckAlert("missing-id", "flight"); !errors.Is(err, ErrAlertNotFound) {
		t.Fatalf("expected ErrAlertNotFound, got %v", err)
	}

	// Ack valid alert with empty acknowledgedBy defaults to "crew"
	acked, err := hub.AckAlert(alertEnv.Alert.ID, "")
	if err != nil {
		t.Fatalf("AckAlert failed: %v", err)
	}
	if !acked.Acknowledged || acked.AcknowledgedBy == nil || *acked.AcknowledgedBy != "crew" {
		t.Fatalf("unexpected acked alert: %+v", acked)
	}
	if got := hub.ActiveAlertsCount(); got != 0 {
		t.Fatalf("ActiveAlertsCount = %d, want 0", got)
	}

	ackEnv := <-client.send
	if ackEnv.Type != "alert_ack" || ackEnv.Alert == nil || !ackEnv.Alert.Acknowledged {
		t.Fatalf("expected alert_ack envelope, got %+v", ackEnv)
	}

	// Advance clock past staleAfter and trigger RunStaleMonitor
	now = now.Add(100 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.RunStaleMonitor(ctx, 10*time.Millisecond)

	select {
	case staleEnv := <-client.send:
		if staleEnv.Type != "stale" || staleEnv.Reading == nil {
			t.Fatalf("expected stale envelope, got %+v", staleEnv)
		}
		if !staleEnv.Reading.IsStale || staleEnv.Reading.Status != model.StatusStale {
			t.Fatalf("expected IsStale=true and StatusStale, got IsStale=%v Status=%q", staleEnv.Reading.IsStale, staleEnv.Reading.Status)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stale envelope")
	}
	cancel()

	// Snapshot reflects stale status
	snap := hub.Snapshot()
	if len(snap.Readings) != 1 || !snap.Readings[0].IsStale || snap.Readings[0].Status != model.StatusStale {
		t.Fatalf("expected stale reading in snapshot, got %+v", snap.Readings)
	}

	// Publishing a new reading clears stale state
	val2 := 98.0
	hub.PublishFrames([]model.TelemetryFrame{
		{
			Source:        "wearable",
			DeviceCode:    "biomonitor_01",
			Environment:   model.EnvIVA,
			MetricType:    1,
			MetricCode:    "spo2",
			UnitCode:      &unitPct,
			ValueNumeric:  &val2,
			Status:        model.StatusNormal,
			RecordedAtUTC: now,
		},
	})
	freshEnv := <-client.send
	if freshEnv.Type != "reading" || freshEnv.Reading == nil || freshEnv.Reading.IsStale || freshEnv.Reading.Status != model.StatusNormal {
		t.Fatalf("expected fresh reading to clear stale flag, got %+v", freshEnv)
	}
}

type mockHydrateQuerier struct {
	db.Querier
	rows []db.ListLatestReadingsByMetricRow
}

func (m mockHydrateQuerier) ListLatestReadingsByMetric(ctx context.Context, astronautID pgtype.UUID) ([]db.ListLatestReadingsByMetricRow, error) {
	return m.rows, nil
}

func TestHydrateFromDB_MarksOldReadingsStaleInMemoryOnly(t *testing.T) {
	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	hub := NewHub(15*time.Second, now.Add(-time.Hour))
	hub.nowFn = func() time.Time { return now }

	dev1 := "cabin_sensor_01"
	unitKpa := "kPa"
	var numOld, numFresh pgtype.Numeric
	_ = numOld.Scan("101.2")
	_ = numFresh.Scan("21.0")

	err := hub.HydrateFromDB(context.Background(), mockHydrateQuerier{
		rows: []db.ListLatestReadingsByMetricRow{
			{
				MetricCode:    "cabin_pressure",
				MetricType:    2,
				DeviceCode:    &dev1,
				Environment:   db.EnvironmentModeIva,
				UnitCode:      &unitKpa,
				RecordedAtUtc: pgtype.Timestamptz{Time: now.Add(-30 * time.Second), Valid: true},
				ValueNumeric:  numOld,
				Status:        db.ReadingStatusNormal,
			},
			{
				MetricCode:    "cabin_o2",
				MetricType:    2,
				DeviceCode:    &dev1,
				Environment:   db.EnvironmentModeIva,
				UnitCode:      &unitKpa,
				RecordedAtUtc: pgtype.Timestamptz{Time: now.Add(-5 * time.Second), Valid: true},
				ValueNumeric:  numFresh,
				Status:        db.ReadingStatusNormal,
			},
		},
	})
	if err != nil {
		t.Fatalf("HydrateFromDB failed: %v", err)
	}

	snap := hub.Snapshot()
	if len(snap.Readings) != 2 {
		t.Fatalf("expected 2 hydrated readings, got %d", len(snap.Readings))
	}

	byCode := map[string]model.LatestReading{}
	for _, r := range snap.Readings {
		byCode[r.MetricCode] = r
	}
	if !byCode["cabin_pressure"].IsStale || byCode["cabin_pressure"].Status != model.StatusStale {
		t.Fatalf("expected aged cabin_pressure to be stale, got %+v", byCode["cabin_pressure"])
	}
	if byCode["cabin_o2"].IsStale || byCode["cabin_o2"].Status != model.StatusNormal {
		t.Fatalf("expected recent cabin_o2 to stay normal, got %+v", byCode["cabin_o2"])
	}
}

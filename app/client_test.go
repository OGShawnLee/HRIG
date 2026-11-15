package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/wailsapp/wails/v3/pkg/application"

	"hrig/shared/model"
)

func TestResolveAPIBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantHTTP string
		wantWS   string
		wantErr  bool
	}{
		{
			name:     "empty uses default",
			raw:      "",
			wantHTTP: "http://127.0.0.1:80",
			wantWS:   "ws://127.0.0.1:80/ws/realtime",
		},
		{
			name:     "whitespace uses default",
			raw:      "   ",
			wantHTTP: "http://127.0.0.1:80",
			wantWS:   "ws://127.0.0.1:80/ws/realtime",
		},
		{
			name:     "http with trailing slash",
			raw:      "http://localhost:8080/",
			wantHTTP: "http://localhost:8080",
			wantWS:   "ws://localhost:8080/ws/realtime",
		},
		{
			name:     "https with multiple trailing slashes",
			raw:      "https://telemetry.example.com:8443///",
			wantHTTP: "https://telemetry.example.com:8443",
			wantWS:   "wss://telemetry.example.com:8443/ws/realtime",
		},
		{
			name:     "uppercase scheme normalized",
			raw:      "HTTPS://telemetry.example.com/api/",
			wantHTTP: "https://telemetry.example.com/api",
			wantWS:   "wss://telemetry.example.com/api/ws/realtime",
		},
		{
			name:    "invalid scheme",
			raw:     "ftp://127.0.0.1:80",
			wantErr: true,
		},
		{
			name:    "missing host",
			raw:     "http://",
			wantErr: true,
		},
		{
			name:    "malformed url",
			raw:     "://bad-url",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotHTTP, gotWS, err := resolveAPIBaseURL(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil (%s, %s)", tc.raw, gotHTTP, gotWS)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.raw, err)
			}
			if gotHTTP != tc.wantHTTP {
				t.Fatalf("httpBase = %q, want %q", gotHTTP, tc.wantHTTP)
			}
			if gotWS != tc.wantWS {
				t.Fatalf("wsURL = %q, want %q", gotWS, tc.wantWS)
			}
		})
	}
}

func TestBackoffDuration(t *testing.T) {
	if got := backoffDuration(0); got != 1*time.Second {
		t.Fatalf("attempt 0 backoff = %v, want 1s", got)
	}
	if got := backoffDuration(1); got != 2*time.Second {
		t.Fatalf("attempt 1 backoff = %v, want 2s", got)
	}
	if got := backoffDuration(2); got != 5*time.Second {
		t.Fatalf("attempt 2 backoff = %v, want 5s", got)
	}
	if got := backoffDuration(10); got != 5*time.Second {
		t.Fatalf("attempt 10 backoff = %v, want capped 5s", got)
	}
}

func TestAppRESTBindingsAndErrorDecoding(t *testing.T) {
	now := time.Date(2026, 11, 15, 21, 0, 0, 0, time.UTC)
	astroID := "11111111-1111-1111-1111-111111111111"
	astroName := "Cmdr. Elena Vance"
	unitCode := "bpm"
	unitSymbol := "bpm"
	hrValue := 72.5
	ackBy := "flight-surgeon"

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", r.Method)
		}
		writeTestJSON(t, w, http.StatusOK, model.HealthStatus{
			Status:          "ok",
			Store:           "postgres",
			TimeUTC:         now,
			ActiveAlerts:    1,
			RealtimeClients: 2,
			IngestClients:   1,
		})
	})

	mux.HandleFunc("/api/astronauts", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, http.StatusOK, []model.Astronaut{
			{
				ID:              astroID,
				FullName:        astroName,
				BirthDate:       time.Date(1990, 4, 12, 0, 0, 0, 0, time.UTC),
				Age:             36,
				NationalityCode: "US",
				IsActive:        true,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		})
	})

	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, http.StatusOK, []model.Metric{
			{
				ID:             "22222222-2222-2222-2222-222222222222",
				Code:           "heart_rate",
				Name:           "Heart Rate",
				Category:       "cardiovascular",
				Tier:           model.TierPrimary,
				MetricType:     1,
				ValueType:      model.ValueNumeric,
				BaseUnitCode:   &unitCode,
				BaseUnitSymbol: &unitSymbol,
				DisplayOrder:   1,
				IsActive:       true,
			},
		})
	})

	mux.HandleFunc("/api/snapshot", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, http.StatusOK, model.Snapshot{
			Readings: []model.LatestReading{
				{
					TelemetryFrame: model.TelemetryFrame{
						Source:        "wearable",
						DeviceCode:    "biovest-01",
						AstronautID:   &astroID,
						Environment:   model.EnvIVA,
						MetricType:    1,
						MetricCode:    "heart_rate",
						UnitCode:      &unitCode,
						ValueNumeric:  &hrValue,
						Status:        model.StatusNormal,
						RecordedAtUTC: now,
						MET:           "000:01:00:00",
					},
					LastSeenUTC: now,
					IsStale:     false,
				},
			},
			Alerts:         []model.AlertEvent{},
			GeneratedAtUTC: now,
		})
	})

	mux.HandleFunc("/api/trends", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("metric_code") == "" {
			writeTestJSON(t, w, http.StatusBadRequest, map[string]string{
				"error": "metric_code is required",
			})
			return
		}
		if q.Get("metric_code") != "heart_rate" || q.Get("astronaut_id") != astroID || q.Get("limit") != "50" {
			t.Fatalf("unexpected trends query: %s", r.URL.RawQuery)
		}
		writeTestJSON(t, w, http.StatusOK, []model.TrendPoint{
			{
				ID:            "33333333-3333-3333-3333-333333333333",
				MeasurementID: "44444444-4444-4444-4444-444444444444",
				MetricID:      "22222222-2222-2222-2222-222222222222",
				MetricCode:    "heart_rate",
				AstronautID:   &astroID,
				DeviceID:      "55555555-5555-5555-5555-555555555555",
				Environment:   model.EnvIVA,
				UnitCode:      &unitCode,
				UnitSymbol:    &unitSymbol,
				ValueNumeric:  &hrValue,
				Status:        model.StatusNormal,
				RecordedAtUTC: now,
				MET:           "000:01:00:00",
			},
		})
	})

	mux.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("astronaut_id") != astroID || q.Get("limit") != "25" || q.Get("offset") != "10" {
			t.Fatalf("unexpected history query: %s", r.URL.RawQuery)
		}
		writeTestJSON(t, w, http.StatusOK, []model.HistoryEntry{
			{
				MeasurementID: "44444444-4444-4444-4444-444444444444",
				ValueID:       "33333333-3333-3333-3333-333333333333",
				AstronautID:   &astroID,
				AstronautName: &astroName,
				DeviceCode:    "biovest-01",
				DeviceName:    "BioVest 01",
				Environment:   model.EnvIVA,
				MetricCode:    "heart_rate",
				MetricName:    "Heart Rate",
				ValueType:     model.ValueNumeric,
				UnitCode:      &unitCode,
				UnitSymbol:    &unitSymbol,
				ValueNumeric:  &hrValue,
				Status:        model.StatusNormal,
				RecordedAtUTC: now,
				MET:           "000:01:00:00",
			},
		})
	})

	mux.HandleFunc("/api/alerts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path == "/api/alerts/missing/ack" {
			writeTestJSON(t, w, http.StatusNotFound, map[string]string{
				"error": "alert not found",
			})
			return
		}
		if r.URL.Path != "/api/alerts/alert-123/ack" {
			t.Fatalf("unexpected ack path %s", r.URL.Path)
		}
		var req model.AckAlertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode ack body: %v", err)
		}
		if req.AcknowledgedBy != ackBy {
			t.Fatalf("acknowledgedBy = %q, want %q", req.AcknowledgedBy, ackBy)
		}
		writeTestJSON(t, w, http.StatusOK, model.AlertEvent{
			ID:                "alert-123",
			AstronautID:       &astroID,
			Source:            "wearable",
			Environment:       model.EnvIVA,
			MetricCode:        "heart_rate",
			ValueNumeric:      &hrValue,
			UnitCode:          &unitCode,
			Status:            model.StatusCritical,
			Message:           "elevated heart rate",
			TriggeredAtUTC:    now,
			MET:               "000:01:00:00",
			Acknowledged:      true,
			AcknowledgedBy:    &ackBy,
			AcknowledgedAtUTC: &now,
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Setenv("HRIG_API_URL", srv.URL)
	a := NewApp()
	if a.Ping() != "pong" {
		t.Fatalf("Ping() = %q, want pong", a.Ping())
	}

	// Ensure RPC methods work safely even before SetContext / ServiceStartup (a.ctx == nil).
	health, err := a.GetHealth()
	if err != nil {
		t.Fatalf("GetHealth failed: %v", err)
	}
	if health.Status != "ok" || health.Store != "postgres" {
		t.Fatalf("unexpected health: %+v", health)
	}

	astronauts, err := a.GetAstronauts()
	if err != nil {
		t.Fatalf("GetAstronauts failed: %v", err)
	}
	if len(astronauts) != 1 || astronauts[0].ID != astroID {
		t.Fatalf("unexpected astronauts: %+v", astronauts)
	}

	metrics, err := a.GetMetrics()
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}
	if len(metrics) != 1 || metrics[0].Code != "heart_rate" {
		t.Fatalf("unexpected metrics: %+v", metrics)
	}

	snapshot, err := a.GetSnapshot()
	if err != nil {
		t.Fatalf("GetSnapshot failed: %v", err)
	}
	if len(snapshot.Readings) != 1 || snapshot.Readings[0].MetricCode != "heart_rate" {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}

	trends, err := a.GetTrends(model.TrendQuery{
		AstronautID: astroID,
		MetricCode:  "heart_rate",
		Limit:       50,
	})
	if err != nil {
		t.Fatalf("GetTrends failed: %v", err)
	}
	if len(trends) != 1 || trends[0].MetricCode != "heart_rate" {
		t.Fatalf("unexpected trends: %+v", trends)
	}

	if _, err := a.GetTrends(model.TrendQuery{}); err == nil || !strings.Contains(err.Error(), "metric_code is required") {
		t.Fatalf("expected 400 error message from GetTrends, got %v", err)
	}

	history, err := a.GetHistory(model.HistoryQuery{
		AstronautID: astroID,
		Limit:       25,
		Offset:      10,
	})
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(history) != 1 || history[0].MeasurementID != "44444444-4444-4444-4444-444444444444" {
		t.Fatalf("unexpected history: %+v", history)
	}

	if _, err := a.AckAlert("   ", ackBy); err == nil {
		t.Fatal("expected validation error for empty alertID")
	}
	if _, err := a.AckAlert("missing", ackBy); err == nil || !strings.Contains(err.Error(), "alert not found") {
		t.Fatalf("expected server error decoding on AckAlert, got %v", err)
	}

	acked, err := a.AckAlert("alert-123", ackBy)
	if err != nil {
		t.Fatalf("AckAlert failed: %v", err)
	}
	if !acked.Acknowledged || acked.ID != "alert-123" {
		t.Fatalf("unexpected acked alert: %+v", acked)
	}
}

func TestRunRealtimeLoopEmitsEventsAndShutsDownCleanly(t *testing.T) {
	now := time.Date(2026, 11, 15, 21, 5, 0, 0, time.UTC)
	hrValue := 145.0
	unitCode := "bpm"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ws/realtime" {
			http.NotFound(w, r)
			return
		}
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return
		}
		defer conn.CloseNow()

		writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		frames := []model.RealtimeEnvelope{
			{
				Type: "snapshot",
				Readings: []model.LatestReading{
					{
						TelemetryFrame: model.TelemetryFrame{
							Source:        "wearable",
							DeviceCode:    "biovest-01",
							Environment:   model.EnvIVA,
							MetricType:    1,
							MetricCode:    "heart_rate",
							UnitCode:      &unitCode,
							ValueNumeric:  &hrValue,
							Status:        model.StatusCritical,
							RecordedAtUTC: now,
							MET:           "000:01:05:00",
						},
						LastSeenUTC: now,
						IsStale:     false,
					},
				},
				SentAtUTC: now,
			},
			{
				Type: "reading",
				Reading: &model.LatestReading{
					TelemetryFrame: model.TelemetryFrame{
						Source:        "wearable",
						DeviceCode:    "biovest-01",
						Environment:   model.EnvIVA,
						MetricType:    1,
						MetricCode:    "heart_rate",
						UnitCode:      &unitCode,
						ValueNumeric:  &hrValue,
						Status:        model.StatusCritical,
						RecordedAtUTC: now,
						MET:           "000:01:05:01",
					},
					LastSeenUTC: now,
					IsStale:     false,
				},
				SentAtUTC: now,
			},
			{
				Type: "alert",
				Alert: &model.AlertEvent{
					ID:             "alert-ws-1",
					Source:         "wearable",
					Environment:    model.EnvIVA,
					MetricCode:     "heart_rate",
					ValueNumeric:   &hrValue,
					UnitCode:       &unitCode,
					Status:         model.StatusCritical,
					Message:        "heart_rate critical",
					TriggeredAtUTC: now,
					MET:            "000:01:05:01",
				},
				SentAtUTC: now,
			},
		}

		for _, frame := range frames {
			if err := wsjson.Write(writeCtx, conn, frame); err != nil {
				return
			}
		}

		// Keep socket open until client cancels and disconnects.
		readCtx := context.Background()
		for {
			if _, _, err := conn.Read(readCtx); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	type emittedEvent struct {
		name string
		data any
	}

	var (
		mu     sync.Mutex
		events []emittedEvent
		doneCh = make(chan struct{})
	)

	client, err := newAPIClient(srv.URL, func(name string, data any) {
		mu.Lock()
		events = append(events, emittedEvent{name: name, data: data})
		envCount := 0
		for _, ev := range events {
			if ev.name == "telemetry:envelope" {
				envCount++
			}
		}
		if envCount == 3 {
			select {
			case <-doneCh:
			default:
				close(doneCh)
			}
		}
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("newAPIClient failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	loopDone := make(chan struct{})
	go func() {
		defer close(loopDone)
		client.RunRealtimeLoop(ctx)
	}()

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("timed out waiting for telemetry:envelope events")
	}

	if status := client.ConnectionStatus(); !status.Connected || status.LastError != "" {
		t.Fatalf("expected connected status, got %+v", status)
	}

	cancel()
	select {
	case <-loopDone:
	case <-time.After(3 * time.Second):
		t.Fatal("RunRealtimeLoop did not exit cleanly after cancel")
	}

	if status := client.ConnectionStatus(); status.Connected {
		t.Fatalf("expected disconnected status after shutdown, got %+v", status)
	}

	mu.Lock()
	defer mu.Unlock()

	var sawConnected bool
	var envelopeTypes []string
	for _, ev := range events {
		switch ev.name {
		case "telemetry:connection":
			state, ok := ev.data.(ConnectionState)
			if !ok {
				t.Fatalf("unexpected telemetry:connection payload type %T", ev.data)
			}
			if state.Connected {
				sawConnected = true
			}
		case "telemetry:envelope":
			env, ok := ev.data.(model.RealtimeEnvelope)
			if !ok {
				t.Fatalf("unexpected telemetry:envelope payload type %T", ev.data)
			}
			envelopeTypes = append(envelopeTypes, env.Type)
		}
	}

	if !sawConnected {
		t.Fatal("expected telemetry:connection with Connected=true")
	}
	wantTypes := []string{"snapshot", "reading", "alert"}
	if len(envelopeTypes) != len(wantTypes) {
		t.Fatalf("envelopeTypes = %v, want %v", envelopeTypes, wantTypes)
	}
	for i := range wantTypes {
		if envelopeTypes[i] != wantTypes[i] {
			t.Fatalf("envelopeTypes[%d] = %q, want %q", i, envelopeTypes[i], wantTypes[i])
		}
	}
}

func TestServiceStartupNonBlockingAndShutdown(t *testing.T) {
	// Point to an unreachable local port so Dial fails without blocking ServiceStartup.
	t.Setenv("HRIG_API_URL", "http://127.0.0.1:1")
	a := NewApp()

	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a.ServiceStartup(ctx, application.ServiceOptions{}); err != nil {
		t.Fatalf("ServiceStartup returned error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("ServiceStartup blocked for %v", elapsed)
	}

	// Give the goroutine a brief moment to enter RunRealtimeLoop, then shut down cleanly.
	time.Sleep(25 * time.Millisecond)

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- a.ServiceShutdown()
	}()

	select {
	case err := <-shutdownDone:
		if err != nil {
			t.Fatalf("ServiceShutdown returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("ServiceShutdown timed out waiting for realtime loop")
	}
}

func TestAppSearchManualBindings(t *testing.T) {
	now := time.Date(2026, 11, 15, 21, 10, 0, 0, time.UTC)
	astroID := "11111111-1111-1111-1111-111111111111"
	envIVA := model.EnvIVA
	unitCode := "%"

	mux := http.NewServeMux()
	mux.HandleFunc("/api/hazards", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", r.Method)
		}
		q := r.URL.Query().Get("q")
		if q == "null_slices" {
			// Ensure nil slices from server normalize to non-nil empty slices.
			_, _ = w.Write([]byte(`[{"id":"h-null","code":"null_test","name":"Null Test","summary":"s","clinicalImpact":"c","relatedMetricCodes":null,"source":"NASA","displayOrder":1}]`))
			return
		}
		if q != "cabin o2 & co2" {
			t.Fatalf("hazards q = %q, want %q (rawQuery=%q)", q, "cabin o2 & co2", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "q=cabin+o2+%26+co2") {
			t.Fatalf("hazards raw query not properly escaped: %q", r.URL.RawQuery)
		}
		writeTestJSON(t, w, http.StatusOK, []model.SpaceHazard{
			{
				ID:                 "haz-1",
				Code:               "hostile_closed_env",
				Name:               "Hostile / Closed Environments",
				Summary:            "Self-contained spacecraft ecosystems.",
				ClinicalImpact:     "Requires continuous dual monitoring.",
				RelatedMetricCodes: []string{"cabin_o2", "cabin_ppco2"},
				Source:             "NASA HRP Five Hazards",
				DisplayOrder:       5,
			},
		})
	})

	mux.HandleFunc("/api/procedures", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", r.Method)
		}
		params := r.URL.Query()
		if params.Get("q") != "air hunger / spo2" ||
			params.Get("category") != "respiratory_cabin" ||
			params.Get("metric_code") != "cabin_o2" {
			t.Fatalf("unexpected procedures query params: %v (raw=%q)", params, r.URL.RawQuery)
		}
		writeTestJSON(t, w, http.StatusOK, []model.MedicalProcedure{
			{
				ID:                 "proc-1",
				Code:               "hypoxia_cabin_o2",
				Title:              "Hypoxia & Low Cabin O2 Response",
				Category:           "respiratory_cabin",
				BodySystem:         "respiratory",
				Summary:            "Operational checklist to verify cabin oxygen delivery.",
				WhenToUse:          "When cabin O2 drops or crew SpO2 trends downward.",
				Escalation:         model.EscalationCMO,
				EstimatedMinutes:   10,
				Environment:        &envIVA,
				Equipment:          []string{"Pulse oximeter"},
				Steps:              []string{"Cross-check cabin O2 sensor telemetry."},
				Warnings:           []string{"Escalate immediately to Flight Surgeon if cyanosis occurs."},
				RelatedMetricCodes: []string{"cabin_o2", "spo2"},
				HazardCodes:        []string{"hostile_closed_env"},
				Keywords:           []string{"hypoxia"},
				Source:             "NASA HRP / ISS Operational Checklist",
				DisplayOrder:       1,
			},
		})
	})

	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", r.Method)
		}
		params := r.URL.Query()
		tab := params.Get("tab")
		if tab == "invalid" {
			writeTestJSON(t, w, http.StatusBadRequest, map[string]string{
				"error": `invalid tab "invalid"`,
			})
			return
		}
		if params.Get("q") != "hypoxia & o2" || tab != "hazards_metrics" {
			t.Fatalf("unexpected search query params: %v (raw=%q)", params, r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "q=hypoxia+%26+o2") {
			t.Fatalf("search raw query not properly escaped: %q", r.URL.RawQuery)
		}
		writeTestJSON(t, w, http.StatusOK, model.SearchResults{
			Query: "hypoxia & o2",
			Tab:   "hazards_metrics",
			Astronauts: []model.Astronaut{
				{
					ID:              astroID,
					FullName:        "Sam Sonia",
					BirthDate:       time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					Age:             36,
					NationalityCode: "US",
					IsActive:        true,
					CreatedAt:       now,
					UpdatedAt:       now,
				},
			},
			Procedures: nil, // verify Normalize() converts nil to empty slice
			Hazards: []model.SpaceHazard{
				{
					ID:                 "haz-1",
					Code:               "hostile_closed_env",
					Name:               "Hostile / Closed Environments",
					Summary:            "Self-contained ecosystems.",
					ClinicalImpact:     "Continuous monitoring.",
					RelatedMetricCodes: nil, // verify nested Normalize()
					Source:             "NASA HRP Five Hazards",
					DisplayOrder:       5,
				},
			},
			Metrics: []model.Metric{
				{
					ID:             "met-1",
					Code:           "spo2",
					Name:           "Blood Oxygen Saturation",
					Category:       "respiratory",
					Tier:           model.TierPrimary,
					MetricType:     1,
					ValueType:      model.ValueNumeric,
					BaseUnitCode:   &unitCode,
					BaseUnitSymbol: &unitCode,
					DisplayOrder:   1,
					IsActive:       true,
				},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Setenv("HRIG_API_URL", srv.URL)
	a := NewApp()

	hazards, err := a.GetHazards("  cabin o2 & co2  ")
	if err != nil {
		t.Fatalf("GetHazards failed: %v", err)
	}
	if len(hazards) != 1 || hazards[0].Code != "hostile_closed_env" || len(hazards[0].RelatedMetricCodes) != 2 {
		t.Fatalf("unexpected hazards: %+v", hazards)
	}

	nullHazards, err := a.GetHazards("null_slices")
	if err != nil {
		t.Fatalf("GetHazards(null_slices) failed: %v", err)
	}
	if len(nullHazards) != 1 || nullHazards[0].RelatedMetricCodes == nil {
		t.Fatalf("expected non-nil normalized slice on hazard, got %+v", nullHazards)
	}

	procs, err := a.GetProcedures("air hunger / spo2", "respiratory_cabin", "cabin_o2")
	if err != nil {
		t.Fatalf("GetProcedures failed: %v", err)
	}
	if len(procs) != 1 || procs[0].Code != "hypoxia_cabin_o2" || procs[0].Escalation != model.EscalationCMO {
		t.Fatalf("unexpected procedures: %+v", procs)
	}

	searchRes, err := a.Search("hypoxia & o2", "hazards_metrics")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if searchRes.Tab != "hazards_metrics" || len(searchRes.Astronauts) != 1 || len(searchRes.Hazards) != 1 || len(searchRes.Metrics) != 1 {
		t.Fatalf("unexpected search results: %+v", searchRes)
	}
	if searchRes.Procedures == nil {
		t.Fatal("expected normalized non-nil Procedures slice")
	}
	if searchRes.Hazards[0].RelatedMetricCodes == nil {
		t.Fatal("expected normalized non-nil RelatedMetricCodes slice on hazard")
	}

	if _, err := a.Search("", "invalid"); err == nil || !strings.Contains(err.Error(), "invalid tab") {
		t.Fatalf("expected 400 error message from Search, got %v", err)
	}
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, status int, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode test json: %v", err)
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"hrig/shared/model"
)

type seedThreshold struct {
	warnMin *float64
	warnMax *float64
	critMin *float64
	critMax *float64
}

func floatPtr(v float64) *float64 { return &v }

// Mirrors db/002_seed.sql numeric limits so tests verify bot nominal ranges stay normal.
var seedThresholds = map[string]seedThreshold{
	"heart_rate":         {warnMin: floatPtr(50), warnMax: floatPtr(100), critMin: floatPtr(40), critMax: floatPtr(200)},
	"spo2":               {warnMin: floatPtr(94), critMin: floatPtr(90)},
	"resp_rate":          {warnMin: floatPtr(10), warnMax: floatPtr(20), critMin: floatPtr(6), critMax: floatPtr(30)},
	"bp_systolic":        {warnMin: floatPtr(90), warnMax: floatPtr(120), critMin: floatPtr(70), critMax: floatPtr(180)},
	"bp_diastolic":       {warnMin: floatPtr(60), warnMax: floatPtr(80), critMin: floatPtr(40), critMax: floatPtr(120)},
	"core_temp":          {warnMin: floatPtr(36.5), warnMax: floatPtr(37.3), critMin: floatPtr(35), critMax: floatPtr(40)},
	"cabin_o2":           {warnMin: floatPtr(19.5), warnMax: floatPtr(23.5), critMin: floatPtr(17), critMax: floatPtr(25)},
	"cabin_ppco2":        {warnMax: floatPtr(3), critMax: floatPtr(5)},
	"cabin_pressure":     {warnMin: floatPtr(97), warnMax: floatPtr(104), critMin: floatPtr(90), critMax: floatPtr(110)},
	"cabin_temp":         {warnMin: floatPtr(18), warnMax: floatPtr(27), critMin: floatPtr(10), critMax: floatPtr(35)},
	"suit_pressure":      {warnMin: floatPtr(28), warnMax: floatPtr(30), critMin: floatPtr(25), critMax: floatPtr(33)},
	"suit_temp":          {warnMin: floatPtr(15), warnMax: floatPtr(35), critMin: floatPtr(5), critMax: floatPtr(45)},
	"hrv_rmssd":          {warnMin: floatPtr(20), critMin: floatPtr(10)},
	"radiation_rate":     {warnMax: floatPtr(0.5), critMax: floatPtr(2)},
	"radiation_dose_cum": {warnMax: floatPtr(20), critMax: floatPtr(50)},
	"sleep_hours":        {warnMin: floatPtr(7), warnMax: floatPtr(9), critMin: floatPtr(4), critMax: floatPtr(14)},
	"mood":               {warnMin: floatPtr(2), warnMax: floatPtr(4), critMin: floatPtr(1), critMax: floatPtr(5)},
	"blood_glucose":      {warnMin: floatPtr(70), warnMax: floatPtr(140), critMin: floatPtr(54), critMax: floatPtr(250)},
	"hydration_intake":   {warnMin: floatPtr(1500), warnMax: floatPtr(3500), critMin: floatPtr(500), critMax: floatPtr(5000)},
	"body_mass":          {warnMin: floatPtr(45), warnMax: floatPtr(110), critMin: floatPtr(35), critMax: floatPtr(140)},
	"rbc_count":          {warnMin: floatPtr(4), warnMax: floatPtr(6), critMin: floatPtr(2.5), critMax: floatPtr(8)},
	"wbc_count":          {warnMin: floatPtr(4), warnMax: floatPtr(11), critMin: floatPtr(1), critMax: floatPtr(30)},
}

func classifyAgainstSeed(code string, val float64) model.ReadingStatus {
	th, ok := seedThresholds[code]
	if !ok {
		return model.StatusNormal
	}
	if th.critMin != nil && val <= *th.critMin {
		return model.StatusCritical
	}
	if th.critMax != nil && val >= *th.critMax {
		return model.StatusCritical
	}
	if th.warnMin != nil && val < *th.warnMin {
		return model.StatusAbnormal
	}
	if th.warnMax != nil && val > *th.warnMax {
		return model.StatusAbnormal
	}
	return model.StatusNormal
}

func TestNominalRangesComplyWithSeedThresholds(t *testing.T) {
	gen := newGenerator(42)

	for iter := 0; iter < 250; iter++ {
		for _, s := range sensors {
			recipients := []string{""}
			if s.Astronaut {
				recipients = crew
			}
			for _, astro := range recipients {
				readings := gen.makeReadings(s, astro)
				for _, r := range readings {
					if r.ValueNumeric != nil {
						val := *r.ValueNumeric
						if status := classifyAgainstSeed(r.MetricCode, val); status != model.StatusNormal {
							t.Fatalf("iter %d metric %s (%s) = %v classified as %s, want normal", iter, r.MetricCode, astro, val, status)
						}
						switch r.MetricCode {
						case "core_temp":
							if val < 36.55 || val > 37.20 {
								t.Fatalf("core_temp = %v outside [36.55, 37.20]", val)
							}
						case "suit_pressure":
							if val < 28.1 || val > 29.9 {
								t.Fatalf("suit_pressure = %v outside [28.1, 29.9]", val)
							}
						case "wbc_count":
							if val < 4.6 || val > 10.0 {
								t.Fatalf("wbc_count = %v outside [4.6, 10.0]", val)
							}
						}
					}
					if r.ValueScale != nil {
						if status := classifyAgainstSeed(r.MetricCode, float64(*r.ValueScale)); status != model.StatusNormal {
							t.Fatalf("iter %d scale metric %s = %d classified as %s, want normal", iter, r.MetricCode, *r.ValueScale, status)
						}
					}
				}
			}
		}
	}
}

func TestOverrideAndCancelTransitions(t *testing.T) {
	gen := newGenerator(99)

	// Inject abnormal core_temp specifically for Liam Okafor
	st, err := gen.setOverride("abnormal", "core_temp", "Liam Okafor")
	if err != nil {
		t.Fatalf("setOverride abnormal: %v", err)
	}
	if st.Mode != modeAbnormal || st.MetricCode != "core_temp" || st.Astronaut != "Liam Okafor" {
		t.Fatalf("unexpected override state: %+v", st)
	}

	biomonitor := sensors[0]
	liamReadings := gen.makeReadings(biomonitor, "Liam Okafor")
	anaReadings := gen.makeReadings(biomonitor, "Ana Torres")

	var liamTemp, anaTemp float64
	for _, r := range liamReadings {
		if r.MetricCode == "core_temp" && r.ValueNumeric != nil {
			liamTemp = *r.ValueNumeric
		}
	}
	for _, r := range anaReadings {
		if r.MetricCode == "core_temp" && r.ValueNumeric != nil {
			anaTemp = *r.ValueNumeric
		}
	}
	if status := classifyAgainstSeed("core_temp", liamTemp); status != model.StatusAbnormal {
		t.Fatalf("expected Liam core_temp (%v) to be abnormal, got %s", liamTemp, status)
	}
	if status := classifyAgainstSeed("core_temp", anaTemp); status != model.StatusNormal {
		t.Fatalf("expected Ana core_temp (%v) to stay normal, got %s", anaTemp, status)
	}

	// Escalate to critical cabin_pressure
	if _, err := gen.setOverride("critical", "cabin_pressure", ""); err != nil {
		t.Fatalf("setOverride critical: %v", err)
	}
	cabinSensor := sensors[3]
	cabinReadings := gen.makeReadings(cabinSensor, "")
	var cabinPressure float64
	for _, r := range cabinReadings {
		if r.MetricCode == "cabin_pressure" && r.ValueNumeric != nil {
			cabinPressure = *r.ValueNumeric
		}
	}
	if status := classifyAgainstSeed("cabin_pressure", cabinPressure); status != model.StatusCritical {
		t.Fatalf("expected cabin_pressure (%v) to be critical, got %s", cabinPressure, status)
	}

	// Cancel restores nominal values for both
	cancelled := gen.cancel()
	if cancelled.Mode != modeNormal || cancelled.MetricCode != "" || cancelled.Astronaut != "" {
		t.Fatalf("unexpected state after cancel: %+v", cancelled)
	}
	for _, r := range gen.makeReadings(biomonitor, "Liam Okafor") {
		if r.ValueNumeric != nil && classifyAgainstSeed(r.MetricCode, *r.ValueNumeric) != model.StatusNormal {
			t.Fatalf("after cancel, %s = %v should be normal", r.MetricCode, *r.ValueNumeric)
		}
	}
	for _, r := range gen.makeReadings(cabinSensor, "") {
		if r.ValueNumeric != nil && classifyAgainstSeed(r.MetricCode, *r.ValueNumeric) != model.StatusNormal {
			t.Fatalf("after cancel, %s = %v should be normal", r.MetricCode, *r.ValueNumeric)
		}
	}
}

func TestControlHTTPEndpointsAndStdinCommands(t *testing.T) {
	gen := newGenerator(7)
	var dispatched atomic.Int32
	handler := newControlHandler(gen, func() {
		dispatched.Add(1)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	// 1. GET /status starts normal
	resp, err := http.Get(srv.URL + "/status")
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	var st overrideState
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		t.Fatalf("decode /status: %v", err)
	}
	resp.Body.Close()
	if st.Mode != modeNormal {
		t.Fatalf("initial mode = %q, want normal", st.Mode)
	}

	// 2. POST /override triggers abnormal & dispatch
	body := bytes.NewBufferString(`{"mode":"critical","metric_code":"bp_systolic","astronaut":"Ana Torres"}`)
	resp, err = http.Post(srv.URL+"/override", "application/json", body)
	if err != nil {
		t.Fatalf("POST /override: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /override status = %d, want 200", resp.StatusCode)
	}
	_ = json.NewDecoder(resp.Body).Decode(&st)
	resp.Body.Close()
	if st.Mode != modeCritical || st.MetricCode != "bp_systolic" || st.Astronaut != "Ana Torres" {
		t.Fatalf("unexpected override response: %+v", st)
	}
	if dispatched.Load() != 1 {
		t.Fatalf("dispatched = %d, want 1", dispatched.Load())
	}

	// 3. POST /cancel restores normal & triggers immediate batch dispatch
	resp, err = http.Post(srv.URL+"/cancel", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /cancel: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /cancel status = %d, want 200", resp.StatusCode)
	}
	_ = json.NewDecoder(resp.Body).Decode(&st)
	resp.Body.Close()
	if st.Mode != modeNormal {
		t.Fatalf("mode after /cancel = %q, want normal", st.Mode)
	}
	if dispatched.Load() != 2 {
		t.Fatalf("dispatched after cancel = %d, want 2", dispatched.Load())
	}

	// 4. Send batch verifies nominal values hit ingest server
	var mu sync.Mutex
	var received []model.IngestRequest
	ingestSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req model.IngestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			mu.Lock()
			received = append(received, req)
			mu.Unlock()
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ingestSrv.Close()

	if err := sendBatch(context.Background(), ingestSrv.Client(), ingestSrv.URL, 1, gen); err != nil {
		t.Fatalf("sendBatch: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(received) == 0 {
		t.Fatal("expected ingest server to receive readings")
	}
	for _, req := range received {
		for _, rd := range req.Readings {
			if rd.ValueNumeric != nil && classifyAgainstSeed(rd.MetricCode, *rd.ValueNumeric) != model.StatusNormal {
				t.Fatalf("post-cancel batch had non-normal %s = %v", rd.MetricCode, *rd.ValueNumeric)
			}
		}
	}

	// 5. Stdin command parsing
	if _, ok := handleCommandLine("abnormal spo2 Sofia Petrov", gen, func() { dispatched.Add(1) }); !ok {
		t.Fatal("expected abnormal stdin command to be handled")
	}
	if cur := gen.status(); cur.Mode != modeAbnormal || cur.MetricCode != "spo2" || cur.Astronaut != "Sofia Petrov" {
		t.Fatalf("unexpected stdin override state: %+v", cur)
	}
	if _, ok := handleCommandLine("normal", gen, func() { dispatched.Add(1) }); !ok {
		t.Fatal("expected normal stdin command to be handled")
	}
	if cur := gen.status(); cur.Mode != modeNormal {
		t.Fatalf("expected normal after stdin normal, got %+v", cur)
	}
}

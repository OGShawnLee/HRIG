package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"hrig/shared/model"
)

// The metric list mirrors db/002_seed.sql. A sensor sends one measurement
// session containing the readings it owns; the server resolves catalog codes
// to database IDs and persists the session and values.
type metric struct {
	Code  string          `json:"metric_code"`
	Unit  string          `json:"unit_code,omitempty"`
	Kind  model.ValueType `json:"-"` // numeric, scale, or image
	Min   float64         `json:"-"`
	Max   float64         `json:"-"`
	Scale int             `json:"-"`
}

type sensor struct {
	Device      string
	Environment model.EnvironmentMode
	Astronaut   bool
	Metrics     []metric
}

type generator struct {
	rng      *rand.Rand
	previous map[string]float64
}

var crew = []string{"Ana Torres", "Liam Okafor", "Sofia Petrov"}

var sensors = []sensor{
	{Device: "biomonitor_01", Environment: model.EnvIVA, Astronaut: true, Metrics: []metric{
		{Code: "spo2", Unit: "%", Kind: model.ValueNumeric, Min: 96, Max: 99},
		{Code: "resp_rate", Unit: "/min", Kind: model.ValueNumeric, Min: 12, Max: 19},
		{Code: "heart_rate", Unit: "/min", Kind: model.ValueNumeric, Min: 55, Max: 95},
		{Code: "hrv_rmssd", Unit: "ms", Kind: model.ValueNumeric, Min: 25, Max: 75},
		{Code: "core_temp", Unit: "Cel", Kind: model.ValueNumeric, Min: 36.4, Max: 37.2},
		{Code: "sleep_hours", Unit: "h", Kind: model.ValueNumeric, Min: 7, Max: 9},
		{Code: "mood", Unit: "{score}", Kind: model.ValueScale, Scale: 4},
		{Code: "hydration_intake", Unit: "mL", Kind: model.ValueNumeric, Min: 1800, Max: 3000},
		{Code: "body_mass", Unit: "kg", Kind: model.ValueNumeric, Min: 60, Max: 85},
	}},
	{Device: "baumanometro_01", Environment: model.EnvIVA, Astronaut: true, Metrics: []metric{
		{Code: "bp_systolic", Unit: "mm[Hg]", Kind: model.ValueNumeric, Min: 105, Max: 120},
		{Code: "bp_diastolic", Unit: "mm[Hg]", Kind: model.ValueNumeric, Min: 65, Max: 80},
	}},
	{Device: "dosimeter_01", Environment: model.EnvIVA, Astronaut: true, Metrics: []metric{
		{Code: "radiation_rate", Unit: "uSv/h", Kind: model.ValueNumeric, Min: 0.05, Max: 0.4},
		{Code: "radiation_dose_cum", Unit: "mSv", Kind: model.ValueNumeric, Min: 1, Max: 18},
	}},
	{Device: "cabin_sensor_01", Environment: model.EnvIVA, Metrics: []metric{
		{Code: "cabin_o2", Unit: "%", Kind: model.ValueNumeric, Min: 20.5, Max: 21.5},
		{Code: "cabin_ppco2", Unit: "mm[Hg]", Kind: model.ValueNumeric, Min: 0.3, Max: 2.5},
		{Code: "cabin_pressure", Unit: "kPa", Kind: model.ValueNumeric, Min: 99, Max: 103},
		{Code: "cabin_temp", Unit: "Cel", Kind: model.ValueNumeric, Min: 20, Max: 24},
	}},
	{Device: "suit_sensor_01", Environment: model.EnvEVA, Metrics: []metric{
		{Code: "suit_pressure", Unit: "kPa", Kind: model.ValueNumeric, Min: 28, Max: 30},
		{Code: "suit_temp", Unit: "Cel", Kind: model.ValueNumeric, Min: 20, Max: 32},
	}},
	{Device: "lab_assay_01", Environment: model.EnvIVA, Astronaut: true, Metrics: []metric{
		{Code: "blood_glucose", Unit: "mg/dL", Kind: model.ValueNumeric, Min: 80, Max: 120},
		{Code: "rbc_count", Unit: "10*6/uL", Kind: model.ValueNumeric, Min: 4.2, Max: 5.8},
		{Code: "wbc_count", Unit: "10*3/uL", Kind: model.ValueNumeric, Min: 4.5, Max: 10},
	}},
	{Device: "ultrasound_01", Environment: model.EnvIVA, Astronaut: true, Metrics: []metric{
		{Code: "ultrasound", Kind: model.ValueImage},
	}},
}

func main() {
	endpoint := env("SENSOR_ENDPOINT", "http://localhost:80/")
	interval, err := time.ParseDuration(env("SENSOR_INTERVAL", "2s"))
	if err != nil || interval <= 0 {
		log.Fatal("SENSOR_INTERVAL debe ser una duración positiva, por ejemplo 5s")
	}
	crewCount, err := strconv.Atoi(env("SENSOR_CREW_COUNT", strconv.Itoa(len(crew))))
	if err != nil || crewCount < 1 || crewCount > len(crew) {
		log.Fatalf("SENSOR_CREW_COUNT debe estar entre 1 y %d", len(crew))
	}
	log.Printf("simulando %d sensores lógicos para %d astronautas; POST %s cada %s", len(sensors), crewCount, endpoint, interval)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	client := &http.Client{Timeout: 10 * time.Second}
	gen := &generator{rng: rand.New(rand.NewSource(time.Now().UnixNano())), previous: make(map[string]float64)}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := sendBatch(ctx, client, endpoint, crewCount, gen); err != nil {
			log.Printf("envío incompleto: %v", err)
		}
		select {
		case <-ctx.Done():
			log.Print("simulador detenido")
			return
		case <-ticker.C:
		}
	}
}

func sendBatch(ctx context.Context, client *http.Client, endpoint string, crewCount int, gen *generator) error {
	var failures []string
	for _, s := range sensors {
		recipients := []string{""}
		if s.Astronaut {
			recipients = crew[:crewCount]
		}
		for _, name := range recipients {
			p := model.IngestRequest{
				DeviceCode:    s.Device,
				Environment:   s.Environment,
				RecordedAtUTC: time.Now().UTC().Format(time.RFC3339Nano),
				Readings:      gen.makeReadings(s, name),
			}
			if name != "" {
				p.AstronautName = &name
			}
			body, err := json.Marshal(p)
			if err != nil {
				failures = append(failures, err.Error())
				continue
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
			if err != nil {
				failures = append(failures, err.Error())
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s: %v", s.Device, err))
				continue
			}
			resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				failures = append(failures, fmt.Sprintf("%s: HTTP %s", s.Device, resp.Status))
			}
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	log.Print("lote de lecturas enviado correctamente")
	return nil
}

func (g *generator) makeReadings(s sensor, astronaut string) []model.IngestReading {
	out := make([]model.IngestReading, 0, len(s.Metrics))
	for _, m := range s.Metrics {
		r := model.IngestReading{MetricCode: m.Code}
		if m.Unit != "" {
			unit := m.Unit
			r.UnitCode = &unit
		}
		switch m.Kind {
		case model.ValueNumeric:
			key := s.Device + "/" + astronaut + "/" + m.Code
			span := m.Max - m.Min
			v, exists := g.previous[key]
			if !exists {
				v = m.Min + g.rng.Float64()*span
			} else if m.Code == "radiation_dose_cum" {
				v += g.rng.Float64() * span * 0.005
				if v > m.Max {
					v = m.Min
				}
			} else {
				step := span * (0.01 + g.rng.Float64()*0.04)
				if g.rng.Intn(2) == 0 {
					v -= step
				} else {
					v += step
				}
				if v < m.Min || v > m.Max {
					v = g.previous[key] - (v - g.previous[key])
				}
			}
			v = float64(int64(v*10000+0.5)) / 10000
			if v == g.previous[key] {
				v += 0.0001
				if v > m.Max {
					v -= 0.0002
				}
			}
			g.previous[key] = v
			r.ValueNumeric = &v
		case model.ValueScale:
			key := s.Device + "/" + astronaut + "/" + m.Code
			v := m.Scale
			if old, exists := g.previous[key]; exists {
				v = int(old)
				if v == m.Scale {
					v = m.Scale - 1
				} else {
					v = m.Scale
				}
			}
			g.previous[key] = float64(v)
			r.ValueScale = &v
		case model.ValueImage:
			options := []string{"simulated normal ultrasound finding", "simulated stable ultrasound finding", "simulated follow-up ultrasound finding"}
			v := options[g.rng.Intn(len(options))]
			r.ValueText = &v
		}
		out = append(out, r)
	}
	return out
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

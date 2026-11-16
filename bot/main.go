package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
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

const (
	modeNormal   = "normal"
	modeAbnormal = "abnormal"
	modeCritical = "critical"

	defaultAbnormalMetric = "bp_systolic"
	defaultCriticalMetric = "cabin_pressure"
)

type overrideState struct {
	Mode         string `json:"mode"`
	MetricCode   string `json:"metric_code,omitempty"`
	Astronaut    string `json:"astronaut,omitempty"`
	UpdatedAtUTC string `json:"updated_at_utc,omitempty"`
}

type deviceControlInfo struct {
	Device       string                `json:"device"`
	Environment  model.EnvironmentMode `json:"environment"`
	Astronaut    bool                  `json:"astronaut"`
	Disconnected bool                  `json:"disconnected"`
	Metrics      []string              `json:"metrics"`
}

type controlStatus struct {
	overrideState
	DisconnectedDevices []string            `json:"disconnected_devices"`
	Devices             []deviceControlInfo `json:"devices"`
}

type anomalyBand struct {
	Abnormal float64
	Critical float64
}

var anomalyTargets = map[string]anomalyBand{
	"spo2":               {Abnormal: 92, Critical: 85},
	"resp_rate":          {Abnormal: 24, Critical: 34},
	"heart_rate":         {Abnormal: 115, Critical: 210},
	"hrv_rmssd":          {Abnormal: 15, Critical: 8},
	"core_temp":          {Abnormal: 38.1, Critical: 40.5},
	"sleep_hours":        {Abnormal: 5.5, Critical: 3.0},
	"mood":               {Abnormal: 1, Critical: 1},
	"hydration_intake":   {Abnormal: 1200, Critical: 400},
	"body_mass":          {Abnormal: 42, Critical: 32},
	"bp_systolic":        {Abnormal: 84, Critical: 65},
	"bp_diastolic":       {Abnormal: 54, Critical: 35},
	"radiation_rate":     {Abnormal: 0.9, Critical: 2.8},
	"radiation_dose_cum": {Abnormal: 25, Critical: 55},
	"cabin_o2":           {Abnormal: 18.8, Critical: 16.0},
	"cabin_ppco2":        {Abnormal: 3.8, Critical: 6.0},
	"cabin_pressure":     {Abnormal: 94.0, Critical: 78.0},
	"cabin_temp":         {Abnormal: 29.0, Critical: 37.0},
	"suit_pressure":      {Abnormal: 27.0, Critical: 24.0},
	"suit_temp":          {Abnormal: 38.0, Critical: 47.0},
	"blood_glucose":      {Abnormal: 165, Critical: 275},
	"rbc_count":          {Abnormal: 3.5, Critical: 2.0},
	"wbc_count":          {Abnormal: 13.5, Critical: 32.0},
}

type generator struct {
	mu           sync.Mutex
	rng          *rand.Rand
	previous     map[string]float64
	override     overrideState
	disconnected map[string]bool
}

func newGenerator(seed int64) *generator {
	return &generator{
		rng:          rand.New(rand.NewSource(seed)),
		previous:     make(map[string]float64),
		override:     overrideState{Mode: modeNormal},
		disconnected: make(map[string]bool),
	}
}

func (g *generator) status() overrideState {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.override
}

func (g *generator) isDisconnected(device string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.disconnected[device]
}

func (g *generator) disconnectedDevices() []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.disconnectedDevicesLocked()
}

func (g *generator) disconnectedDevicesLocked() []string {
	out := make([]string, 0, len(g.disconnected))
	for _, s := range sensors {
		if g.disconnected[s.Device] {
			out = append(out, s.Device)
		}
	}
	return out
}

func (g *generator) controlSnapshot() controlStatus {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.controlSnapshotLocked()
}

func (g *generator) controlSnapshotLocked() controlStatus {
	devices := make([]deviceControlInfo, 0, len(sensors))
	for _, s := range sensors {
		metricCodes := make([]string, 0, len(s.Metrics))
		for _, m := range s.Metrics {
			metricCodes = append(metricCodes, m.Code)
		}
		devices = append(devices, deviceControlInfo{
			Device:       s.Device,
			Environment:  s.Environment,
			Astronaut:    s.Astronaut,
			Disconnected: g.disconnected[s.Device],
			Metrics:      metricCodes,
		})
	}
	return controlStatus{
		overrideState:       g.override,
		DisconnectedDevices: g.disconnectedDevicesLocked(),
		Devices:             devices,
	}
}

func resolveDeviceCode(device string) (string, bool) {
	norm := strings.ToLower(strings.TrimSpace(device))
	for _, s := range sensors {
		if s.Device == norm {
			return s.Device, true
		}
	}
	return "", false
}

func (g *generator) disconnect(device string) (controlStatus, error) {
	target := strings.ToLower(strings.TrimSpace(device))
	if target == "" {
		return controlStatus{}, fmt.Errorf("dispositivo requerido: indique un dispositivo válido, 'all' o '*'")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if target == "all" || target == "*" {
		for _, s := range sensors {
			g.disconnected[s.Device] = true
		}
		return g.controlSnapshotLocked(), nil
	}

	canonical, ok := resolveDeviceCode(target)
	if !ok {
		return controlStatus{}, fmt.Errorf("dispositivo desconocido %q", device)
	}
	g.disconnected[canonical] = true
	return g.controlSnapshotLocked(), nil
}

func (g *generator) reconnect(device string) (controlStatus, error) {
	target := strings.ToLower(strings.TrimSpace(device))

	g.mu.Lock()
	defer g.mu.Unlock()

	if target == "" || target == "all" || target == "*" {
		clear(g.disconnected)
		return g.controlSnapshotLocked(), nil
	}

	canonical, ok := resolveDeviceCode(target)
	if !ok {
		return controlStatus{}, fmt.Errorf("dispositivo desconocido %q", device)
	}
	delete(g.disconnected, canonical)
	return g.controlSnapshotLocked(), nil
}

func (g *generator) setOverride(mode, metricCode, astronaut string) (overrideState, error) {
	normalizedMode := strings.ToLower(strings.TrimSpace(mode))
	if normalizedMode == "cancel" {
		normalizedMode = modeNormal
	}
	if normalizedMode != modeNormal && normalizedMode != modeAbnormal && normalizedMode != modeCritical {
		return overrideState{}, fmt.Errorf("modo inválido %q: use normal, abnormal o critical", mode)
	}

	metricCode = strings.TrimSpace(metricCode)
	astronaut = strings.TrimSpace(astronaut)
	now := time.Now().UTC().Format(time.RFC3339)

	g.mu.Lock()
	defer g.mu.Unlock()

	if normalizedMode == modeNormal {
		g.override = overrideState{
			Mode:         modeNormal,
			UpdatedAtUTC: now,
		}
		g.resetOutOfBoundsLocked()
		return g.override, nil
	}

	if metricCode == "" {
		if normalizedMode == modeCritical {
			metricCode = defaultCriticalMetric
		} else {
			metricCode = defaultAbnormalMetric
		}
	}
	if _, ok := anomalyTargets[metricCode]; !ok {
		return overrideState{}, fmt.Errorf("métrica desconocida o no numérica %q", metricCode)
	}

	g.override = overrideState{
		Mode:         normalizedMode,
		MetricCode:   metricCode,
		Astronaut:    astronaut,
		UpdatedAtUTC: now,
	}
	return g.override, nil
}

func (g *generator) cancel() overrideState {
	st, _ := g.setOverride(modeNormal, "", "")
	return st
}

func (g *generator) resetOutOfBoundsLocked() {
	for _, s := range sensors {
		for _, m := range s.Metrics {
			if m.Kind != model.ValueNumeric {
				continue
			}
			prefix := s.Device + "/"
			suffix := "/" + m.Code
			for key, prev := range g.previous {
				if strings.HasPrefix(key, prefix) && strings.HasSuffix(key, suffix) {
					if prev < m.Min || prev > m.Max {
						delete(g.previous, key)
					}
				}
			}
		}
	}
}

func (g *generator) overrideValueLocked(s sensor, astronaut string, m metric) (float64, bool) {
	if g.override.Mode == modeNormal || g.override.MetricCode != m.Code {
		return 0, false
	}
	if s.Astronaut && g.override.Astronaut != "" && !matchesAstronaut(g.override.Astronaut, astronaut) {
		return 0, false
	}
	band, ok := anomalyTargets[m.Code]
	if !ok {
		return 0, false
	}
	if g.override.Mode == modeCritical {
		return band.Critical, true
	}
	return band.Abnormal, true
}

func matchesAstronaut(target, candidate string) bool {
	t := strings.ToLower(strings.TrimSpace(target))
	if t == "" || t == "*" || t == "all" {
		return true
	}
	c := strings.ToLower(strings.TrimSpace(candidate))
	return c == t || strings.Contains(c, t)
}

var crew = []string{"Ana Torres", "Liam Okafor", "Sofia Petrov"}

var sensors = []sensor{
	{Device: "biomonitor_01", Environment: model.EnvIVA, Astronaut: true, Metrics: []metric{
		{Code: "spo2", Unit: "%", Kind: model.ValueNumeric, Min: 96, Max: 99},
		{Code: "resp_rate", Unit: "/min", Kind: model.ValueNumeric, Min: 12, Max: 19},
		{Code: "heart_rate", Unit: "/min", Kind: model.ValueNumeric, Min: 55, Max: 95},
		{Code: "hrv_rmssd", Unit: "ms", Kind: model.ValueNumeric, Min: 25, Max: 75},
		{Code: "core_temp", Unit: "Cel", Kind: model.ValueNumeric, Min: 36.55, Max: 37.20},
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
		{Code: "suit_pressure", Unit: "kPa", Kind: model.ValueNumeric, Min: 28.1, Max: 29.9},
		{Code: "suit_temp", Unit: "Cel", Kind: model.ValueNumeric, Min: 20, Max: 32},
	}},
	{Device: "lab_assay_01", Environment: model.EnvIVA, Astronaut: true, Metrics: []metric{
		{Code: "blood_glucose", Unit: "mg/dL", Kind: model.ValueNumeric, Min: 80, Max: 120},
		{Code: "rbc_count", Unit: "10*6/uL", Kind: model.ValueNumeric, Min: 4.2, Max: 5.8},
		{Code: "wbc_count", Unit: "10*3/uL", Kind: model.ValueNumeric, Min: 4.6, Max: 10.0},
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
	controlAddr := env("BOT_CONTROL_ADDR", "127.0.0.1:8090")
	log.Printf("simulando %d sensores lógicos para %d astronautas; POST %s cada %s", len(sensors), crewCount, endpoint, interval)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	client := &http.Client{Timeout: 10 * time.Second}
	gen := newGenerator(time.Now().UnixNano())

	triggerCh := make(chan struct{}, 1)
	triggerBatch := func() {
		select {
		case triggerCh <- struct{}{}:
		default:
		}
	}

	if !strings.EqualFold(controlAddr, "off") {
		srv := &http.Server{
			Addr:              controlAddr,
			Handler:           newControlHandler(gen, triggerBatch),
			ReadHeaderTimeout: 5 * time.Second,
		}
		ln, err := net.Listen("tcp", controlAddr)
		if err != nil {
			log.Printf("advertencia: no se pudo iniciar control HTTP en %s: %v", controlAddr, err)
		} else {
			log.Printf("control HTTP activo en http://%s (GET /, GET /status, POST /override, POST /cancel, POST /disconnect, POST /reconnect)", ln.Addr().String())
			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("error en servidor de control: %v", err)
				}
			}()
			go func() {
				<-ctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_ = srv.Shutdown(shutdownCtx)
			}()
		}
	}

	go runStdinLoop(ctx, os.Stdin, gen, triggerBatch)

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
		case <-triggerCh:
		}
	}
}

type overrideRequest struct {
	Mode       string `json:"mode"`
	MetricCode string `json:"metric_code"`
	Astronaut  string `json:"astronaut"`
}

type deviceRequest struct {
	Device string `json:"device"`
}

func newControlHandler(gen *generator, onDispatch func()) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, controlPanelHTML)
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, gen.controlSnapshot())
	})
	mux.HandleFunc("/override", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req overrideRequest
		if err := json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido: " + err.Error()})
			return
		}
		st, err := gen.setOverride(req.Mode, req.MetricCode, req.Astronaut)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		log.Printf("override HTTP actualizado: mode=%s metric=%s astronaut=%q", st.Mode, st.MetricCode, st.Astronaut)
		if onDispatch != nil {
			onDispatch()
		}
		writeJSON(w, http.StatusOK, gen.controlSnapshot())
	})
	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = gen.cancel()
		log.Print("override cancelado vía HTTP; restaurando telemetría nominal")
		if onDispatch != nil {
			onDispatch()
		}
		writeJSON(w, http.StatusOK, gen.controlSnapshot())
	})
	mux.HandleFunc("/disconnect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req deviceRequest
		if err := json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido: " + err.Error()})
			return
		}
		snap, err := gen.disconnect(req.Device)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		log.Printf("dispositivo desconectado vía HTTP: target=%q disconnected=%v", req.Device, snap.DisconnectedDevices)
		writeJSON(w, http.StatusOK, snap)
	})
	mux.HandleFunc("/reconnect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req deviceRequest
		if r.Body != nil {
			if err := json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido: " + err.Error()})
				return
			}
		}
		snap, err := gen.reconnect(req.Device)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		log.Printf("dispositivo reconectado vía HTTP: target=%q disconnected=%v", req.Device, snap.DisconnectedDevices)
		if onDispatch != nil {
			onDispatch()
		}
		writeJSON(w, http.StatusOK, snap)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func runStdinLoop(ctx context.Context, r io.Reader, gen *generator, onDispatch func()) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if msg, ok := handleCommandLine(scanner.Text(), gen, onDispatch); ok && msg != "" {
			log.Print(msg)
		}
	}
}

func handleCommandLine(line string, gen *generator, onDispatch func()) (string, bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", false
	}
	cmd := strings.ToLower(fields[0])
	switch cmd {
	case modeAbnormal, modeCritical:
		var metricCode, astronaut string
		if len(fields) >= 2 {
			metricCode = fields[1]
		}
		if len(fields) >= 3 {
			astronaut = strings.Join(fields[2:], " ")
		}
		st, err := gen.setOverride(cmd, metricCode, astronaut)
		if err != nil {
			return fmt.Sprintf("comando inválido: %v", err), true
		}
		if onDispatch != nil {
			onDispatch()
		}
		return fmt.Sprintf("override activo: mode=%s metric=%s astronaut=%q", st.Mode, st.MetricCode, st.Astronaut), true
	case "cancel", modeNormal:
		st := gen.cancel()
		if onDispatch != nil {
			onDispatch()
		}
		return fmt.Sprintf("override cancelado: mode=%s", st.Mode), true
	case "disconnect", "offline":
		if len(fields) < 2 {
			return "uso: disconnect <device|all>", true
		}
		snap, err := gen.disconnect(fields[1])
		if err != nil {
			return fmt.Sprintf("comando inválido: %v", err), true
		}
		return fmt.Sprintf("dispositivo(s) desconectado(s): disconnected=%v", snap.DisconnectedDevices), true
	case "reconnect", "online":
		target := ""
		if len(fields) >= 2 {
			target = fields[1]
		}
		snap, err := gen.reconnect(target)
		if err != nil {
			return fmt.Sprintf("comando inválido: %v", err), true
		}
		if onDispatch != nil {
			onDispatch()
		}
		return fmt.Sprintf("dispositivo(s) reconectado(s): disconnected=%v", snap.DisconnectedDevices), true
	case "status":
		snap := gen.controlSnapshot()
		return fmt.Sprintf("estado actual: mode=%s metric=%s astronaut=%q updated=%s disconnected=%v", snap.Mode, snap.MetricCode, snap.Astronaut, snap.UpdatedAtUTC, snap.DisconnectedDevices), true
	case "help", "?":
		return "comandos: abnormal [metric] [astronaut] | critical [metric] [astronaut] | cancel | normal | disconnect <device|all> | reconnect [device|all] | status | help", true
	default:
		return fmt.Sprintf("comando desconocido %q (use 'help')", cmd), true
	}
}

func sendBatch(ctx context.Context, client *http.Client, endpoint string, crewCount int, gen *generator) error {
	var failures []string
	for _, s := range sensors {
		if gen.isDisconnected(s.Device) {
			continue
		}
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
	g.mu.Lock()
	defer g.mu.Unlock()

	out := make([]model.IngestReading, 0, len(s.Metrics))
	for _, m := range s.Metrics {
		r := model.IngestReading{MetricCode: m.Code}
		if m.Unit != "" {
			unit := m.Unit
			r.UnitCode = &unit
		}
		switch m.Kind {
		case model.ValueNumeric:
			if overrideVal, ok := g.overrideValueLocked(s, astronaut, m); ok {
				v := overrideVal
				r.ValueNumeric = &v
				break
			}
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
			v = clampFloat(float64(int64(v*10000+0.5))/10000, m.Min, m.Max)
			if v == g.previous[key] {
				v += 0.0001
				if v > m.Max {
					v -= 0.0002
				}
				v = clampFloat(v, m.Min, m.Max)
			}
			g.previous[key] = v
			r.ValueNumeric = &v
		case model.ValueScale:
			if overrideVal, ok := g.overrideValueLocked(s, astronaut, m); ok {
				v := int(overrideVal)
				r.ValueScale = &v
				break
			}
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

func clampFloat(v, minVal, maxVal float64) float64 {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

const controlPanelHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>HRIG Telemetry Bot Control</title>
<style>
  :root {
    --bg: #070b14;
    --panel: #0f172a;
    --panel-alt: #131c31;
    --border: #1e293b;
    --text: #e2e8f0;
    --muted: #94a3b8;
    --accent: #38bdf8;
    --ok: #22c55e;
    --warn: #f59e0b;
    --crit: #ef4444;
  }
  * { box-sizing: border-box; }
  body {
    margin: 0;
    font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, sans-serif;
    background: radial-gradient(circle at top right, #0e1a33 0%, var(--bg) 55%);
    color: var(--text);
    min-height: 100vh;
    padding: 24px;
  }
  .container { max-width: 1040px; margin: 0 auto; display: grid; gap: 18px; }
  header {
    display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 12px;
    background: var(--panel); border: 1px solid var(--border); border-radius: 12px; padding: 16px 20px;
  }
  h1 { margin: 0; font-size: 1.15rem; letter-spacing: 0.04em; text-transform: uppercase; color: var(--accent); }
  .subtitle { color: var(--muted); font-size: 0.85rem; margin-top: 4px; }
  .pill {
    display: inline-flex; align-items: center; gap: 6px; padding: 4px 10px; border-radius: 999px;
    font-size: 0.78rem; font-weight: 600; font-variant-numeric: tabular-nums; border: 1px solid var(--border);
    background: var(--panel-alt);
  }
  .pill.ok { color: var(--ok); border-color: rgba(34,197,94,0.35); background: rgba(34,197,94,0.1); }
  .pill.warn { color: var(--warn); border-color: rgba(245,158,11,0.35); background: rgba(245,158,11,0.12); }
  .pill.crit { color: var(--crit); border-color: rgba(239,68,68,0.4); background: rgba(239,68,68,0.12); }
  .card {
    background: var(--panel); border: 1px solid var(--border); border-radius: 12px; padding: 18px 20px;
  }
  .card-head {
    display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 10px;
    margin-bottom: 14px;
  }
  .card-title { margin: 0; font-size: 0.95rem; text-transform: uppercase; letter-spacing: 0.05em; color: #cbd5e1; }
  .form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px; align-items: end; }
  label { display: block; font-size: 0.78rem; color: var(--muted); margin-bottom: 5px; }
  select, input {
    width: 100%; background: var(--bg); color: var(--text); border: 1px solid var(--border);
    border-radius: 8px; padding: 8px 10px; font-size: 0.88rem;
  }
  .btn-row { display: flex; flex-wrap: wrap; gap: 8px; }
  button {
    cursor: pointer; border-radius: 8px; border: 1px solid var(--border); padding: 8px 14px;
    font-size: 0.84rem; font-weight: 600; color: var(--text); background: var(--panel-alt);
    transition: border-color 0.15s, background 0.15s;
  }
  button:hover { border-color: var(--accent); }
  button.primary { background: rgba(56,189,248,0.16); border-color: rgba(56,189,248,0.5); color: #7dd3fc; }
  button.warn { background: rgba(245,158,11,0.16); border-color: rgba(245,158,11,0.5); color: #fcd34d; }
  button.danger { background: rgba(239,68,68,0.16); border-color: rgba(239,68,68,0.5); color: #fca5a5; }
  button.success { background: rgba(34,197,94,0.16); border-color: rgba(34,197,94,0.5); color: #86efac; }
  .device-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 12px; }
  .device-card {
    background: var(--panel-alt); border: 1px solid var(--border); border-radius: 10px; padding: 14px;
    display: flex; flex-direction: column; gap: 10px;
  }
  .device-card.disconnected { border-color: rgba(239,68,68,0.45); background: rgba(239,68,68,0.06); }
  .device-top { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
  .device-name { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 0.92rem; font-weight: 700; }
  .meta { font-size: 0.78rem; color: var(--muted); }
  .metrics { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 0.75rem; color: #94a3b8; line-height: 1.4; }
  .feedback { min-height: 20px; font-size: 0.84rem; color: var(--muted); }
  .feedback.err { color: #fca5a5; }
  .feedback.ok { color: #86efac; }
</style>
</head>
<body>
<div class="container">
  <header>
    <div>
      <h1>HRIG Telemetry Bot Control</h1>
      <div class="subtitle">Manual Anomaly Override &amp; Sensor Disconnection Simulator</div>
    </div>
    <div class="btn-row">
      <span id="mode-pill" class="pill ok">MODE: normal</span>
      <span id="conn-pill" class="pill ok">ONLINE: 7/7</span>
      <button type="button" onclick="refreshStatus()">Refresh</button>
    </div>
  </header>

  <section class="card">
    <div class="card-head">
      <h2 class="card-title">Anomaly Override</h2>
      <span id="override-meta" class="meta">Active target: nominal</span>
    </div>
    <div class="form-grid">
      <div>
        <label for="ov-mode">Mode</label>
        <select id="ov-mode">
          <option value="abnormal">abnormal</option>
          <option value="critical">critical</option>
          <option value="normal">normal</option>
        </select>
      </div>
      <div>
        <label for="ov-metric">Metric</label>
        <select id="ov-metric">
          <option value="bp_systolic">bp_systolic</option>
          <option value="bp_diastolic">bp_diastolic</option>
          <option value="cabin_pressure">cabin_pressure</option>
          <option value="cabin_o2">cabin_o2</option>
          <option value="cabin_ppco2">cabin_ppco2</option>
          <option value="cabin_temp">cabin_temp</option>
          <option value="suit_pressure">suit_pressure</option>
          <option value="suit_temp">suit_temp</option>
          <option value="spo2">spo2</option>
          <option value="resp_rate">resp_rate</option>
          <option value="heart_rate">heart_rate</option>
          <option value="hrv_rmssd">hrv_rmssd</option>
          <option value="core_temp">core_temp</option>
          <option value="sleep_hours">sleep_hours</option>
          <option value="mood">mood</option>
          <option value="hydration_intake">hydration_intake</option>
          <option value="body_mass">body_mass</option>
          <option value="radiation_rate">radiation_rate</option>
          <option value="radiation_dose_cum">radiation_dose_cum</option>
          <option value="blood_glucose">blood_glucose</option>
          <option value="rbc_count">rbc_count</option>
          <option value="wbc_count">wbc_count</option>
        </select>
      </div>
      <div>
        <label for="ov-astro">Astronaut</label>
        <select id="ov-astro">
          <option value="">All Crew</option>
          <option value="Ana Torres">Ana Torres</option>
          <option value="Liam Okafor">Liam Okafor</option>
          <option value="Sofia Petrov">Sofia Petrov</option>
        </select>
      </div>
      <div class="btn-row">
        <button type="button" class="warn" onclick="applyOverride()">Apply Override</button>
        <button type="button" class="success" onclick="cancelOverride()">Cancel / Normal</button>
      </div>
    </div>
  </section>

  <section class="card">
    <div class="card-head">
      <h2 class="card-title">Device Connection Control</h2>
      <div class="btn-row">
        <button type="button" class="danger" onclick="disconnectDevice('all')">Disconnect All</button>
        <button type="button" class="success" onclick="reconnectDevice('all')">Reconnect All</button>
      </div>
    </div>
    <div id="device-grid" class="device-grid"></div>
  </section>

  <div id="feedback" class="feedback">Ready.</div>
</div>
<script>
function setFeedback(msg, kind) {
  var el = document.getElementById('feedback');
  el.textContent = msg;
  el.className = 'feedback' + (kind ? ' ' + kind : '');
}

async function apiJSON(url, options) {
  var resp = await fetch(url, options);
  var data = await resp.json();
  if (!resp.ok) {
    throw new Error(data.error || ('HTTP ' + resp.status));
  }
  return data;
}

function renderStatus(data) {
  var mode = data.mode || 'normal';
  var modePill = document.getElementById('mode-pill');
  modePill.textContent = 'MODE: ' + mode;
  modePill.className = 'pill ' + (mode === 'critical' ? 'crit' : (mode === 'abnormal' ? 'warn' : 'ok'));

  var metaParts = [];
  if (mode === 'normal') {
    metaParts.push('Active target: nominal');
  } else {
    metaParts.push('Active target: ' + (data.metric_code || 'default') + (data.astronaut ? (' (' + data.astronaut + ')') : ' (all crew)'));
  }
  if (data.updated_at_utc) {
    metaParts.push('Updated: ' + data.updated_at_utc);
  }
  document.getElementById('override-meta').textContent = metaParts.join(' | ');

  var devices = data.devices || [];
  var discCount = (data.disconnected_devices || []).length;
  var total = devices.length;
  var online = total - discCount;
  var connPill = document.getElementById('conn-pill');
  connPill.textContent = 'ONLINE: ' + online + '/' + total;
  connPill.className = 'pill ' + (discCount === 0 ? 'ok' : (online === 0 ? 'crit' : 'warn'));

  var grid = document.getElementById('device-grid');
  grid.innerHTML = '';
  devices.forEach(function(dev) {
    var card = document.createElement('div');
    card.className = 'device-card' + (dev.disconnected ? ' disconnected' : '');

    var top = document.createElement('div');
    top.className = 'device-top';

    var name = document.createElement('span');
    name.className = 'device-name';
    name.textContent = dev.device;

    var statusPill = document.createElement('span');
    statusPill.className = 'pill ' + (dev.disconnected ? 'crit' : 'ok');
    statusPill.textContent = dev.disconnected ? 'DISCONNECTED' : 'ONLINE';

    top.appendChild(name);
    top.appendChild(statusPill);

    var meta = document.createElement('div');
    meta.className = 'meta';
    meta.textContent = 'Env: ' + dev.environment + ' | Scope: ' + (dev.astronaut ? 'Per-Astronaut' : 'Ambient');

    var metrics = document.createElement('div');
    metrics.className = 'metrics';
    metrics.textContent = 'Metrics: ' + (dev.metrics || []).join(', ');

    var btnRow = document.createElement('div');
    btnRow.className = 'btn-row';
    var btn = document.createElement('button');
    btn.type = 'button';
    if (dev.disconnected) {
      btn.className = 'success';
      btn.textContent = 'Reconnect';
      btn.onclick = function() { reconnectDevice(dev.device); };
    } else {
      btn.className = 'danger';
      btn.textContent = 'Disconnect';
      btn.onclick = function() { disconnectDevice(dev.device); };
    }
    btnRow.appendChild(btn);

    card.appendChild(top);
    card.appendChild(meta);
    card.appendChild(metrics);
    card.appendChild(btnRow);
    grid.appendChild(card);
  });
}

async function refreshStatus() {
  try {
    var data = await apiJSON('/status');
    renderStatus(data);
  } catch (err) {
    setFeedback('Status refresh failed: ' + err.message, 'err');
  }
}

async function applyOverride() {
  try {
    var payload = {
      mode: document.getElementById('ov-mode').value,
      metric_code: document.getElementById('ov-metric').value,
      astronaut: document.getElementById('ov-astro').value
    };
    var data = await apiJSON('/override', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    renderStatus(data);
    setFeedback('Override updated: ' + data.mode, 'ok');
  } catch (err) {
    setFeedback('Override failed: ' + err.message, 'err');
  }
}

async function cancelOverride() {
  try {
    var data = await apiJSON('/cancel', { method: 'POST' });
    renderStatus(data);
    setFeedback('Override cancelled; restored normal telemetry.', 'ok');
  } catch (err) {
    setFeedback('Cancel failed: ' + err.message, 'err');
  }
}

async function disconnectDevice(device) {
  try {
    var data = await apiJSON('/disconnect', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ device: device })
    });
    renderStatus(data);
    setFeedback('Disconnected: ' + device, 'ok');
  } catch (err) {
    setFeedback('Disconnect failed: ' + err.message, 'err');
  }
}

async function reconnectDevice(device) {
  try {
    var data = await apiJSON('/reconnect', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ device: device })
    });
    renderStatus(data);
    setFeedback('Reconnected: ' + device, 'ok');
  } catch (err) {
    setFeedback('Reconnect failed: ' + err.message, 'err');
  }
}

refreshStatus();
setInterval(refreshStatus, 2000);
</script>
</body>
</html>
`

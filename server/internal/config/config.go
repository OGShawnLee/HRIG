package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultStaleAfter  = 15 * time.Second
	defaultDatabaseURL = "postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable"
)

type Config struct {
	APIAddr         string
	AllowedOrigins  []string
	DatabaseURL     string
	StaleAfter      time.Duration
	MissionEpochUTC time.Time
}

func Load() Config {
	dotEnv := loadDotEnvFiles(".env", filepath.Join("server", ".env"))
	getEnv := func(key string) string {
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return dotEnv[key]
	}

	addr := strings.TrimSpace(getEnv("HRIG_API_ADDR"))
	if addr == "" {
		addr = ":80"
	}

	rawOrigins := strings.TrimSpace(getEnv("HRIG_ALLOWED_ORIGINS"))
	if rawOrigins == "" {
		rawOrigins = "http://127.0.0.1:9245,http://localhost:9245,wails://wails"
	}

	parts := strings.Split(rawOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		origins = []string{"http://127.0.0.1:9245", "http://localhost:9245", "wails://wails"}
	}

	databaseURL := strings.TrimSpace(getEnv("DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = strings.TrimSpace(getEnv("HRIG_DATABASE_URL"))
	}
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	staleAfter := defaultStaleAfter
	if rawStale := strings.TrimSpace(getEnv("HRIG_STALE_AFTER")); rawStale != "" {
		if parsed, err := time.ParseDuration(rawStale); err == nil && parsed > 0 {
			staleAfter = parsed
		}
	}

	missionEpoch := time.Now().UTC()
	if rawEpoch := strings.TrimSpace(getEnv("HRIG_MISSION_EPOCH_UTC")); rawEpoch != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, rawEpoch); err == nil {
			missionEpoch = parsed.UTC()
		}
	}

	return Config{
		APIAddr:         addr,
		AllowedOrigins:  origins,
		DatabaseURL:     databaseURL,
		StaleAfter:      staleAfter,
		MissionEpochUTC: missionEpoch,
	}
}

func loadDotEnvFiles(paths ...string) map[string]string {
	values := make(map[string]string)
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		parseDotEnv(file, values)
		_ = file.Close()
	}
	return values
}

func parseDotEnv(file *os.File, values map[string]string) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(rest)
		} else if rest, ok := strings.CutPrefix(line, "export\t"); ok {
			line = strings.TrimSpace(rest)
		}

		key, rawVal, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := values[key]; exists {
			continue
		}

		val := strings.TrimSpace(rawVal)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		values[key] = val
	}
}

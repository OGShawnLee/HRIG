package config

import (
	"os"
	"strings"
)

type Config struct {
	APIAddr        string
	AllowedOrigins []string
}

func Load() Config {
	addr := strings.TrimSpace(os.Getenv("HRIG_API_ADDR"))
	if addr == "" {
		addr = ":8080"
	}

	rawOrigins := strings.TrimSpace(os.Getenv("HRIG_ALLOWED_ORIGINS"))
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

	return Config{
		APIAddr:        addr,
		AllowedOrigins: origins,
	}
}

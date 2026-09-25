package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	AppEnv   string
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Verify   VerifyConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	URL      string
	MaxConns int32
	MinConns int32
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	Origin string
}

// VerifyConfig holds verification-related settings.
type VerifyConfig struct {
	BaseURL string
}

// Load reads configuration from environment variables and validates required fields.
// It also automatically loads key=value pairs from a local .env file if present.
func Load() (*Config, error) {
	loadDotEnv(".env")

	cfg := &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port: getEnv("PORT", "8081"),
		},
		CORS: CORSConfig{
			Origin: getEnv("CORS_ORIGIN", "http://localhost:3000"),
		},
		Verify: VerifyConfig{
			BaseURL: getEnv("VERIFICATION_BASE_URL", "http://localhost:3000/verify"),
		},
	}

	// Database — required
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required")
	}
	cfg.Database.URL = dbURL

	maxConns, err := getEnvInt32("DB_MAX_CONNS", 25)
	if err != nil {
		return nil, fmt.Errorf("config: DB_MAX_CONNS: %w", err)
	}
	cfg.Database.MaxConns = maxConns

	minConns, err := getEnvInt32("DB_MIN_CONNS", 5)
	if err != nil {
		return nil, fmt.Errorf("config: DB_MIN_CONNS: %w", err)
	}
	cfg.Database.MinConns = minConns

	// JWT — required
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}
	cfg.JWT.Secret = jwtSecret

	jwtExp := getEnv("JWT_EXPIRATION", "24h")
	dur, err := time.ParseDuration(jwtExp)
	if err != nil {
		return nil, fmt.Errorf("config: JWT_EXPIRATION invalid duration %q: %w", jwtExp, err)
	}
	cfg.JWT.Expiration = dur

	return cfg, nil
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt32 returns an int32 env var or a default.
func getEnvInt32(key string, fallback int32) (int32, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q", v)
	}
	return int32(n), nil
}

// loadDotEnv parses key-value pairs from a .env file and sets them in the process environment.
func loadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}

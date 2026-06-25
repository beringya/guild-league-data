package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName           string
	AppEnv            string
	ListenAddr        string
	PublicURL         string
	DatabaseDSN       string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	SessionSecret     string
	SessionTTL        time.Duration
	CookieSecure      bool
	CSRFHeader        string
	UploadDir         string
	BackupDir         string
	StaticDir         string
	PreviewTTL        time.Duration
	LoginFailureLimit int
}

func Load() Config {
	return Config{
		AppName:           env("APP_NAME", "nsh-guild-analytics"),
		AppEnv:            env("GIN_MODE", "release"),
		ListenAddr:        env("LISTEN_ADDR", ":8080"),
		PublicURL:         env("PUBLIC_URL", "http://localhost:18080"),
		DatabaseDSN:       requiredAny("DATABASE_DSN", "DATABASE_URL"),
		RedisAddr:         env("REDIS_ADDR", "redis:6379"),
		RedisPassword:     env("REDIS_PASSWORD", ""),
		RedisDB:           envInt("REDIS_DB", 0),
		SessionSecret:     required("SESSION_SECRET"),
		SessionTTL:        envDuration("SESSION_TTL", 8*time.Hour),
		CookieSecure:      envBool("COOKIE_SECURE", false),
		CSRFHeader:        env("CSRF_HEADER", "X-CSRF-Token"),
		UploadDir:         env("UPLOAD_DIR", "/app/data/uploads"),
		BackupDir:         env("BACKUP_DIR", "/app/backups"),
		StaticDir:         env("STATIC_DIR", "/app/public"),
		PreviewTTL:        envDuration("IMPORT_PREVIEW_TTL", 30*time.Minute),
		LoginFailureLimit: envInt("LOGIN_FAILURE_LIMIT", 5),
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func required(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		panic(fmt.Sprintf("missing required environment variable %s", key))
	}
	return value
}

func requiredAny(keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	panic(fmt.Sprintf("missing required environment variable %s", strings.Join(keys, " or ")))
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	return raw == "1" || raw == "true" || raw == "yes" || raw == "on"
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

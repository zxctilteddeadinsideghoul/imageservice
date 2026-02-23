package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr             string
	StorageDir             string
	PublicURL              string
	MaxFileSize            int64
	DownloadTimeout        time.Duration
	MaxConcurrentDownloads int
}

func Load() *Config {
	return &Config{
		ListenAddr:             envOrDefault("LISTEN_ADDR", ":8080"),
		StorageDir:             envOrDefault("STORAGE_DIR", "./data"),
		PublicURL:              strings.TrimRight(envOrDefault("PUBLIC_URL", "http://localhost:8080"), "/"),
		MaxFileSize:            envOrDefaultInt64("MAX_FILE_SIZE", 10<<20),
		DownloadTimeout:        envOrDefaultDuration("DOWNLOAD_TIMEOUT", 10*time.Second),
		MaxConcurrentDownloads: envOrDefaultInt("MAX_CONCURRENT_DOWNLOADS", 20),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envOrDefaultInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}

func envOrDefaultDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

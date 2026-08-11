package config

import (
	"log/slog"
	"os"
	"strconv"
)

const (
	DefaultAccessKey = "auxio"
	DefaultSecretKey = "auxio-secret-key"
)

type Config struct {
	AccessKey          string
	SecretKey          string
	UsingDefaultKeys   bool
	AdminUsername      string
	AdminPassword      string
	DataDir            string
	Bind               string
	HttpPort           string
	Region             string
	ImagingEnabled     bool
	ImagingMaxWidth    int
	ImagingMaxHeight   int
	UploadCleanupHours int
	LogLevel           string
}

func Load() (*Config, error) {
	accessKey := os.Getenv("AUXIO_ACCESS_KEY")
	secretKey := os.Getenv("AUXIO_SECRET_KEY")
	usingDefaultKeys := false

	if accessKey == "" {
		accessKey = DefaultAccessKey
		usingDefaultKeys = true
	}
	if secretKey == "" {
		secretKey = DefaultSecretKey
		usingDefaultKeys = true
	}

	adminUsername := getEnv("AUXIO_USERNAME", "admin")
	adminPassword := getEnv("AUXIO_PASSWORD", "password")

	dataDir := getEnv("AUXIO_DATA_DIR", "./data")
	bind := getEnv("AUXIO_BIND", "127.0.0.1")
	httpPort := getEnv("AUXIO_HTTP_PORT", "9000")
	region := getEnv("AUXIO_REGION", "eu-north-1")
	imagingEnabled := getBoolEnv("AUXIO_IMAGING_ENABLED", true)
	imagingMaxWidth := getIntEnv("AUXIO_IMAGING_MAX_WIDTH", 4096)
	imagingMaxHeight := getIntEnv("AUXIO_IMAGING_MAX_HEIGHT", 4096)
	uploadCleanupHours := getIntEnv("AUXIO_UPLOAD_CLEANUP_HOURS", 24)
	logLevel := getEnv("AUXIO_LOG_LEVEL", "info")

	return &Config{
		AccessKey:          accessKey,
		SecretKey:          secretKey,
		UsingDefaultKeys:   usingDefaultKeys,
		AdminUsername:      adminUsername,
		AdminPassword:      adminPassword,
		DataDir:            dataDir,
		Bind:               bind,
		HttpPort:           httpPort,
		Region:             region,
		ImagingEnabled:     imagingEnabled,
		ImagingMaxWidth:    imagingMaxWidth,
		ImagingMaxHeight:   imagingMaxHeight,
		UploadCleanupHours: uploadCleanupHours,
		LogLevel:           logLevel,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		b, err := strconv.ParseBool(value)
		if err != nil {
			slog.Warn("invalid value for env var, using default", "key", key, "value", value, "default", defaultValue)
			return defaultValue
		}
		return b
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		i, err := strconv.Atoi(value)
		if err != nil {
			slog.Warn("invalid value for env var, using default", "key", key, "value", value, "default", defaultValue)
			return defaultValue
		}
		return i
	}
	return defaultValue
}

// UsingDefaultDashboardCreds reports whether the dashboard admin credentials
// are still the insecure built-in defaults.
func (c *Config) UsingDefaultDashboardCreds() bool {
	return c.AdminUsername == "admin" && c.AdminPassword == "password"
}

// IsLoopbackBind reports whether the configured bind address is loopback-only,
// i.e. the kernel will refuse off-machine connections.
func (c *Config) IsLoopbackBind() bool {
	switch c.Bind {
	case "127.0.0.1", "::1", "localhost":
		return true
	default:
		return false
	}
}

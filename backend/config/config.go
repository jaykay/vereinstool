package config

import "os"

type Config struct {
	AppURL            string
	DBPath            string
	SessionSecret     string
	Port              string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPassword      string
	SMTPFrom          string
	SeedAdminEmail    string
	SeedAdminPassword string
}

func Load() *Config {
	return &Config{
		AppURL:            getEnv("APP_URL", "http://localhost:5173"),
		DBPath:            getEnv("DB_PATH", "./dev.db"),
		SessionSecret:     getEnv("SESSION_SECRET", "dev-secret-not-for-production"),
		Port:              getEnv("PORT", "8080"),
		SMTPHost:          getEnv("SMTP_HOST", "localhost"),
		SMTPPort:          getEnv("SMTP_PORT", "1025"),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPassword:      getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:          getEnv("SMTP_FROM", "dev@localhost"),
		SeedAdminEmail:    getEnv("SEED_ADMIN_EMAIL", ""),
		SeedAdminPassword: getEnv("SEED_ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

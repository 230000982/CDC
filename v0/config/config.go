package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Email    EmailConfig
	Session  SessionConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// EmailConfig holds email configuration
type EmailConfig struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort string
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Secret string
}

// Load loads configuration from environment variables or defaults
func Load() (*Config, error) {
	// Set defaults and override with environment variables if available
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "beso")
	dbPassword := getEnv("DB_PASSWORD", "beso")
	dbName := getEnv("DB_NAME", "concurso")

	serverPort := getEnv("SERVER_PORT", "8080")

	emailFrom := getEnv("EMAIL_FROM", "notbeso2000@gmail.com")
	emailPassword := getEnv("EMAIL_PASSWORD", "wmfhdtxnsegwhnzj")
	emailSMTPHost := getEnv("EMAIL_SMTP_HOST", "smtp.gmail.com")
	emailSMTPPort := getEnv("EMAIL_SMTP_PORT", "587")

	sessionSecret := getEnv("SESSION_SECRET", "secret-key")

	return &Config{
		Database: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			DBName:   dbName,
		},
		Server: ServerConfig{
			Port: serverPort,
		},
		Email: EmailConfig{
			From:     emailFrom,
			Password: emailPassword,
			SMTPHost: emailSMTPHost,
			SMTPPort: emailSMTPPort,
		},
		Session: SessionConfig{
			Secret: sessionSecret,
		},
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.User, c.Password, c.Host, c.Port, c.DBName)
}

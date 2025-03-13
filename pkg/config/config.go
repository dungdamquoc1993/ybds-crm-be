package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	AccountDB      DatabaseConfig
	NotificationDB DatabaseConfig
	OrderDB        DatabaseConfig
	ProductDB      DatabaseConfig
	Server         ServerConfig
	JWT            JWTConfig
	Upload         UploadConfig
}

// DatabaseConfig holds all database related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// ServerConfig holds all server related configuration
type ServerConfig struct {
	Port string
	Env  string
}

// JWTConfig holds all JWT related configuration
type JWTConfig struct {
	Secret string
	Expiry string
}

// UploadConfig holds all upload related configuration
type UploadConfig struct {
	Dir       string
	MaxSizeMB int
}

// LoadConfig loads the configuration from .env file and environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize Viper
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set default values
	setDefaults(v)

	// Create config instance
	config := &Config{
		AccountDB: DatabaseConfig{
			Host:     v.GetString("db.host"),
			Port:     v.GetString("db.port"),
			User:     v.GetString("db.user"),
			Password: v.GetString("db.pass"),
			Name:     v.GetString("db.account.name"),
			SSLMode:  v.GetString("db.ssl_mode"),
		},
		NotificationDB: DatabaseConfig{
			Host:     v.GetString("db.host"),
			Port:     v.GetString("db.port"),
			User:     v.GetString("db.user"),
			Password: v.GetString("db.pass"),
			Name:     v.GetString("db.notification.name"),
			SSLMode:  v.GetString("db.ssl_mode"),
		},
		OrderDB: DatabaseConfig{
			Host:     v.GetString("db.host"),
			Port:     v.GetString("db.port"),
			User:     v.GetString("db.user"),
			Password: v.GetString("db.pass"),
			Name:     v.GetString("db.order.name"),
			SSLMode:  v.GetString("db.ssl_mode"),
		},
		ProductDB: DatabaseConfig{
			Host:     v.GetString("db.host"),
			Port:     v.GetString("db.port"),
			User:     v.GetString("db.user"),
			Password: v.GetString("db.pass"),
			Name:     v.GetString("db.product.name"),
			SSLMode:  v.GetString("db.ssl_mode"),
		},
		Server: ServerConfig{
			Port: v.GetString("server.port"),
			Env:  v.GetString("env"),
		},
		JWT: JWTConfig{
			Secret: v.GetString("jwt.secret"),
			Expiry: v.GetString("jwt.expiry"),
		},
		Upload: UploadConfig{
			Dir:       v.GetString("upload.dir"),
			MaxSizeMB: v.GetInt("upload.max_size"),
		},
	}

	// Ensure upload directory exists
	if err := ensureUploadDir(config.Upload.Dir); err != nil {
		return nil, err
	}

	return config, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Database defaults
	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", "5432")
	v.SetDefault("db.user", "postgres")
	v.SetDefault("db.pass", "postgres")
	v.SetDefault("db.account.name", "ybds_user")
	v.SetDefault("db.notification.name", "ybds_notification")
	v.SetDefault("db.order.name", "ybds_order_payment")
	v.SetDefault("db.product.name", "ybds_product_inventory")
	v.SetDefault("db.ssl_mode", "disable")

	// Server defaults
	v.SetDefault("server.port", "3000")
	v.SetDefault("env", "development")

	// JWT defaults
	v.SetDefault("jwt.expiry", "24h")

	// Upload defaults
	v.SetDefault("upload.dir", "./uploads")
	v.SetDefault("upload.max_size", 10) // 10MB

	// Map environment variables to viper keys
	mapEnvToConfig(v)
}

// mapEnvToConfig maps environment variables to viper configuration keys
func mapEnvToConfig(v *viper.Viper) {
	// Database mapping
	v.BindEnv("db.host", "DB_HOST")
	v.BindEnv("db.port", "DB_PORT")
	v.BindEnv("db.user", "DB_USER")
	v.BindEnv("db.pass", "DB_PASS")
	v.BindEnv("db.account.name", "DB_ACCOUNT_NAME")
	v.BindEnv("db.notification.name", "DB_NOTIFICATION_NAME")
	v.BindEnv("db.order.name", "DB_ORDER_NAME")
	v.BindEnv("db.product.name", "DB_PRODUCT_NAME")
	v.BindEnv("db.ssl_mode", "DB_SSL_MODE")

	// Server mapping
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("env", "ENV")

	// JWT mapping
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("jwt.expiry", "JWT_EXPIRY")

	// Upload mapping
	v.BindEnv("upload.dir", "UPLOAD_DIR")
	v.BindEnv("upload.max_size", "MAX_UPLOAD_SIZE")
}

// ensureUploadDir ensures that the upload directory exists
func ensureUploadDir(dir string) error {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("failed to create upload directory: %w", err)
		}
	}

	return nil
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

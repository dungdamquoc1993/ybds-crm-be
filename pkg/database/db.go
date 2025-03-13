package database

import (
	"fmt"
	"log"
	"time"

	"github.com/ybds/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConnections holds all database connections
type DBConnections struct {
	AccountDB      *gorm.DB
	NotificationDB *gorm.DB
	OrderDB        *gorm.DB
	ProductDB      *gorm.DB
}

// NewDatabaseConnections creates new database connections
func NewDatabaseConnections(cfg *config.Config) (*DBConnections, error) {
	// Initialize account database
	accountDB, err := newDatabase(&cfg.AccountDB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to account database: %w", err)
	}

	// Initialize notification database
	notificationDB, err := newDatabase(&cfg.NotificationDB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification database: %w", err)
	}

	// Initialize order database
	orderDB, err := newDatabase(&cfg.OrderDB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order database: %w", err)
	}

	// Initialize product database
	productDB, err := newDatabase(&cfg.ProductDB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product database: %w", err)
	}

	return &DBConnections{
		AccountDB:      accountDB,
		NotificationDB: notificationDB,
		OrderDB:        orderDB,
		ProductDB:      productDB,
	}, nil
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	return newDatabase(cfg)
}

// newDatabase creates a new database connection (internal function)
func newDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open connection to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get generic database object SQL
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping database to verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to database %s successfully", cfg.Name)
	return db, nil
}

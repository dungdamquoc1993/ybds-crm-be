package database

import (
	"log"

	"github.com/ybds/internal/models/account"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/internal/models/order"
	"github.com/ybds/internal/models/product"
	"github.com/ybds/pkg/database"
	"gorm.io/gorm"
)

// InitDatabases initializes all databases by auto-migrating their respective models
func InitDatabases(dbConn *database.DBConnections) error {
	log.Println("Initializing databases...")

	// Auto-migrate account models
	if err := migrateAccountModels(dbConn.AccountDB); err != nil {
		return err
	}

	// Auto-migrate notification models
	if err := migrateNotificationModels(dbConn.NotificationDB); err != nil {
		return err
	}

	// Auto-migrate order models
	if err := migrateOrderModels(dbConn.OrderDB); err != nil {
		return err
	}

	// Auto-migrate product models
	if err := migrateProductModels(dbConn.ProductDB); err != nil {
		return err
	}

	log.Println("Database initialization completed successfully")
	return nil
}

// InitDatabase initializes a single database by auto-migrating all models (legacy support)
func InitDatabase(db *gorm.DB) error {
	log.Println("Initializing database (legacy mode)...")

	// Auto-migrate account models
	if err := migrateAccountModels(db); err != nil {
		return err
	}

	// Auto-migrate notification models
	if err := migrateNotificationModels(db); err != nil {
		return err
	}

	// Auto-migrate order models
	if err := migrateOrderModels(db); err != nil {
		return err
	}

	// Auto-migrate product models
	if err := migrateProductModels(db); err != nil {
		return err
	}

	log.Println("Database initialization completed successfully")
	return nil
}

// migrateAccountModels auto-migrates account-related models
func migrateAccountModels(db *gorm.DB) error {
	log.Println("Migrating account models...")
	return db.AutoMigrate(
		&account.User{},
		&account.Role{},
		&account.UserRole{},
		&account.Guest{},
		&account.Address{},
	)
}

// migrateNotificationModels auto-migrates notification-related models
func migrateNotificationModels(db *gorm.DB) error {
	log.Println("Migrating notification models...")
	return db.AutoMigrate(
		&notification.Notification{},
		&notification.Channel{},
	)
}

// migrateOrderModels auto-migrates order-related models
func migrateOrderModels(db *gorm.DB) error {
	log.Println("Migrating order models...")
	return db.AutoMigrate(
		&order.Order{},
		&order.OrderItem{},
		&order.Shipment{},
	)
}

// migrateProductModels auto-migrates product-related models
func migrateProductModels(db *gorm.DB) error {
	log.Println("Migrating product models...")
	return db.AutoMigrate(
		&product.Product{},
		&product.Inventory{},
		&product.Price{},
		&product.InventoryTransaction{},
	)
}

package repositories

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models/order"
	"gorm.io/gorm"
)

// OrderRepository handles database operations for orders
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new instance of OrderRepository
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

// GetOrderByID retrieves an order by ID with all relations
func (r *OrderRepository) GetOrderByID(id uuid.UUID) (*order.Order, error) {
	var o order.Order
	err := r.db.Where("id = ?", id).
		Preload("Items").
		Preload("Shipment").
		First(&o).Error
	return &o, err
}

// GetOrderByTrackingNumber retrieves an order by shipment tracking number
func (r *OrderRepository) GetOrderByTrackingNumber(trackingNumber string) (*order.Order, error) {
	var o order.Order
	err := r.db.Joins("JOIN shipments ON orders.id = shipments.order_id").
		Where("shipments.tracking_number = ? AND shipments.deleted_at IS NULL", trackingNumber).
		Preload("Items").
		Preload("Shipment").
		First(&o).Error
	return &o, err
}

// GetAllOrders retrieves all orders with pagination and filtering
func (r *OrderRepository) GetAllOrders(page, pageSize int, filters map[string]interface{}) ([]order.Order, int64, error) {
	var orders []order.Order
	var total int64

	query := r.db.Model(&order.Order{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "payment_method":
			query = query.Where("payment_method = ?", value)
		case "payment_status":
			query = query.Where("payment_status = ?", value)
		case "order_status":
			query = query.Where("order_status = ?", value)
		case "created_by":
			query = query.Where("created_by = ?", value)
		case "from_date":
			query = query.Where("orders.created_at >= ?", value)
		case "to_date":
			query = query.Where("orders.created_at <= ?", value)
		case "phone_number":
			query = query.Where("customer_phone LIKE ?", "%"+value.(string)+"%")
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Preload("Items").
		Preload("Shipment").
		Find(&orders).Error

	return orders, total, err
}

// CreateOrder creates a new order
func (r *OrderRepository) CreateOrder(o *order.Order) error {
	return r.db.Create(o).Error
}

// UpdateOrder updates an existing order
func (r *OrderRepository) UpdateOrder(o *order.Order) error {
	return r.db.Save(o).Error
}

// DeleteOrder deletes an order by ID
func (r *OrderRepository) DeleteOrder(id uuid.UUID) error {
	return r.db.Delete(&order.Order{}, id).Error
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(id uuid.UUID, status order.OrderStatus) error {
	return r.db.Model(&order.Order{}).Where("id = ?", id).Update("order_status", status).Error
}

// GetOrderItemByID retrieves an order item by ID
func (r *OrderRepository) GetOrderItemByID(id uuid.UUID) (*order.OrderItem, error) {
	var item order.OrderItem
	err := r.db.Joins("JOIN orders ON order_items.order_id = orders.id").
		Where("order_items.id = ? AND orders.deleted_at IS NULL", id).
		First(&item).Error
	return &item, err
}

// GetOrderItemsByOrderID retrieves all items for an order
func (r *OrderRepository) GetOrderItemsByOrderID(orderID uuid.UUID) ([]order.OrderItem, error) {
	var items []order.OrderItem

	// Check if order exists and is not deleted
	var count int64
	if err := r.db.Model(&order.Order{}).Where("id = ? AND deleted_at IS NULL", orderID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	err := r.db.Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

// CreateOrderItem creates a new order item
func (r *OrderRepository) CreateOrderItem(item *order.OrderItem) error {
	return r.db.Create(item).Error
}

// UpdateOrderItem updates an existing order item
func (r *OrderRepository) UpdateOrderItem(item *order.OrderItem) error {
	return r.db.Save(item).Error
}

// DeleteOrderItem deletes an order item by ID
func (r *OrderRepository) DeleteOrderItem(id uuid.UUID) error {
	return r.db.Delete(&order.OrderItem{}, id).Error
}

// GetShipmentByID retrieves a shipment by ID
func (r *OrderRepository) GetShipmentByID(id uuid.UUID) (*order.Shipment, error) {
	var shipment order.Shipment
	err := r.db.Joins("JOIN orders ON shipments.order_id = orders.id").
		Where("shipments.id = ? AND orders.deleted_at IS NULL", id).
		First(&shipment).Error
	return &shipment, err
}

// GetShipmentByOrderID retrieves a shipment by order ID
func (r *OrderRepository) GetShipmentByOrderID(orderID uuid.UUID) (*order.Shipment, error) {
	// Check if order exists and is not deleted
	var count int64
	if err := r.db.Model(&order.Order{}).Where("id = ? AND deleted_at IS NULL", orderID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var shipment order.Shipment
	err := r.db.Where("order_id = ?", orderID).First(&shipment).Error
	return &shipment, err
}

// CreateShipment creates a new shipment
func (r *OrderRepository) CreateShipment(shipment *order.Shipment) error {
	return r.db.Create(shipment).Error
}

// UpdateShipment updates an existing shipment
func (r *OrderRepository) UpdateShipment(shipment *order.Shipment) error {
	return r.db.Save(shipment).Error
}

// DeleteShipment deletes a shipment by ID
func (r *OrderRepository) DeleteShipment(id uuid.UUID) error {
	return r.db.Delete(&order.Shipment{}, id).Error
}

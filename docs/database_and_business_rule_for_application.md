# Database and Business Rules Documentation

This document outlines both database validation requirements and business rules that need to be implemented at the application level.

## Cross-Database Relationships

Since our application uses multiple databases (account, notification, order, product), traditional database foreign key constraints cannot be enforced between databases. The following cross-database relationships must be validated at the application level:

1. **OrderItem to Inventory**: Validate that `InventoryID` in `OrderItem` exists in the `inventory` table in the product database.
2. **Order to User/Guest**: Validate that `CustomerID` in `Order` exists in either the `users` or `guests` table based on the `CustomerType`.
3. **Notification to Recipients**: Validate that `RecipientID` in `Notification` exists in the appropriate table based on the `RecipientType`.

## Business Rules and Validations

### Order and Payment Processing

#### Payment Method Rules
1. **Cash Payment**:
   - Payment is handled in person
   - No payment tracking system needed
   - Order can proceed to shipping once payment is confirmed
   - Status should be marked as "completed" immediately upon confirmation

2. **Bank Transfer**:
   - Manual verification of bank transfer required
   - Staff should verify the transfer before marking as "completed"
   - Order processing can begin after payment verification

3. **Cash on Delivery (COD)**:
   - No upfront payment required
   - Payment is collected by shipping carrier
   - Carrier handles payment collection and transfer
   - Final payment status updated based on delivery outcome

#### Payment Status Transitions
- **Pending**: Initial state for all orders
- **Completed**:
  - Cash: Set immediately upon staff confirmation
  - Bank Transfer: Set after transfer verification
  - COD: Set after successful delivery and payment collection
- **Failed**:
  - Order is cancelled
  - COD delivery fails
  - Customer refuses delivery/payment
  - Order is returned

#### Order Status Flow
1. **New Order Creation**:
   - Validate customer exists
   - Check inventory availability
   - Calculate total amount
   - Set initial payment status based on payment method

2. **Order Processing**:
   - Confirm payment (except COD)
   - Reserve inventory
   - Create shipment request
   - Update inventory quantities

3. **Order Completion**:
   - Update payment status
   - Release or convert inventory reservation
   - Send notifications

### Inventory Management

#### Inventory Transaction System
The inventory transaction system tracks all changes to product inventory without requiring client-side input. It is handled automatically by the application layer.

#### Transaction Types
1. **Inbound**:
   - New stock arrivals
   - Returns from customers
   - Inventory adjustments (positive)

2. **Outbound**:
   - Sales
   - Damages
   - Inventory adjustments (negative)

3. **Reservation**:
   - Temporary holds during checkout
   - Released if checkout fails
   - Converted to outbound if order completes

#### Implementation Guidelines
1. **Service Layer Responsibility**:
   ```go
   // Example: Processing an order
   func (s *OrderService) ProcessOrder(order *Order) error {
       tx := s.db.Begin()
       
       // Deduct inventory and log transaction
       for _, item := range order.Items {
           if err := s.inventoryService.DeductStock(item.InventoryID, item.Quantity); err != nil {
               tx.Rollback()
               return err
           }
       }
       
       return tx.Commit()
   }
   ```

2. **Inventory Service**:
   ```go
   func (s *InventoryService) DeductStock(inventoryID uuid.UUID, quantity int) error {
       // Update inventory
       if err := s.updateQuantity(inventoryID, -quantity); err != nil {
           return err
       }
       
       // Log transaction (automatic)
       return s.createTransaction(InventoryTransaction{
           InventoryID: inventoryID,
           Quantity: -quantity,
           Type: TransactionOutbound,
           Reason: ReasonSale,
       })
   }
   ```

### Other Model Validations

#### Address Model
- Ensure an address is associated with either a User or a Guest, but not both or neither.
- When a User or Guest is deleted, their associated addresses should be handled appropriately.

#### Price Model
- Prevent overlapping price periods for the same product.
- Ensure `EndDate` is after `StartDate` when `EndDate` is provided.

#### Notification System
- Validate recipient existence before sending notifications.
- Implement retry logic for failed notification attempts.

## Technical Considerations

### Audit and Security
- Implement comprehensive audit logging for critical operations.
- Use `created_by` and `updated_by` fields to track user actions.
- Implement data versioning for critical entities that require historical tracking.

### Performance Optimization
- Add application-level caching for frequently accessed data.
- Implement database connection pooling for each database.
- Consider read replicas for high-traffic scenarios. 
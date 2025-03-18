Order & Payment Management
Overview
This section describes the database tables used for managing orders, and shipments. The relationships between tables are defined to ensure scalability and maintainability in the future.

📌 Table: orders
The orders table contains general information about the order.
Column	        Data Type	Description
id	            UUID	Unique order ID
customer_id	    UUID	Customer ID (can be users.id or guests.id)
customer_type	  ENUM	Type of customer (user, guest)
payment         ENUM (cash, cod, bank_transfer)
discount        DECIMAL(10,2) discount for the order
total_price	    DECIMAL(10,2)	Total price of the order
order_status	  ENUM	Order status 
created_at	    TIMESTAMP	Order creation timestamp
updated_at	    TIMESTAMP	Last order update timestamp
Notes:
The order_status only stores the most recent status (no need for a separate order_status_logs table).
order_status ENUM (
    'pending_confirmation',   -- 🟡 Chờ xác nhận đơn hàng (khách để lại thông tin)
    'confirmed',             -- 🟢 Đã xác nhận (khách chốt đơn)
    'shipment_requested',    -- 📦 Đã tạo vận đơn (gửi yêu cầu bên vận chuyển)
    'packing',               -- 📦 Đóng hàng, chờ vận chuyển lấy hàng
    'shipped',               -- 🚚 Bên vận chuyển đã đến lấy hàng
    'delivered',             -- ✅ Giao hàng thành công
    'return_requested',      -- 🔄 Khách yêu cầu trả hàng
    'return_processing',     -- 🔄 Đơn hàng đang trong quá trình hoàn trả
    'returned',              -- 🔄 Hàng đã về kho (hoàn thành hoàn hàng)
    'canceled'               -- ❌ Đơn hàng bị hủy
);
packing => minus from inventory
returned => plus from inventory
Nếu trạng thái trước đó là packing or shipped rối mới chuyển thành canceled => plus from inventory


📌 Table: order_items
The order_items table stores details about the products in each order.
Column	          Data          Type	Description
id	              UUID	        Unique order item ID
order_id	        UUID	        Order ID (foreign key referencing orders)
inventory_id	    UUID	        Product ID (foreign key referencing products)
quantity	        INT	          Quantity of the product ordered
price_at_order	  DECIMAL(10,2)	Price of the product at the time the order was placed
Notes:
Stores the price at the time of order placement to avoid issues when the product price changes.
Each order can have multiple products, and each product will have a corresponding row in order_items.


📌 Table: shipments
The shipments table stores information about the shipment related to an order.
Column	          Data Type	      Description
id	              UUID	          Unique shipment ID
order_id	        UUID	          Order ID (foreign key referencing orders)
tracking_number	  VARCHAR(100)	  Shipment tracking number (e.g., from GHN, GHTK, etc.)
carrier	          VARCHAR(50)	    Shipping carrier (e.g., GHN, GHTK, Viettel Post, etc.)
created_at	      TIMESTAMP	      Shipment creation timestamp
Notes:
The status field includes additional statuses like returned and return completed for tracking returns.
Each order will have a corresponding shipment, but shipments can be tracked independently.


📌 Relationships:
1 customer - N orders
A customer can have multiple orders, whether they are a registered user or a guest.

1 order - N order_items
An order can contain multiple items, as it may include different products.

1 order - 1 shipment
Each order will have exactly one shipment record.

External relationship
N orders - 1 user/guest
N order_times  - 1 inventory

ERD Diagram for human read:
                                  +---------------------+
                                  |     shipments       |
                                  +---------------------+
                                  | id (PK)             |
                                  | order_id (FK)       |
                                  | tracking_number     |
                                  | created_at          |
                                  +---------------------+
                                         |
                                         |
+--------------------+             +------------------+             +--------------------+ 
|     customers      | 1         N |    orders        | 1         N |    order_items     |
+--------------------+             +------------------+             +--------------------+
| id (PK)            | <---------- | customer_id (FK) | <---------- | order_id (FK)      |
| name               |             | total_price      |             | product_id (FK)    |
| email              |                                               | quantity           |
| phone              |             | order_status     |             | price_at_order     |
| created_at         |             | created_at       |             |                    |
+--------------------+             | updated_at       |             +--------------------+
                                   +------------------+          
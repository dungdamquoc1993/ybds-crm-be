Product & Inventory Management
Overview
This section describes the database tables used for managing products, inventory, and prices. The relationships between tables are also defined to support easy expansion in the future.


📌 Table: products
The products table contains information about the products being sold, such as the name, description, and SKU.
Column	    Data Type	    Description
id	        UUID	        Unique product ID
name	    VARCHAR(255)	Product name
description	TEXT	        Product description
sku	        VARCHAR(50)	    SKU (Stock Keeping Unit)
category	VARCHAR(100)	Product category (e.g., football shoes, NFT)
image_url	TEXT	        Product image URL
created_at	TIMESTAMP	    Timestamp when the product was created
updated_at	TIMESTAMP	    Timestamp when the product was last updated
Note:
If you expand into selling NFTs or meme coins, you can add new categories.
The sku helps in uniquely identifying each product.


📌 Table: inventory
The inventory table tracks the available stock of each product, including variations in size, color, and location within the warehouse.
Column	    Data Type	    Description
id	        UUID	        Unique inventory entry ID
product_id	UUID	        Product ID (foreign key linking to products)
size	    VARCHAR(10)	    Shoe size (e.g., 41, 42, 43, etc.)
color	    VARCHAR(50)	    Shoe color (e.g., red, blue, etc.)
quantity	INT	Quantity    of the product in stock
location	VARCHAR(255)	Location in the warehouse (e.g., shelf, floor, etc.)
created_at	TIMESTAMP	    Timestamp when the inventory entry was created
updated_at	TIMESTAMP	    Timestamp when the inventory entry was last updated
Note:
If a product comes in multiple sizes and colors, multiple entries will exist in the inventory table for each combination.
The location field helps in tracking the product’s physical position in the warehouse (shelf, floor, etc.).


📌 Table: prices
The prices table stores the historical prices for each product, supporting price changes over time.
Column	    Data Type	    Description
id	        UUID	        Unique price entry ID
product_id	UUID	        Product ID (foreign key linking to products)
price	    DECIMAL(10,2)	Price of the product at this point in time
currency	VARCHAR(10)	    Currency unit (e.g., VND, USD, etc.)
start_date	TIMESTAMP	    Date when the price becomes effective
end_date	TIMESTAMP	    Date when the price ends (NULL if still active)
created_at	TIMESTAMP	    Timestamp when the price record was created
Note:
Every time the price changes, a new entry is added to the prices table instead of updating the price directly in the products table.
To get the current price of a product, find the row with end_date = NULL or the row with the latest start_date.
If there is a promotion, a temporary price can be set with an end_date when the promotion ends.


📌 Relationships:
Here is the relationship between the tables:
1 product - N inventory
A product can have multiple inventory records corresponding to different sizes, colors, or warehouse locations.

1 product - N prices
A product can have multiple price records to reflect price changes over time.

External relationship:
1 inventory - N order_itmes

ERD Diagram for human read:
+--------------------+             +------------------+             +--------------------+
|     products       | 1         N |    inventory     | 1         N |      prices        |
+--------------------+             +------------------+             +--------------------+
| id (PK)            | <---------- | product_id (FK)  | <---------- | product_id (FK)    |
| name               |             | size             |             | price              |
| description        |             | color            |             | currency           |
| sku                |             | quantity         |             | start_date         |
| category           |             | location         |             | end_date           |
| image_url          |             | created_at       |             | created_at         |
| created_at         |             | updated_at       |             |                    |
| updated_at         |             +------------------+             +--------------------+
+--------------------+


📌 Summary
The products table holds the basic details of the product.
The inventory table tracks the product’s variants (size, color) and its stock level.
The prices table keeps the historical pricing of each product.
This design is scalable and can accommodate new product categories, size/color combinations, and price changes easily as your business grows. The database schema also ensures that the system can manage products and inventory effectively with flexibility for future changes (e.g., adding NFTs or other digital goods).


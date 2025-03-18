Customer & Account Management
Overview
This section describes the database tables used for managing customer information, user accounts, roles, and addresses. It also provides flexibility for managing both registered users and guest customers, allowing for future expansion.

📌 Table: users
The users table stores information about users (including admins and registered customers).
Column	        Data Type	Description
id	            UUID	Unique user ID
username        VARCHAR(100)
email           VARCHAR(100)
phone           VARCHAR(100)
password_hash   TEXT NOT NULL,
salt            TEXT NOT NULL,
is_active       BOOLEAN NOT NULL DEFAULT TRUE,
created_at	    TIMESTAMP	User creation timestamp
updated_at	    TIMESTAMP	Last user update timestamp
Notes:
Admin and staff users are manually created.
A customer will only enter this table if they register an account.
If a customer has an account, customer_id refers to users.id and customer_type = 'user'.
If a customer does not have an account, customer_id refers to guests.id and customer_type = 'guest'.
No strict foreign key between orders and users or guests, allowing flexibility with hybrid users.


📌 Table: roles
The roles table defines the different roles available in the system.
Column	    Data Type	Description
id	        UUID	Unique role ID
name	    ENUM	Name of the role (admin, staff, customer, agent) // agent is AI Agent 
Notes:
Initially, the roles are limited to admin and staff. Over time, customer roles and more specific roles can be added.


📌 Table: user_roles
The user_roles table manages the association between users and their roles, supporting role-based access control (RBAC).
Column	Data Type	Description
id	UUID	Unique ID
user_id	UUID	User ID (foreign key referencing users)
role_id	UUID	Role ID (foreign key referencing roles)
Notes:
If needed, this table can be expanded to support more advanced RBAC (Role-Based Access Control), allowing users to have multiple roles with different permissions.


📌 Table: user_addresses
The user_addresses table stores the shipping addresses for users. It supports multiple addresses for each user.
Column	        Data Type	    Description
id	            UUID	        Unique address ID
user_id	        UUID	        User ID (foreign key, can be NULL if the address is for a guest)
guest_id	    UUID	        Guest ID (foreign key, can be NULL if the address is for a user)
address	        TEXT	        Full address
ward/town       VARCHAR(100)    ward/town
district        VARCHAR(100)    District
city/province	VARCHAR(100)	City
country	        VARCHAR(100)	Country (default is Vietnam)
is_default	    BOOLEAN	        Default shipping address flag
created_at	    TIMESTAMP	    Address creation timestamp
Notes:
Each address can be linked to either user_id or guest_id depending on whether the customer has registered.
This allows customers (whether users or guests) to have their shipping addresses stored and referenced easily.


📌 Table: guests
The guests table stores information about customers who have not yet registered as users but have placed orders.
Column	        Data Type	    Description
id	            UUID	        Unique guest ID
name	        VARCHAR(255)	Guest name
phone	        VARCHAR(20)	    Guest's phone number
email	        VARCHAR(255)	Guest's email address (optional)
created_at	    TIMESTAMP	    Timestamp when the guest entry was created
Notes:
If a guest later creates an account, their information can be migrated from the guests table to the users table.
Orders for guests will reference the guests table via guest_id instead of user_id.
When a guest creates an account, their guest data can be converted to a user, linking their previous orders to their new account.

📌 Relationships:
1 user - N addresses
A user can have multiple addresses for different delivery locations.

1 guest - N addresses
A guest can place multiple orders before registering as a user.

1 user - N roles
A user can have multiple roles depending on the system's access control needs.

External relationship:
1 user/guest - N orders
1 user - N notifications

ERD Diagram for human read:
+-----------------+                 +----------------+                 +-------------------+
|      users       | 1            N |  user_addresses |               1 |      guests       |
+--------------------+             +------------------+             +--------------------+
| id (PK)            | <----------- | user_id (FK)     |             | id (PK)            |
| name               |              | guest_id (FK)    |-----------> | name               |
| email              |              | address          |             | email              |
| phone              |              | city             |             | phone              |
| created_at         |              | country          |             | address_raw        |
| updated_at         |              | is_default       |             | created_at         |
+--------------------+              | created_at       |             +--------------------+
                                    +------------------+
                                             
+--------------------+               +------------------+               +-----------------+  
|      roles         | 1          N  |    user_roles    | N          1  |      users      |  
+--------------------+               +------------------+               +-----------------+  
| id (PK)            | <------------ | role_id (FK)     |               | id (PK)         |
| name               |               | user_id (FK)     |-------------> | name            |
+--------------------+               +------------------+               +-----------------+  

📌 Summary
The users table stores information about registered users and admin accounts.
The roles and user_roles tables define and manage user roles, allowing for future RBAC (Role-Based Access Control).
The user_addresses table stores shipping addresses for users and guests.
The guests table stores information about customers who haven't registered yet.
Flexible hybrid model where orders can reference either registered users (users.id) or guests (guests.id).
This system is designed to be flexible and scalable, allowing you to manage users, guest customers, and orders seamlessly while enabling future growth and user registration.


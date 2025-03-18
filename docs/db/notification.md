Notification System
Overview
This section describes the database tables used for the notification system. It allows for sending notifications to different types of recipients, including registered users, guests, potential customers, and partners. The system also supports multiple notification channels such as email, WebSocket, SMS, and Telegram.


📌 Table: notifications
The notifications table stores the basic information about the notification, including its content, recipient, and whether the notification has been read.
Column	            Data Type	    Description
id	                UUID	        Unique notification ID
recipient_id	    UUID	        ID of the recipient (can be users.id, guests.id, potential_customers.id, or partners.id)
recipient_type	    ENUM	        Type of recipient (user, guest, potential_customer, partner, other)
title	            VARCHAR(255)	Title of the notification
message	            TEXT	        Content of the notification
status	            ENUM	        Status of the notification (pending, sent, failed)
metadata	        JSON	        Additional data (e.g., { "order_id": "ABC123" })
is_read	            BOOLEAN	        Whether the notification has been read (default: false)
created_at	        TIMESTAMP	    imestamp when the notification was created
Notes:
recipient_id = NULL means the notification is sent to a non-registered recipient, such as a potential customer or partner.
The recipient_type allows flexibility, enabling notifications to be sent to any type of entity (user, guest, potential customer, etc.).
The status field tracks whether the notification is still in the queue (pending), successfully sent (sent), or failed (failed).
Metadata is used to store additional information (such as order ID or payment status) that can be useful for processing the notification.


📌 Table: notification_channels
The notification_channels table stores details about each channel through which the notification will be sent (WebSocket, email, SMS, Telegram, etc.).
Column	            Data Type	    Description
id	                UUID	        Unique channel ID
notification_id	    UUID	        Notification ID (foreign key linking to notifications)
channel	            ENUM	        Channel used for the notification (websocket, email, telegram, sms)
status	            ENUM	        Status of the notification sending (pending, sent, failed)
attempts	        INT	            Number of attempts to send the notification (used for retries)
response	        JSON	        Response from the sending system (e.g., SMTP response, API response)
created_at	        TIMESTAMP	    Timestamp when the channel entry was created
Notes:
Each notification can be sent through multiple channels (e.g., WebSocket for real-time notifications and email for asynchronous notifications).
If sending fails (status = failed), the system can retry based on the number of attempts.
The response field stores feedback from the system responsible for sending the notification (e.g., email server, SMS API, etc.).
Backend, AI Automation, or AI Agents will be responsible for processing the data and determining the appropriate sending method.


📌 Relationships:
1 notification - N notification_channels
A notification can be sent through multiple channels (e.g., WebSocket, email, SMS, Telegram).

1 notification - 1 recipient
A notification is directed at a recipient (which could be a user, guest, potential customer, or other).

External relationship: 
N notification - 1 User

ERD Diagram for human read:
+---------------------+             +-----------------------+  
|    notifications    | 1         N |   notification_channels|  
+---------------------+             +-----------------------+  
| id (PK)             | <---------- | notification_id (FK)   |  
| recipient_id (FK)   |             | channel                |  
| recipient_type      |             | status                 |  
| title               |             | attempts               |  
| message             |             | response               |  
                                    | created_at             |  
| status              |             +-----------------------+  
| is_read             |                                     
| metadata            |                                     
| created_at          |                                     
+---------------------+        

📌 Summary
The notifications table stores all the basic information about a notification, including its recipient, content, and metadata.
The notification_channels table tracks how the notification is delivered (via WebSocket, email, SMS, Telegram, etc.) and its status.
is_read is used to track whether the recipient has seen or interacted with the notification.
Metadata and status help manage notifications and keep track of their delivery status.
This design is flexible and allows the system to send notifications to different types of recipients (users, guests, potential customers, etc.) through different channels.

-- Connect to the account database
\c ybds_user

-- Add telegram_id column to users table
ALTER TABLE users
ADD COLUMN telegram_id BIGINT DEFAULT NULL;

-- Create index on telegram_id for faster lookups
CREATE INDEX idx_users_telegram_id ON users(telegram_id);

-- Comment on table and column
COMMENT ON COLUMN users.telegram_id IS 'Telegram chat ID for sending notifications'; 
-- Connect to the account database
\c ybds_user

-- Drop index first
DROP INDEX IF EXISTS idx_users_telegram_id;

-- Remove telegram_id column from users table
ALTER TABLE users
DROP COLUMN IF EXISTS telegram_id; 
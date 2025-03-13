#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# Check if psql is installed
if ! command -v psql &> /dev/null; then
  echo "Error: psql is not installed. Please install PostgreSQL client tools."
  exit 1
fi

# Set default values if not provided in .env
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASS=${DB_PASS:-postgres}
DB_NAME=${DB_NAME:-ybds}

# Function to run a migration file
run_migration() {
  local file=$1
  echo "Running migration: $file"
  
  PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $file
  
  if [ $? -eq 0 ]; then
    echo "Migration successful: $file"
  else
    echo "Migration failed: $file"
    exit 1
  fi
}

# Create database if it doesn't exist
echo "Checking if database exists..."
PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
  PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME"

# Run all migration files in order
echo "Running migrations..."
for file in $(find ./migrations -name "*.sql" | sort); do
  run_migration $file
done

echo "All migrations completed successfully!" 
#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Testing Authentication API ===${NC}"

# Base URL
BASE_URL="http://localhost:3000/api"

# Test 1: Register a new user
echo -e "\n${BLUE}Test 1: Register a new user${NC}"
RANDOM_EMAIL="user_$(date +%s)@example.com"
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${RANDOM_EMAIL}\", \"password\": \"password123\"}")

echo "Register Response:"
echo "$REGISTER_RESPONSE"

# Extract user_id from registration response
USER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"user_id":"[^"]*' | cut -d'"' -f4)
USERNAME=$(echo "$REGISTER_RESPONSE" | grep -o '"username":"[^"]*' | cut -d'"' -f4)

if [[ "$REGISTER_RESPONSE" == *"Registration successful"* ]]; then
  echo -e "${GREEN}✓ Registration successful${NC}"
  echo "User ID: $USER_ID"
  echo "Username: $USERNAME"
  echo "Email: $RANDOM_EMAIL"
else
  echo -e "${RED}✗ Registration failed${NC}"
fi

# Test 2: Login with the newly registered user
echo -e "\n${BLUE}Test 2: Login with newly registered user${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"${USERNAME}\", \"password\": \"password123\"}")

echo "Login Response:"
echo "$LOGIN_RESPONSE"

# Extract token from login response
TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [[ "$LOGIN_RESPONSE" == *"Authentication successful"* ]]; then
  echo -e "${GREEN}✓ Login successful${NC}"
  echo "Token received: ${TOKEN:0:20}..."
else
  echo -e "${RED}✗ Login failed${NC}"
fi

# Test 3: Login with admin user
echo -e "\n${BLUE}Test 3: Login with admin user${NC}"
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}')

echo "Admin Login Response:"
echo "$ADMIN_LOGIN_RESPONSE"

# Extract admin token
ADMIN_TOKEN=$(echo "$ADMIN_LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [[ "$ADMIN_LOGIN_RESPONSE" == *"Authentication successful"* ]]; then
  echo -e "${GREEN}✓ Admin login successful${NC}"
  echo "Admin token received: ${ADMIN_TOKEN:0:20}..."
else
  echo -e "${RED}✗ Admin login failed${NC}"
fi

# Test 4: Login with invalid credentials
echo -e "\n${BLUE}Test 4: Login with invalid credentials${NC}"
INVALID_LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "nonexistent", "password": "wrongpassword"}')

echo "Invalid Login Response:"
echo "$INVALID_LOGIN_RESPONSE"

if [[ "$INVALID_LOGIN_RESPONSE" == *"Authentication failed"* ]]; then
  echo -e "${GREEN}✓ Invalid login correctly rejected${NC}"
else
  echo -e "${RED}✗ Invalid login test failed${NC}"
fi

# Test 5: Register with existing email
echo -e "\n${BLUE}Test 5: Register with existing email${NC}"
DUPLICATE_REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${RANDOM_EMAIL}\", \"password\": \"password123\"}")

echo "Duplicate Register Response:"
echo "$DUPLICATE_REGISTER_RESPONSE"

if [[ "$DUPLICATE_REGISTER_RESPONSE" == *"Email or phone number already registered"* ]]; then
  echo -e "${GREEN}✓ Duplicate registration correctly rejected${NC}"
else
  echo -e "${RED}✗ Duplicate registration test failed${NC}"
fi

echo -e "\n${BLUE}=== Authentication API Tests Completed ===${NC}" 
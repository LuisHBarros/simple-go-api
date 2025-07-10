#!/bin/bash

# Simple API test script
BASE_URL="http://127.0.0.1:8080/api/v1"

echo "?? Testing SmarApp API"
echo "======================"

# Test health check
echo "1. Testing health check..."
curl -s "http://127.0.0.1:8080/health" | jq .
echo ""

# Register admin user
echo "2. Registering admin user..."
ADMIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "password123",
    "role": "admin"
  }')

ADMIN_TOKEN=$(echo $ADMIN_RESPONSE | jq -r '.token')
echo "Admin registered. Token: ${ADMIN_TOKEN:0:20}..."
echo ""

# Register regular user
echo "3. Registering regular user..."
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user1",
    "email": "user1@example.com",
    "password": "password123"
  }')

USER_TOKEN=$(echo $USER_RESPONSE | jq -r '.token')
echo "User registered. Token: ${USER_TOKEN:0:20}..."
echo ""

# Create a product (admin only)
echo "4. Creating a product (admin)..."
PRODUCT_RESPONSE=$(curl -s -X POST "$BASE_URL/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "Laptop",
    "description": "High-performance laptop",
    "price": 999.99,
    "stock": 10
  }')

PRODUCT_ID=$(echo $PRODUCT_RESPONSE | jq -r '.id')
echo "Product created with ID: $PRODUCT_ID"
echo ""

# List products (public)
echo "5. Listing products (public)..."
curl -s "$BASE_URL/products" | jq .
echo ""

# Create an order (user)
echo "6. Creating an order (user)..."
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d '{
    "product_id": '$PRODUCT_ID',
    "quantity": 2
  }' | jq .
echo ""

# Get user orders
echo "7. Getting user orders..."
curl -s "$BASE_URL/orders" \
  -H "Authorization: Bearer $USER_TOKEN" | jq .
echo ""

# Get chat history
echo "8. Getting chat history..."
curl -s "$BASE_URL/chat/history" \
  -H "Authorization: Bearer $USER_TOKEN" | jq .
echo ""

echo "? API tests completed!"
echo "?? To test WebSocket chat, connect to: ws://127.0.0.1:8080/api/v1/chat/ws"
echo "   Remember to include Authorization header with Bearer token"

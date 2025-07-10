# SmarApp API

A simple Go API with JWT authentication, role-based access control, WebSocket chat, and product management.

## Features

- **JWT Authentication**: Secure user authentication with role-based access control
- **User Roles**: Admin and User roles with different permissions
- **Product CRUD**: Complete product management (admin only for create/update/delete)
- **Order System**: Users can purchase products with automatic stock management
- **WebSocket Chat**: Real-time chat between users with message persistence
- **SQLite Database**: Lightweight database for development

## Quick Start

1. **Run the server:**
   ```bash
   cd go-api
   go run cmd/server/main.go
   ```

2. **The API will be available at:** `http://127.0.0.1:8080` (or set PORT environment variable)

3. **Access Swagger Documentation:** `http://127.0.0.1:8080/docs/index.html`

## API Documentation

The API includes comprehensive Swagger documentation available at `/docs/index.html` when the server is running.

### Key Features:
- **Interactive API Explorer**: Test endpoints directly from the browser
- **Request/Response Examples**: See exactly what data to send and expect
- **Authentication Support**: Built-in JWT token authentication
- **Model Definitions**: Complete data structure documentation

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user
- `GET /api/v1/profile` - Get user profile (protected)

### Products
- `GET /api/v1/products` - List all products (public)
- `GET /api/v1/products/:id` - Get product by ID (public)
- `POST /api/v1/products` - Create product (admin only)
- `PUT /api/v1/products/:id` - Update product (admin only)
- `DELETE /api/v1/products/:id` - Delete product (admin only)

### Orders
- `POST /api/v1/orders` - Create order (buy product)
- `GET /api/v1/orders` - Get user's orders
- `GET /api/v1/orders/:id` - Get specific order
- `GET /api/v1/admin/orders` - Get all orders (admin only)

### Chat
- `GET /api/v1/chat/ws` - WebSocket connection for real-time chat
- `GET /api/v1/chat/history` - Get chat history

## Example Usage

### 1. Register an Admin User
```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "password123",
    "role": "admin"
  }'
```

### 2. Login
```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

### 3. Create a Product (Admin only)
```bash
curl -X POST http://127.0.0.1:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Laptop",
    "description": "High-performance laptop",
    "price": 999.99,
    "stock": 10
  }'
```

### 4. Register a Regular User
```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user1",
    "email": "user1@example.com",
    "password": "password123"
  }'
```

### 5. Buy a Product
```bash
curl -X POST http://127.0.0.1:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer USER_JWT_TOKEN" \
  -d '{
    "product_id": 1,
    "quantity": 2
  }'
```

## WebSocket Chat

Connect to the WebSocket endpoint with authentication:
```javascript
const token = "YOUR_JWT_TOKEN";
const ws = new WebSocket(`ws://127.0.0.1:8080/api/v1/chat/ws`, [], {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

// Send a message
ws.send(JSON.stringify({
  message: "Hello, everyone!"
}));
```

**For easy WebSocket testing, open `websocket_test.html` in your browser.**

## Environment Variables

- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - SQLite database file path (default: ./smarapp.db)
- `JWT_SECRET` - JWT signing secret (default: your-secret-key-change-this-in-production)

## Database Schema

The API automatically creates the following tables:
- `users` - User accounts with roles
- `products` - Product catalog
- `orders` - Purchase orders
- `chat_messages` - Chat message history

## CORS Configuration

The API is configured with permissive CORS settings for development:
- ? **All Origins Allowed**: Any domain can make requests to the API
- ? **Credentials Supported**: JWT tokens work with cross-origin requests
- ? **All Methods**: GET, POST, PUT, DELETE, OPTIONS, PATCH
- ? **Authorization Headers**: Supports Bearer token authentication

## Security Notes

- Change the JWT secret in production
- Configure CORS appropriately for your frontend domain in production
- Use HTTPS in production
- Consider rate limiting for production use
- Current CORS settings are permissive for development - restrict in production

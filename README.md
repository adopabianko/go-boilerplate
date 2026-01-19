# Go Boilerplate

A comprehensive Go boilerplate project following Clean Architecture, featuring Gin, Gorm, RabbitMQ, Redis, PostgreSQL, and Swagger documentation.

## Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make
- [Air](https://github.com/air-verse/air) (for live reload)
- [Golang Migrate](https://github.com/golang-migrate/migrate) (CLI)

## Setup

1.  **Clone the repository**

2.  **Environment Variables**
    Copy the example environment file:
    ```bash
    cp .env.example .env
    ```
    (Ensure `.env` contains correct credentials. Default `docker-compose` credentials should match).

3.  **Start Infrastructure**
    Start PostgreSQL, Redis, RabbitMQ, and Minio:
    ```bash
    make docker-up
    ```

4.  **Run Migrations**
    Initialize the database schema:
    ```bash
    make migrate-up
    ```

## Running the Application

### Development Mode (Live Reload)
```bash
make air
```

### Standard Run
```bash
go run cmd/api/main.go
```

The server will start at `http://localhost:8080`.

## API Documentation

Swagger documentation is available at:
`http://localhost:8080/swagger/index.html`

To regenerate docs:
```bash
make swagger
```

## Testing with Curl

Here are example commands to test the endpoints.

### 1. Health Check
```bash
curl -i http://localhost:8080/health
```

### 2. Register User
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 3. Login User
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```
*Note: Copy the `token` from the response for the following requests.*

### 4. List Users (Pagination)
Replace `<TOKEN>` with the JWT token obtained from login.
```bash
curl -i -X GET "http://localhost:8080/api/v1/users?page=1&limit=5&order=created_at%20desc" \
  -H "Authorization: Bearer <TOKEN>"
```

### 5. Get Current User Profile
```bash
curl -i -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <TOKEN>"
```

### 6. User CRUD Operations (Protected)

**Get User Detail:**
```bash
curl -i -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer <TOKEN>"
```

**Update User:**
```bash
curl -i -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"email": "updated@example.com"}'
```

**Delete User:**
```bash
curl -i -X DELETE http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer <TOKEN>"
```

### 7. Get Products (External API)
Fetch products from DummyJSON integration.
```bash
curl -i -X GET "http://localhost:8080/api/v1/products?limit=5&page=1"
```

### 8. Check Payment Status (gRPC)
Prerequisite: Run the dummy gRPC server first.
```bash
make run-dummy-grpc
```

Then call the API:
```bash
curl -i -X GET http://localhost:8080/api/v1/payments/TRX-123
```
Test cases:
- `TRX-123` -> SUCCESS
- `fail` -> FAILED
- `pending` -> PENDING

## gRPC Code Generation
If you modify `.proto` files in `api/proto/`, run:
```bash
make proto
```

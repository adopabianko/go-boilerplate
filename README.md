# Go Boilerplate

A comprehensive Go boilerplate project following Clean Architecture, featuring Gin, Gorm, RabbitMQ, Redis, PostgreSQL, and Swagger documentation.

## Main Directory Structure

### `/cmd`
Main entry points of the application.
- **/api**: Contains `main.go` which serves as the *bootstrapper* to run the API server. Dependencies (database, config, router, etc.) are initialized here.
- **/dummy_grpc_server**: (Optional) Entry point for a dummy gRPC server if present.

### `/internal`
Contains private application code that should not be imported by external projects. This is the core of clean architecture.
- **/config**: Module for loading application configuration (e.g., from `.env` files or environment variables).
- **/entity**: Contains data structure definitions (Structs) representing business domain objects (e.g., `User`, `Product`) and database tables.
- **/repository**: _Data Access Layer_. Focuses solely on database queries or data storage (SQL, Redis, etc.). Repositories implement interfaces defined for data access.
- **/usecase**: _Business Logic Layer_. Contains the main business logic. Usecases combine data from repositories and perform validation or business processes before data is sent to the delivery layer.
- **/delivery**: _Presentation Layer_.
  - **/http**: Handles HTTP requests (REST API). Consists of:
    - **handler**: Receives requests, invokes usecases, and returns JSON responses.
    - **middleware**: Interceptor logic such as Auth, Logging, Recovery.
    - **router**: URL route definitions and mapping to handlers.
- **/dto**: _Data Transfer Objects_. Simple structs used to define API request (Input) and response (Output) data shapes, separating input validation from core entities.
- **/gateway**: Clients for interacting with external services (Third-party APIs, other Microservices).

### `/pkg`
Contains public libraries or utilities that can be reused in other parts of the project (or even other projects).
- **/auth**: Authentication utilities (JWT, Password hashing).
- **/database**: Database connection configurations (Postgres, MySQL).
- **/logger**: Wrapper for logging systems (e.g., Zap, Logrus).
- **/response**: Helper for standardizing API response formats (Success/Error wrapping).
- **/errors**: Custom error definitions.
- **/redis**: Redis connection helper.
- **/rabbitmq**: RabbitMQ connection helper (Message Broker).
- **/minio**: Helper for uploading files to Object Storage (MinIO/S3).
- **/pb**: Generated code for Protocol Buffers (gRPC).

### `/api`
Contains API contract definitions, usually OpenAPI/Swagger specification files or `.proto` files.

### `/docs`
Additional project documentation, such as architectural diagrams, technical guides, or this structure explanation.

### `/migrations`
Database schema manager. Contains SQL files to create (up) or delete (down) database tables to keep the schema synchronized across environments.

## Data Flow

In general, the data flow for an API request is unidirectional:

1. **Incoming Request** → `Delivery (HTTP/Handler)`
2. `Handler` calls → `Usecase`
3. `Usecase` processes logic & calls → `Repository`
4. `Repository` fetches data from → `Database`
5. **Return Data** → `Repository` → `Usecase` → `Delivery` → **JSON Response**

This separation ensures that changes in one layer (e.g., switching Databases) do not break other layers (e.g., Business Logic).

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


4.  **Generate RSA Keys**
    Generate keys for JWT signing:
    ```bash
    make cert
    ```

5.  **Run Migrations**
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

### 3. Login User (Get Tokens)
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```
**Response:**
```json
{
  "access_token": "ey...<SHORT_LIVED_TOKEN>...",
  "refresh_token": "ey...<LONG_LIVED_TOKEN>..."
}
```
*Note: Copy the `access_token` for authenticating requests. Use `refresh_token` to get a new access token when it expires.*

### 4. Refresh Access Token
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<YOUR_REFRESH_TOKEN>"
  }'
```

### 5. Access Protected API (List Users)
Replace `<ACCESS_TOKEN>` with the token obtained from login.
```bash
curl -i -X GET "http://localhost:8080/api/v1/users?page=1&limit=5&order=created_at%20desc" \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
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

# Assessment Backend

REST API mini e-commerce built with Go, Gin, and PostgreSQL.

The API supports customer registration and login, JWT authentication, role-based access control, product catalog management, transactional checkout, stock validation, order history, and Swagger documentation.

## Tech Stack

- Go
- Gin
- PostgreSQL 16
- pgx
- JWT
- Argon2id password hashing
- Docker Compose
- Swaggo / Swagger UI

## Features

- Customer registration and login
- Secure password hashing with Argon2id
- JWT authentication
- Role-based access control: `admin` and `customer`
- Product catalog with category, pagination, filters, and sorting
- Admin product and category management
- Transactional checkout
- Atomic stock deduction with row locking
- Order history for customers
- Admin order listing and status management
- Order lifecycle: `PENDING`, `COMPLETED`, and `CANCELLED`
- Automatic stock restoration for cancelled orders
- Idempotency key support for safe checkout retries
- Swagger API documentation
- Automatic database migrations on application startup

## Architecture

```text
routes → handler → service → repository → PostgreSQL
```

- `routes`: endpoint and middleware registration
- `handler`: HTTP request/response handling
- `service`: business rules
- `repo`: database queries and transactions
- `models`: request, response, and domain models

## Entity Relationship Diagram

```mermaid
erDiagram
    USERS {
        uuid id PK
        varchar name
        varchar email UK
        varchar password_hash
        varchar role
        timestamptz created_at
    }

    CATEGORIES {
        uuid id PK
        varchar name UK
    }

    PRODUCTS {
        uuid id PK
        uuid category_id FK
        varchar name
        text description
        bigint price
        int stock
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    ORDERS {
        uuid id PK
        uuid user_id FK
        bigint total_amount
        varchar status
        varchar idempotency_key
        timestamptz created_at
    }

    ORDER_ITEMS {
        uuid id PK
        uuid order_id FK
        uuid product_id FK
        int quantity
        bigint unit_price
        bigint subtotal
    }

    USERS ||--o{ ORDERS : places
    CATEGORIES o|--o{ PRODUCTS : classifies
    ORDERS ||--|{ ORDER_ITEMS : contains
    PRODUCTS ||--o{ ORDER_ITEMS : purchased_as
```

Notes:

- Product deletion uses soft delete through `deleted_at`.
- A customer can only access their own orders.
- Each order can only contain one row per product.
- Duplicate product items in a checkout request are aggregated before processing.
- `(user_id, idempotency_key)` is unique to prevent duplicate checkout processing.
- New orders are created with `PENDING` status.
- An admin can change a `PENDING` order to `COMPLETED` or `CANCELLED`.
- `COMPLETED` and `CANCELLED` are final statuses; cancelling an order restores its product stock atomically.

## Prerequisites

- Go 1.19+
- Docker and Docker Compose
- PostgreSQL is provided through Docker Compose

## Getting Started

Clone the repository:

```bash
git clone https://github.com/muhammadwahyupratamaa/assessment-backend.git
cd assessment-backend
```

Create the environment file:

```bash
cp .env.example .env
```

Example `.env` configuration:

```env
APP_ENV=development
APP_PORT=8081

POSTGRES_USER=shop
POSTGRES_PASSWORD=shop_pass
POSTGRES_DB=shop

DATABASE_URL=postgres://shop:shop_pass@localhost:5433/shop?sslmode=disable

JWT_SECRET=change-this-to-a-long-random-secret
JWT_EXPIRES_HOURS=24

ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=AdminPassword123!
```

Start PostgreSQL:

```bash
docker compose up -d --wait db
```

Run the API:

```bash
go run ./cmd/api
```

Database migrations run automatically when the application starts. The configured admin account is also seeded automatically if it does not already exist.

## API Documentation

Swagger UI is available while the server is running:

```text
http://localhost:8081/swagger/index.html
```

To regenerate Swagger documentation after changing annotations:

```fish
/home/pra/go/bin/swag fmt
/home/pra/go/bin/swag init -g ./cmd/api/main.go --parseInternal -o ./docs
```

## Main Endpoints

| Method | Endpoint | Access | Description |
| --- | --- | --- | --- |
| GET | `/health` | Public | Health check |
| POST | `/api/v1/auth/register` | Public | Register a customer |
| POST | `/api/v1/auth/login` | Public | Login and receive JWT |
| GET | `/api/v1/me` | Authenticated | Get current user profile |
| GET | `/api/v1/categories` | Public | Get all categories |
| POST | `/api/v1/categories` | Admin | Create category |
| GET | `/api/v1/products` | Public | Get product list |
| GET | `/api/v1/products/:id` | Public | Get product detail |
| POST | `/api/v1/products` | Admin | Create product |
| PUT | `/api/v1/products/:id` | Admin | Update product |
| DELETE | `/api/v1/products/:id` | Admin | Soft-delete product |
| POST | `/api/v1/orders` | Customer | Checkout order |
| GET | `/api/v1/orders` | Customer | Get own order history |
| GET | `/api/v1/orders/:id` | Customer | Get own order detail |
| GET | `/api/v1/admin/orders` | Admin | Get all orders |
| PATCH | `/api/v1/admin/orders/:id/status` | Admin | Update a pending order to `COMPLETED` or `CANCELLED` |

## Authentication

Protected endpoints require a JWT token:

```http
Authorization: Bearer <your_jwt_token>
```

The admin account uses the values from `ADMIN_EMAIL` and `ADMIN_PASSWORD` in the environment file.

## Product Query Parameters

```text
GET /api/v1/products?page=1&limit=10&category_id=<uuid>&min_price=10000&max_price=1000000&in_stock=true&sort=price&order=asc
```

Supported query parameters:

| Parameter | Description |
| --- | --- |
| `page` | Page number, default `1` |
| `limit` | Items per page, default `10`, maximum `100` |
| `category_id` | Filter by category UUID |
| `min_price` | Minimum price |
| `max_price` | Maximum price |
| `in_stock` | Filter products with stock |
| `sort` | `price` or `created_at` |
| `order` | `asc` or `desc` |

## Checkout Behavior

Checkout is handled in a single PostgreSQL transaction.

1. Duplicate product IDs are aggregated.
2. Product rows are locked using `FOR UPDATE`.
3. Product stock is validated.
4. Product stock is reduced atomically.
5. An order is created with `PENDING` status, along with its order items.
6. Any failure causes the complete transaction to roll back.

An admin can update a `PENDING` order to:

- `COMPLETED`: the order is finalized and stock remains deducted.
- `CANCELLED`: stock is restored in the same database transaction.

Both `COMPLETED` and `CANCELLED` are final statuses and cannot be changed again.

An optional `Idempotency-Key` header can be sent when checking out. Repeating the same request with the same key returns the existing order and does not reduce stock again.

```http
Idempotency-Key: checkout-unique-key-001
```

## Testing

Run all tests:

```bash
go test ./...
```

Run formatting:

```bash
gofmt -w cmd/api internal
```

## Security Notes

- Passwords are hashed using Argon2id.
- JWTs are signed using a server-side secret.
- Role-based middleware protects admin-only endpoints.
- Input validation is applied at the request boundary.
- Checkout uses database transactions and row locking to prevent overselling.
- Sensitive environment values are excluded from version control.
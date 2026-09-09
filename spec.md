# Specification: Smart AI POS & Merchant Engine

This document provides a production-grade, architectural, and engineering specification for the **Smart AI POS & Merchant Engine**. The system is built with a backend in Golang utilizing Clean/Hexagonal Architecture, a frontend in React/Next.js (App Router), an AI-driven receipt scanner for automated stock/expense ledger entry, PostgreSQL as the system of record, Redis for distributed concurrency & idempotency locks, and modular Terraform for cloud infrastructure.

---

## 1. System & Hexagonal Architecture

The backend follows the principles of Clean/Hexagonal (Ports & Adapters) Architecture. This isolates business logic (Domain & Ports) from technology-specific infrastructure details (Adapters).

```
                  +--------------------------------------------+
                  |                 ADAPTERS                   |
                  |                                            |
                  |   +------------------------------------+   |
                  |   |              HANDLERS              |   |
                  |   |  - HTTP REST (gin/fiber)           |   |
                  |   |  - WebSockets (real-time sync)     |   |
                  |   +-----------------+------------------+   |
                  |                     |                      |
                  |                     | calls                |
                  |                     v                      |
                  |   +------------------------------------+   |
                  |   |            DRIVER PORTS            |   |
                  |   |  - OrderUseCase                    |   |
                  |   |  - ProductUseCase                  |   |
                  |   |  - ReceiptUseCase                  |   |
                  |   +-----------------+------------------+   |
                  |                     |                      |
                  |                     | implements           |
                  |                     v                      |
                  |   +------------------------------------+   |
                  |   |            DOMAIN CORE             |   |
                  |   |  - Business Rules                  |   |
                  |   |  - Domain Entities (Product, Order)|   |
                  |   +-----------------+------------------+   |
                  |                     |                      |
                  |                     | uses                 |
                  |                     v                      |
                  |   +------------------------------------+   |
                  |   |            DRIVEN PORTS            |   |
                  |   |  - ProductRepository (DB SPI)      |   |
                  |   |  - OrderRepository (DB SPI)        |   |
                  |   |  - LockService (Redis SPI)         |   |
                  |   |  - VisionService (LLM SPI)         |   |
                  |   +-----------------+------------------+   |
                  |                     |                      |
                  |                     | implemented by       |
                  |                     v                      |
                  |   +------------------------------------+   |
                  |   |         INFRASTRUCTURE             |   |
                  |   |  - PG Repository (pgx / GORM)      |   |
                  |   |  - Redis Cache & Lock Client       |   |
                  |   |  - Gemini/OpenAI API Client        |   |
                  |   +------------------------------------+   |
                  |                                            |
                  +--------------------------------------------+
```

### Directory Structure

The complete repository file layout is defined below:

```
/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                 # Application entry point & dependency injection wire-up
│   ├── internal/
│   │   ├── domain/                     # Pure domain models (no external dependencies, framework agnostic)
│   │   │   ├── product.go              # Product and stock models
│   │   │   ├── order.go                # Order & OrderItem models
│   │   │   ├── receipt.go              # Receipt & OCR result models
│   │   │   └── payment.go              # Payment transactions and webhook event models
│   │   ├── ports/                      # Inbound (Driver) & Outbound (Driven) interfaces
│   │   │   ├── inbound/                # Use Cases (Driver Ports)
│   │   │   │   ├── product_usecase.go
│   │   │   │   ├── order_usecase.go
│   │   │   │   └── receipt_usecase.go
│   │   │   └── outbound/               # Repositories & Clients (Driven Ports / SPI)
│   │   │       ├── product_repo.go
│   │   │       ├── order_repo.go
│   │   │       ├── lock_service.go
│   │   │       └── vision_client.go
│   │   └── adapters/                   # Technology implementations
│   │       ├── handlers/               # Driving Adapters
│   │       │   ├── rest/               # Gin or Fiber HTTP server & handlers
│   │       │   │   ├── server.go
│   │       │   │   ├── product_handler.go
│   │       │   │   ├── order_handler.go
│   │       │   │   └── receipt_handler.go
│   │       │   └── ws/                 # WebSocket handler for stock push updates
│   │       │       └── hub.go
│   │       └── infrastructure/         # Driven Adapters
│   │           ├── pg/                 # PostgreSQL implementation using pgx/v5 or GORM
│   │           │   ├── db.go
│   │           │   ├── product_repo_impl.go
│   │           │   └── order_repo_impl.go
│   │           ├── redis/              # Redis locking and rate-limiting implementation
│   │           │   ├── client.go
│   │           │   └── lock_service_impl.go
│   │           └── vision/             # OpenAI Vision or Google Gemini client
│   │               └── client_impl.go
│   ├── pkg/                            # Shared libraries (logger, config, custom errors)
│   │   ├── logger/
│   │   └── config/
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── app/                            # Next.js App Router root
│   │   ├── layout.tsx                  # Global HTML wrapper, theme-provider, React Query Provider
│   │   ├── page.tsx                    # Landing / redirect to Dashboard
│   │   ├── dashboard/                  # Merchant Dashboard page
│   │   │   └── page.tsx
│   │   ├── checkout/                   # POS Checkout register page
│   │   │   └── page.tsx
│   │   └── receipts/                   # Receipt uploader / AI auditing page
│   │       └── page.tsx
│   ├── components/                     # Shared UI components (Radix / Tailwind)
│   │   ├── ui/                         # Atomic components (button, input, toast, dialog)
│   │   ├── checkout-terminal.tsx       # Real-time stock sync grid & cart
│   │   ├── sales-chart.tsx             # TanStack-Query driven dashboard analytics
│   │   └── receipt-uploader.tsx        # Upload component with OCR previews
│   ├── hooks/                          # Custom React Hooks
│   │   ├── use-websocket.ts            # Hook for handling real-time push events
│   │   └── use-checkout.ts             # Cart & state mutation hook
│   ├── services/                       # API layer using Axios / TanStack Query
│   │   └── api.ts
│   ├── types/                          # TypeScript definitions mirroring Go entities
│   │   └── index.ts
│   ├── package.json
│   ├── tailwind.config.js
│   └── tsconfig.json
├── terraform/
│   ├── modules/                        # Reusable modular cloud infrastructure (AWS provider)
│   │   ├── vpc/
│   │   ├── rds/                        # PostgreSQL RDS instances (Aurora Serverless v2 or RDS Multi-AZ)
│   │   ├── redis/                      # ElastiCache Redis cluster
│   │   ├── s3/                         # S3 bucket for Receipt image uploads with KMS encryption
│   │   └── app_runner/                 # App Runner or ECS container runtime for Go server
│   └── environments/
│       └── staging/                    # Main test environment deployment
│           ├── main.tf
│           ├── variables.tf
│           └── outputs.tf
└── .github/
    └── workflows/
        └── ci.yml                      # CI build, lint, test (Go & NextJS), terraform-validate pipeline
```

---

## 2. Database Schema & Concurrency Management

### PostgreSQL Schema (The SSOT)

To guarantee ACID properties and consistent stock during parallel sales checkouts, we define a structured schema with relational foreign keys and indices.

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Products Table
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(12, 2) NOT NULL CHECK (price >= 0),
    stock_quantity INT NOT NULL CHECK (stock_quantity >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_sku ON products(sku);

-- Orders Table
CREATE TYPE order_status AS ENUM ('PENDING', 'COMPLETED', 'CANCELLED', 'REFUNDED');

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id VARCHAR(128) UNIQUE NOT NULL, -- External txn ID
    total_amount NUMERIC(12, 2) NOT NULL CHECK (total_amount >= 0),
    status order_status NOT NULL DEFAULT 'PENDING',
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orders_transaction ON orders(transaction_id);

-- Order Items Table (Junction Table)
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12, 2) NOT NULL CHECK (unit_price >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- Expense Audits (Receipt Data Captured from OCR)
CREATE TABLE receipt_audits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_name VARCHAR(255) NOT NULL,
    receipt_date TIMESTAMP WITH TIME ZONE NOT NULL,
    total_amount NUMERIC(12, 2) NOT NULL,
    ocr_raw_json JSONB NOT NULL,
    s3_image_url VARCHAR(512) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### High-Concurrency Stock Protections

During high-traffic checkout scenarios (e.g., flash sales, peak hours), multiple clerks or terminals can attempt to buy the same item, leading to **double-spend** or **negative stock** anomalies if read-then-write updates are uncoordinated.

#### Go Implementation Pattern (Pessimistic Locking)
We utilize PostgreSQL **Pessimistic Locking (`SELECT FOR UPDATE`)** executed strictly within an isolated database transaction to guarantee inventory consistency.

```go
// Inside pg/product_repo_impl.go
func (r *ProductRepositoryImpl) DeductStockWithLock(ctx context.Context, tx *sql.Tx, sku string, quantity int) error {
    var currentStock int
    var productID uuid.UUID

    // Lock the specific product row for writes until transaction commits or rolls back
    query := `SELECT id, stock_quantity FROM products WHERE sku = $1 FOR UPDATE`
    err := tx.QueryRowContext(ctx, query, sku).Scan(&productID, &currentStock)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return domain.ErrProductNotFound
        }
        return err
    }

    if currentStock < quantity {
        return domain.ErrInsufficientStock
    }

    newStock := currentStock - quantity
    updateQuery := `UPDATE products SET stock_quantity = $1, updated_at = NOW() WHERE id = $2`
    _, err = tx.ExecContext(ctx, updateQuery, newStock, productID)
    return err
}
```

### Redis Distributed Locks for Webhook Idempotency

When payment gateways (like Stripe, Adyen, or PayPal) post-payment webhooks to confirm an invoice completion, network retries can send the identical webhook payload multiple times. To guarantee **exactly-once processing (idempotency)** and prevent double-fulfillment, we implement Redis distributed locks.

#### Idempotency Key Mechanics
1. On payment webhook arrival, extract the unique webhook payment reference or `X-Idempotency-Key` header.
2. Attempt to acquire an exclusive lock in Redis:
   - Key: `lock:payment:idempotency:<idempotency_key>`
   - Value: `PROCESSING`
   - Expiration (TTL): 15 seconds (bounds processing time).
3. If acquisition fails, respond with `409 Conflict` or retry block.
4. If acquisition succeeds, check the PostgreSQL `orders` table to see if `idempotency_key` is already marked `COMPLETED`.
5. Execute payment fulfillment, update DB.
6. Set key `payment:idempotency:<idempotency_key>:status` to `SUCCESS` with 24-hour TTL, and release the processing lock.

```go
// Inside redis/lock_service_impl.go
func (s *LockServiceImpl) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
    // SET key value NX PX ttl
    success, err := s.redisClient.SetNX(ctx, key, "locked", ttl).Result()
    if err != nil {
        return false, err
    }
    return success, nil
}

func (s *LockServiceImpl) ReleaseLock(ctx context.Context, key string) error {
    _, err := s.redisClient.Del(ctx, key).Result()
    return err
}
```

---

## 3. Go Clean Interfaces & Ports

To facilitate testability, clean separation, and dependency injection, we establish strict interface declarations.

### Core Domain Entities

```go
package domain

import (
	"time"
	"github.com/google/uuid"
)

type Product struct {
	ID            uuid.UUID `json:"id"`
	SKU           string    `json:"sku"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Order struct {
	ID             uuid.UUID   `json:"id"`
	TransactionID  string      `json:"transaction_id"`
	TotalAmount    float64     `json:"total_amount"`
	Status         string      `json:"status"` // PENDING, COMPLETED, CANCELLED
	IdempotencyKey string      `json:"idempotency_key"`
	Items          []OrderItem `json:"items"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
}
```

### Driven Ports (Outbound Interfaces - SPI)

```go
package outbound

import (
	"context"
	"database/sql"
	"pos-one-repo/backend/internal/domain"
)

type ProductRepository interface {
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	GetBySKUWithLock(ctx context.Context, tx *sql.Tx, sku string) (*domain.Product, error)
	UpdateStock(ctx context.Context, tx *sql.Tx, sku string, newQty int) error
	ListAll(ctx context.Context) ([]domain.Product, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type LockService interface {
	AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, key string) error
}

type VisionClient interface {
	AnalyzeReceipt(ctx context.Context, imageBytes []byte) (*domain.ReceiptOCRResult, error)
}
```

### Driver Ports (Inbound Interfaces - Use Cases)

```go
package inbound

import (
	"context"
	"pos-one-repo/backend/internal/domain"
)

type CheckoutRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Items []struct {
		SKU      string `json:"sku" binding:"required"`
		Quantity int    `json:"quantity" binding:"required,gt=0"`
	} `json:"items" binding:"required,dive"`
}

type OrderUseCase interface {
	Checkout(ctx context.Context, req CheckoutRequest) (*domain.Order, error)
}
```

---

## 4. Next.js Dashboard UI Specs

The frontend is a single-page management application containing three views structured via Next.js App Router folders.

### State & Server Mutation: TanStack Query (React Query)
To ensure reliable, cached state synchronization and simple mutation management, TanStack Query is used instead of standard `useEffect` API calls.

```typescript
// Example custom hook inside hooks/use-checkout.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
import { CheckoutRequest } from '../types';

export function useCheckout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: CheckoutRequest) => {
      const response = await axios.post('/api/orders/checkout', payload, {
        headers: {
          'X-Idempotency-Key': payload.idempotency_key
        }
      });
      return response.data;
    },
    onSuccess: () => {
      // Invalidate products to force query refetch & reflect updated stock count
      queryClient.invalidateQueries({ queryKey: ['products'] });
    }
  });
}
```

### Real-time Sync Hub: WebSocket Adapter
To update prices and stock levels instantly across multiple checkouts, the Next.js UI establishes a persistent WebSocket connection to the Go backend. When stock falls or an items updates, the Go server broadcasts details:

```json
{
  "event": "stock_updated",
  "data": {
    "sku": "PROD-101",
    "stock_quantity": 42
  }
}
```

On reception, React Query interceptively overrides cache:
```typescript
queryClient.setQueryData(['products'], (oldProducts: Product[] | undefined) => {
  if (!oldProducts) return [];
  return oldProducts.map(p => p.sku === data.sku ? { ...p, stock_quantity: data.stock_quantity } : p);
});
```

---

## 5. AI OCR Vision Pipeline Specification

To automate stock updates and capture inventory purchase receipts, we integrate an AI Vision Model (Gemini Pro Vision or OpenAI GPT-4o).

```
 +------------------+      +-------------------------+      +---------------------------+
 | Receipt Uploader | ---> |   Multipart-Form POST   | ---> |  Verify/KMS Secure Temp   |
 |  (React UI Panel)|      |   /api/receipts/scan    |      |  S3 Receipt Bucket Storage|
 +------------------+      +-------------------------+      +---------------------------+
                                                                          |
                                                                          v
 +------------------+      +-------------------------+      +---------------------------+
 | Write PG Audit   | <--- | Match Strict Struct JSON| <--- | Outbound Vision API Call  |
 | and Return JSON  |      |   (Line Items, Prices)  |      | (Gemini-1.5-Pro Structured|
 +------------------+      +-------------------------+      +---------------------------+
```

### Structured Prompt Configuration
The adapter targets Gemini / OpenAI APIs with systematic instructions forcing a deterministic JSON matching Schema:

```json
{
  "type": "object",
  "properties": {
    "merchant_name": { "type": "string" },
    "date": { "type": "string", "format": "date-time" },
    "items": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "quantity": { "type": "integer" },
          "price": { "type": "number" }
        },
        "required": ["name", "quantity", "price"]
      }
    },
    "total_amount": { "type": "number" }
  },
  "required": ["merchant_name", "date", "items", "total_amount"]
}
```

---

## 6. Infrastructure-as-Code (Terraform)

The cloud-native architecture is built entirely via modular, clean Terraform scripts under `/terraform`, enforcing security, networking, and high-availability standards.

```
                  +-------------------------------------------------+
                  |                   AWS VPC                       |
                  |                                                 |
                  |    +---------------------------------------+    |
                  |    |            PUBLIC SUBNET              |    |
                  |    |  - NAT Gateways                       |    |
                  |    |  - ALB (Application Load Balancer)    |    |
                  |    +------------------+--------------------+    |
                  |                       |                         |
                  |                       v routes traffic          |
                  |    +---------------------------------------+    |
                  |    |            PRIVATE SUBNET             |    |
                  |    |  - App Runner (Go ECS Service Container)|  |
                  |    +------------------+--------------------+    |
                  |                       |                         |
                  |                       v secure interfaces       |
                  |    +------------------+--------------------+    |
                  |    |           DATABASE SUBNET             |    |
                  |    |  - RDS PostgreSQL (SSOT)              |    |
                  |    |  - ElastiCache Redis Cluster          |    |
                  |    +---------------------------------------+    |
                  |                                                 |
                  +-------------------------------------------------+
```

### Module Layout
1. **vpc**: Dual-AZ public, private, and database subnet isolation.
2. **rds**: PostgreSQL RDS engine configurations with secure database groups.
3. **redis**: Managed ElastiCache single-node or multi-AZ cluster.
4. **s3**: Private S3 buckets with CORS policies configured strictly to UI domain and AES-256 server-side encryption.
5. **app_runner**: Direct Container service runner pointing to VPC Connector for private DB routes.

---

## 7. Quality Assurance & Test Criteria

To satisfy enterprise reliability levels, the codebase maintains code test rules:
* **Backend Unit Tests**: Mock out repositories using `golang/mock` or handwritten structs to test business logic (`OrderUseCase.Checkout`) isolated from database connectivity.
* **Concurrency Race Testing**: Implement Go parallel tests (`go test -race -v ./...`) executing simultaneous database checkout functions across 20+ goroutines targeting a live test PostgreSQL database container (using testcontainers-go or isolated database schemas).
* **Distributed Lock Mocks**: Inject fake mock Redis clients during unit tests to verify that idempotency failures respond precisely with structured HTTP error codes.

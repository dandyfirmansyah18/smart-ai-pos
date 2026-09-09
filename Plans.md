# Phased Implementation Plan: Smart AI POS & Merchant Engine

This execution roadmap breaks down the **Smart AI POS & Merchant Engine** into small, atomic, test-driven phases. Each task is self-contained with exact file paths, interfaces, and testing instructions to ensure high code quality, predictability, and minimal token usage during development.

---

## Phase 1: Backend Domain Models & Core Ports (Interfaces)

In this phase, we establish the domain boundaries. There are no external framework dependencies in `/internal/domain/`.

### Task 1.1: Product Domain Models & Repository Interface
* **Files to create/modify**:
  * `backend/internal/domain/product.go`
  * `backend/internal/ports/outbound/product_repo.go`
* **Interfaces & Structs**:
  ```go
  // domain/product.go
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

  // ports/outbound/product_repo.go
  type ProductRepository interface {
      GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
      GetBySKUWithLock(ctx context.Context, tx *sql.Tx, sku string) (*domain.Product, error)
      UpdateStock(ctx context.Context, tx *sql.Tx, sku string, newQty int) error
      ListAll(ctx context.Context) ([]domain.Product, error)
  }
  ```
* **Testing Requirements**:
  * No database connection here. Write a simple mock or stub for `ProductRepository` in `backend/internal/ports/outbound/mock_product_repo.go` to prove interfaces compile.

---

### Task 1.2: Order & OrderItem Domain Models & Repository Interface
* **Files to create/modify**:
  * `backend/internal/domain/order.go`
  * `backend/internal/ports/outbound/order_repo.go`
* **Interfaces & Structs**:
  ```go
  // domain/order.go
  type Order struct {
      ID             uuid.UUID   `json:"id"`
      TransactionID  string      `json:"transaction_id"`
      TotalAmount    float64     `json:"total_amount"`
      Status         string      `json:"status"` // PENDING, COMPLETED, CANCELLED, REFUNDED
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

  // ports/outbound/order_repo.go
  type OrderRepository interface {
      CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error
      GetByID(ctx context.Context, id string) (*domain.Order, error)
      GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
      UpdateStatus(ctx context.Context, id string, status string) error
  }
  ```
* **Testing Requirements**:
  * Ensure compiler safety. Write interface satisfaction tests.

---

### Task 1.3: Inbound Ports (Use Cases / Drivers)
* **Files to create/modify**:
  * `backend/internal/ports/inbound/order_usecase.go`
* **Interfaces & Structs**:
  ```go
  // ports/inbound/order_usecase.go
  type CheckoutItem struct {
      SKU      string `json:"sku" binding:"required"`
      Quantity int    `json:"quantity" binding:"required,gt=0"`
  }

  type CheckoutRequest struct {
      IdempotencyKey string         `json:"idempotency_key" binding:"required"`
      Items          []CheckoutItem `json:"items" binding:"required,dive"`
  }

  type OrderUseCase interface {
      Checkout(ctx context.Context, req CheckoutRequest) (*domain.Order, error)
  }
  ```
* **Testing Requirements**:
  * Compilation verification.

---

## Phase 2: Database Adapters & Concurrency Locking

We implement PostgreSQL infrastructure. Database queries use SQL transactions (`*sql.Tx`) and pessimistic concurrency controls.

### Task 2.1: PostgreSQL Connection Pool Setup & Viper Configuration
* **Files to create/modify**:
  * `backend/internal/adapters/infrastructure/pg/db.go`
  * `backend/config/config.go`
  * `backend/config/config.yaml`
  * `backend/config/config.example.yaml`
* **Details**:
  * Implement configuration management in `backend/config/config.go` using Viper (`github.com/spf13/viper`) to parse `config.yaml` with automatic fallbacks and environment variable overrides.
  * Implement standard database driver pool initialization (using `github.com/jackc/pgx/v5/stdlib`).
  * Configure connection limits: `SetMaxOpenConns(25)`, `SetMaxIdleConns(25)`, `SetConnMaxLifetime(15 * time.Minute)`.
* **Testing Requirements**:
  * `db_test.go` ensuring that the connection pool correctly reads credentials from Viper configuration settings.

---

### Task 2.2: Product PG Repository Implementation (Pessimistic Locking)
* **Files to create/modify**:
  * `backend/internal/adapters/infrastructure/pg/product_repo_impl.go`
* **Details**:
  * Implement `ProductRepository` interfaces.
  * In `GetBySKUWithLock`, use `SELECT id, sku, name, description, price, stock_quantity, created_at, updated_at FROM products WHERE sku = $1 FOR UPDATE`. This row-level lock blocks other transactions attempting to read/lock this SKU for write until the tx completes.
* **Testing Requirements**:
  * Concurrency Integration Test: Spawn 10 concurrent goroutines inside `product_repo_impl_test.go`. Each must run `GetBySKUWithLock` inside its own transaction and attempt to decrement the inventory of a product. Verify that final stock counts are correct and no negative stock levels occur.

---

### Task 2.3: Order PG Repository Implementation
* **Files to create/modify**:
  * `backend/internal/adapters/infrastructure/pg/order_repo_impl.go`
* **Details**:
  * Implement `CreateOrder` (saves both Order and related OrderItems inside the active SQL Transaction `*sql.Tx`).
  * Implement lookup queries (`GetByID`, `GetByIdempotencyKey`, `UpdateStatus`).
* **Testing Requirements**:
  * Integration tests verifying cascading order item insertion and transaction rollbacks if order insertion fails midway.

---

### Task 2.4: Database Migrations & Seed Data Setup
* **Files to create/modify**:
  * `backend/migrations/000001_create_products_table.up.sql`
  * `backend/migrations/000001_create_products_table.down.sql`
  * `backend/migrations/000002_create_orders_table.up.sql`
  * `backend/migrations/000002_create_orders_table.down.sql`
  * `backend/migrations/000003_create_receipt_audits_table.up.sql`
  * `backend/migrations/000003_create_receipt_audits_table.down.sql`
  * `backend/cmd/migrate/main.go`
  * `backend/cmd/seed/main.go`
* **Details**:
  * Create raw SQL migration up/down scripts for `products`, `orders`, `order_items`, `receipt_audits`, and custom `order_status` ENUM type.
  * Implement CLI command `cmd/migrate/main.go` to execute database schema migrations against PostgreSQL.
  * Implement CLI seeder `cmd/seed/main.go` to populate initial product catalog stock items for local development & testing.
* **Testing Requirements**:
  * Verify migration scripts execute cleanly up & down without SQL syntax errors.


---

## Phase 3: Redis Integration (Cache, Lock, and Idempotency)

Setting up Redis storage and distributed mutual exclusion locks (`SET NX EX`) to protect payments and webhook processors.

### Task 3.1: Redis Lock Service
* **Files to create/modify**:
  * `backend/internal/ports/outbound/lock_service.go`
  * `backend/internal/adapters/infrastructure/redis/client.go`
  * `backend/internal/adapters/infrastructure/redis/lock_service_impl.go`
* **Interfaces & Structs**:
  ```go
  // ports/outbound/lock_service.go
  type LockService interface {
      AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error)
      ReleaseLock(ctx context.Context, key string) error
  }
  ```
* **Details**:
  * Use `github.com/redis/go-redis/v9` client.
  * `AcquireLock` executes: `redisClient.SetNX(ctx, key, "locked", expiration).Result()`.
  * `ReleaseLock` deletes the key.
* **Testing Requirements**:
  * Integration test using Redis-Mock or live Redis container. Verify that if Client A holds a lock, Client B's `AcquireLock` returns `false` until Client A releases it or TTL expires.

---

## Phase 4: Order Checkout Use Case Implementation

This is the core business logic hub where ports are wired together under a database transaction wrapper.

### Task 4.1: Order Checkout Service
* **Files to create/modify**:
  * `backend/internal/ports/inbound/order_usecase_impl.go`
* **Details**:
  * Implements `OrderUseCase` interface.
  * **Workflow inside Checkout**:
    1. Check payment webhook/order idempotency using Redis: `LockService.AcquireLock(ctx, "lock:checkout:"+req.IdempotencyKey, 15*time.Second)`. If false, abort (already processing).
    2. Start PostgreSQL Transaction (`tx, err := db.BeginTx(ctx, nil)`).
    3. Verify if `IdempotencyKey` order already exists in Postgres. If yes, commit transaction, release Redis lock, and return existing order details (safely handles retries).
    4. For each item in `req.Items`:
       - Query product using `ProductRepository.GetBySKUWithLock(ctx, tx, item.SKU)`.
       - Validate: Product exists, requested quantity is $\le$ stock quantity. If failed, rollback transaction, release Redis lock, return error (e.g. `domain.ErrInsufficientStock`).
       - Calculate: `newStock = stock - item.Quantity`.
       - Update product stock: `ProductRepository.UpdateStock(ctx, tx, item.SKU, newStock)`.
    5. Save Order & OrderItems: `OrderRepository.CreateOrder(ctx, tx, order)`.
    6. Commit transaction.
    7. Release Redis lock.
* **Testing Requirements**:
  * **Unit Tests (`order_usecase_impl_test.go`)**: Use Mock repositories for database/Redis. Test:
    - Successful checkouts.
    - Aborted checkouts due to insufficient stock (assert stock is unmodified, transaction rolled back).
    - Repeated checkouts with identical idempotency key (asserting duplicate is blocked / returns original order).

---

## Phase 5: REST API Handlers & Routing Adapter

We implement our driving HTTP adapter using a routing framework like Gin or Fiber.

### Task 5.1: Router & Product Handlers
* **Files to create/modify**:
  * `backend/internal/adapters/handlers/rest/server.go`
  * `backend/internal/adapters/handlers/rest/product_handler.go`
* **API contracts**:
  * `GET /api/products` -> returns list of all products (status 200).
* **Testing Requirements**:
  * Use `net/http/httptest` to record HTTP requests and assert expected JSON array outputs.

---

### Task 5.2: Checkout Handler
* **Files to create/modify**:
  * `backend/internal/adapters/handlers/rest/order_handler.go`
* **API contracts**:
  * `POST /api/orders/checkout` -> accepts `CheckoutRequest` body, checks custom header `X-Idempotency-Key`, triggers Use Case, returns completed Order (status 201).
* **Testing Requirements**:
  * Request payload validation tests (ensure bad quantities trigger 400 Bad Request).
  * Successful checkout testing via simulated REST calls.

---

## Phase 6: Real-time WebSockets Sync

Ensures POS registers and terminal UIs instantly synchronize stock updates when checkouts succeed.

### Task 6.1: WebSocket Server & Broadcaster Hub
* **Files to create/modify**:
  * `backend/internal/adapters/handlers/ws/hub.go`
* **Details**:
  * Build a WebSocket client hub maintaining current connections.
  * Create `Hub.Broadcast(event []byte)` to stream events to all active clients.
  * Integrate into `Checkout` pipeline: once transaction commits successfully, trigger `hub.Broadcast(StockUpdateEvent{SKU: sku, Stock: stock})` asynchronously.
* **Testing Requirements**:
  * `hub_test.go` verifying that registered clients successfully receive mock payloads when a broadcast is triggered.

---

## Phase 7: AI Vision OCR Adapter

Integrate an LLM Vision API client (Gemini 1.5 Pro or OpenAI GPT-4o) to scan paper receipt images and extract structured JSON expenses.

### Task 7.1: Receipt Domain & Port Interface
* **Files to create/modify**:
  * `backend/internal/domain/receipt.go`
  * `backend/internal/ports/outbound/vision_client.go`
* **Details**:
  * Define Domain types for Receipts and OCR outputs:
  ```go
  // domain/receipt.go
  type OCRItem struct {
      Name     string  `json:"name"`
      Quantity int     `json:"quantity"`
      Price    float64 `json:"price"`
  }

  type ReceiptOCRResult struct {
      MerchantName string    `json:"merchant_name"`
      Date         time.Time `json:"date"`
      Items        []OCRItem `json:"items"`
      TotalAmount  float64   `json:"total_amount"`
  }

  // ports/outbound/vision_client.go
  type VisionClient interface {
      AnalyzeReceipt(ctx context.Context, imageBytes []byte) (*domain.ReceiptOCRResult, error)
  }
  ```

---

### Task 7.2: LLM Vision Adapter Client (Gemini / OpenAI API)
* **Files to create/modify**:
  * `backend/internal/adapters/infrastructure/vision/client_impl.go`
* **Details**:
  * Construct API POST payload to external endpoints (`https://generativelanguage.googleapis.com` or `https://api.openai.com/v1`).
  * Embed a System Prompt asking to analyze the image binary bytes and format outputs exactly to the structured JSON schema defined in Section 5 of `spec.md`.
* **Testing Requirements**:
  * Mock HTTP testing (using `gock` or standard `httptest` Server mock) to simulate the external LLM responses, ensuring parsing logic handles timezone-specific dates, integer items, and float prices without errors.

---

### Task 7.3: Receipt Handler & Database Entry Use Case
* **Files to create/modify**:
  * `backend/internal/ports/inbound/receipt_usecase.go`
  * `backend/internal/adapters/handlers/rest/receipt_handler.go`
* **Details**:
  * Endpoint `POST /api/receipts/scan`
  * Accepts Multipart Form upload file (`image`).
  * Process receipt: sends bytes to `VisionClient`, saves parsed fields to PG `receipt_audits` table, and returns structured data to the client UI.
* **Testing Requirements**:
  * Upload integration tests utilizing mock multipart requests.

---

## Phase 8: Frontend Setup & Core Layout

Establishing the Next.js foundation, styling, and server query configurations.

### Task 8.1: NextJS Configuration, Theme, & Radix
* **Files to create/modify**:
  * `frontend/app/layout.tsx`
  * `frontend/tailwind.config.js`
  * `frontend/package.json`
* **Details**:
  * Install Tailwind CSS, Lucide icons, TanStack Query (`@tanstack/react-query`), and Axios.
  * Create the global NextJS Layout wrapped inside custom React Query providers.

---

### Task 8.2: TypeScript Interfaces & Client Services
* **Files to create/modify**:
  * `frontend/types/index.ts`
  * `frontend/services/api.ts`
* **Details**:
  * Create TS types matching Go domain entities (`Product`, `Order`, `ReceiptOCRResult`).
  * Configure default Axios client pointers targeting `/api/` matching server routing ports.

---

## Phase 9: Frontend Checkout Terminal & WebSocket Integration

Building real-time synced store catalog layout grids for merchants and cashier clerks.

### Task 9.1: WebSocket Live Sync React Hook
* **Files to create/modify**:
  * `frontend/hooks/use-websocket.ts`
* **Details**:
  * Build custom hook creating a persistent connection: `new WebSocket("ws://localhost:8080/ws")`.
  * On message arrival, trigger React Query cache invalidation (`queryClient.setQueryData`) to live-patch stock numbers directly on screen without site refresh.
* **Testing Requirements**:
  * Verify connection handler handles automatic disconnect retries gracefully (exponential backoff).

---

### Task 9.2: Checkout Grid & Terminal Layout
* **Files to create/modify**:
  * `frontend/components/checkout-terminal.tsx`
  * `frontend/app/checkout/page.tsx`
* **Details**:
  * Render a sidebar containing the running checkout cart list, and a central grid showing the active catalog.
  * Implement quantity addition, subtractions, and checkout actions.
  * Incorporate visual checkout toast alerts mapping `useMutation` checkout states.

---

### Task 9.3: Sales Dashboard UI
* **Files to create/modify**:
  * `frontend/components/sales-chart.tsx`
  * `frontend/app/dashboard/page.tsx`
* **Details**:
  * Display summary statistics: Total Revenue, Orders count, Out of stock products indicator alert cards.
  * Integrate simple dashboards graphs utilizing CSS flex/SVGs or simple charts displaying real-time cash desk metrics.

---

## Phase 10: AI Vision OCR Receipt Uploader Component

Connecting Next.js users to the visual receipt scanner pipeline.

### Task 10.1: Receipt Scan Screen
* **Files to create/modify**:
  * `frontend/components/receipt-uploader.tsx`
  * `frontend/app/receipts/page.tsx`
* **Details**:
  * Drag-and-drop receipt file input.
  * Post upload multipart file to `/api/receipts/scan` endpoint.
  * Render preview overlay cards showing the AI scanned items, merchant name, date, and final parsed sum.
* **Testing Requirements**:
  * Simulate file-uploads to assert that loaded states correctly prompt parsing preview tables.

---

## Phase 11: Infrastructure-as-Code (Terraform)

Automating highly available enterprise cloud configurations.

### Task 11.1: Database (RDS & Redis ElastiCache) modules
* **Files to create/modify**:
  * `terraform/modules/rds/main.tf`
  * `terraform/modules/redis/main.tf`
* **Details**:
  * RDS: Configure `aws_db_instance` using PostgreSQL 16 engine, encrypted via custom KMS keys. Enable auto-minor-version-upgrade.
  * Redis: Configure `aws_elasticache_cluster` using Redis engine version 7, configured inside private DB subnet security groups.

---

### Task 11.2: App Container Platform module (AWS App Runner)
* **Files to create/modify**:
  * `terraform/modules/app_runner/main.tf`
  * `terraform/modules/s3/main.tf`
* **Details**:
  * AWS S3: Safe, locked bucket with strict CORS blocking and lifecycle rule transition objects to Glacier after 90 days.
  * AWS App Runner: Deploy container pointing to regional ECR images, configured inside security private networks utilizing a VPC Connector to route db traffic privately.

---

### Task 11.3: Terraform Staging Environment
* **Files to create/modify**:
  * `terraform/environments/staging/main.tf`
  * `terraform/environments/staging/variables.tf`
  * `terraform/environments/staging/outputs.tf`
* **Details**:
  * Stitch modules together (VPC -> RDS -> Redis -> S3 -> App Runner) in staging configuration files.
* **Testing Requirements**:
  * Run `terraform validate` inside staging directory to ensure syntactic logic validity.

---

## Phase 12: Continuous Integration (CI/CD)

Enforcing quality controls automatically with GitHub Actions.

### Task 12.1: Automated CI Pipeline workflow
* **Files to create/modify**:
  * `.github/workflows/ci.yml`
* **Details**:
  * Create a workflow triggered on push/pull requests to default branches.
  * **Jobs configuration**:
    1. **Backend**: Install Go, lint code using `golangci-lint`, run tests including race-detector: `go test -race -v ./...`.
    2. **Frontend**: Install NodeJS, run `npm ci`, verify build checks `npm run build` and UI linter rules.
    3. **Infrastructure**: Setup Terraform, execute `terraform fmt -check`, run `terraform validate`.

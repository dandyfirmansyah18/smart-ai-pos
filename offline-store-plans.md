# Architectural Specification & Implementation Plan: Offline Store with SQLite Sync Engine

This document provides a production-grade blueprint for integrating an **Offline Store** capability into the Smart AI POS Engine. 

By leveraging Clean/Hexagonal Architecture principles, we can seamlessly support dual-database operations. The POS can operate entirely offline using a local, zero-configuration **SQLite** database, and automatically synchronize data with the central **PostgreSQL** cloud database when internet connectivity is available.

---

## 1. High-Level Architecture & Synchronization Model

The system operates in a hybrid online/offline model, maintaining a clean separation of concerns by utilizing the Hexagonal Port-Adapter boundaries.

```
                     +---------------------------------------+
                     |             SMART AI POS              |
                     |             (FRONTEND)                |
                     +-------------------+-------------------+
                                         |
                                         | API Requests
                                         v
                     +-------------------+-------------------+
                     |            REST SERVER                |
                     |           (BACKEND API)               |
                     +---+-------------------------------+---+
                         |                               |
        (Online Mode)    |                               |    (Offline Mode)
        Uses Postgres    |                               |    Uses SQLite
                         v                               v
             +-----------+-----------+       +-----------+-----------+
             |  POSTGRESQL ADAPTERS  |       |    SQLITE ADAPTERS    |
             |  - ProductPGRepository|       |  - ProductSQLiteRepo  |
             |  - OrderPGRepository  |       |  - OrderSQLiteRepo    |
             |  - CashShiftPGRepo    |       |  - CashShiftSQLiteRepo|
             +-----------+-----------+       +-----------+-----------+
                         |                               |
                         |           SYNC ENGINE         |
                         |      (Bidirectional Syncer)   |
                         +------------->   <-------------+
                                     PULL  /  PUSH
```

### Modes of Operation
1. **Online Mode (Default):** Reads and writes directly from PostgreSQL. If Redis is available, it handles locks and cache.
2. **Offline Mode (Local-First):** Reads and writes from a local SQLite database file (`pos_local.db`). No external services (PostgreSQL or Redis) are required.
3. **Automatic Fallback (Resilience Mode):** If PostgreSQL ping fails on startup or during a transaction timeout, the backend automatically transitions to SQLite adapters and logs an alert.

---

## 2. Core Architectural Strategy

### A. Zero-CGO SQLite Driver
To ensure cross-platform compatibility without complex native C development toolchains, we will use **`modernc.org/sqlite`** (the pure Go port of SQLite). This allows instant compiles across Darwin (macOS), Linux, and Windows.

```bash
go get modernc.org/sqlite
```

### B. UUID Safety for Collision-Free ID Generation
Because all entities (Products, Orders, Payments, Cash Shifts) utilize **UUIDv4** for IDs, we completely eliminate primary key collisions during synchronization. The local database can safely generate unique IDs offline, which will never conflict with other terminals or the central server.

### C. Sync Status Tracking
To manage which records have been pushed to the cloud, we introduce a `sync_status` tracking layer. We will add a `sync_status` column to the following transaction tables on SQLite:
* `orders` (`sync_status`: `PENDING`, `SYNCED`, `FAILED`)
* `cash_shifts` (`sync_status`: `PENDING`, `SYNCED`, `FAILED`)
* `order_payments` (`sync_status`: `PENDING`, `SYNCED`, `FAILED`)

---

## 3. SQLite Schema Mapping & Migration Strategy

SQLite does not support PostgreSQL-specific features like `CREATE TYPE ... AS ENUM` or certain extensions like `uuid-ossp`. We must define equivalent SQLite table creation scripts.

### Mapping Conversions
* **UUID / UUID-OSSP:** Mapped to `TEXT PRIMARY KEY` with UUID strings generated in Go application code (using `github.com/google/uuid`).
* **ENUMs (order_status, user_role, shift_status, payment_status):** Mapped to `TEXT NOT NULL` in SQLite with `CHECK` constraints to ensure values are restricted.
* **NUMERIC / DECIMAL:** Mapped to `REAL` or `NUMERIC` in SQLite.
* **TIMESTAMP WITH TIME ZONE:** Mapped to `TEXT` or `DATETIME` in SQLite, storing ISO-8601 strings or integers.
* **FOR UPDATE Locks:** Ignored or handled using SQLite's native single-writer transaction locks (`BEGIN IMMEDIATE` transactions).

---

## 4. Phase-by-Phase Execution Plan

### Phase 1: Configuration, Driver installation, and DB Initialization
* **Files to create/modify**:
  * `backend/config/config.go`
  * `backend/config/config.example.yaml`
  * `backend/internal/adapters/infrastructure/sqlite/db.go`
* **Tasks**:
  1. Add configuration variables:
     * `DATABASE_TYPE` (string: `postgres` or `sqlite`)
     * `SQLITE_PATH` (string, default: `./pos_local.db`)
     * `AUTO_FAILOVER` (bool, default: `true`)
  2. Implement SQLite DB initializer `NewSQLiteDB(path string) (*sql.DB, error)`.
  3. Write an embedded schema initializer inside `sqlite/db.go` to create SQLite tables on startup if they do not exist.

---

### Phase 2: SQLite Repository Adapters
* **Files to create**:
  * `backend/internal/adapters/infrastructure/sqlite/product_repo_impl.go`
  * `backend/internal/adapters/infrastructure/sqlite/order_repo_impl.go`
  * `backend/internal/adapters/infrastructure/sqlite/user_repo_impl.go`
  * `backend/internal/adapters/infrastructure/sqlite/cash_shift_repo_impl.go`
  * `backend/internal/adapters/infrastructure/sqlite/access_menu_repo_impl.go`
  * `backend/internal/adapters/infrastructure/sqlite/payment_repo_impl.go`
* **Tasks**:
  1. Implement each repository matching the ports defined in `backend/internal/ports/outbound/`.
  2. Map PostgreSQL-specific queries. For instance, replace Postgres query placeholders `$1, $2` with standard SQLite placeholders, and strip `FOR UPDATE` clauses since SQLite handles concurrent writes through file-level database locking.
  3. Map the results into clean domain entities.

---

### Phase 3: Dependency Injection & Failover Engine
* **Files to modify**:
  * `backend/cmd/api/main.go`
* **Tasks**:
  1. Refactor `main.go` to support runtime adapter selection.
  2. If `DATABASE_TYPE` is set to `sqlite` or if connection to PostgreSQL fails and `AUTO_FAILOVER` is enabled:
     * Initialize the local SQLite database.
     * Inject SQLite repositories into Use Cases instead of PostgreSQL repositories.
     * Disable Redis locking service and inject a local, memory-based mock/lock service.

```go
// main.go pseudo-code logic
var db *sql.DB
var productRepo outbound.ProductRepository
var orderRepo outbound.OrderRepository

if cfg.DBType == "sqlite" {
    db, err = sqlite.NewSQLiteDB(cfg.SQLitePath)
    productRepo = sqlite.NewProductSQLiteRepository(db)
    orderRepo = sqlite.NewOrderSQLiteRepository(db)
} else {
    db, err = pg.NewPostgresDB(cfg)
    if err != nil && cfg.AutoFailover {
        log.Println("Postgres failed, switching to fallback SQLite")
        db, err = sqlite.NewSQLiteDB(cfg.SQLitePath)
        productRepo = sqlite.NewProductSQLiteRepository(db)
        orderRepo = sqlite.NewOrderSQLiteRepository(db)
    } else {
        productRepo = pg.NewProductPGRepository(db)
        orderRepo = pg.NewOrderPGRepository(db)
    }
}
```

---

### Phase 4: Sync Engine & Conflict Resolution
* **Files to create**:
  * `backend/internal/ports/inbound/sync_usecase.go`
  * `backend/internal/ports/inbound/sync_usecase_impl.go`
  * `backend/internal/adapters/handlers/rest/sync_handler.go`
* **Sync Strategy**:
  1. **Sync Master Data (Pull):** Pull master records (users, products, access menus) from PostgreSQL to SQLite. Since cloud data is authoritative, existing records in SQLite are upserted.
  2. **Sync Transactions (Push):** Pull unsynced orders, payments, and cash shifts (where `sync_status = 'PENDING'`) from SQLite, format them, and write them to PostgreSQL. Upon successful database commits on PostgreSQL, update SQLite's `sync_status` to `'SYNCED'`.
  3. **Conflict Resolution:** Since transactions are uniquely identified by UUIDv4, there are no primary key clashes. For stock levels, the SQLite local cache will synchronize quantities to represent actual sales.

---

### Phase 5: REST APIs & Sync Endpoints
We expose dedicated endpoints for managing and inspecting the synchronization status.

* **Endpoints**:
  * `GET /api/sync/status`: Returns current DB mode (Postgres/SQLite), count of unsynced transactions, and the timestamp of the last successful sync.
  * `POST /api/sync/trigger`: Forcefully initiates a bidirectional synchronization routine.

---

### Phase 6: Frontend Offline Detection & Sync UI
* **Files to create/modify**:
  * `frontend/components/OfflineIndicator.tsx` (Status bar showing Online vs. Offline)
  * `frontend/pages/sync.tsx` or `frontend/components/SyncSettings.tsx` (Dashboard for triggering syncs and seeing pending sync queue sizes)
* **Strategy**:
  1. Use standard browser APIs (`navigator.onLine` and `window.addEventListener('offline')`) to track network connectivity.
  2. Display a prominent top banner when the application drops offline, letting the user know they are using local backup mode.
  3. In the Settings or Navigation menu, show a "Pending Sync: X Orders" counter. When online, provide a button to invoke `POST /api/sync/trigger`.

---

## 5. Testing & Verification Plan

### 1. SQLite Repository Tests
Write unit tests for the SQLite repositories in `backend/internal/adapters/infrastructure/sqlite/` using a temporary memory database (`:memory:`) to verify queries execute flawlessly without a Postgres server.

### 2. Network Interruption Simulation
1. Start the API Server in online mode (Postgres active).
2. Create an order and verify it is written to PostgreSQL.
3. Simulate an internet outage by shutting down PostgreSQL.
4. Create new orders on the POS. Verify the server gracefully falls back to SQLite and registers the orders with `sync_status = 'PENDING'`.
5. Restart PostgreSQL.
6. Trigger the sync API endpoint (`POST /api/sync/trigger`) and verify all pending orders are uploaded to PostgreSQL and marked as `SYNCED` in SQLite.

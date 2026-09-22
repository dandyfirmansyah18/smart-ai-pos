# Smart AI POS System

A high-performance Point of Sale (POS) backend and frontend application built with strict Hexagonal Architecture in Go and Next.js / TypeScript, featuring offline-first SQLite resilience, real-time WebSocket updates, AI receipt scanning, cash shift management (buka-tutup kasir), and Midtrans payment gateway integration.

---

## 📋 Prerequisites

Before running the project, ensure you have the following installed on your system:
- **Go** (v1.22 or higher)
- **Node.js** (v18 or higher) & npm / yarn
- **PostgreSQL** (v14 or higher) - *Optional if running in offline SQLite mode*
- **Redis** (v6 or higher) - *Optional for distributed idempotency locking*

---

## ⚙️ Configuration

### 1. Backend Configuration
Navigate to `backend/config/`:
- Copy `config.example.yaml` to `config.yaml`:
  ```bash
  cp backend/config/config.example.yaml backend/config/config.yaml
  ```
- Update database settings (`postgres` or `sqlite`), server port, JWT secret, and Midtrans keys in `config.yaml`.

### 2. Frontend Configuration
Navigate to `frontend/`:
- Copy `.env.example` to `.env`:
  ```bash
  cp frontend/.env.example frontend/.env
  ```
- Configure `NEXT_PUBLIC_API_URL` (default: `http://localhost:8080/api/v1`) and WebSocket URL.

---

## 🚀 How to Run

### 1. Run Backend Server
```bash
cd backend
go mod download
go run cmd/api/main.go
```
The backend API server will start on port `8080` by default, with real-time WebSocket support at `/ws`.

### 2. Run Frontend Development Server
Open a new terminal session:
```bash
cd frontend
npm install
npm run dev
```
The Next.js frontend will start on port `3000` (or `3001` depending on port availability). Open `http://localhost:3000` in your browser.

---

## 📦 Features & Architecture

- **Hexagonal Architecture (Ports & Adapters)**: Clean separation of domain logic, inbound use cases, and outbound persistence/infrastructure adapters.
- **Offline-First Resilience**: Automatic failover to local SQLite database when PostgreSQL is unreachable, accompanied by a background sync worker.
- **Real-Time Kitchen Display System (KDS)**: WebSocket hub broadcasting live order placement, status changes, and cash shifts.
- **Payment Gateway**: Integrated with Midtrans QRIS / Snap for automated payment status handling via webhooks.
- **Cash Shift Management (Buka-Tutup Kasir)** & Profit/Loss tracking.

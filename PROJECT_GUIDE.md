# Nuvriq AI - SDE 1 Golang Backend Assessment: Complete Technical Guide & Specification

> **Project Name**: Unified Endpoint Management (UEM) Device Service  
> **Target Execution Time**: 2 Hours  
> **Tech Stack**: Golang | PostgreSQL | REST API | Docker & Docker Compose  

---

## 1. Project Overview & Problem Statement

### What is the Problem Statement? (Simple & Easy Explanation)
In large enterprises, companies issue hundreds or thousands of laptops, desktops, and servers to employees across different operating systems (Windows, macOS, Linux). 

To manage and secure these devices, IT administrators need a **Unified Endpoint Management (UEM) System**. The core problem this system solves is **visibility and inventory management**:
1. **Device Inventory**: Registering every device that belongs to the enterprise with details like hostname, platform, OS version, and assigned employee.
2. **Health Monitoring (Heartbeats)**: Devices periodically ping ("sync" or send heartbeats to) the server to report that they are online and functioning.
3. **Inactivity Detection**: Identifying lost, stolen, broken, or turned-off devices that haven't communicated with the server within a specified threshold (e.g., 30 or 60 days).
4. **Never-Synchronized Detection**: Catching devices that were registered in the inventory but have *never* connected for their initial setup.
5. **Fleet Diagnostics**: Providing IT admins with high-level aggregate summaries (e.g., total fleet size, active vs. inactive breakdown, OS distribution).

---

## 2. Tech Stack, Architecture & Database Selection

### Recommended Architecture & Tech Ideas

| Component | Technology | Rationale / Implementation Idea |
| :--- | :--- | :--- |
| **Language** | **Golang (1.22+)** | Native concurrency (goroutines/channels), high performance, small memory footprint for backend microservices. |
| **HTTP Router** | **Standard `net/http`** or **`github.com/go-chi/chi/v5`** | Idiomatic Go, lightweight, zero extra dependencies with standard library, or `chi` for clean URL parameter parsing (`/devices/{id}`). |
| **Database** | **PostgreSQL 15+** | Relational integrity, ACID compliance, robust indexing capabilities, and high-concurrency connection pooling. |
| **DB Driver** | **`github.com/lib/pq`** or **`pgx/v5`** | High-performance PostgreSQL driver for Go `database/sql`. |
| **Concurrency** | **`database/sql` Connection Pool** | Configured `SetMaxOpenConns(25)` & `SetMaxIdleConns(10)` to handle thousands of concurrent heartbeats. |
| **Simulator** | **Goroutines + Worker Pool** | Spawns $N$ concurrent routines sending HTTP POST sync requests to simulate real device agents. |

---

## 3. Database Schema & Indexing Strategy

### PostgreSQL Schema (`schema.sql`)

```sql
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    serial_number VARCHAR(100) NOT NULL UNIQUE,
    hostname VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL CHECK (platform IN ('windows', 'macos', 'linux')),
    os_version VARCHAR(100) NOT NULL,
    owner_email VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'never_synced')),
    last_synced_at TIMESTAMPTZ NULL, -- NULL indicates the device has NEVER synchronized
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance Indexes for High Concurrency & Fast Filter Queries
CREATE INDEX IF NOT EXISTS idx_devices_platform ON devices(platform);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);
CREATE INDEX IF NOT EXISTS idx_devices_last_synced_at ON devices(last_synced_at);
CREATE INDEX IF NOT EXISTS idx_devices_serial_number ON devices(serial_number);
```

---

## 4. API Endpoints Specification & Sample Results

Below are the 6 core REST API endpoints required for the service, plus a system health check.

---

### Endpoint 1: Register a New Device
* **HTTP Method**: `POST`
* **Path**: `/api/v1/devices`
* **Description**: Registers a new device in the enterprise inventory. Initial status is `never_synced` until first heartbeat is received (or `active` if initial sync timestamp is provided).

#### Sample Request Body:
```json
{
  "serial_number": "SN-WIN-98214",
  "hostname": "workstation-john-01",
  "platform": "windows",
  "os_version": "Windows 11 Pro 23H2",
  "owner_email": "john.doe@company.com"
}
```

#### Sample Response (`201 Created`):
```json
{
  "status": "success",
  "data": {
    "id": "e4a2c510-5321-4f1d-910e-0a5682bc8141",
    "serial_number": "SN-WIN-98214",
    "hostname": "workstation-john-01",
    "platform": "windows",
    "os_version": "Windows 11 Pro 23H2",
    "owner_email": "john.doe@company.com",
    "status": "never_synced",
    "last_synced_at": null,
    "created_at": "2026-08-29T10:30:00Z",
    "updated_at": "2026-08-29T10:30:00Z"
  }
}
```

#### Error Responses:
* `400 Bad Request` (Validation Failure):
  ```json
  { "error": "invalid platform 'android', allowed: windows, macos, linux" }
  ```
* `409 Conflict` (Duplicate Registration):
  ```json
  { "error": "device with serial_number 'SN-WIN-98214' already exists" }
  ```

---

### Endpoint 2: Device Synchronization / Heartbeat
* **HTTP Method**: `POST`
* **Path**: `/api/v1/devices/{id}/sync`
* **Description**: Called by a device agent to send periodic heartbeats. Updates `last_synced_at` to the current time and sets status to `active`.

#### Sample Request Body (Optional Telemetry / Timestamp):
```json
{
  "timestamp": "2026-08-29T10:32:00Z"
}
```
*(If timestamp is omitted, the backend uses the server's current UTC time).*

#### Sample Response (`200 OK`):
```json
{
  "status": "success",
  "message": "Device synchronized successfully",
  "data": {
    "id": "e4a2c510-5321-4f1d-910e-0a5682bc8141",
    "status": "active",
    "last_synced_at": "2026-08-29T10:32:00Z"
  }
}
```

#### Error Responses:
* `404 Not Found`:
  ```json
  { "error": "device with ID 'e4a2c510-5321-4f1d-910e-0a5682bc8141' not found" }
  ```
* `400 Bad Request`:
  ```json
  { "error": "sync timestamp cannot be in the future" }
  ```

---

### Endpoint 3: Retrieve Device Details
* **HTTP Method**: `GET`
* **Path**: `/api/v1/devices/{id}`
* **Description**: Fetches detailed information for a single specific device by its UUID.

#### Sample Response (`200 OK`):
```json
{
  "status": "success",
  "data": {
    "id": "e4a2c510-5321-4f1d-910e-0a5682bc8141",
    "serial_number": "SN-WIN-98214",
    "hostname": "workstation-john-01",
    "platform": "windows",
    "os_version": "Windows 11 Pro 23H2",
    "owner_email": "john.doe@company.com",
    "status": "active",
    "last_synced_at": "2026-08-29T10:32:00Z",
    "created_at": "2026-08-29T10:30:00Z",
    "updated_at": "2026-08-29T10:32:00Z"
  }
}
```

#### Error Response:
* `404 Not Found`:
  ```json
  { "error": "device not found" }
  ```

---

### Endpoint 4: List & Filter Fleet (With Pagination)
* **HTTP Method**: `GET`
* **Path**: `/api/v1/devices`
* **Query Parameters**:
  * `platform` (optional): `windows`, `macos`, `linux`
  * `status` (optional): `active`, `inactive`, `never_synced`
  * `page` (optional, default `1`)
  * `limit` (optional, default `10`, max `100`)
* **Example URL**: `/api/v1/devices?platform=macos&page=1&limit=2`

#### Sample Response (`200 OK`):
```json
{
  "status": "success",
  "pagination": {
    "total_records": 15,
    "page": 1,
    "limit": 2,
    "total_pages": 8
  },
  "data": [
    {
      "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
      "serial_number": "SN-MAC-40291",
      "hostname": "macbook-air-sarah",
      "platform": "macos",
      "os_version": "macOS Sonoma 14.5",
      "owner_email": "sarah.connor@company.com",
      "status": "active",
      "last_synced_at": "2026-08-29T10:15:00Z"
    },
    {
      "id": "8f3c7a10-2b11-4f99-88aa-112233445566",
      "serial_number": "SN-MAC-10923",
      "hostname": "macbook-pro-alex",
      "platform": "macos",
      "os_version": "macOS Sequoia 15.0",
      "owner_email": "alex.smith@company.com",
      "status": "active",
      "last_synced_at": "2026-08-29T09:45:00Z"
    }
  ]
}
```

---

### Endpoint 5: Detect Inactive & Never-Synced Devices
* **HTTP Method**: `GET`
* **Path**: `/api/v1/devices/inactive`
* **Query Parameters**:
  * `days` (optional, integer, default `60`): Inactivity threshold in days.
* **Description**: Returns devices that have not synchronized within the past `N` days, as well as devices that have `never_synced` (`last_synced_at IS NULL`).

#### Example URL: `/api/v1/devices/inactive?days=60`

#### Sample Response (`200 OK`):
```json
{
  "status": "success",
  "threshold_days": 60,
  "summary": {
    "total_inactive_or_unSynced": 3,
    "inactive_count": 2,
    "never_synced_count": 1
  },
  "data": [
    {
      "id": "11111111-2222-3333-4444-555555555555",
      "serial_number": "SN-LIN-00012",
      "hostname": "dev-server-old",
      "platform": "linux",
      "owner_email": "ops@company.com",
      "status": "inactive",
      "last_synced_at": "2026-06-01T08:00:00Z",
      "inactivity_reason": "Not synchronized in 89 days"
    },
    {
      "id": "22222222-3333-4444-5555-666666666666",
      "serial_number": "SN-WIN-00099",
      "hostname": "laptop-unassigned",
      "platform": "windows",
      "owner_email": "it-stock@company.com",
      "status": "never_synced",
      "last_synced_at": null,
      "inactivity_reason": "Device registered but never synchronized"
    }
  ]
}
```

---

### Endpoint 6: Fleet Summary Diagnostics
* **HTTP Method**: `GET`
* **Path**: `/api/v1/fleet/summary`
* **Description**: Returns high-level metrics and platform distribution for dashboard display.

#### Sample Response (`200 OK`):
```json
{
  "status": "success",
  "data": {
    "total_devices": 20,
    "status_counts": {
      "active": 15,
      "inactive": 3,
      "never_synced": 2
    },
    "platform_breakdown": {
      "windows": 8,
      "macos": 7,
      "linux": 5
    },
    "health_percentage": 75.0
  }
}
```

---

### Endpoint 7: Service Health Check
* **HTTP Method**: `GET`
* **Path**: `/health`

#### Sample Response (`200 OK`):
```json
{
  "status": "UP",
  "database": "connected",
  "timestamp": "2026-08-29T10:33:00Z"
}
```

---

## 5. Device Simulator CLI Specification

Per requirement 6, you must create a Go simulator CLI program:
```bash
go run ./cmd/simulator --devices=100 --url=http://localhost:8080
```

### Simulator Implementation Logic
1. **Goroutines**: Spawns $N$ concurrent workers.
2. **Device Registration**: Sends `POST /api/v1/devices` to register initial batch.
3. **Heartbeat Simulation Loop**:
   - **80% Healthy Devices**: Continuously ping `/api/v1/devices/{id}/sync` every 2-5 seconds.
   - **10% Inactive Devices**: Send 1 heartbeat, then immediately stop pinging (to demonstrate inactivity detection).
   - **10% Never-Synced Devices**: Register with the backend, but skip heartbeat calls entirely.

---

## 6. Project `.gitignore` File

Here is the `.gitignore` created in the workspace root:

```gitignore
# Binaries for programs and plugins
*.exe
*.exe-
*.dll
*.so
*.dylib

# Test binary built with `go test -c`
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (vendor/ is not required if using Go modules)
vendor/

# Go build output
bin/
dist/
/nuvriq-uem-service
/simulator

# Environment files
.env
.env.local
.env.*.local

# IDE & Editor files
.idea/
.vscode/
*.swp
*.swo
*~
.DS_Store

# Logs
*.log
logs/

# Database / Storage data
pgdata/
tmp/
```

---

## 7. Execution Checklist for 2-Hour Plan

1. **Step 1 (15 mins)**: Set up `docker-compose.yml` (Go app + PostgreSQL) & database `schema.sql`.
2. **Step 2 (30 mins)**: Write domain models, DB repository layer, and DB connection pooling in Go.
3. **Step 3 (35 mins)**: Implement REST handlers (`register`, `sync`, `get`, `list`, `inactive`, `summary`).
4. **Step 4 (20 mins)**: Build the Go simulator in `cmd/simulator/main.go` with concurrent goroutines.
5. **Step 5 (20 mins)**: Write unit/integration tests (`device_test.go`) & update README with evaluation instructions.

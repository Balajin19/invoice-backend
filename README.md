# Invoice Generator Backend

A Go backend service for generating and managing invoices using PostgreSQL.

## Project Structure

- `cmd/` - Main application entry point
- `pkg/` - Public packages and libraries
- `internal/` - Internal packages (models, handlers, database)
- `api/` - API specifications and contracts
- `config/` - Configuration files
- `scripts/` - Utility scripts for development and deployment
- `tests/` - Integration and end-to-end tests
- `migrations/` - Database migrations

## Prerequisites

- Go 1.21+
- PostgreSQL 12+

## Setup

### 1. PostgreSQL Database Setup

Create a new PostgreSQL database:

```bash
createdb invoice_db
```

Then run the migration to create tables:

```bash
psql -U postgres -d invoice_db -f migrations/001_create_invoices.sql
```

### 2. Environment Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` with your PostgreSQL credentials:

```
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=your_password
DB_NAME=invoice_db
SERVER_PORT=8080
```

### 3. Install Dependencies

```bash
go mod download
go mod tidy
```

## Getting Started

### Build and Run

```bash
go run ./cmd/
```

Or build an executable:

```bash
go build -o invoice-generator ./cmd/
./invoice-generator
```

Or use the Makefile:

```bash
make run
```

On Windows PowerShell (without `make`), run:

```powershell
go run ./cmd/
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check

```bash
GET /health
```

### Get All Invoices

```bash
GET /invoices
```

Response:

```json
[
  {
    "id": 1,
    "client_name": "Acme Corp",
    "amount": 5000.0,
    "description": "Web Development Services",
    "issued_date": "2026-04-02T00:00:00Z",
    "due_date": "2026-05-02T00:00:00Z",
    "status": "pending",
    "created_at": "2026-04-02T00:00:00Z",
    "updated_at": "2026-04-02T00:00:00Z"
  }
]
```

### Create Invoice

```bash
POST /invoices
```

Request body:

```json
{
  "client_name": "New Client",
  "amount": 2500.0,
  "description": "Service Description",
  "issued_date": "2026-04-02T00:00:00Z",
  "due_date": "2026-05-02T00:00:00Z",
  "status": "pending"
}
```

## Development

Run tests:

```bash
make test
```

On Windows PowerShell (without `make`), run:

```powershell
go test -v ./...
```

Clean build artifacts:

```bash
make clean
```

## Database Schema

### Invoices Table

- `id` - Primary key (auto-increment)
- `client_name` - Client name (required)
- `amount` - Invoice amount (required)
- `description` - Invoice description
- `issued_date` - Date invoice was issued (required)
- `due_date` - Date invoice is due (required)
- `status` - Invoice status: pending, paid, overdue (default: pending)
- `created_at` - Record creation timestamp
- `updated_at` - Record update timestamp

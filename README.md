# Manager service

A highly relational, strictly-typed Headless CMS built to manage portfolio data.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.26.4-00add8?style=flat-square&logo=go)](https://golang.org/)
[![gRPC](https://img.shields.io/badge/gRPC-API-244c5a?style=flat-square&logo=grpc)](https://grpc.io/)
[![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL-4169E1?style=flat-square&logo=postgresql)](https://www.postgresql.org/)

## Overview

The manager service operates as the core Headless Portfolio CMS. It models and manages user profiles, work experiences, showcase projects, technologies, and incoming contact messages. It orchestrates complex timeline histories and relational tagging purely through locked-down PostgreSQL stored procedures.

> [!NOTE]
> The database schema and service architecture natively support multi-tenancy and multiple users, allowing seamless scaling if the platform expands to additional accounts.

## Architecture

This section explains the technologies and physical layout of the service.

- **Framework**: Standard library `net` with `google.golang.org/grpc` for the RPC server
- **Database driver**: Direct PostgreSQL interactions via `jackc/pgx/v5` with custom composite type registration
- **Validation**: Incoming RPC requests are intercepted by a custom middleware using `buf.build/go/protovalidate`
- **Logging**: Idiomatic structured logging via `log/slog`

### Project structure

- `controller/`: gRPC service implementations
- `db/`: Database connection setup, pooling, and custom pg type registration
- `internal/`: Internal models, faults, and timeout configurations
- `middlewares/`: gRPC interceptors for logging and validation
- `sql/`: Schema definitions, stored procedures, and role grants
- `main.go`: Application entrypoint and dependency wiring
- `go.mod`: Dependency management

## Features

This section outlines the capabilities of the manager service.

- **High-performance API**: Concurrent gRPC server built with modern Go.
- **Automated validation**: Real-time payload validation via a gRPC unary interceptor to reject invalid requests early.
- **Relational integrity**: Uses PostgreSQL stored procedures to safely handle complex writes, cascading deletions, and relational upserts.
- **Graceful shutdown**: Safely drains active connections and closes database pools on OS signals.

## Database security

- Operates strictly under the `portfolio` schema
- Connects using the `manager_service` database role, which does **not** have direct `INSERT`, `UPDATE`, or `DELETE` permissions on tables
- All complex writes are handled safely via strictly defined database functions

## Getting started

This section explains how to run the manager service locally.

### Prerequisites

- Go 1.26.4 or higher
- PostgreSQL 18+ database instance
- Protocol buffer compilation tools if modifying definitions

### Configuration

Export these variables directly in your shell environment:

| Variable | Description | Required |
| :--- | :--- | :---: |
| `POSTGRES_URL` | Connection string for the PostgreSQL database | **Yes** |
| `PORT` | The port on which the gRPC server will listen (Default: `7296`) | No |
| `INTERNAL_BIND_IP` | IP to bind the server to (Default: `0.0.0.0`) | No |

### Running locally

Run the service using the Go CLI:

```bash
go run main.go
```

To watch for file changes and automatically rebuild and restart the server, use [air](https://github.com/cosmtrek/air):

```bash
go install github.com/air-verse/air@latest
air
```

## Deployment

The service is containerized and leverages GitHub Actions (`.github/workflows/main.yml`) for automated builds and deployment to a remote VM using Podman.

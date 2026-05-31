<div align="center">

# Portfolio Manager Service
*A high-performance gRPC backend for portfolio and content management*

[![Go Version](https://img.shields.io/badge/Go->=1.25.3-00add8?style=flat-square&logo=go)](https://golang.org/)
[![gRPC](https://img.shields.io/badge/gRPC-API-244c5a?style=flat-square&logo=grpc)](https://grpc.io/)
[![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL-4169E1?style=flat-square&logo=postgresql)](https://www.postgresql.org/)

⭐ If you find this project useful, star it on GitHub!

[Overview](#overview) • [Architecture & Tech Stack](#architecture--tech-stack) • [Features](#features) • [Getting Started](#getting-started) • [Polyrepo Local Setup](#polyrepo-local-setup) • [API Summary](#api-summary) • [Database Architecture](#database-architecture)

</div>

## Overview

The **Manager Service** is a high-performance gRPC backend written in Go. Operating as a core component within the infrastructure microservices ecosystem, it is responsible for handling data persistence and business logic across domains like user management, portfolio tracking, experiences, technologies, and messages.

It relies on strict protocol buffer definitions shared via a common contracts repository, ensuring type-safe and validated communication across services.

## Architecture & Tech Stack

- **Framework**: Standard library `net` with `google.golang.org/grpc` for the RPC server.
- **Database Driver**: Direct interactions with a PostgreSQL database via `jackc/pgx/v5` for high-performance connection pooling.
- **Validation**: Incoming RPC requests are intercepted by a custom middleware using `buf.build/go/protovalidate`.
- **Logging**: Idiomatic structured logging is implemented via `log/slog`.
- **Contracts**: Shares proto definitions through a common Go module.

### Project Structure

```text
.
├── controller/    # gRPC service implementations (user, portfolio, etc.)
├── db/            # Database connection setup and connection pooling logic
├── lib/           # Shared libraries and internal core packages
├── middlewares/   # gRPC interceptors (e.g., validation)
├── sql/           # Schema definitions, stored procedures, and role grants (experience, master, etc.)
├── utils/         # Helper functions and constants (e.g., timeout configurations)
├── main.go        # Application entrypoint and dependency wiring
└── go.mod         # Dependency management
```

## Features

- ⚡ **High-Performance API**: Exposes a fast, concurrent gRPC server built with modern Go.
- 🛡️ **Automated Validation**: Real-time payload validation via a gRPC unary interceptor. Invalid requests are rejected before they hit business logic.
- 🗄️ **Robust Data Storage**: Uses PostgreSQL for reliable data persistence. Raw SQL statements are maintained in the `sql/` directory for optimized query execution and clear separation of concerns.
- 🚦 **Graceful Shutdown**: Handles OS signals (e.g., `SIGTERM`, `SIGINT`) to safely drain active connections, close database pools, and shut down the server gracefully.
- 📝 **Structured Logging**: Built-in JSON structured logging using Go's standard `log/slog` for excellent observability.
- 🔥 **Hot Reloading**: Configuration ready for hot-reloading via `air` for rapid local development.

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.25.3 or higher
- A running [PostgreSQL](https://www.postgresql.org/download/) 18+ database instance
- The `common` contracts module (see Polyrepo setup below).
- Protocol buffer compilation tools (e.g., `buf`) if modifying definitions.

### Installation

Clone the repository:
```bash
git clone https://github.com/Aditya-0011/manager.git manager
cd manager
go mod tidy
```

### Configuration

The service relies on environment variables for configuration. You can export these directly in your shell or place them in a `.env` file at the root of the project.

| Variable | Description | Default | Required |
| :--- | :--- | :--- | :---: |
| `POSTGRES_URL` | Connection string for the PostgreSQL database (e.g. `postgres://user:pass@host/db`) | - | **Yes** |
| `PORT` | The port on which the gRPC server will listen | `7296` | No |

### Running the Service

For standard execution, use the Go CLI:

```bash
go run main.go
```

> [!TIP]
> If you have `air` installed globally, or want to use the `run.ps1` script for local development, run `air` or `./run.ps1` to watch for file changes and automatically rebuild and restart the server.

## Polyrepo Local Setup

This project uses a polyrepo architecture. The `manager` service is in its own repository, but it depends on the schemas defined in the `common` repository.

> [!IMPORTANT]
> When developing locally, keep all repositories (`common`, `auth`, `manager`) in the same parent folder. Use a Go Workspace file (`go.work`) in the parent directory to safely resolve the local `common` module without altering `go.mod` files.

Example local setup:
```text
infrastructure/
├── common/
├── auth/
├── manager/
└── go.work # Should include 'use ./common', 'use ./auth', 'use ./manager'
```

## API Summary

The service implements a multi-domain gRPC interface focused on content management:
- **User Management**: Updating profile metadata, about descriptions, and avatar links.
- **Projects**: Managing showcase projects, linked repositories, and featured status flags.
- **Experiences**: Orchestrating complex timeline histories, including deeply nested positions and roles.
- **Technologies**: Mapping a taxonomy of tech-stack skills to specific projects and experiences.

## Database Architecture

The manager service relies on a highly relational PostgreSQL structure that is completely locked down at the database layer.
- **Dedicated Schema**: It operates under the `portfolio` schema, managing entities like `user`, `project`, `experience`, `position`, and `technology` mappings.
- **Restricted Access**: The backend application connects using the `manager_service` database role, which does **not** have direct `INSERT`, `UPDATE`, or `DELETE` permissions on tables.
- **Stored Procedures**: All complex writes, cascading deletions, and relational upserts (e.g., updating an experience timeline and linking project positions) are handled safely via strictly defined database functions like `portfolio.edit_experience` and `portfolio.delete_project`.

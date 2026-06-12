# Portfolio Manager Service

A highly relational, strictly-typed Headless CMS built to manage the developer's personal portfolio.

[![Go Version](https://img.shields.io/badge/Go->=1.25.3-00add8?style=flat-square&logo=go)](https://golang.org/)
[![gRPC](https://img.shields.io/badge/gRPC-API-244c5a?style=flat-square&logo=grpc)](https://grpc.io/)
[![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL-4169E1?style=flat-square&logo=postgresql)](https://www.postgresql.org/)

## Overview

The Manager Service operates as the core Headless Portfolio CMS. This service is designed to model and manage user professional entities: user profiles (about sections, cover images), work experiences (companies and nested positions), showcase projects, technologies, and incoming contact messages. It orchestrates complex timeline histories and relational tagging purely through locked-down PostgreSQL stored procedures.

## Architecture & Tech Stack

- **Framework**: Standard library `net` with `google.golang.org/grpc` for the RPC server.
- **Database Driver**: Direct PostgreSQL interactions via `jackc/pgx/v5` for high-performance connection pooling.
- **Validation**: Incoming RPC requests are intercepted by a custom middleware using `buf.build/go/protovalidate`.
- **Logging**: Idiomatic structured logging via `log/slog`.

### Project Structure

```text
.
├── controller/    # gRPC service implementations (user, portfolio, etc.)
├── db/            # Database connection setup and connection pooling logic
├── lib/           # Shared libraries and internal core packages
├── middlewares/   # gRPC interceptors (e.g., validation)
├── sql/           # Schema definitions, stored procedures, and role grants
├── utils/         # Helper functions and constants
├── main.go        # Application entrypoint and dependency wiring
└── go.mod         # Dependency management
```

## Features

- ⚡ **High-Performance API**: Exposes a fast, concurrent gRPC server built with modern Go.
- 🛡️ **Automated Validation**: Real-time payload validation via a gRPC unary interceptor. Invalid requests are rejected before they hit business logic.
- 🗄️ **Robust Data Storage**: Uses PostgreSQL for reliable data persistence. Raw SQL statements are maintained in the `sql/` directory for optimized query execution.
- 🚦 **Graceful Shutdown**: Handles OS signals to safely drain active connections and close database pools.

## API Summary

The service implements a multi-domain gRPC interface tailored exclusively for resume and portfolio management:
- **User Profile**: Endpoints for updating the user's main about description and avatar links.
- **Experiences**: Orchestrating complex timeline histories, including companies and deeply nested positions.
- **Projects**: Managing showcase projects, linked repositories, and featured status flags.
- **Technologies**: Mapping a taxonomy of tech-stack skills dynamically to projects and experiences.
- **Messages**: Storing and retrieving contact form submissions sent from the public website.

## Database Architecture

- **Dedicated Schema**: Operates under the `portfolio` schema, managing entities like `user`, `project`, `experience`, `position`, `message`, and `technology`.
- **Restricted Access**: Connects using the `manager_service` database role, which does **not** have direct `INSERT`, `UPDATE`, or `DELETE` permissions on tables.
- **Stored Procedures**: All complex writes, cascading deletions, and relational upserts (e.g., updating an experience timeline and linking project positions) are handled safely via strictly defined database functions (e.g., `portfolio.edit_experience`).

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.25.3 or higher
- A running [PostgreSQL](https://www.postgresql.org/download/) 18+ database instance
- Protocol buffer compilation tools (e.g., `buf`) if modifying definitions.

### Configuration

The service relies on environment variables for configuration. You should export these directly in your shell environment.

| Variable | Description | Default | Required |
| :--- | :--- | :--- | :---: |
| `POSTGRES_URL` | Connection string for the PostgreSQL database | - | **Yes** |
| `PORT` | The port on which the gRPC server will listen | `7296` | No |

### Polyrepo Local Setup

This project uses a polyrepo architecture. Services like `auth`, `manager`, and `gateway` rely on the `common` repository. To run this locally without dependency errors, clone all repositories side-by-side into the same parent directory so that relative paths resolve correctly.

Example local setup:
```text
infrastructure/
├── common/
├── auth/
├── manager/
└── gateway/
```

### Running the Service

For standard execution, use the Go CLI:

```bash
go run main.go
```

#### Hot Reloading for Local Development

To watch for file changes and automatically rebuild and restart the server, you can use [air](https://github.com/cosmtrek/air). Since `air` is written purely in Go, it compiles to a single binary and runs natively on Windows, macOS, and Linux without any OS-specific shell scripts.

```bash
go install github.com/air-verse/air@latest
air
```

## CI/CD & Production Deployment

The service is fully containerized and leverages GitHub Actions (`.github/workflows/main.yml`) for automated builds and deployment.

### 1. Server Initialization
Before the CI/CD pipeline can deploy the application for the first time, you must initialize the host environment on your VM.

Create the shared network that the containers will use to communicate:
```bash
podman network create infra-network
```

Create the environment variable configuration file:
```bash
touch ~/manager-service-production.conf
# Edit the file to include your POSTGRES_URL
```

### 2. Automated Build (GHCR)
The workflow uses Docker Buildx with QEMU to cross-compile a `linux/arm64` container image. 
Because the project depends on private GitHub modules (`common`), the workflow injects a `PAT_TOKEN` secret during the build stage to securely download the private contracts. The resulting image is pushed to the GitHub Container Registry (`ghcr.io/aditya-0011/manager-service`).

### 3. Production Execution (Podman)
In production, the service is deployed to a remote Linux VM using **Podman**. The CI pipeline automatically SSHs into the server, pulls the latest GHCR image, and seamlessly swaps the running container.

The container is orchestrated with the following constraints:
- Attached to the shared internal network (`--network infra-network`).
- Reads configuration purely from the host-level environment file (`--env-file ~/manager-service-production.conf`).
- Configured to restart automatically (`--restart unless-stopped`).

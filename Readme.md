![GoForge Logo](assets/logo_without_bg.png)

> [!NOTE]
> A comprehensive, production-ready Go application framework and CLI for rapid database, gRPC, and service scaffolding.

The **GoForge CLI** is a powerful command-line tool designed to accompany the GoForge application framework template. It provides a universal utility for rapidly scaffolding new services, generating database migrations, compiling protocol buffers, and managing application-specific console commands effortlessly out-of-the-box.

## ✨ Features

- **🚀 Zero-Configuration Awareness**: The CLI dynamically analyzes your project's `.env` and `go.mod` files for context; it executes tasks directly against your application state without complicated tooling setup.
- **🏗️ Service Scaffolding**: Automatically generate complete service skeletons, including database models, service contracts, gRPC stubs, API handlers, and routes.
- **🗄️ Database Tooling**: Effortlessly integrate with `atlas` and `sqlc` for clean database migrations and typed SQL query generation!
- **🔌 Extensibility**: Bootstrap custom local CLI tools to provide application specific console actions (e.g. running daily cronjobs, admin backfills, cache clearers, etc).

---

## 🏗️ Project Structure

A typical GoForge project follows a clean, modular architecture:

```text
.
├── cmd/                # Application entry points (main.go)
├── internal/           # Private application code
│   ├── cache/          # Caching drivers (Redis, Ristretto, Multi-tier)
│   ├── client/         # internal gRPC and HTTP clients
│   ├── config/         # Configuration loaders and types
│   ├── console/        # Application-specific CLI commands
│   ├── database/       # Database core, migrations, and SQLC queries
│   ├── server/         # HTTP and gRPC server setups & middleware
│   └── services/       # Domain-specific business logic & services
├── proto/              # Protocol Buffer definitions
├── tests/              # Unit, integration, and client tests
├── air.toml            # Air configuration for live reloading
├── atlas.hcl           # Atlas migration configuration
├── goforge.sh          # Helper shell script for CLI proxying
└── sqlc.yaml           # SQLC configuration
```

---

## 🏁 Getting Started

### Prerequisites

Ensure you have the following installed:

- [Go](https://go.dev/doc/install) (1.21+)
- [Atlas](https://atlasgo.io/getting-started/) (for migrations)
- [SQLC](https://sqlc.dev/) (for type-safe SQL)
- [Protoc](https://grpc.io/docs/protoc-installation/) (for gRPC)

### Installation

Install the **GoForge CLI** globally:

```bash
go install github.com/mmycin/GoForge@latest
```

Ensure your `$GOPATH/bin` is in your `$PATH`.

### Initializing a Project

```bash
goforge new my-awesome-project
cd my-awesome-project
```

---

## 🛠️ Usage & Commands

### 🔑 Core Utilities

| Command           | Description                                      |
| :---------------- | :----------------------------------------------- |
| `goforge new`     | Create a new GoForge project from the template.  |
| `goforge gen:key` | Generates and sets a secure `APP_KEY` in `.env`. |
| `goforge rem:key` | Removes the active `APP_KEY` from `.env`.        |
| `goforge readme`  | Displays the recommended GoForge workflow.       |
| `goforge version` | Displays the CLI version.                        |

### 🛠️ Service Generation

Scaffold domain logic rapidly:

- **`goforge gen:service [name]`**: Scaffolds a complete service in `internal/services/<name>`.
- **`goforge rem:service [name]`**: Safely removes a service and un-registers it.

### 📡 Protocol Buffers

- **`goforge gen:proto [name]`**: Compiles specific or all `.proto` files using `protoc`.
- **`goforge rem:proto`**: Scrubs all generated `.pb.go` files.

### 💻 Custom Application Commands

Bridge the gap between global tools and project-specific execution:

- **`goforge gen:command [name]`**: Creates a Cobra CLI command in `internal/console`.
- **`goforge app serve [args]`**: Proxies the local compiler. Running `goforge app serve` invokes `go run cmd/main.go serve`, loading the full environment.

### 🗄️ Database Operations

Native support for `gorm`, `atlas`, and `sqlc`:

- **`goforge migrate`**: Applies SQL migrations using Atlas.
- **`goforge gen:migration [name]`**: Generates schema differences using Atlas and GORM.
- **`goforge rem:migration`**: Reverts the most recent migration.
- **`goforge gen:sqlc`**: Compiles `internal/database/queries/*.sql` into type-safe Go.

---

## 📄 License

The `goforge` CLI is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for more details.

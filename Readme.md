# GoForge CLI

> A comprehensive, production-ready Go application framework and CLI for rapid database, gRPC, and service scaffolding.

The `GoForge` CLI is a powerful command-line tool designed to accompany the GoForge application framework template. It provides a universal utility for rapidly scaffolding new services, generating database migrations, compiling protocol buffers, and managing application-specific console commands effortlessly out-of-the-box.

## Features

- **Zero-Configuration Awareness**: The CLI dynamically analyzes your project's `.env` and `go.mod` files for context; it executes tasks directly against your application state without complicated tooling setup.
- **Service Scaffolding**: Automatically generate complete service skeletons, including database models, service contracts, gRPC stubs, API handlers, and routes.
- **Database Tooling**: Effortlessly integrate with `atlasschema` and `sqlc` for clean database migrations and typed SQL query generation!
- **Extensibility**: Bootstrap custom local CLI tools to provide application specific console actions (e.g. running daily cronjobs, admin backfills, cache clearers, etc).

## Installation

You can install the `GoForge` CLI directly from source using `go install`:

```bash
go install github.com/mmycin/GoForge@latest
```

_(adjust the GitHub URL above to the correct module path if necessary)_

Ensure that your `$GOPATH/bin` directory is available in your system's `$PATH` variable so you can run the `goforge` command globally!

To verify installation, run:

```bash
goforge version
```

## Usage & Commands

The CLI operates relative to the active working directory, assuming it is inside a valid `goforge-template` structured project.

### Core Utilities

- `goforge gen:key` - Generates a secure, cryptographically random `APP_KEY` and sets it in your `.env` file.
- `goforge rem:key` - Removes the active `APP_KEY` from your `.env` file.
- `goforge version` - Displays the framework CLI version and credits.

### Service Generation

- `goforge gen:service [name]` - Scaffolds a complete boilerplate API and gRPC service in `internal/services/<name>`. Includes routing, docs, interfaces, and proto definitions!
- `goforge rem:service [name]` - Permanently deletes the `<name>` service directory, its protobufs, and un-registers it from the application kernel.

### Protocol Buffers

- `goforge gen:proto [name]` - Compiles the specific `.proto` file utilizing `protoc` and generates the gRPC implementation structure. Omit the `name` argument to run compilation on all active services globally.
- `goforge rem:proto` - Scrubs all generated `.pb.go` and `_grpc.pb.go` implementations safely.

### Custom Application Commands

GoForge helps you bridge the gap between global scaffolding tools and project-specific execution!

- `goforge gen:command [name]` - Creates a standard Cobra CLI command boilerplate within your project's `internal/console` directory (e.g., `make:admin`).
- `goforge rem:command [name]` - Removes a generated Cobra command from your project.
- `goforge app run [command] [...args]` - This command proxies the local Go compiler! Executing `goforge app run cache:clear` will seamlessly invoke `go run cmd/main.go cache:clear` from within the active project directory, ensuring your custom commands load your full backend environment dependencies automatically.

### Database Operations

GoForge natively supports `gorm`, `atlas`, and `sqlc` code generations using these proxy commands:

- `goforge migrate` - Applies the current SQL migrations against the active database using Atlas.
- `goforge gen:migration [name]` - Parses the active GORM schema mappings via a dynamic `loader` and compares them to the active dev database to generate precise schema differences using Atlas!
- `goforge rem:migration` - Reverts and deletes the most recent database migration iteration.
- `goforge loader` - (Internal) Dynamically parses active models to output raw schema queries used during `gen:migration`.
- `goforge gen:sqlc` - Parses your custom queries in `internal/database/queries/*.sql` into type-safe Go bindings via `sqlc` and injects them seamlessly into the active `internal/database/database.go` core.
- `goforge rem:sqlc` - Removes `sqlc` hooks from the project core and deletes generated struct files.

## License

The `goforge` CLI is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for more details.

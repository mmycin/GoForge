# GoForge CLI Commands

A complete reference to all available GoForge CLI commands.

## Core Commands
These commands handle project setup and configuration.

### `goforge new`
Create a new GoForge project.
```bash
goforge new <project-name>
```

### `goforge gen:key`
Generate and set a secure `APP_KEY` in your `.env` file.
```bash
goforge gen:key
```

### `goforge rem:key`
Remove the current `APP_KEY` from your `.env` file.
```bash
goforge rem:key
```

### `goforge version`
Display the installed GoForge CLI version.
```bash
goforge version
```

## Service Commands
Commands for working with services.

### `goforge gen:service`
Generate a new service with all necessary files.
```bash
goforge gen:service <service-name>
```
This command creates:
- `internal/services/<service-name>/model.go`: The database model
- `internal/services/<service-name>/service.go`: Business logic layer
- `internal/services/<service-name>/handler.go`: HTTP handlers
- `internal/services/<service-name>/routes.go`: Route definitions
- `internal/services/<service-name>/docs.go`: API documentation setup
- `internal/services/<service-name>/grpc.go`: gRPC stub
- Updates `internal/services/kernel.go` with service registration

### `goforge rem:service`
Remove an existing service and clean up registrations.
```bash
goforge rem:service <service-name>
```

## Database Commands
Commands for managing your database.

### `goforge gen:migration`
Generate a new database migration by comparing GORM models to the current schema.
```bash
goforge gen:migration <migration-name>
```

### `goforge migrate`
Apply all pending migrations to your database.
```bash
goforge migrate
```

### `goforge rem:migration`
Revert the last applied migration.
```bash
goforge rem:migration
```

### `goforge gen:sqlc`
Compile SQL queries from `internal/database/queries` into type-safe Go code using SQLC.
```bash
goforge gen:sqlc
```

### `goforge loader`
Display the current GORM schema as interpreted by Atlas.
```bash
goforge loader
```

## gRPC Commands
Commands for working with Protocol Buffers and gRPC.

### `goforge gen:proto`
Compile Protocol Buffer files into Go gRPC stubs.
```bash
goforge gen:proto [service-name]
```
If no service name is provided, all `.proto` files in `proto/` are compiled.

### `goforge rem:proto`
Remove all generated `.pb.go` files from your project.
```bash
goforge rem:proto
```

## Custom Command
Commands for creating custom application console commands.

### `goforge gen:command`
Generate a new custom console command in `internal/console`.
```bash
goforge gen:command <command-name>
```

## Application Proxy Commands
Commands to run your application's own console commands.

### `goforge app`
Proxy commands to your application's own console. For example:
```bash
# Start your application's server
goforge app serve
```
This runs `go run cmd/main.go serve` in your project.

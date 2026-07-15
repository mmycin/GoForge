# Getting Started with GoForge

This guide walks you through setting up and using GoForge CLI to create your first GoForge project.

## Prerequisites
- **Go** 1.21 or later ([Download Go](https://go.dev/doc/install))
- **Atlas** for database migrations ([Install Atlas](https://atlasgo.io/getting-started/))
- **SQLC** for type-safe SQL queries ([Install SQLC](https://sqlc.dev/))
- **Protoc** for gRPC support ([Install Protoc](https://grpc.io/docs/protoc-installation/))

## Step 1: Install GoForge CLI

```bash
go install github.com/mmycin/GoForge@latest
goforge version
```

## Step 2: Create a New Project

```bash
goforge new my-project
cd my-project
```

## Step 3: Configure Your Environment

```bash
cp .env.example .env
goforge gen:key
```

## Step 4: Project Structure

```text
my-project/
├── app/                         # Entry point
│   └── main.go
├── boot/                        # Framework bootstrap (do not edit)
│   ├── kernel.go
│   ├── client/
│   └── server/
│       └── middleware/
├── core/                        # Framework packages — import freely, do not edit
│   ├── config/
│   ├── database/
│   ├── errors/
│   ├── validator/
│   └── console/
├── internal/                    # Your code — edit freely
│   ├── services/
│   │   ├── kernel.go            # Auto-generated — do not hand-edit
│   │   └── <name>/              # One directory per service
│   ├── database/
│   │   ├── migrations/
│   │   └── queries/
│   ├── proto/
│   │   └── <name>/
│   └── tests/
├── .env.example
├── atlas.hcl
├── sqlc.yaml
├── air.toml
└── Dockerfile
```

| Directory | Owner | Edit? |
|-----------|-------|-------|
| `boot/`   | Framework | No |
| `core/`   | Framework | No |
| `internal/services/` | You | Yes |
| `internal/database/` | You | Yes |
| `internal/proto/`    | You | Yes |
| `internal/tests/`    | You | Yes |

## Step 5: Start the server

```bash
goforge app serve
```

## Next Steps
- [Commands reference](./commands.md)
- [Configuration guide](./configuration.md)

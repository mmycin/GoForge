# Getting Started with GoForge

This guide will walk you through setting up and using GoForge CLI to create your first GoForge project.

## Prerequisites
Before you begin, ensure you have the following installed:
- **Go**: 1.21 or later ([Download Go](https://go.dev/doc/install))
- **Atlas**: For database migrations ([Install Atlas](https://atlasgo.io/getting-started/))
- **SQLC**: For type-safe SQL queries ([Install SQLC](https://sqlc.dev/))
- **Protoc**: For gRPC support ([Install Protoc](https://grpc.io/docs/protoc-installation/))

## Step 1: Install GoForge CLI
First, install the GoForge CLI globally on your system:

```bash
go install github.com/mmycin/GoForge@latest
```

Verify the installation:
```bash
goforge version
```

## Step 2: Create a New Project
Use the `new` command to generate a new GoForge project:

```bash
goforge new my-awesome-project
cd my-awesome-project
```

This will create a fully configured project with all the necessary files and directories.

## Step 3: Configure Your Environment
Copy the example environment file and generate a secure application key:

```bash
cp .env.example .env
goforge gen:key
```

## Step 4: Explore Your Project
Check out your new project structure:

```text
my-awesome-project/
├── boot/               # Core framework bootstrapping
├── cmd/                # Application entry point
├── internal/           # Private application code
│   ├── config/         # Configuration
│   ├── console/        # Custom commands
│   ├── database/       # Database setup
│   └── services/       # Domain services
├── proto/              # gRPC definitions
├── tests/              # Test files
├── .env.example        # Example environment file
├── Dockerfile          # Docker configuration
├── air.toml            # Live-reloading config
├── atlas.hcl           # Atlas migration config
└── sqlc.yaml           # SQLC config
```

## Step 5: Next Steps
Now you're ready to start building! Check out:
- [How to Generate a Service](./commands.md#gen-service)
- [Database Migrations](./commands.md#database-commands)

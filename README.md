<p align="center">
  <img src="assets/logo_without_bg.png" alt="GoForge Logo" width="200">
</p>

<h1 align="center">GoForge CLI</h1>

<p align="center">
  A powerful, production-ready CLI for Go application development with built-in database migrations, service scaffolding, and gRPC support.
</p>

<p align="center">
  <a href="https://github.com/mmycin/GoForge/actions"><img src="https://github.com/mmycin/GoForge/actions/workflows/release.yml/badge.svg" alt="Build Status"></a>
  <a href="https://github.com/mmycin/GoForge/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/mmycin/GoForge"><img src="https://goreportcard.com/badge/github.com/mmycin/GoForge" alt="Go Report Card"></a>
</p>

---

## 🚀 Quick Start
Install the GoForge CLI globally:
```bash
go install github.com/mmycin/GoForge@latest
```

Create your first GoForge project:
```bash
goforge new my-awesome-project
cd my-awesome-project
```

## ✨ Features
- **⚡ Service Scaffolding**: Generate complete service stacks (models, handlers, gRPC stubs, docs) with dependency injection in seconds
- **🗄️ Database Tools**: Built-in support for Atlas migrations and SQLC type-safe queries
- **📡 gRPC Support**: Generate and compile proto files with a single command
- **🔌 Extensibility**: Create custom application-specific console commands
- **🧪 Testable Code**: Dependency injection by default for easy testing

## 📖 Documentation
For detailed documentation, please visit [GoForge Docs](./docs/index.md)

## 📄 License
GoForge CLI is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for more details.

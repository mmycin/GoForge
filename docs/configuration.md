# Configuration

GoForge uses a `.env` file to manage configuration. This file is generated when you create a new project and is located at the root of your project.

## Example `.env` File
```env
# Application
APP_NAME=GoForge
APP_VERSION=1.0.0
APP_MODULE=github.com/mmycin/goforge
APP_DEBUG=true
APP_KEY=
APP_HOST=localhost
APP_PORT=8080

# Database
DB_CONNECTION=sqlite
DB_HOST=localhost
DB_PORT=5432
DB_NAME=database.db
DB_USER=
DB_PASSWORD=
DB_MIGRATOR=atlas

# gRPC
GRPC_ENABLE=false
GRPC_REFLECTION=false
GRPC_HOST=localhost
GRPC_PORT=9090

# HTTP
HTTP_RATE_LIMIT_PER_MINUTE=60
CORS_ALLOWED_ORIGINS=*

# Encryption
ENCRYPTION_KEY=
ENCRYPTION_ROUNDS=4

# Logging
LOG_TYPE=stdout
LOG_LEVEL=debug
LOG_PATH=
LOG_FORMAT=text
```

## Configuration Options

### Application Options
| Option          | Description                                  |
|-----------------|----------------------------------------------|
| `APP_NAME`      | Name of your application                     |
| `APP_VERSION`   | Version of your application                  |
| `APP_MODULE`    | Go module name of your application           |
| `APP_DEBUG`     | Debug mode (`true` or `false`)               |
| `APP_KEY`       | Secret key for encryption                    |
| `APP_HOST`      | Host address for HTTP server                 |
| `APP_PORT`      | Port for HTTP server                         |

### Database Options
| Option              | Description                                  |
|---------------------|----------------------------------------------|
| `DB_CONNECTION`     | Database driver (`sqlite`, `postgres`, `mysql`, `sqlserver`) |
| `DB_HOST`           | Database host address                        |
| `DB_PORT`           | Database port                                |
| `DB_NAME`           | Database name or file path for SQLite        |
| `DB_USER`           | Database username                            |
| `DB_PASSWORD`       | Database password                            |
| `DB_MIGRATOR`       | Migration tool to use (`atlas` or `gorm`)    |

### gRPC Options
| Option               | Description                                  |
|----------------------|----------------------------------------------|
| `GRPC_ENABLE`        | Enable gRPC server (`true` or `false`)       |
| `GRPC_REFLECTION`    | Enable gRPC reflection (`true` or `false`)   |
| `GRPC_HOST`          | gRPC server host                             |
| `GRPC_PORT`          | gRPC server port                             |

### HTTP Options
| Option                   | Description                              |
|--------------------------|------------------------------------------|
| `HTTP_RATE_LIMIT_PER_MINUTE` | Requests per minute per IP address       |
| `CORS_ALLOWED_ORIGINS`      | Comma-separated list of allowed origins  |

### Logging Options
| Option        | Description                                  |
|---------------|----------------------------------------------|
| `LOG_TYPE`    | Log output (`stdout`, `file`)                |
| `LOG_LEVEL`   | Log level (`debug`, `info`, `warn`, `error`) |
| `LOG_PATH`    | Path to log file if using file output        |
| `LOG_FORMAT`  | Log format (`text`, `json`)                  |

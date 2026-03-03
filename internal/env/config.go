package env

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Module       string
	DBConnection string
	DBName       string
	DBUsername   string
	DBPassword   string
	DBHost       string
	DBPort       string
	DBDevName    string
}

// Load reads the local go.mod and .env files to construct a configuration object
// that the CLI can use to operate on the local project directory.
func Load() (*Config, error) {
	cfg := &Config{}

	// Load Module name from go.mod
	modBytes, err := os.ReadFile("go.mod")
	if err == nil {
		lines := strings.Split(string(modBytes), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				cfg.Module = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading go.mod: %w", err)
	}

	// Load DB Config from .env
	envBytes, err := os.ReadFile(".env")
	if err == nil {
		lines := strings.Split(string(envBytes), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
				// Strip surrounding quotes if present
				val = strings.Trim(val, `"'`)
				switch key {
				case "DB_CONNECTION":
					cfg.DBConnection = val
				case "DB_NAME":
					cfg.DBName = val
				case "DB_USERNAME":
					cfg.DBUsername = val
				case "DB_PASSWORD":
					cfg.DBPassword = val
				case "DB_HOST":
					cfg.DBHost = val
				case "DB_PORT":
					cfg.DBPort = val
				case "DB_DEV_NAME":
					cfg.DBDevName = val
				}
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error reading .env: %w", err)
	}

	return cfg, nil
}



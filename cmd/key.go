package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(genKeyCmd)
	rootCmd.AddCommand(remKeyCmd)
}

var genKeyCmd = &cobra.Command{
	Use:   "gen:key",
	Short: "Generate a new APP_KEY",
	Long:  `Generate a cryptographically secure 32-byte key and save it to the .env file.`,
	Run: func(cmd *cobra.Command, args []string) {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			fmt.Printf("Error generating key: %v\n", err)
			os.Exit(1)
		}

		encodedKey := base64.StdEncoding.EncodeToString(key)
		fmt.Printf("→ Generated Key: %s\n", encodedKey)

		if err := updateEnvKey("APP_KEY", encodedKey); err != nil {
			fmt.Printf("Error updating .env file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ APP_KEY successfully updated in .env")
	},
}

var remKeyCmd = &cobra.Command{
	Use:   "rem:key",
	Short: "Remove the APP_KEY from .env",
	Long:  `Clear the APP_KEY value in your .env file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Removing APP_KEY from .env...")
		if err := updateEnvKey("APP_KEY", ""); err != nil {
			fmt.Printf("Error clearing APP_KEY: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ APP_KEY cleared from .env")
	},
}

func updateEnvKey(key, value string) error {
	content, err := os.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(".env", []byte(fmt.Sprintf("%s=%s\n", key, value)), 0644)
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = fmt.Sprintf("%s=%s", key, value)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	return os.WriteFile(".env", []byte(strings.Join(lines, "\n")), 0644)
}

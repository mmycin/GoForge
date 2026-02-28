package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const banner = `
   ____       ______                    
  / ___| ___ |  ____|___  _ __ __ _  ___ 
 | |  _ / _ \| |_  / _ \| '__/ _` + "`" + ` |/ _ \
 | |_| | (_) |  _|| (_) | | | (_| |  __/
  \____|\___/|_|   \___/|_|  \__, |\___|
                             |___/      
`

var rootCmd = &cobra.Command{
	Use:   "GoForge",
	Short: "A comprehensive, production-ready Go application framework CLI",
	Long: banner + `
GoForge is a comprehensive, production-ready Go application framework designed to provide robust tooling for database migrations, code generation, caching, and service scaffolding. It comes with built-in support for multiple SQL databases, gRPC and HTTP servers, tiered caching, and a powerful CLI to speed up development.`,
	// Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

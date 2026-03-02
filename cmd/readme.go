package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(readmeCmd)
}

var readmeCmd = &cobra.Command{
	Use:   "readme",
	Short: "Show the GoForge recommended workflow",
	Long:  `Display the step-by-step workflow for creating and managing a GoForge project.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GoForge Recommended Workflow:")
		fmt.Println("  1. goforge new              - Create a new project")
		fmt.Println("  2. goforge gen:key          - Generate application key")
		fmt.Println("  3. goforge gen:migration init - Initialize migrations")
		fmt.Println("  4. goforge migrate          - Run migrations")
		fmt.Println("  5. goforge gen:proto        - Generate gRPC code")
		fmt.Println("  6. goforge gen:sqlc         - Generate SQL bindings")
		fmt.Println("  7. goforge app serve        - Start the application")
		fmt.Println("")
		fmt.Println("To create a new service:")
		fmt.Println("  goforge gen:service [name]")
	},
}

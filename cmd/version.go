package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GoForge",
	Long:  `All software has versions. This is GoForge's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GoForge CLI v0.1.0")
		fmt.Println("Created by Tahcin Ul Karim Mycin a professional software engineer")
		fmt.Println("Description: A comprehensive, production-ready Go application framework CLI")
	},
}

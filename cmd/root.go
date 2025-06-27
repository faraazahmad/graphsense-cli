package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "graphsense-cli",
	Short: "GraphSense Multi-Instance Deployment CLI",
	Long: `GraphSense CLI for managing multiple GraphSense instances using Docker Compose.
This tool allows you to deploy, manage, and monitor GraphSense instances for different repositories.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(cleanupCmd)
}

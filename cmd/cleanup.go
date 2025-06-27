package cmd

import (
	"graphsense-cli/internal"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up stopped containers and unused volumes",
	Long:  "Remove all stopped containers and unused volumes to free up disk space.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cleanup()
	},
}

func cleanup() error {
	internal.Log.Info("Cleaning up stopped containers and unused volumes...")
	
	// Clean up stopped containers
	err := internal.RunDockerCompose([]string{"container", "prune", "-f"}, nil)
	if err != nil {
		internal.Log.Warning("Failed to clean up containers, continuing...")
	}

	// Clean up unused volumes  
	err = internal.RunDockerCompose([]string{"volume", "prune", "-f"}, nil)
	if err != nil {
		internal.Log.Warning("Failed to clean up volumes, continuing...")
	}

	internal.Log.Success("Cleanup completed.")
	return nil
}

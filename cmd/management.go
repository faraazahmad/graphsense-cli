package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"graphsense-cli/internal"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <instance_name>",
	Short: "Stop a GraphSense instance",
	Long:  "Stop a running GraphSense instance without removing it.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopInstance(args[0])
	},
}

var startCmd = &cobra.Command{
	Use:   "start <instance_name>",
	Short: "Start a GraphSense instance",
	Long:  "Start a stopped GraphSense instance.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return startInstance(args[0])
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <instance_name>",
	Short: "Remove a GraphSense instance",
	Long:  "Permanently remove a GraphSense instance and all its data.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeInstance(args[0])
	},
}

func stopInstance(instanceName string) error {
	if !internal.InstanceExists(instanceName) {
		return fmt.Errorf("instance '%s' does not exist", instanceName)
	}

	internal.Log.Info(fmt.Sprintf("Stopping instance: %s", instanceName))

	envVars := map[string]string{
		"COMPOSE_PROJECT_NAME": instanceName,
	}

	// Find the docker-compose.yml in any existing container for this instance
	// We'll use docker-compose without specifying -f since the project name is set
	err := internal.RunDockerCompose([]string{
		"stop",
	}, envVars)
	if err != nil {
		return fmt.Errorf("failed to stop instance %s: %v", instanceName, err)
	}

	internal.Log.Success(fmt.Sprintf("Instance '%s' stopped.", instanceName))
	return nil
}

func startInstance(instanceName string) error {
	if !internal.InstanceExists(instanceName) {
		return fmt.Errorf("instance '%s' does not exist", instanceName)
	}

	internal.Log.Info(fmt.Sprintf("Starting instance: %s", instanceName))

	envVars := map[string]string{
		"COMPOSE_PROJECT_NAME": instanceName,
	}

	// Use docker-compose without specifying -f since the project name is set
	err := internal.RunDockerCompose([]string{
		"start",
	}, envVars)
	if err != nil {
		return fmt.Errorf("failed to start instance %s: %v", instanceName, err)
	}

	internal.Log.Success(fmt.Sprintf("Instance '%s' started.", instanceName))
	return nil
}

func removeInstance(instanceName string) error {
	if !internal.InstanceExists(instanceName) {
		return fmt.Errorf("instance '%s' does not exist", instanceName)
	}

	internal.Log.Warning(fmt.Sprintf("This will permanently remove instance '%s' and all its data.", instanceName))
	fmt.Print("Are you sure? (y/N): ")
	
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		internal.Log.Info("Cancelled.")
		return nil
	}

	internal.Log.Info(fmt.Sprintf("Removing instance: %s", instanceName))

	envVars := map[string]string{
		"COMPOSE_PROJECT_NAME": instanceName,
	}

	// Stop and remove containers
	err = internal.RunDockerCompose([]string{
		"down", "-v", "--remove-orphans",
	}, envVars)
	if err != nil {
		internal.Log.Warning("Failed to cleanly remove instance with docker-compose, trying manual cleanup...")
		
		// Manual cleanup as fallback
		if err := internal.RunDockerCompose([]string{
			"ps", "-a", "--filter", fmt.Sprintf("label=com.docker.compose.project=%s", instanceName), "-q",
		}, nil); err == nil {
			internal.RunDockerCompose([]string{"rm", "-f"}, nil)
		}
	}

	// Remove associated volumes
	internal.Log.Info("Removing associated volumes...")
	internal.RunDockerCompose([]string{
		"volume", "ls", "-q", "|", "grep", fmt.Sprintf("^%s_", instanceName), "|", "xargs", "-r", "docker", "volume", "rm",
	}, nil)

	internal.Log.Success(fmt.Sprintf("Instance '%s' removed.", instanceName))
	return nil
}

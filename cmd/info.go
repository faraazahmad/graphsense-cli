package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"graphsense-cli/internal"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all GraphSense instances",
	Long:  "List all running and stopped GraphSense instances.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listInstances()
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs <instance_name> [service]",
	Short: "Show logs for a GraphSense instance",
	Long:  "Show logs for a GraphSense instance. Optionally specify a service (app, postgres, neo4j).",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceName := args[0]
		var service string
		if len(args) > 1 {
			service = args[1]
		}
		return showLogs(instanceName, service)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status <instance_name>",
	Short: "Show status of a GraphSense instance",
	Long:  "Show the status and details of a GraphSense instance.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return showStatus(args[0])
	},
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Show debug information",
	Long:  "Show port usage and debug information for troubleshooting.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return debugPorts()
	},
}

func listInstances() error {
	internal.Log.Info("GraphSense Instances:")
	fmt.Println()

	// Get all containers with graphsense in their name
	cmd := exec.Command("docker", "ps", "--format", "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list containers: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var graphsenseContainers []string
	
	for _, line := range lines {
		if strings.Contains(line, "graphsense-") {
			graphsenseContainers = append(graphsenseContainers, line)
		}
	}

	if len(graphsenseContainers) == 0 {
		internal.Log.Info("No instances found.")
		return nil
	}

	for _, container := range graphsenseContainers {
		fmt.Println(container)
	}

	return nil
}

func showLogs(instanceName, service string) error {
	if !internal.InstanceExists(instanceName) {
		return fmt.Errorf("instance '%s' does not exist", instanceName)
	}

	envVars := map[string]string{
		"COMPOSE_PROJECT_NAME": instanceName,
	}

	args := []string{
		"logs", "-f",
	}

	if service != "" {
		args = append(args, service)
	}

	return internal.RunDockerCompose(args, envVars)
}

func showStatus(instanceName string) error {
	if !internal.InstanceExists(instanceName) {
		return fmt.Errorf("instance '%s' does not exist", instanceName)
	}

	internal.Log.Info("Container details:")
	
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("label=com.docker.compose.project=%s", instanceName), "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

func debugPorts() error {
	internal.Log.Info("Port Usage Debug Information")
	fmt.Println()

	// Show currently listening ports (GraphSense related)
	internal.Log.Info("Currently listening ports (GraphSense related):")
	cmd := exec.Command("sh", "-c", "netstat -an 2>/dev/null | grep LISTEN | grep -E ':(808[0-9]|5[0-9][0-9][0-9]|74[0-9][0-9]|76[0-9][0-9])' | sort -n -k4 -t:")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		fmt.Println("No GraphSense ports detected")
	} else {
		fmt.Print(string(output))
	}

	fmt.Println()
	internal.Log.Info("Docker containers with port mappings:")
	cmd = exec.Command("sh", "-c", "docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Ports}}' | grep -E '(graphsense|neo4j|postgres)'")
	output, err = cmd.Output()
	if err != nil || len(output) == 0 {
		fmt.Println("No GraphSense containers running")
	} else {
		fmt.Print(string(output))
	}

	fmt.Println()
	internal.Log.Info("GraphSense Docker Compose projects:")
	cmd = exec.Command("sh", "-c", "docker ps --filter 'label=com.docker.compose.project' --format 'table {{.Names}}\t{{.Label \"com.docker.compose.project\"}}\t{{.Ports}}' | grep graphsense")
	output, err = cmd.Output()
	if err != nil || len(output) == 0 {
		fmt.Println("No GraphSense compose projects detected")
	} else {
		fmt.Print(string(output))
	}

	fmt.Println()
	internal.Log.Info("Available port ranges starting from common bases:")
	
	basePorts := []int{8080, 8090, 8100, 8110, 8120}
	for _, basePort := range basePorts {
		appPort := basePort
		postgresPort := basePort + 100
		neo4jBoltPort := basePort + 200

		conflicts := ""
		if internal.IsPortInUse(appPort) {
			conflicts += fmt.Sprintf(" APP:%d", appPort)
		}
		if internal.IsPortInUse(postgresPort) {
			conflicts += fmt.Sprintf(" PG:%d", postgresPort)
		}
		if internal.IsPortInUse(neo4jBoltPort) {
			conflicts += fmt.Sprintf(" NEO4J-BOLT:%d", neo4jBoltPort)
		}

		if conflicts == "" {
			fmt.Printf("  Base %d: ✅ AVAILABLE (App:%d, PG:%d, Neo4j:%d)\n", basePort, appPort, postgresPort, neo4jBoltPort)
		} else {
			fmt.Printf("  Base %d: ❌ CONFLICTS -%s\n", basePort, conflicts)
		}
	}

	fmt.Println()
	internal.Log.Info("Next available base port:")
	nextPort, err := internal.FindAvailablePortSet(8080)
	if err != nil {
		return fmt.Errorf("failed to find available port: %v", err)
	}
	
	fmt.Printf("  Recommended base port: %d\n", nextPort)
	fmt.Println("  Ports that will be used:")
	fmt.Printf("    - MCP Server: %d\n", nextPort)
	fmt.Printf("    - PostgreSQL: %d\n", nextPort+100)
	fmt.Printf("    - Neo4j Bolt: %d\n", nextPort+200)

	return nil
}



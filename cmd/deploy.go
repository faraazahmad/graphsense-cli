package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"graphsense-cli/internal"

	"github.com/spf13/cobra"
)

var (
	port int
)

var deployCmd = &cobra.Command{
	Use:   "deploy <repo_path> [instance_name]",
	Short: "Deploy a new GraphSense instance",
	Long: `Deploy a new GraphSense instance for the given repository.
If instance_name is not provided, it will be generated from the repository name.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := args[0]
		var instanceName string
		
		if len(args) > 1 {
			instanceName = args[1]
		}

		return deployInstance(repoPath, instanceName, port)
	},
}

func init() {
	deployCmd.Flags().IntVar(&port, "port", 0, "Base port for the instance (default: auto-assigned)")
}

func deployInstance(repoPath, instanceName string, basePort int) error {
	// Validate repo path
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository path does not exist: %s", repoPath)
	}

	// Convert to absolute path
	absRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Generate instance name if not provided
	if instanceName == "" {
		instanceName = internal.GenerateInstanceName(absRepoPath)
	}

	// Sanitize instance name
	instanceName = internal.SanitizeInstanceName(instanceName)

	internal.Log.Info(fmt.Sprintf("Deploying instance: %s for repository: %s", instanceName, absRepoPath))

	// Check if instance already exists
	if internal.InstanceExists(instanceName) {
		return fmt.Errorf("instance '%s' already exists. Use 'remove' command first", instanceName)
	}

	// Get available ports
	appPort, err := internal.FindAvailablePortSet(basePort)
	if err != nil {
		return fmt.Errorf("failed to find available ports: %v", err)
	}
	
	postgresPort := appPort + 100
	neo4jBoltPort := appPort + 200

	// Load API keys from ~/.graphsense/.env
	coAPIKey, anthropicAPIKey, err := internal.LoadAPIKeys()
	if err != nil {
		return fmt.Errorf("failed to load API keys: %v", err)
	}

	// Create deployment configuration
	config := &internal.DeployConfig{
		RepoPath:         absRepoPath,
		InstanceName:     instanceName,
		AppPort:          appPort,
		PostgresPort:     postgresPort,
		Neo4jBoltPort:    neo4jBoltPort,
		CoAPIKey:         coAPIKey,
		AnthropicAPIKey:  anthropicAPIKey,
	}

	// Create temporary environment file
	envFile, err := internal.CreateTempEnvFile(config)
	if err != nil {
		return fmt.Errorf("failed to create environment file: %v", err)
	}
	defer os.Remove(envFile)

	// Create instance-specific docker-compose override
	composeOverride, err := internal.CreateComposeOverride(config)
	if err != nil {
		return fmt.Errorf("failed to create compose override: %v", err)
	}
	defer os.Remove(composeOverride)

	// Deploy the instance using the docker-compose.yml in the target repository
	internal.Log.Info(fmt.Sprintf("Starting services for instance: %s", instanceName))

	envVars := map[string]string{
		"COMPOSE_PROJECT_NAME": instanceName,
	}

	// Use the docker-compose.yml from ~/oss/code-graph-rag/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	
	composeFile := filepath.Join(homeDir, "oss", "code-graph-rag", "docker-compose.yml")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found at: %s", composeFile)
	}

	err = internal.RunDockerCompose([]string{
		"-f", composeFile,
		"-f", composeOverride,
		"--env-file", envFile,
		"up", "-d",
	}, envVars)
	if err != nil {
		return fmt.Errorf("failed to deploy instance %s: %v", instanceName, err)
	}

	// Wait for services to be healthy
	if err := internal.WaitForHealthy(instanceName, 60); err != nil {
		internal.Log.Warning("Health check failed, but continuing...")
	}

	// Store container information in database
	if err := internal.StoreInstanceContainers(config); err != nil {
		internal.Log.Warning(fmt.Sprintf("Failed to store container information: %v", err))
	}

	internal.Log.Success(fmt.Sprintf("Instance '%s' deployed successfully!", instanceName))
	internal.Log.Info("Access URLs:")
	internal.Log.Info(fmt.Sprintf("  MCP Server: http://localhost:%d", appPort))
	internal.Log.Info(fmt.Sprintf("  PostgreSQL: localhost:%d", postgresPort))
	internal.Log.Info(fmt.Sprintf("  Neo4j Bolt: bolt://localhost:%d", neo4jBoltPort))

	return nil
}

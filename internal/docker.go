package internal

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBasePort     = 8080
	DefaultPostgresPort = 5432
	DefaultNeo4jPort    = 7687
)

type Logger struct{}

func (l *Logger) Info(msg string) {
	fmt.Printf("\033[0;34m[INFO]\033[0m %s\n", msg)
}

func (l *Logger) Success(msg string) {
	fmt.Printf("\033[0;32m[SUCCESS]\033[0m %s\n", msg)
}

func (l *Logger) Warning(msg string) {
	fmt.Printf("\033[1;33m[WARNING]\033[0m %s\n", msg)
}

func (l *Logger) Error(msg string) {
	fmt.Printf("\033[0;31m[ERROR]\033[0m %s\n", msg)
}

var Log = &Logger{}

// FindAvailablePortSet finds the next available base port where all required ports are free
func FindAvailablePortSet(basePort int) (int, error) {
	if basePort == 0 {
		basePort = DefaultBasePort
	}

	port := basePort

	for {
		appPort := port
		postgresPort := port + 100
		neo4jBoltPort := port + 200

		// Check if any of the required ports are in use
		if isPortInUse(appPort) || isPortInUse(postgresPort) || isPortInUse(neo4jBoltPort) {
			port += 10 // Skip by 10 to avoid conflicts
		} else {
			break
		}

		// Safety check to avoid infinite loop
		if port > 65000 {
			return 0, fmt.Errorf("unable to find available port set starting from %d", basePort)
		}
	}

	return port, nil
}

// isPortInUse checks if a port is currently in use
func isPortInUse(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

// IsPortInUse checks if a port is currently in use (exported version)
func IsPortInUse(port int) bool {
	return isPortInUse(port)
}

// GenerateInstanceName generates an instance name from a repository path
func GenerateInstanceName(repoPath string) string {
	repoName := filepath.Base(repoPath)
	// Convert to lowercase and replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	sanitized := reg.ReplaceAllString(strings.ToLower(repoName), "-")
	return "graphsense-" + strings.Trim(sanitized, "-")
}

// SanitizeInstanceName sanitizes an instance name
func SanitizeInstanceName(name string) string {
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	return reg.ReplaceAllString(strings.ToLower(name), "-")
}

// InstanceExists checks if a Docker Compose instance exists
func InstanceExists(instanceName string) bool {
	cmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("label=com.docker.compose.project=%s", instanceName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// CreateTempEnvFile creates a temporary environment file for Docker Compose
func CreateTempEnvFile(config *DeployConfig) (string, error) {
	tmpFile, err := os.CreateTemp("", "graphsense-env-*.env")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	content := fmt.Sprintf(`# Repository Configuration
REPO_PATH=%s

# Port Configuration
PORT=%d
POSTGRES_PORT=%d
NEO4J_BOLT_PORT=%d

# Database Configuration
POSTGRES_DB=graphsense
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

# Neo4j Configuration
NEO4J_AUTH=none
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=

# Application Configuration
NODE_ENV=production
LOG_LEVEL=info
INDEX_FROM_SCRATCH=true

# Security Configuration
CORS_ORIGIN=*
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW=900000
`, config.RepoPath, config.AppPort, config.PostgresPort, config.Neo4jBoltPort)

	if config.CoAPIKey != "" {
		content += fmt.Sprintf("CO_API_KEY=%s\n", config.CoAPIKey)
	}

	if config.AnthropicAPIKey != "" {
		content += fmt.Sprintf("ANTHROPIC_API_KEY=%s\n", config.AnthropicAPIKey)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// CreateComposeOverride creates a Docker Compose override file
func CreateComposeOverride(config *DeployConfig) (string, error) {
	tmpFile, err := os.CreateTemp("", "graphsense-compose-*.yml")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	content := fmt.Sprintf(`version: "3.8"

services:
  postgres:
    container_name: %s-postgres
    volumes:
      - %s_postgres_data:/var/lib/postgresql/data
    networks:
      - %s-network

  neo4j:
    container_name: %s-neo4j
    volumes:
      - %s_neo4j_data:/data
      - %s_neo4j_logs:/logs
      - %s_neo4j_plugins:/plugins
      - %s_neo4j_conf:/conf
    networks:
      - %s-network

  app:
    container_name: %s-app
    volumes:
      - %s_app_repos:/app/.graphsense
      - %s:/home/repo:ro
    ports:
      - "%d:8080"
    networks:
      - %s-network
    environment:
      - POSTGRES_URL=postgresql://postgres:postgres@%s-postgres:5432/${POSTGRES_DB}
      - NEO4J_URI=bolt://%s-neo4j:7687
      - LOCAL_REPO_PATH=/home/repo

networks:
  %s-network:
    driver: bridge

volumes:
  %s_postgres_data:
    name: %s_postgres_data
  %s_neo4j_data:
    name: %s_neo4j_data
  %s_neo4j_logs:
    name: %s_neo4j_logs
  %s_neo4j_plugins:
    name: %s_neo4j_plugins
  %s_neo4j_conf:
    name: %s_neo4j_conf
  %s_app_repos:
    name: %s_app_repos
`,
		config.InstanceName, config.InstanceName, config.InstanceName,
		config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName,
		config.InstanceName, config.InstanceName, config.RepoPath, config.AppPort, config.InstanceName, config.InstanceName, config.InstanceName,
		config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName, config.InstanceName)

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// RunDockerCompose runs a docker-compose command
func RunDockerCompose(args []string, envVars map[string]string) error {
	cmd := exec.Command("docker-compose", args...)

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// WaitForHealthy waits for services to become healthy
func WaitForHealthy(instanceName string, maxAttempts int) error {
	Log.Info("Waiting for services to be healthy...")

	for attempt := 0; attempt < maxAttempts; attempt++ {
		cmd := exec.Command("docker-compose", "ps")
		cmd.Env = append(os.Environ(), fmt.Sprintf("COMPOSE_PROJECT_NAME=%s", instanceName))

		output, err := cmd.Output()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		// Simple check - if we see "Up" or "healthy" in the output, consider it healthy
		outputStr := string(output)
		if strings.Contains(outputStr, "Up") {
			return nil
		}

		time.Sleep(5 * time.Second)
		Log.Info(fmt.Sprintf("Waiting for health checks... (%d/%d)", attempt+1, maxAttempts))
	}

	Log.Warning("Not all services became healthy within timeout, but continuing...")
	return nil
}

// DeployConfig holds configuration for deployment
type DeployConfig struct {
	RepoPath        string
	InstanceName    string
	AppPort         int
	PostgresPort    int
	Neo4jBoltPort   int
	CoAPIKey        string
	AnthropicAPIKey string
}

// GetRunningInstances returns a list of running GraphSense instances
func GetRunningInstances() ([]string, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var instances []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "graphsense-") {
			instances = append(instances, line)
		}
	}

	return instances, nil
}

// GetPortsInUse returns a list of ports currently in use
func GetPortsInUse() ([]int, error) {
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var ports []int
	re := regexp.MustCompile(`:(\d+)\s`)
	matches := re.FindAllStringSubmatch(string(output), -1)

	for _, match := range matches {
		if len(match) > 1 {
			port, err := strconv.Atoi(match[1])
			if err == nil {
				ports = append(ports, port)
			}
		}
	}

	return ports, nil
}

// LoadAPIKeys loads API keys from ~/.graphsense/.env
func LoadAPIKeys() (coAPIKey, anthropicAPIKey string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get home directory: %v", err)
	}

	envFile := filepath.Join(homeDir, ".graphsense", ".env")
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return "", "", fmt.Errorf("API keys file not found: %s", envFile)
	}

	file, err := os.Open(envFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to open API keys file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "CO_API_KEY":
			coAPIKey = value
		case "ANTHROPIC_API_KEY":
			anthropicAPIKey = value
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("failed to read API keys file: %v", err)
	}

	return coAPIKey, anthropicAPIKey, nil
}

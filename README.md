# GraphSense CLI

A Go-based command-line interface for managing GraphSense multi-instance deployments using Docker Compose.

## Features

- Deploy new GraphSense instances for different repositories
- Manage instance lifecycle (start, stop, remove)
- Monitor instances (logs, status, list)
- Debug port conflicts and system state
- Clean up unused Docker resources

## Installation

### Build from Source

```bash
git clone <repository-url>
cd graphsense-cli
go mod tidy
go build -o graphsense-cli
```

### Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for building from source)
- `netstat` command available on your system

## Usage

### Deploy a New Instance

```bash
# Deploy with auto-generated instance name
./graphsense-cli deploy /path/to/repository

# Deploy with custom instance name
./graphsense-cli deploy /path/to/repository my-analysis

# Deploy with specific port and API keys
./graphsense-cli deploy /path/to/repository my-analysis --port 8090 --co-api-key YOUR_KEY --anthropic-api-key YOUR_KEY
```

### Manage Instances

```bash
# List all instances
./graphsense-cli list

# Stop an instance
./graphsense-cli stop my-analysis

# Start a stopped instance
./graphsense-cli start my-analysis

# Remove an instance permanently
./graphsense-cli remove my-analysis
```

### Monitor Instances

```bash
# Show logs for all services
./graphsense-cli logs my-analysis

# Show logs for specific service
./graphsense-cli logs my-analysis app

# Show instance status
./graphsense-cli status my-analysis
```

### Debug and Cleanup

```bash
# Show port usage and debug information
./graphsense-cli debug

# Clean up stopped containers and unused volumes
./graphsense-cli cleanup
```

## Port Configuration

The CLI automatically assigns ports to avoid conflicts:

- **App Port**: Base port (default: 8080)
- **PostgreSQL**: Base port + 100 (default: 8180)
- **Neo4j Bolt**: Base port + 200 (default: 8280)

The CLI will automatically find the next available port set if the default ports are in use.

## Commands Reference

| Command | Description | Arguments |
|---------|-------------|-----------|
| `deploy` | Deploy a new instance | `<repo_path> [instance_name]` |
| `stop` | Stop an instance | `<instance_name>` |
| `start` | Start a stopped instance | `<instance_name>` |
| `remove` | Remove an instance permanently | `<instance_name>` |
| `list` | List all instances | - |
| `logs` | Show instance logs | `<instance_name> [service]` |
| `status` | Show instance status | `<instance_name>` |
| `debug` | Show debug information | - |
| `cleanup` | Clean up Docker resources | - |

## Options

| Option | Description | Commands |
|--------|-------------|----------|
| `--port` | Base port for the instance | `deploy` |
| `--co-api-key` | Cohere API key | `deploy` |
| `--anthropic-api-key` | Anthropic API key | `deploy` |

## Configuration Files

The CLI expects each GraphSense repository to contain its own `docker-compose.yml` file with the service definitions for that specific application.

## Error Handling

The CLI provides colored output for different message types:
- **INFO**: Blue text for informational messages
- **SUCCESS**: Green text for successful operations
- **WARNING**: Yellow text for warnings
- **ERROR**: Red text for errors

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

[Add your license information here]

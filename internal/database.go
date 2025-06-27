package internal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Instance represents a deployed GraphSense instance
type Instance struct {
	ID            int    `json:"id"`
	InstanceName  string `json:"instance_name"`
	ContainerName string `json:"container_name"`
	RepoPath      string `json:"repo_path"`
	AppPort       int    `json:"app_port"`
	PostgresPort  int    `json:"postgres_port"`
	Neo4jBoltPort int    `json:"neo4j_bolt_port"`
	CreatedAt     string `json:"created_at"`
}

// InitDB initializes the SQLite database
func InitDB() (*sql.DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	graphsenseDir := filepath.Join(homeDir, ".graphsense")
	if err := os.MkdirAll(graphsenseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .graphsense directory: %v", err)
	}

	dbPath := filepath.Join(graphsenseDir, "instances.db")
	
	// Check if database file exists and create if not
	dbExists := true
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbExists = false
		Log.Info(fmt.Sprintf("Creating new database at: %s", dbPath))
	}
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	
	if !dbExists {
		Log.Info("Database file created successfully")
	}

	// Create the instances table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS instances (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		instance_name TEXT NOT NULL,
		container_name TEXT NOT NULL,
		repo_path TEXT NOT NULL,
		app_port INTEGER NOT NULL,
		postgres_port INTEGER NOT NULL,
		neo4j_bolt_port INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(instance_name, container_name)
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create instances table: %v", err)
	}

	return db, nil
}

// StoreInstanceContainers stores container names for a deployed instance
func StoreInstanceContainers(config *DeployConfig) error {
	db, err := InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Container names based on the compose override pattern
	containerNames := []string{
		fmt.Sprintf("%s-app", config.InstanceName),
		fmt.Sprintf("%s-postgres", config.InstanceName),
		fmt.Sprintf("%s-neo4j", config.InstanceName),
	}

	// Insert each container
	insertSQL := `
	INSERT OR REPLACE INTO instances 
	(instance_name, container_name, repo_path, app_port, postgres_port, neo4j_bolt_port) 
	VALUES (?, ?, ?, ?, ?, ?)`

	for _, containerName := range containerNames {
		_, err := db.Exec(insertSQL, 
			config.InstanceName, 
			containerName, 
			config.RepoPath, 
			config.AppPort, 
			config.PostgresPort, 
			config.Neo4jBoltPort,
		)
		if err != nil {
			return fmt.Errorf("failed to store container %s: %v", containerName, err)
		}
	}

	Log.Info(fmt.Sprintf("Stored %d containers for instance %s in database", len(containerNames), config.InstanceName))
	return nil
}

// GetInstanceContainers retrieves all containers for a given instance
func GetInstanceContainers(instanceName string) ([]Instance, error) {
	db, err := InitDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
	SELECT id, instance_name, container_name, repo_path, app_port, postgres_port, neo4j_bolt_port, created_at 
	FROM instances 
	WHERE instance_name = ?
	ORDER BY container_name`

	rows, err := db.Query(query, instanceName)
	if err != nil {
		return nil, fmt.Errorf("failed to query containers: %v", err)
	}
	defer rows.Close()

	var instances []Instance
	for rows.Next() {
		var instance Instance
		err := rows.Scan(
			&instance.ID,
			&instance.InstanceName,
			&instance.ContainerName,
			&instance.RepoPath,
			&instance.AppPort,
			&instance.PostgresPort,
			&instance.Neo4jBoltPort,
			&instance.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// RemoveInstanceContainers removes all containers for a given instance from the database
func RemoveInstanceContainers(instanceName string) error {
	db, err := InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	deleteSQL := `DELETE FROM instances WHERE instance_name = ?`
	result, err := db.Exec(deleteSQL, instanceName)
	if err != nil {
		return fmt.Errorf("failed to remove containers for instance %s: %v", instanceName, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	Log.Info(fmt.Sprintf("Removed %d containers for instance %s from database", rowsAffected, instanceName))
	return nil
}

// GetAllInstances retrieves all instances from the database
func GetAllInstances() ([]Instance, error) {
	db, err := InitDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
	SELECT id, instance_name, container_name, repo_path, app_port, postgres_port, neo4j_bolt_port, created_at 
	FROM instances 
	ORDER BY instance_name, container_name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all instances: %v", err)
	}
	defer rows.Close()

	var instances []Instance
	for rows.Next() {
		var instance Instance
		err := rows.Scan(
			&instance.ID,
			&instance.InstanceName,
			&instance.ContainerName,
			&instance.RepoPath,
			&instance.AppPort,
			&instance.PostgresPort,
			&instance.Neo4jBoltPort,
			&instance.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

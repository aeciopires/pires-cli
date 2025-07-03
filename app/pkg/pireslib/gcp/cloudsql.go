// Package gcp have public and private functions to connect to GCP services, like: IAM, CloudSQL, GKE, etc.
package gcp

import (
	"strings"

	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
)

// CreateGCPCloudSQLUser creates a new user in a Cloud SQL instance using gcloud command.
// host defaults to '%' if empty.
func CreateGCPCloudSQLUser(projectID, instanceID, userName, password, host string) {
	if projectID == "" || instanceID == "" || userName == "" {
		common.Logger("fatal", "projectID, instanceID and userName are required to create SQL user in CreateGCPCloudSQLUser function.")
	}
	// Password can be empty for some DB types or if managed externally (e.g., IAM DB auth)

	if host == "" {
		host = "%" // Default to allow connection from any host for gcloud
	}

	common.Logger("info", "Creating SQL user '%s' for instance '%s' on project '%s' (source-host: '%s')...", userName, instanceID, projectID, host)

	args := []string{
		"sql", "users", "create", userName,
		"--instance", instanceID,
		"--host", host,
		"--project", projectID,
		// gcloud sql users create will prompt for password if not provided and TTY is available.
		// If running non-interactively, password should be provided or IAM auth used.
	}
	if password != "" {
		args = append(args, "--password", password)
	} else {
		common.Logger("fatal", "No password provided for SQL user '%s'. `gcloud` might prompt if interactive, or creation might expect IAM authentication / no password.", userName)
	}

	_, stderr, err := RunGcloudCommand(args...)
	if err != nil {
		// Check stderr for common issues like user already exists
		if strings.Contains(stderr, "already exists") {
			common.Logger("warning", "SQL user '%s'@'%s' already exists on instance '%s' on project '%s'.", userName, host, instanceID, projectID)
		} else {
			common.Logger("fatal", "Failed to create SQL user '%s' on instance '%s' on project '%s': %w. Stderr: %s", userName, instanceID, projectID, err, stderr)
		}
	}

	common.Logger("info", "SQL user '%s'@'%s' created successfully for instance '%s' on project '%s'.", userName, host, instanceID, projectID)
}

// CreateGCPCloudSQLDatabase creates a new database in a Cloud SQL instance using gcloud command.
func CreateGCPCloudSQLDatabase(projectID, instanceID, dbName, charset, collation string) {
	if projectID == "" || instanceID == "" || dbName == "" {
		common.Logger("fatal", "projectID, instanceID, and dbName are required to create SQL database in CreateGCPCloudSQLDatabase function.")
	}

	common.Logger("info", "Creating SQL database '%s' for instance '%s' on project '%s' ...", dbName, instanceID, projectID)

	args := []string{
		"sql", "databases", "create", dbName,
		"--instance", instanceID,
		"--project", projectID,
	}
	if charset != "" {
		args = append(args, "--charset", charset)
	}
	if collation != "" {
		args = append(args, "--collation", collation)
	}

	_, stderr, err := RunGcloudCommand(args...)
	if err != nil {
		if strings.Contains(stderr, "already exists") {
			common.Logger("warning", "SQL database '%s' already exists on instance '%s' on project '%s'.", dbName, instanceID, projectID)
		} else {
			common.Logger("fatal", "Failed to create SQL database '%s' on instance '%s' on project '%s': %w. Stderr: %s", dbName, instanceID, projectID, err, stderr)
		}
	}

	common.Logger("info", "SQL database '%s' created successfully for instance '%s' on project '%s'.", dbName, instanceID, projectID)
}

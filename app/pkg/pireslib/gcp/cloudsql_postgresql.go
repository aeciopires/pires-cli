// Package gcp have public and private functions to connect to GCP services, like: IAM, CloudSQL, GKE, etc.
package gcp

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
)

// ExportPostgresUsersAndPermissions connects to a PostgreSQL Cloud SQL instance
// using the psql CLI, iterates through all databases (except those matching excludePattern or cloudsqladmin),
// and exports a detailed list of user permissions per table to a TXT file.
func ExportPostgresUsersAndPermissions(projectID, instanceID, dbHost, dbPort, dbUser, dbPassword, outputDir, excludePattern string, sslRequired bool) {
	common.Logger("info", "Exporting user permissions from instance '%s' in project '%s'\n", instanceID, projectID)

	// Compile regex if provided
	var excludeRegex *regexp.Regexp
	var err error
	if excludePattern != "" {
		excludeRegex, err = regexp.Compile(excludePattern)
		if err != nil {
			common.Logger("fatal", "Invalid exclude pattern regex '%s': %v", excludePattern, err)
		}
	}

	// Ensure output dir exists
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, config.PermissionDir); err != nil {
			common.Logger("fatal", "Failed to create output directory '%s': %v", outputDir, err)
		}
	}

	var output strings.Builder
	timestamp := time.Now().Format("20060102-150405")
	output.WriteString(fmt.Sprintf("User and Role Permissions Report for Instance: '%s' in project: '%s'. Generated is: '%s'\n\n", instanceID, projectID, timestamp))

	sslMode := "disable"
	if sslRequired {
		// if user forces sslmode, use require
		sslMode = "require"
	}

	runPSQL := func(dbName, sql string) (string, error) {
		args := []string{
			fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbPassword, dbName, sslMode),
			"-At",
			"-c", sql,
		}

		stdout, stderr, err := RunPsqlCommand(args...)
		common.Logger("debug", "Executing command: psql %s", strings.Join(args, " "))
		if err != nil {
			common.Logger("fatal", "psql command 'psql %s' failed: %w\nStderr: %s", strings.Join(args, " "), err, stderr)
		}

		if stderr != "" {
			common.Logger("fatal", "psql command stderr (exit code 0):\n%s", stderr)
		}

		return stdout, nil
	}

	// List databases
	dbListSQL := `SELECT datname FROM pg_database WHERE datistemplate = false;`
	dbListOut, err := runPSQL("postgres", dbListSQL)
	if err != nil {
		common.Logger("fatal", "Failed to list databases: %v", err)
	}

	dbNames := strings.Fields(dbListOut)

	// Iterate each database
	for _, dbName := range dbNames {
		if dbName == "cloudsqladmin" {
			common.Logger("info", "Skipping internal database 'cloudsqladmin'")
			continue
		}
		if excludeRegex != nil && excludeRegex.MatchString(dbName) {
			common.Logger("info", "Skipping database '%s' (matches exclude pattern)", dbName)
			continue
		}

		common.Logger("info", "Checking permissions in database: %s", dbName)

		output.WriteString(fmt.Sprintf("========================================\n"))
		output.WriteString(fmt.Sprintf(" DATABASE: %s\n", dbName))
		output.WriteString(fmt.Sprintf("========================================\n\n"))

		permSQL := `
SELECT 
    grantee || '|' || table_schema || '.' || table_name || '|' || privilege_type
FROM 
    information_schema.role_table_grants 
WHERE 
    grantee != 'postgres' AND grantee NOT LIKE 'pg_%' AND grantee NOT LIKE 'cloudsql%'
ORDER BY 
    grantee, table_schema, table_name;
`
		permOut, err := runPSQL(dbName, permSQL)
		if err != nil {
			output.WriteString(fmt.Sprintf("Could not query permissions in %s: %v\n\n", dbName, err))
			continue
		}

		if strings.TrimSpace(permOut) == "" {
			output.WriteString("No specific user permissions found on tables in this database.\n\n")
			continue
		}

		lines := strings.Split(strings.TrimSpace(permOut), "\n")

		currentUser := ""
		for _, line := range lines {
			parts := strings.Split(line, "|")
			if len(parts) != 3 {
				continue
			}
			grantee, table, privilege := parts[0], parts[1], parts[2]

			if grantee == "PUBLIC" {
				// Skip PUBLIC role
				continue
			}

			if grantee != currentUser {
				output.WriteString(fmt.Sprintf("  User/Role: %s\n", grantee))
				currentUser = grantee
			}
			output.WriteString(fmt.Sprintf("    - Table: %s\n", table))
			output.WriteString(fmt.Sprintf("      Permission: %s\n", privilege))
		}

		output.WriteString("\n")
	}

	// Write report to file
	fileName := fmt.Sprintf("%s_%s_database_permissions_%s.txt", projectID, instanceID, timestamp)
	filePath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(filePath, []byte(output.String()), config.PermissionFile); err != nil {
		common.Logger("fatal", "Failed to write permissions report to file '%s': %v", filePath, err)
	}

	common.Logger("info", "Successfully exported detailed database permissions to: %s\n", filePath)
}

// ExportPostgresAuditLogs fetches logs for INSERT, UPDATE, and DELETE statements
// from a Cloud SQL instance using the gcloud logging command.
// This requires the 'cloudsql.enable_pgaudit' flag to be enabled on the instance.
// More details: https://cloud.google.com/sql/docs/postgres/flags and
// https://cloud.google.com/sql/docs/postgres/pg-audit
// The logs are saved to a specified output directory with a timestamped filename.
func ExportPostgresAuditLogs(projectID, instanceID, outputDir string) {
	common.Logger("info", "Exporting audit logs for instance '%s' in project '%s'", instanceID, projectID)

	// Build the filter to get logs for DML statements.
	// This requires the 'pgaudit' flag to be configured on the Cloud SQL instance.
	// We look for statements containing the DML keywords.
	filter := fmt.Sprintf(`
resource.type="cloudsql_database"
resource.labels.database_id="%s:%s"
logName="projects/%s/logs/cloudsql.googleapis.com%%2Fpostgres.log"
(textPayload:"statement: INSERT" OR textPayload:"statement: UPDATE" OR textPayload:"statement: DELETE")
`, projectID, instanceID, projectID)

	fmt.Printf("Using log filter:\n%s\n", filter)

	// Define arguments for the gcloud command
	args := []string{
		"logging",
		"read",
		filter,
		"--project", projectID,
		"--format=value(timestamp,textPayload)",
	}

	// Run the gcloud command
	stdout, stderr, err := RunGcloudCommand(args...)
	if err != nil {
		common.Logger("fatal", "Failed to read audit logs for instance '%s' in project '%s': %w", instanceID, projectID, stderr)
	}

	if stdout == "" {
		common.Logger("fatal", "No audit logs found. Ensure the 'cloudsql.enable_pgaudit' flag is enabled on your Cloud SQL instance. More details: https://cloud.google.com/sql/docs/postgres/flags and https://cloud.google.com/sql/docs/postgres/pg-audit")
	}

	// Create the output directory if it doesn't exist
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, config.PermissionDir); err != nil {
			common.Logger("fatal", "Failed to create custom output directory '%s': %w", outputDir, err)
		}
	}

	// Generate the filename
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("%s_%s_audit_logs_%s.txt", projectID, instanceID, timestamp)
	filePath := filepath.Join(outputDir, fileName)

	// Write the output to the file
	if err := os.WriteFile(filePath, []byte(stdout), config.PermissionFile); err != nil {
		common.Logger("fatal", "Failed to write audit logs to file '%s': %w", filePath, err)
	}

	common.Logger("info", "Successfully exported audit logs to: %s\n", filePath)
}

// Package gcp have public and private functions to connect to GCP services, like: IAM, CloudSQL, GKE, etc.
package gcp

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// ExportPostgresUsersAndPermissions connects to a Cloud SQL for PostgreSQL instance,
// iterates through all databases, and exports a detailed list of user permissions
// on a per-table basis to a .txt file.
func ExportPostgresUsersAndPermissions(projectID, instanceID, user, password, outputDir string) {
	common.Logger("info", "Exporting user permissions from instance '%s' in project '%s'\n", instanceID, projectID)

	ctx := context.Background()

	// Use the Cloud SQL Go Connector to securely connect to the database.
	d, err := cloudsqlconn.NewDialer(ctx)
	if err != nil {
		common.Logger("fatal", "cloudsqlconn.NewDialer: %w", err)
	}
	defer d.Close()

	// Function to create a database connection pool for a specific database
	getDB := func(dbName string) (*sql.DB, error) {
		// Custom dialer function to connect to Cloud SQL with IAM authentication
		// and public IP. This is necessary for connecting to Cloud SQL instances.
		// If you want to use private IP, you can remove the WithPublicIP()
		// option and ensure your environment is set up for private IP access.
		// Note: WithIAMAuthN() is used for IAM authentication, which requires
		// the Cloud SQL Admin API to be enabled and the user to have the
		// appropriate IAM roles (e.g., Cloud SQL Client).
		// If you want to use a service account, you can use WithServiceAccount()
		// instead of WithIAMAuthN().
		customDialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.Dial(ctx, fmt.Sprintf("%s:%s", projectID, instanceID), cloudsqlconn.WithPublicIP())
		}

		// parse pgx config
		dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbName)
		pgxConfig, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			common.Logger("fatal", "pgxpool.ParseConfig: %v", err)
		}

		// override DialFunc with Cloud SQL Dialer
		pgxConfig.ConnConfig.DialFunc = customDialer

		// open database
		db := stdlib.OpenDB(*pgxConfig.ConnConfig)

		return db, nil
	}

	// Connect to the default 'postgres' database to get a list of all databases
	db, err := getDB("postgres")
	if err != nil {
		common.Logger("fatal", "Failed to connect to 'postgres' db to list databases: %w", err)
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false;")
	if err != nil {
		common.Logger("fatal", "Failed to query for database list: %w", err)
	}
	defer rows.Close()

	var dbNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			common.Logger("fatal", "Failed to scan database name: %w", err)
		}
		dbNames = append(dbNames, name)
	}
	rows.Close()
	db.Close() // Close connection to 'postgres' db

	var output strings.Builder
	output.WriteString(fmt.Sprintf("User and Role Permissions Report for Instance: '%s'\n\n", instanceID))

	// Iterate through each database to get permissions
	for _, dbName := range dbNames {
		common.Logger("info", "Checking permissions in database: %s\n", dbName)
		output.WriteString(fmt.Sprintf("========================================\n") +
			fmt.Sprintf(" DATABASE: %s\n", dbName) +
			fmt.Sprintf("========================================\n\n"))

		db, err := getDB(dbName)
		if err != nil {
			output.WriteString(fmt.Sprintf("Could not connect to database %s: %v\n\n", dbName, err))
			continue // Skip to the next database
		}
		defer db.Close()

		// Query for table-level grants for all roles
		query := `
SELECT 
    grantee, 
    table_schema, 
    table_name, 
    privilege_type 
FROM 
    information_schema.role_table_grants 
WHERE 
    grantee != 'postgres' AND grantee NOT LIKE 'pg_%' AND grantee NOT LIKE 'cloudsql%'
ORDER BY 
    grantee, table_schema, table_name;
`
		permRows, err := db.QueryContext(ctx, query)
		if err != nil {
			output.WriteString(fmt.Sprintf("Could not query permissions in %s: %v\n\n", dbName, err))
			db.Close()
			continue
		}
		defer permRows.Close()

		permissions := make(map[string]map[string][]string) // user -> table -> [perms]
		for permRows.Next() {
			var grantee, tableSchema, tableName, privilegeType string
			if err := permRows.Scan(&grantee, &tableSchema, &tableName, &privilegeType); err != nil {
				common.Logger("warning", "Failed to scan permission row in %s: %v\n", dbName, err)
				continue
			}
			fullTableName := fmt.Sprintf("%s.%s", tableSchema, tableName)
			if permissions[grantee] == nil {
				permissions[grantee] = make(map[string][]string)
			}
			permissions[grantee][fullTableName] = append(permissions[grantee][fullTableName], privilegeType)
		}

		if len(permissions) == 0 {
			output.WriteString("No specific user permissions found on tables in this database.\n\n")
		} else {
			for user, tables := range permissions {
				output.WriteString(fmt.Sprintf("  User/Role: %s\n", user))
				for table, perms := range tables {
					output.WriteString(fmt.Sprintf("    - Table: %s\n", table))
					output.WriteString(fmt.Sprintf("      Permissions: %s\n", strings.Join(perms, ", ")))
				}
				output.WriteString("\n")
			}
		}
		db.Close()
	}

	// Create the output directory if it doesn't exist
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, config.PermissionDir); err != nil {
			common.Logger("fatal", "Failed to create custom output directory '%s': %w", outputDir, err)
		}
	}

	// Generate the filename
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("%s_%s_database_permissions_%s.txt", projectID, instanceID, timestamp)
	filePath := filepath.Join(outputDir, fileName)

	// Write the output to the file
	if err := os.WriteFile(filePath, []byte(output.String()), config.PermissionFile); err != nil {
		common.Logger("fatal", "Failed to write permissions report to file '%s': %w", filePath, err)
	}

	common.Logger("fatal", "Successfully exported detailed database permissions to: %s\n", filePath)
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

package cmd

import (
	"reflect"
	"syscall"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
	"github.com/aeciopires/pires-cli/pkg/pireslib/gcp"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Local variables
var (
	cloudsqlInstanceID    string
	cloudsqlUserName      string
	cloudsqlPassword      string // Consider prompting for password
	cloudsqlHost          string
	cloudsqlAddress       string
	cloudsqlPort          string
	cloudsqlDBName        string
	cloudsqlDBCharset     string
	cloudsqlDBCollation   string
	cloudsqlDBIgnoreRegex string
	cloudsqlSSLRequired   bool
	outputReportDir       string

	// cloudsqlCmd represents the cloudsql command
	cloudsqlCmd = &cobra.Command{
		Use:   "cloudsql",
		Short: "Manage Cloud SQL instances, users, and databases",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {

			// Debug message is displayed if -D option was passed
			common.Logger("debug", "====> Values loaded in cmd/gcp-cloudsql subcommand")
			auxValue := reflect.ValueOf(config.Properties)
			auxType := reflect.TypeOf(config.Properties)

			// Interate over the fields of the struct
			for i := 0; i < auxValue.NumField(); i++ {
				fieldName := auxType.Field(i).Name
				fieldValue := auxValue.Field(i).Interface()
				common.Logger("debug", "Field: %s, Value: %v", fieldName, fieldValue)
			}

			// GCP Admin Permissions Check
			common.Logger("debug", "Performing admin permission checks as requested...")
			gcp.CheckGcloudAdminPermissions(config.Properties.DefaultGCPProject)
		},
	}

	// --- Create User Subcommand ---
	cloudsqlCreateUserCmd = &cobra.Command{
		Use:   "create-user",
		Short: "Create a new user in a Cloud SQL instance",
		RunE: func(cmd *cobra.Command, args []string) error {

			gcp.CreateGCPCloudSQLUser(config.Properties.DefaultGCPProject, cloudsqlInstanceID, cloudsqlUserName, cloudsqlPassword, cloudsqlHost)
			return nil
		},
	}

	// --- Create Database Subcommand ---
	cloudsqlCreateDatabaseCmd = &cobra.Command{
		Use:   "create-database",
		Short: "Create a new database in a Cloud SQL instance",
		RunE: func(cmd *cobra.Command, args []string) error {

			gcp.CreateGCPCloudSQLDatabase(config.Properties.DefaultGCPProject, cloudsqlInstanceID, cloudsqlDBName, cloudsqlDBCharset, cloudsqlDBCollation)
			return nil
		},
	}

	// --- Export PostgreSQL Users Permissions Subcommand ---
	exportPostgreSQLUsersPermissionsCmd = &cobra.Command{
		Use:   "export-postgresql-users-permissions",
		Short: "Exports PostgreSQL users and permissions from a Cloud SQL instance.",
		Long: `Connects to a specified PostgreSQL database within a Cloud SQL instance and
	exports a list of all roles (users), their attributes, and memberships to a .txt file.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Prompt for password if not provided via flag for better security
			if cloudsqlPassword == "" {
				common.Logger("info", "Enter password for user '%s': ", cloudsqlUserName)

				// ReadPassword takes a file descriptor (int) as input.
				// syscall.Stdin represents the standard input file descriptor.
				bytePassword, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					common.Logger("fatal", "Error reading password: %v", err)
				}

				// Convert the byte slice to a string for use.
				cloudsqlPassword = string(bytePassword)
			}

			gcp.ExportPostgresUsersAndPermissions(config.Properties.DefaultGCPProject, cloudsqlInstanceID, cloudsqlAddress, cloudsqlPort, cloudsqlUserName, cloudsqlPassword, outputReportDir, cloudsqlDBIgnoreRegex, cloudsqlSSLRequired)
		},
	}

	// --- Export PostgreSQL Audit Logs Subcommand ---
	exportPostgreSQLAuditLogsCmd = &cobra.Command{
		Use:   "export-postgresql-audit-logs",
		Short: "Exports DML audit logs (INSERT, UPDATE, DELETE) from a Cloud SQL instance.",
		Long: `Fetches logs from Google Cloud Logging for a specific Cloud SQL instance,
	filtering for INSERT, UPDATE, and DELETE statements. This requires the 'cloudsql.enable_pgaudit'
	database flag to be enabled on the instance. More details: https://cloud.google.com/sql/docs/postgres/flags and
	https://cloud.google.com/sql/docs/postgres/pg-audit`,
		Run: func(cmd *cobra.Command, args []string) {
			gcp.ExportPostgresAuditLogs(config.Properties.DefaultGCPProject, cloudsqlInstanceID, outputReportDir)
		},
	}
)

func init() {
	gcpCmd.AddCommand(cloudsqlCmd) // Add cloudsql to parent gcp command

	// Add subcommands to cloudsqlCmd
	cloudsqlCmd.AddCommand(cloudsqlCreateUserCmd)
	cloudsqlCmd.AddCommand(cloudsqlCreateDatabaseCmd)
	cloudsqlCmd.AddCommand(exportPostgreSQLUsersPermissionsCmd)
	cloudsqlCmd.AddCommand(exportPostgreSQLAuditLogsCmd)

	// Flags for 'cloudsql create-user'
	cloudsqlCreateUserCmd.Flags().StringVarP(&cloudsqlInstanceID, "instance", "i", "", "Cloud SQL instance ID (e.g. nonprod-psql) (required)")
	cloudsqlCreateUserCmd.Flags().StringVarP(&cloudsqlUserName, "username", "u", "", "Username for the new SQL user (e.g. app-name) (required)")
	cloudsqlCreateUserCmd.Flags().StringVarP(&cloudsqlPassword, "password", "p", "", "Password for the new SQL user (prompt if not provided, or use IAM auth) (e.g. changeme) (required)")
	cloudsqlCreateUserCmd.Flags().StringVarP(&cloudsqlHost, "source-host", "s", "%", "Host from which the user can connect (e.g., '%', 'localhost', '1.2.3.4') (optional)")

	// Flags are required
	_ = cloudsqlCreateUserCmd.MarkFlagRequired("instance")
	_ = cloudsqlCreateUserCmd.MarkFlagRequired("username")
	_ = cloudsqlCreateUserCmd.MarkFlagRequired("password")

	// Flags for 'cloudsql create-database'
	cloudsqlCreateDatabaseCmd.Flags().StringVarP(&cloudsqlInstanceID, "instance", "i", "", "Cloud SQL instance ID (e.g. nonprod-psql) (required)")
	cloudsqlCreateDatabaseCmd.Flags().StringVarP(&cloudsqlDBName, "dbname", "d", "", "Name for the new database (e.g. app-name-db) (required)")
	cloudsqlCreateDatabaseCmd.Flags().StringVarP(&cloudsqlDBCharset, "charset", "c", "UTF8", "Character set for the new database (e.g., utf8, UTF8) (optional)")
	cloudsqlCreateDatabaseCmd.Flags().StringVarP(&cloudsqlDBCollation, "collation", "l", "en_US.UTF8", "Collation for the new database (e.g., en_US.UTF8) (optional)")

	// Flags are required
	_ = cloudsqlCreateDatabaseCmd.MarkFlagRequired("instance")
	_ = cloudsqlCreateDatabaseCmd.MarkFlagRequired("dbname")

	// Flags for 'cloudsql export-postgresql-users-permissions'
	exportPostgreSQLUsersPermissionsCmd.Flags().StringVarP(&cloudsqlInstanceID, "instance", "i", "", "Cloud SQL instance ID (e.g. nonprod-psql) (required)")
	exportPostgreSQLUsersPermissionsCmd.Flags().StringVarP(&cloudsqlUserName, "username", "u", "", "Username for the new SQL user (e.g. app-name) (required)")
	exportPostgreSQLUsersPermissionsCmd.Flags().StringVarP(&cloudsqlPassword, "password", "p", "", "Password for the new SQL user (prompt if not provided, or use IAM auth) (e.g. changeme) (required)")
	exportPostgreSQLUsersPermissionsCmd.Flags().StringVarP(&outputReportDir, "output-dir", "o", "", "Custom output directory for the permissions report (default is current directory)")
	exportPostgreSQLUsersPermissionsCmd.Flags().StringVarP(&cloudsqlPort, "port", "t", "5432", "Port for the PostgreSQL instance (e.g 5432)")
	exportPostgreSQLUsersPermissionsCmd.Flags().StringVarP(&cloudsqlAddress, "address", "a", "mydb.example.com", "Address (IP or DNS) of the PostgreSQL instance (e.g. 'mydb.example.com')")
	exportPostgreSQLUsersPermissionsCmd.Flags().StringVarP(&cloudsqlDBIgnoreRegex, "regex-ignore-databases", "r", "^prisma_migrate", "Regular expression to ignore specific databases (e.g. '^prisma_migrate')")
	exportPostgreSQLUsersPermissionsCmd.Flags().BoolVarP(&cloudsqlSSLRequired, "ssl-required", "s", false, "Force SSL connection to the PostgreSQL instance (default is false)")

	// Flags are required
	_ = exportPostgreSQLUsersPermissionsCmd.MarkFlagRequired("instance")
	_ = exportPostgreSQLUsersPermissionsCmd.MarkFlagRequired("username")
	_ = exportPostgreSQLUsersPermissionsCmd.MarkFlagRequired("address")

	// Flags for 'cloudsql export-postgresql-audit-logs'
	exportPostgreSQLAuditLogsCmd.Flags().StringVarP(&cloudsqlInstanceID, "instance", "i", "", "Cloud SQL instance ID (e.g. nonprod-psql) (required)")
	exportPostgreSQLAuditLogsCmd.Flags().StringVarP(&outputReportDir, "output-dir", "o", "", "Custom output directory for the audit logs (default is current directory)")

	// Flags are required
	_ = exportPostgreSQLAuditLogsCmd.MarkFlagRequired("instance")

}

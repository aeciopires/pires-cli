package cmd

import (
	"reflect"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
	"github.com/aeciopires/pires-cli/pkg/pireslib/gcp"
	"github.com/spf13/cobra"
)

// Local variables
var (
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
	cloudsqlInstanceID string
	cloudsqlUserName   string
	cloudsqlPassword   string // Consider prompting for password
	cloudsqlHost       string

	cloudsqlCreateUserCmd = &cobra.Command{
		Use:   "create-user",
		Short: "Create a new user in a Cloud SQL instance",
		RunE: func(cmd *cobra.Command, args []string) error {

			gcp.CreateGCPCloudSQLUser(config.Properties.DefaultGCPProject, cloudsqlInstanceID, cloudsqlUserName, cloudsqlPassword, cloudsqlHost)
			return nil
		},
	}

	// --- Create Database Subcommand ---
	cloudsqlDBName      string
	cloudsqlDBCharset   string
	cloudsqlDBCollation string

	cloudsqlCreateDatabaseCmd = &cobra.Command{
		Use:   "create-database",
		Short: "Create a new database in a Cloud SQL instance",
		RunE: func(cmd *cobra.Command, args []string) error {

			gcp.CreateGCPCloudSQLDatabase(config.Properties.DefaultGCPProject, cloudsqlInstanceID, cloudsqlDBName, cloudsqlDBCharset, cloudsqlDBCollation)
			return nil
		},
	}
)

func init() {
	gcpCmd.AddCommand(cloudsqlCmd) // Add cloudsql to parent gcp command

	// Add subcommands to cloudsqlCmd
	cloudsqlCmd.AddCommand(cloudsqlCreateUserCmd)
	cloudsqlCmd.AddCommand(cloudsqlCreateDatabaseCmd)

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
}

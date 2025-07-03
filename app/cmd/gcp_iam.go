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
	// iamCmd represents the iam command
	iamCmd = &cobra.Command{
		Use:   "iam",
		Short: "Manage GCP IAM resources (service accounts, roles, permissions)",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// This runs before any iam subcommand

			// Debug message is displayed if -D option was passed
			common.Logger("debug", "====> Values loaded in cmd/gcp-iam subcommand")
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

	iamCreateSaAccountID   string
	iamCreateSaDescription string

	// --- Create Service Account Subcommand ---
	iamCreateSaCmd = &cobra.Command{
		Use:   "create-sa",
		Short: "Create a new service account",
		RunE: func(cmd *cobra.Command, args []string) error {

			gcp.CreateGCPIAMServiceAccount(config.Properties.DefaultGCPProject, iamCreateSaAccountID, iamCreateSaDescription)
			return nil
		},
	}

	// --- Grant Role Subcommand ---
	iamGrantRoleMember string
	iamGrantRoleName   string

	iamGrantRoleCmd = &cobra.Command{
		Use:   "grant-role",
		Short: "Grant an IAM role to a member on the project",
		Long: `Grants a specified IAM role to a member.
	Member format:
	  - user:{emailid} (e.g., user:name.surname@company.com)
	  - serviceAccount:{emailid} (e.g., serviceAccount:app-name-gsa@change-project.iam.gserviceaccount.com)
	  - group:{emailid} (e.g., group:admins@company.com)
	  - domain:{domain} (e.g., domain:company.com)
	Role format:
	  - roles/{SERVICE_NAME}.{ROLE_NAME} (e.g., roles/storage.objectViewer)
	  - projects/{PROJECT_ID}/roles/{CUSTOM_ROLE_ID} for custom roles`,
		RunE: func(cmd *cobra.Command, args []string) error {

			gcp.GrantGCPIAMRoleToMember(config.Properties.DefaultGCPProject, iamGrantRoleMember, iamGrantRoleName)
			return nil
		},
	}
)

func init() {
	gcpCmd.AddCommand(iamCmd) // Add iam to parent gcp command

	// Add subcommands to iamCmd
	iamCmd.AddCommand(iamCreateSaCmd)
	iamCmd.AddCommand(iamGrantRoleCmd)

	// Flags for 'iam create-sa'
	iamCreateSaCmd.Flags().StringVarP(&iamCreateSaAccountID, "service-account-id", "s", "", "Unique ID for the new service account (e.g., app-name-gsa) (required)")
	iamCreateSaCmd.Flags().StringVarP(&iamCreateSaDescription, "sa-description", "d", "", "Description for the service account (optional)")

	// Flags are required
	_ = iamCreateSaCmd.MarkFlagRequired("service-account-id")

	// Flags for 'iam grant-role'
	iamGrantRoleCmd.Flags().StringVarP(&iamGrantRoleMember, "member", "m", "", "Member to grant the role to (e.g., user:name.surname@company.com, serviceAccount:app-name-gsa@change-project.iam.gserviceaccount.com) (required)")
	iamGrantRoleCmd.Flags().StringVarP(&iamGrantRoleName, "role", "r", "roles/cloudsql.editor", "IAM role to grant (e.g., roles/storage.admin) (required)")

	// Flags are required
	_ = iamGrantRoleCmd.MarkFlagRequired("member")
	_ = iamGrantRoleCmd.MarkFlagRequired("role")

}

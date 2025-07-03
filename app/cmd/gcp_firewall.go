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
	// firewallCmd represents the firewall command
	firewallCmd = &cobra.Command{
		Use:   "firewall",
		Short: "Manage GCP Firewall rules",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// This runs before any firewall subcommand

			// Debug message is displayed if -D option was passed
			common.Logger("debug", "====> Values loaded in cmd/gcp-firewall subcommand")
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

	outputDir string

	// --- Export fireall rules Subcommand ---
	exportFirewallRulesCmd = &cobra.Command{
		Use:   "export-rules",
		Short: "Export GCP firewall rules",
		RunE: func(cmd *cobra.Command, args []string) error {

			if config.GCPFirewallRulesOutputType != "csv" {
				common.Logger("fatal", "Unsupported output type '%s'. Only 'csv' is supported.", config.GCPFirewallRulesOutputType)
			} else {
				gcp.ExportGCPFirewallRulesToCSV(config.Properties.DefaultGCPProject, outputDir)
			}
			return nil
		},
	}
)

func init() {
	gcpCmd.AddCommand(firewallCmd) // Add firewall to parent gcp command

	// Add subcommands to firewallCmd
	firewallCmd.AddCommand(exportFirewallRulesCmd)

	// Flags for 'firewall export-rules'
	exportFirewallRulesCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Custom output directory for the CSV file (default is current directory)")
	exportFirewallRulesCmd.Flags().StringVarP(&config.GCPFirewallRulesOutputType, "output-type", "t", config.GCPFirewallRulesOutputType, "Output type for file rules")

	// Flags are required
	_ = exportFirewallRulesCmd.MarkFlagRequired("output-dir")

}

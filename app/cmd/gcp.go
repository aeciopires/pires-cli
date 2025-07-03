package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Local variables
var (
	// gcpCmd represents the base gcp command
	gcpCmd = &cobra.Command{
		Use:   "gcp",
		Short: "Perform Google Cloud Platform operations",
		Long:  `Provides commands to interact with GCP services like Cloud SQL, IAM, etc.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// This runs before any gcp subcommand

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("GCP command requires a subcommand (e.g., cloudsql, iam).")
			cmd.Help()
		},
	}
)

func init() {
	rootCmd.AddCommand(gcpCmd) // Add gcpCmd to the root command

}

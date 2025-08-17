package cmd

import (
	"fmt"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/internal/update"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
	"github.com/spf13/cobra"
)

// Local variables
var (
	githubRepo string = "aeciopires/pires-cli"

	// updateCmd represents the update command
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Check for a new version and update the application.",
		Long: `Checks for the latest release on GitHub. If a newer version is found
for your operating system and architecture, it downloads and replaces the
current application binary.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.Logger("info", "Checking for updates...")

			release := update.CheckForUpdate(githubRepo)

			if release == nil {
				common.Logger("warning", "You are already on the latest version: %s\n", config.CLIVersion)
				return
			}

			common.Logger("info", "A new version is available: %s. Do you want to update? (y/n): ", release.TagName)
			var response string
			fmt.Scanln(&response)

			if response != "y" && response != "Y" {
				common.Logger("fatal", "Update cancelled.")
			}

			common.Logger("info", "Updating to version: %s", release.TagName)
			update.ApplyUpdate(release)

			common.Logger("info", "Update complete! Please run the CLI again.")
		},
	}
)

func init() {
	rootCmd.AddCommand(updateCmd) // Add update to parent root command
	// Add flags to the update command if needed
}

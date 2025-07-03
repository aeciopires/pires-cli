// Package gcp have public and private functions to connect to GCP services, like: IAM, CloudSQL, GKE, etc.
package gcp

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
)

// ExportGCPFirewallRulesToCSV exports all firewall rules from a given GCP project to a CSV file.
// The filename includes the project ID and a timestamp.
// The file can be saved to a custom directory.
func ExportGCPFirewallRulesToCSV(projectID, outputDir string) error {
	common.Logger("debug", "====> Exporting firewall rules for GCP project: %s", projectID)

	// Define arguments for the gcloud command
	args := []string{
		"compute",
		"firewall-rules",
		"list",
		"--project",
		projectID,
		"--format=csv(name,network,direction,priority,sourceRanges.list():label=SOURCE_RANGES,destinationRanges.list():label=DESTINATION_RANGES,allowed.list():label=ALLOWED,denied.list():label=DENIED,sourceTags.list():label=SOURCE_TAGS,targetTags.list():label=TARGET_TAGS,disabled)",
	}

	// Run the gcloud command
	stdout, stderr, err := RunGcloudCommand(args...)
	if err != nil {
		common.Logger("fatal", "Failed to export firewall rules for project '%s'... Stdout: %s, Stderr: %s", projectID, stdout, stderr)
	}

	if stdout == "" {
		common.Logger("warning", "gcloud command returned no firewall rules for project '%s'. The output file will be empty.", projectID)
	}

	// Create the output directory if it doesn't exist
	if outputDir != "" {
		if errMkdir := os.MkdirAll(outputDir, config.PermissionDir); errMkdir != nil {
			common.Logger("fatal", "Failed to create custom output directory '%s': %s", outputDir, errMkdir)
		}
	}

	// Generate the filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("%s-%s-%s.csv", config.GCPFirewallRulesPrefix, projectID, timestamp)
	// If outputDir is "", it joins to the current dir
	filePath := filepath.Join(outputDir, fileName)

	// Write the CSV output to the file
	errWrite := os.WriteFile(filePath, []byte(stdout), config.PermissionFile)
	if errWrite != nil {
		common.Logger("fatal", "Failed to write firewall rules to file '%s': %s", filePath, errWrite)
	}

	common.Logger("info", "Successfully exported firewall rules for project '%s' to: %s", projectID, filePath)

	// Return nil if everything went well
	return nil
}

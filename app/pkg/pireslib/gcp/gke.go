// Package gcp have public and private functions to connect to GCP services, like: IAM, CloudSQL, GKE, etc.
package gcp

import (
	"strings"

	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
)

// ConnectToGKECluster uses gcloud command to configure kubectl to connect to the specified GKE cluster.
func ConnectToGKECluster(projectID, location, clusterName string) {
	if projectID == "" || location == "" || clusterName == "" {
		common.Logger("fatal", "projectID, location (region/zone), and clusterName are required to connect to GKE cluster")
	}

	common.Logger("info", "Attempting to configure kubectl for GKE cluster '%s' in region/zone '%s' (project: '%s')...", clusterName, location, projectID)

	args := []string{
		"container", "clusters", "get-credentials", clusterName,
		"--project", projectID,
	}

	// Add --zone or --region based on whether location contains '-' (typical for zones)
	if strings.Contains(location, "-") && (strings.Count(location, "-") == 2) { // Heuristic for zone, e.g., us-central1-a
		args = append(args, "--zone", location)
	} else {
		args = append(args, "--region", location)
	}

	stdout, stderr, err := RunGcloudCommand(args...)
	if err != nil {
		common.Logger("fatal", "Failed to get GKE cluster credentials for '%s' in  region/zone '%s' (project: '%s')... Stdout: %s, Stderr: %s", clusterName, location, projectID, stdout, stderr)
	}

	// gcloud get-credentials output usually includes "Fetching cluster endpoint and auth data."
	// and "kubeconfig entry generated for <cluster_name>."
	if stdout != "" {
		common.Logger("debug", "gcloud get-credentials stdout: %s", stdout)
	}
	common.Logger("info", "Successfully configured kubectl for GKE cluster '%s' in region/zone '%s' (project: '%s')...", clusterName, location, projectID)
}

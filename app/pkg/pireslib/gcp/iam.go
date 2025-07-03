// Package gcp have public and private functions to connect to GCP services, like: IAM, CloudSQL, GKE, etc.
package gcp

import (
	"fmt"
	"strings"

	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
)

// CreateGCPIAMServiceAccount creates a new service account in the specified project using gcloud command.
func CreateGCPIAMServiceAccount(projectID, accountID, description string) {
	if projectID == "" || accountID == "" {
		common.Logger("fatal", "projectID and accountID are required to create a service account on CreateGCPIAMServiceAccount function")
	}

	common.Logger("info", "Creating service account '%s' in project '%s'...", accountID, projectID)

	args := []string{
		"iam", "service-accounts", "create", accountID,
		"--display-name", accountID,
		"--project", projectID,
	}
	if description != "" {
		args = append(args, "--description", description)
	}

	// gcloud iam service-accounts create prints the email of the created SA to stdout on success,
	// or an error to stderr.
	_, stderr, err := RunGcloudCommand(args...)
	if err != nil {
		// Check if SA already exists
		if strings.Contains(stderr, "already exists") {
			saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountID, projectID)
			common.Logger("warning", "Service account '%s' already exists.", saEmail)
		} else {
			common.Logger("fatal", "Failed to create service account '%s' on project '%s': %w. Stderr: %s", accountID, projectID, err, stderr)
		}
	}

	// Expected output on success: "Created service account [sa-id]."
	// And for email, we can construct it or try to parse stdout if gcloud changes its output.
	// For now, constructing it is safer.
	createdSAEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountID, projectID)
	common.Logger("info", "Service account '%s' created successfully. Email: %s on project '%s'.", accountID, createdSAEmail, projectID)
}

// GrantGCPIAMRoleToMember grants a specific IAM role to a member on a project using gcloud command.
// Member format: "user:email@example.com", "serviceAccount:sa-email@project.iam.gserviceaccount.com", etc.
// Role format: "roles/rolename" (e.g., "roles/storage.objectViewer")
func GrantGCPIAMRoleToMember(projectID, member, role string) {
	if projectID == "" || member == "" || role == "" {
		common.Logger("fatal", "projectID, member, and role are required to grant IAM role on GrantGCPIAMRoleToMember function")
	}

	common.Logger("info", "Granting role '%s' to member '%s' on project '%s'...", role, member, projectID)

	args := []string{
		"projects", "add-iam-policy-binding", projectID,
		"--member", member,
		"--role", role,
		"--condition=None", // Explicitly set no condition for simplicity, can be parameterized
		"--project", projectID,
	}

	// `add-iam-policy-binding` is idempotent. If the binding already exists, it won't error.
	_, stderr, err := RunGcloudCommand(args...)
	if err != nil {
		// Check stderr for specific permission denied errors for the operation itself
		if strings.Contains(stderr, "PERMISSION_DENIED") && strings.Contains(stderr, "resourcemanager.projects.setIamPolicy") {
			common.Logger("fatal", "Permission denied to set IAM policy for project '%s': %w. Stderr: %s", projectID, err, stderr)
		}
		common.Logger("fatal", "Failed to grant role '%s' to member '%s' on project '%s': %w. Stderr: %s", role, member, projectID, err, stderr)
	}

	common.Logger("info", "Successfully granted (or ensured) role '%s' to member '%s' on project '%s'.", role, member, projectID)
}

// Package gcp have public and private functions to connect to GCP services, like: IAM, CloudSQL, GKE, etc.
package gcp

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
)

// RunGcloudCommand executes a gcloud command with the given arguments.
// It captures and returns stdout and stderr.
// Assumes gcloud is in the system PATH.
func RunGcloudCommand(args ...string) (stdout string, stderr string, err error) {
	// Proceed with running the command
	cmd := exec.Command("gcloud", args...)

	// Buffers to capture stdout and stderr
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	common.Logger("debug", "Executing command: gcloud %s", strings.Join(args, " "))
	err = cmd.Run()

	stdout = outb.String()
	stderr = errb.String()

	if err != nil {
		return stdout, stderr, fmt.Errorf("gcloud command 'gcloud %s' failed: %w\nStderr: %s", strings.Join(args, " "), err, stderr)
	}

	if stderr != "" {
		common.Logger("info", "gcloud command stderr (exit code 0):\n%s", stderr)
	}

	return stdout, stderr, nil
}

// RunPsqlCommand executes a psql command with the given arguments.
// It captures and returns stdout and stderr.
// Assumes psql is in the system PATH.
func RunPsqlCommand(args ...string) (stdout string, stderr string, err error) {
	// Proceed with running the command
	cmd := exec.Command("psql", args...)

	// Buffers to capture stdout and stderr
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	common.Logger("debug", "Executing command: psql %s", strings.Join(args, " "))
	err = cmd.Run()

	stdout = outb.String()
	stderr = errb.String()

	if err != nil {
		return stdout, stderr, fmt.Errorf("psql command 'psql %s' failed: %w\nStderr: %s", strings.Join(args, " "), err, stderr)
	}

	if stderr != "" {
		common.Logger("info", "psql command stderr (exit code 0):\n%s", stderr)
	}

	return stdout, stderr, nil
}

// CheckGcloudAuth verifies if gcloud is authenticated by checking the active account.
func CheckGcloudAuth() string {
	common.Logger("debug", "Checking gcloud authentication status...")
	// `gcloud auth list --filter=status:ACTIVE --format="value(account)"` is more robust
	// but `gcloud config get-value account` is simpler for a basic check.
	stdout, stderr, err := RunGcloudCommand("config", "get-value", "account")
	activeAccount := strings.TrimSpace(stdout)
	if err != nil || activeAccount == "" {
		common.Logger("error", "Failed to check gcloud auth status. Ensure gcloud is installed and authenticated using account: %w. Stderr: %s", err, stderr)
		common.Logger("fatal", "Please run 'gcloud auth login' \n 'gcloud auth application-default login' commands.")
	}

	common.Logger("debug", "gcloud is authenticated with account: %s", activeAccount)
	return activeAccount
}

// CheckGcloudAdminPermissions verifies if the current gcloud credentials have a set of administrative permissions on the project.
// This function uses `gcloud projects test-iam-permissions`.
func CheckGcloudAdminPermissions(projectID string) {
	if projectID == "" {
		common.Logger("fatal", "Project ID is required to check admin permissions in CheckGcloudAdminPermissions function.")
	}
	common.Logger("debug", "Checking if current gcloud user has '%s' on project '%s'...", config.GCPRequiredRole, projectID)

	// Get the currently authenticated gcloud account email
	activeAccount := CheckGcloudAuth()
	memberIdentifier := "user:" + activeAccount
	common.Logger("debug", "Checking '%s' for member: %s", config.GCPRequiredRole, memberIdentifier)

	// Command to check if the member has the 'roles/owner' role
	// gcloud projects get-iam-policy <PROJECT_ID> \
	//   --flatten="bindings[].members" \
	//   --filter="bindings.role:roles/owner AND bindings.members:<MEMBER_IDENTIFIER>" \
	//   --format="value(bindings.role)"
	// If the user has the role, this command will output "roles/owner". Otherwise, it will be empty.
	args := []string{
		"projects", "get-iam-policy", projectID,
		"--flatten=bindings[].members",
		fmt.Sprintf("--filter=bindings.role:%s AND bindings.members:%s", config.GCPRequiredRole, memberIdentifier),
		"--format=value(bindings.role)",
	}

	stdout, stderrCmd, errCmd := RunGcloudCommand(args...)
	if errCmd != nil {
		// This error means the `gcloud projects get-iam-policy` command itself failed.
		// This could be due to the project not existing, or the user not having
		// even 'resourcemanager.projects.getIamPolicy' permission.
		common.Logger("fatal", "Execution of 'gcloud projects get-iam-policy' command for project '%s' failed. \nReview stderr output from gcloud for details. \nStdout: %w . \nStderr from gcloud: %s", projectID, errCmd, stderrCmd)
	}

	// Check the output
	outputRole := strings.TrimSpace(stdout)
	if outputRole != config.GCPRequiredRole {
		common.Logger("fatal", "Current gcloud user ('%s') does NOT have '%s' on project '%s'. Insufficient permissions for administrative tasks.", activeAccount, config.GCPRequiredRole, projectID)
	}

	common.Logger("debug", "Current gcloud user ('%s') has '%s' on project '%s'. Administrative permissions check passed.", activeAccount, config.GCPRequiredRole, projectID)
}

// Package update have public and private functions to updatete the application.
package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
)

// Package-level variables.

// GitHubReleaseAsset represents an asset in a GitHub release.
type GitHubReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

// GitHubRelease represents a GitHub release.
type GitHubRelease struct {
	TagName string               `json:"tag_name"`
	Assets  []GitHubReleaseAsset `json:"assets"`
}

// CheckForUpdate checks for a new version of the application on GitHub.
// It returns the release info if an update is available, otherwise nil.
func CheckForUpdate(repo string) *GitHubRelease {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	common.Logger("debug", "Checking for updates at: %s", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		common.Logger("fatal", "Failed to fetch latest release from GitHub %s: %w", apiURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		common.Logger("fatal", "Failed to get latest release from %s: GitHub API returned status %s", apiURL, resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		common.Logger("fatal", "Failed to parse GitHub release JSON: %w", err)
	}

	latestVersion := release.TagName
	currentVersion := config.CLIVersion

	common.Logger("info", "Current version: %s, Latest version on GitHub: %s", currentVersion, latestVersion)

	if currentVersion != latestVersion {
		return &release
	}

	return nil // No update available
}

// ApplyUpdate downloads and applies a new binary from a GitHub release.
func ApplyUpdate(release *GitHubRelease) {
	// Determine the asset name based on OS and architecture
	assetName := fmt.Sprintf("%s-%s-%s", config.CLIName, runtime.GOOS, runtime.GOARCH)
	common.Logger("debug", "Looking for asset: %s", assetName)

	var binaryAsset *GitHubReleaseAsset
	var checksumsAsset *GitHubReleaseAsset

	for i, asset := range release.Assets {
		if asset.Name == assetName {
			binaryAsset = &release.Assets[i]
		}
		if asset.Name == "checksums.txt" {
			checksumsAsset = &release.Assets[i]
		}
	}

	if binaryAsset == nil {
		common.Logger("fatal", "Could not find a release asset for your platform (%s/%s)", runtime.GOOS, runtime.GOARCH)
	}
	if checksumsAsset == nil {
		common.Logger("fatal", "Could not find checksums.txt in the release assets")
	}

	common.Logger("info", "Downloading checksums from %s...", checksumsAsset.DownloadURL)
	checksums, err := DownloadFile(checksumsAsset.DownloadURL)
	if err != nil {
		common.Logger("fatal", "Failed to download checksums: %w", err)
	}

	// Download the new binary to a temporary file
	common.Logger("info", "Downloading new version from %s...", binaryAsset.DownloadURL)
	newBinaryBytes, err := DownloadFile(binaryAsset.DownloadURL)
	if err != nil {
		common.Logger("fatal", "Failed to download new binary: %w", err)
	}

	// Verify the checksum
	expectedChecksum, err := ParseChecksum(string(checksums), assetName)
	if err != nil {
		common.Logger("fatal", "Failed to find checksum for asset %s: %w", assetName, err)
	}

	actualChecksum := sha256.Sum256(newBinaryBytes)
	actualChecksumStr := hex.EncodeToString(actualChecksum[:])

	if actualChecksumStr != expectedChecksum {
		common.Logger("fatal", "Checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksumStr)
	}
	common.Logger("info", "Checksum verified successfully.")

	// Replace the current executable
	executablePath, err := os.Executable()
	if err != nil {
		common.Logger("fatal", "Could not determine executable path: %w", err)
	}

	// Create a temporary file with the new binary content
	tmpFile, err := os.CreateTemp(filepath.Dir(executablePath), "update-*.tmp")
	if err != nil {
		common.Logger("fatal", "Could not create temporary file for update: %w", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(newBinaryBytes); err != nil {
		common.Logger("fatal", "Failed to write new binary to temporary file: %w", err)
	}
	tmpFile.Close() // Close the file so we can rename it

	// Set executable permissions on the new binary
	if err := os.Chmod(tmpFile.Name(), config.PermissionBinary); err != nil {
		common.Logger("fatal", "Failed to set executable permission on new binary: %w", err)
	}

	// Rename the old binary
	oldPath := executablePath + ".old"
	if err := os.Rename(executablePath, oldPath); err != nil {
		common.Logger("fatal", "Failed to rename old binary: %w", err)
	}

	// Move the new binary into place
	if err := os.Rename(tmpFile.Name(), executablePath); err != nil {
		// Attempt to restore the old binary if the final rename fails
		os.Rename(oldPath, executablePath)
		common.Logger("fatal", "Failed to move new binary into place: %w", err)
	}

	common.Logger("info", "Update successful! The old binary is at %s. It can be removed manually.", oldPath)
}

// DownloadFile is a helper to download a file from a URL.
func DownloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// ParseChecksum finds the checksum for a specific file from the checksums.txt content.
func ParseChecksum(checksumsContent, fileName string) (string, error) {
	lines := strings.Split(checksumsContent, "\n")
	for _, line := range lines {
		// Format is "checksum  filename"
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == fileName {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("checksum for %s not found", fileName)
}

// Package fileeditor have public and private functions to edit files
package fileeditor

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
	"gopkg.in/yaml.v3"
)

// ATTENTION!!!
// The go embed directive statement must be outside of function body
// embed must be used in the same package where the files are needed
// You cannot use '..' in the path to access files in parent directories.
// This limitation is by design for security and to avoid ambiguity.

// Embed the 'internalembeds' directory.
// This directory should be structured as follows:
// internalembeds/
// |-- yq (the yq executable)
//
//go:embed all:internalembeds
var internalFS embed.FS

// Package-level variables.
var (
	foundYqPath string    // Stores the path to the extracted yq executable
	findYqOnce  sync.Once // Ensures yq extraction runs only once
	err         error
	expression  string // yq expression, reused by various functions
)

// SearchForYq extracts the embedded yq executable to a temporary file
// and makes it executable. This function is run once by GetYqPath.
func SearchForYq() {
	foundYqPath = "" // Ensure path is empty
	common.Logger("debug", "Preparing embedded yq executable from internalFS")

	// Path to yq within the embedded FS
	embeddedYqPath := "internalembeds/yq"
	yqEmbeddedBytes, errCmd := internalFS.ReadFile(embeddedYqPath)
	if errCmd != nil {
		common.Logger("fatal", "Failed to read embedded yq binary from '%s': %v", embeddedYqPath, errCmd)
	}

	if len(yqEmbeddedBytes) == 0 {
		common.Logger("fatal", "Embedded yq binary '%s' is empty.", embeddedYqPath)
	}

	tmpFile, errCreate := os.CreateTemp("", "yq-*")
	if errCreate != nil {
		common.Logger("fatal", "Failed to create temporary file for yq: %v", errCreate)
	}
	// Defer close here to ensure it's closed even if subsequent steps fail before explicit close.
	// Store name before potential close if needed, though tmpFile.Name() is fine until remove.
	tempFilePath := tmpFile.Name()

	if _, errWrite := tmpFile.Write(yqEmbeddedBytes); errWrite != nil {
		tmpFile.Close()         // Close before removing
		os.Remove(tempFilePath) // Clean up
		common.Logger("fatal", "Failed to write embedded yq to temporary file '%s': %v", tempFilePath, errWrite)
	}

	// Close the file before changing permissions, especially on Windows.
	if errClose := tmpFile.Close(); errClose != nil {
		common.Logger("fatal", "Failed to close temporary yq file '%s' before chmod: %v", tempFilePath, errClose)
	}

	// Make it executable
	if errChmod := os.Chmod(tempFilePath, config.PermissionBinary); errChmod != nil {
		os.Remove(tempFilePath) // Clean up
		common.Logger("fatal", "Failed to make temporary yq file '%s' executable: %v", tempFilePath, errChmod)
	}

	common.Logger("debug", "Embedded yq executable prepared at: %s", tempFilePath)
	foundYqPath = tempFilePath
	// Note: The temporary file persists for the application's lifetime or until OS cleanup.
}

// GetYqPath returns the path to the (potentially extracted) yq executable.
// The extraction logic (SearchForYq) is run only once.
func GetYqPath() string {
	findYqOnce.Do(func() {
		SearchForYq()
		// Package-level 'err' is set by SearchForYq if an error occurs.
		// If 'err' is not nil here, 'foundYqPath' will likely be empty.
	})
	return foundYqPath
}

// RunYqCommand executes the yq command with the given arguments.
// It uses the yq executable obtained from GetYqPath.
func RunYqCommand(args ...string) (string, error) {
	// Get the validated path to yq (search runs only once)
	execPath := GetYqPath()
	if execPath == "" {
		return "", errors.New("[ERROR] yq executable path is not set or yq preparation failed. Review logs from SearchForYq function")
	}

	// Proceed with running the command
	cmd := exec.Command(execPath, args...)

	// Buffers to capture stdout and stderr
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	// Run the command
	runCmdErr := cmd.Run()

	stdout := outb.String()
	stderr := errb.String()
	combinedOutput := stdout + stderr // Combine for context in case of error

	if runCmdErr != nil {
		exitCode := -1
		if exitError, ok := runCmdErr.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		return combinedOutput, fmt.Errorf("[ERROR] yq command failed (exit code %d): %w\nStderr: %s", exitCode, runCmdErr, stderr)
	}

	// Check if yq wrote anything to stderr, even if exit code is 0 (might indicate warnings)
	if stderr != "" {
		common.Logger("warning", "yq command stderr (exit code 0):\n%s\n", stderr)
	}
	return strings.TrimSpace(stdout), nil // Return trimmed stdout on success
}

// GetYamlValue reads a value from a YAML file using a yq expression.
// Example expression: ".spec.replicas" or ".metadata.name"
//
// Reference:
//
// https://mikefarah.gitbook.io/yq
func GetYamlValue(filePath string, expression string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("[ERROR] File path cannot be empty")
	}
	if expression == "" {
		return "", fmt.Errorf("[ERROR] yq expression cannot be empty")
	}

	// Arguments for yq: eval '<expression>' <filePath
	args := []string{"eval", expression, filePath}
	output, cmdErr := RunYqCommand(args...)
	if cmdErr != nil {
		return "", fmt.Errorf("[ERROR] Failed to get value from '%s' using expression '%s': %w", filePath, expression, cmdErr)
	}
	return output, nil
}

// ModifyYamlInPlace modifies a YAML file in-place using a full yq expression.
// If the target file or its directory structure does not exist, they will be created before modification.
// The caller is responsible for providing a valid yq expression string.
// This allows for setting values, deleting keys, merging objects/arrays, etc.
//
// Examples of expressions:
//   - Set string value:   `.metadata.name = "new-name"`
//   - Set numeric value:  `.spec.replicas = 3`
//   - Set boolean value:  `.spec.enabled = true`
//   - Delete a key:       `del(.metadata.annotations)`
//   - Merge an object:    `.metadata.labels += {"new-label": "value"}`
//   - Append to array:    `.spec.containers += [{"name": "sidecar", "image": "sidecar:latest"}]`
//   - Complex selection:  `select(.kind == "Deployment") | .spec.template.spec.serviceAccountName = "new-sa"`
//
// filePath: The path to the YAML file to modify.
// fullExpression: The complete yq expression string to evaluate.
//
// Reference:
//
// https://mikefarah.gitbook.io/yq
func ModifyYamlInPlace(filePath string, fullExpression string) error {
	if filePath == "" {
		return fmt.Errorf("[ERROR] File path cannot be empty")
	}

	// We still check for empty expression as it's likely an error,
	// although yq might handle it gracefully (e.g., as a no-op).
	if fullExpression == "" {
		return fmt.Errorf("[ERROR] yq expression cannot be empty")
	}

	// --- Ensure directory exists ---
	dirPath := filepath.Dir(filePath)
	// Check if directory exists. os.Stat returns an error if path doesn't exist.
	// We only care about creating it if it specifically *doesn't* exist.
	if _, statErr := os.Stat(dirPath); errors.Is(statErr, os.ErrNotExist) {
		common.Logger("debug", "Directory '%s' not found, creating it.", dirPath)
		// Create the directory and any necessary parents.
		if mkdirErr := os.MkdirAll(dirPath, config.PermissionDir); mkdirErr != nil {
			return fmt.Errorf("[ERROR] failed to create directory '%s': %w", dirPath, mkdirErr)
		}
	} else if statErr != nil {
		// Another error occurred trying to Stat the directory (e.g., permissions)
		return fmt.Errorf("[ERROR] Failed to check status of directory '%s': %w", dirPath, statErr)
	}
	// --- Directory exists or was just created ---

	// --- Check if file exists, create if not ---
	// This check remains necessary even after creating the directory.
	if _, statErr := os.Stat(filePath); errors.Is(statErr, os.ErrNotExist) {
		// File does not exist, attempt to create it
		common.Logger("debug", "File '%s' not found, creating it.", filePath)
		// Create an empty file. os.Create truncates existing files, but we already know it doesn't exist.
		fileHandle, createErr := os.Create(filePath)
		if createErr != nil {
			return fmt.Errorf("[ERROR] Failed to create non-existent file '%s': %w", filePath, createErr)
		}
		// Close the file immediately after creation
		if closeErr := fileHandle.Close(); closeErr != nil {
			// Log warning, but proceed as the file likely exists now.
			common.Logger("warning", "Failed to close newly created file handle for '%s': %v\n", filePath, closeErr)
		}
	} else if statErr != nil {
		// Another error occurred during Stat (e.g., permission denied)
		return fmt.Errorf("[ERROR] Failed to check status of file '%s': %w", filePath, statErr)
	}
	// --- File exists or was just created ---

	// Arguments for yq: eval -i '<fullExpression>' <filePath>
	// The '-i' flag modifies the file in-place.
	args := []string{"eval", "-i", fullExpression, filePath}
	// Run the command. Output might contain errors/warnings from yq.
	output, runErr := RunYqCommand(args...)
	if runErr != nil {
		// Include the expression and any yq output in the error message for context
		return fmt.Errorf("[ERROR] Failed to modify file '%s' using expression '%s': %w\nOutput:\n%s", filePath, fullExpression, runErr, output)
	}
	// If yq exits successfully (runErr == nil), the modification is assumed complete.
	return nil
}

// HasAnySuffix check short and long suffix like this: foo.bar.baz.tar.gz and foo.bar
func HasAnySuffix(name string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// ApplyYqExpressionRecursively applies a yq expression in-place to all YAML files
// under the given directory and its subdirectories.
// It uses the RunYqCommand helper to execute the yq command with proper logging and error handling.
func ApplyYqExpressionRecursively(rootDir string, expressionToApply string) error {
	if rootDir == "" {
		return fmt.Errorf("[ERROR] Root directory path cannot be empty")
	}
	if expressionToApply == "" {
		return fmt.Errorf("[ERROR] yq expression cannot be empty")
	}

	// Traverse the directory tree and apply the expression to each .yaml/.yml file
	return filepath.WalkDir(rootDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("[ERROR] Unable to access path '%s': %w", path, walkErr)
		}
		// Skip directories
		if d.IsDir() {
			return nil
		}
		// Skip non-YAML files
		if !HasAnySuffix(path, ".yaml", ".yml", ".patch.yaml", ".patch.yml") {
			common.Logger("debug", "Skipping non-YAML file: %s", path)
			return nil
		}

		// Construct the in-place edit command: yq eval -i '<expression>' <filePath>
		args := []string{"eval", "-i", expressionToApply, path}
		// Run yq with custom wrapper to capture output and errors
		output, cmdErr := RunYqCommand(args...)
		if cmdErr != nil {
			return fmt.Errorf("[ERROR] Failed to apply yq to '%s': %w\nOutput:\n%s", path, cmdErr, output)
		}
		common.Logger("debug", "Successfully applied yq expression to: %s\n", path)
		return nil
	})
}

// CopyTemplateFiles copies files from an embedded source directory to a destination on disk.
// embeddedSourceDirRelToInternalEmbeds is the path within 'internalFS' relative to its root 'internalembeds',
// e.g., "templates/common".
func CopyTemplateFiles(embeddedSourceDirRelToInternalEmbeds string, destDir string) error {
	// Construct the full path within the embed.FS (e.g., "internalembeds/templates/common")
	fullEmbedSourcePath := path.Join("internalembeds", embeddedSourceDirRelToInternalEmbeds)

	if _, statErr := os.Stat(destDir); os.IsNotExist(statErr) {
		if mkdirErr := os.MkdirAll(destDir, config.PermissionDir); mkdirErr != nil {
			return fmt.Errorf("[ERROR] Failed to create destination directory %s: %w", destDir, mkdirErr)
		}
	}

	// Walk files from source directory
	return fs.WalkDir(internalFS, fullEmbedSourcePath, func(embedPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("[ERROR] Error accessing embedded path %s: %w", embedPath, walkErr)
		}

		// Calculate path relative to the source directory being copied (e.g., "templates/common")
		// embedPath is like "internalembeds/templates/common/file.yaml"
		// fullEmbedSourcePath is like "internalembeds/templates/common"
		// relPath should be like "file.yaml" or "subdir/file.yaml"
		relPath, errRel := filepath.Rel(fullEmbedSourcePath, embedPath)
		if errRel != nil {
			return fmt.Errorf("[ERROR] Error calculating relative path for %s from %s: %w", embedPath, fullEmbedSourcePath, errRel)
		}

		// Skip the root directory itself, only copy its contents
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		// Ensures that the destination directory exists
		if d.IsDir() {
			// Create corresponding directory in destination
			if errMkdir := os.MkdirAll(destPath, config.PermissionDir); errMkdir != nil {
				return fmt.Errorf("[ERROR] Error creating destination directory %s: %w", destPath, errMkdir)
			}
			return nil
		}

		// It's a file, copy it
		fileData, errRead := internalFS.ReadFile(embedPath)
		if errRead != nil {
			return fmt.Errorf("[ERROR] Error reading embedded file %s: %w", embedPath, errRead)
		}

		// Write to destination file
		if errWrite := os.WriteFile(destPath, fileData, config.PermissionFile); errWrite != nil {
			return fmt.Errorf("[ERROR] Error writing file to %s: %w", destPath, errWrite)
		}
		return nil
	})
}

// CopyFile copies a single file from source to destination.
func CopyFile(srcFile, destFile string) error {
	src, openErr := os.Open(srcFile)
	if openErr != nil {
		return openErr
	}
	defer src.Close()

	dest, createErr := os.Create(destFile)
	if createErr != nil {
		return createErr
	}
	defer dest.Close()

	_, copyErr := io.Copy(dest, src)
	return copyErr
}

// MergeYAMLFiles merges two YAML files and returns the result as a YAML string.
// No changes needed.
func MergeYAMLFiles(filePath1, filePath2 string) (string, error) {
	yamlData1, errRead1 := os.ReadFile(filePath1)
	if errRead1 != nil {
		return "", fmt.Errorf("[ERROR] Could not read file %s: %w", filePath1, errRead1)
	}
	yamlData2, errRead2 := os.ReadFile(filePath2)
	if errRead2 != nil {
		return "", fmt.Errorf("[ERROR] Could not read file %s: %w", filePath2, errRead2)
	}

	var rootNode1, rootNode2 yaml.Node
	if errUnmarshal1 := yaml.Unmarshal(yamlData1, &rootNode1); errUnmarshal1 != nil {
		return "", fmt.Errorf("[ERROR] Failed to parse YAML from %s: %w", filePath1, errUnmarshal1)
	}
	if errUnmarshal2 := yaml.Unmarshal(yamlData2, &rootNode2); errUnmarshal2 != nil {
		return "", fmt.Errorf("[ERROR] Failed to parse YAML from %s: %w", filePath2, errUnmarshal2)
	}

	// Merge the root nodes
	mergedNode, errMerge := MergeRootDocumentNodes(&rootNode1, &rootNode2)
	if errMerge != nil {
		return "", fmt.Errorf("[ERROR] Failed to merge YAML nodes: %w", errMerge)
	}

	var buffer bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buffer)
	yamlEncoder.SetIndent(2)
	if errEncode := yamlEncoder.Encode(mergedNode); errEncode != nil {
		return "", fmt.Errorf("[ERROR] Failed to encode merged YAML: %w", errEncode)
	}
	yamlEncoder.Close()

	return buffer.String(), nil
}

// MergeRootDocumentNodes merges two YAML DocumentNodes and returns the resulting merged node.
func MergeRootDocumentNodes(docNode1, docNode2 *yaml.Node) (*yaml.Node, error) {
	if docNode1.Kind != yaml.DocumentNode || docNode2.Kind != yaml.DocumentNode {
		return nil, fmt.Errorf("[ERROR] Expected both nodes to be DocumentNode")
	}

	map1 := ConvertMappingNodeToMap(docNode1.Content[0])
	map2 := ConvertMappingNodeToMap(docNode2.Content[0])
	mergedMappingNode := MergeMappingPreservingKeyOrder(map1, map2)

	return &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{mergedMappingNode},
	}, nil
}

// ConvertMappingNodeToMap builds a map from a YAML mapping node.
func ConvertMappingNodeToMap(mappingNode *yaml.Node) map[string]*yaml.Node {
	result := make(map[string]*yaml.Node)
	if mappingNode.Kind != yaml.MappingNode {
		return result
	}

	// Iterate through the content of a YAML to convert a YAML mappingNode
	// into a more convenient Go map[string]*yaml.Node. This makes it easier
	// to work with the key-value pairs in the YAML data
	for i := 0; i < len(mappingNode.Content); i += 2 {
		// Ensure there's a corresponding value for the key
		if i+1 >= len(mappingNode.Content) {
			common.Logger("warning", "Odd number of elements in MappingNode.Content. Skipping last key.")
			break // Exit the loop if there's no value
		}
		key := mappingNode.Content[i].Value
		value := mappingNode.Content[i+1]
		// Store the key-value pair in the map
		result[key] = value
	}
	return result
}

// MergeMappingPreservingKeyOrder merges two YAML maps preserving a specific key order.
func MergeMappingPreservingKeyOrder(primaryMap, secondaryMap map[string]*yaml.Node) *yaml.Node {
	preferredKeyOrder := config.K8sYamlManifestsPreferredKeyOrder
	mergedNode := &yaml.Node{Kind: yaml.MappingNode}
	seenKeys := map[string]bool{}

	// Defines an anonymous function (also known as a closure).
	// This function is designed to add a key-value pair to a YAML mapping node (mergedNode)
	// while also keeping track of which keys have already been added (seenKeys).
	// This function is used to build up the mergedNode by iterating through the preferredKeyOrder
	// and then any remaining keys. The seenKeys map ensures that keys are not added multiple times.
	addKeyValue := func(key string, value *yaml.Node) {
		mergedNode.Content = append(mergedNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
			value,
		)
		seenKeys[key] = true
	}

	// This loop is to iterate through the preferredKeyOrder and, for each key,
	// determine how to handle it based on its presence in the two input maps.
	// It prioritizes keys in the primaryMap but also incorporates keys from the secondaryMap
	// if they are not present in the primaryMap. If a key exists in both maps,
	// it merges their values using the MergeValuesForKey function.
	for _, key := range preferredKeyOrder {
		// Check if the key exists in the primary map
		if node1, exists := primaryMap[key]; exists {
			// Check if the key also exists in the secondary map
			if node2, exists2 := secondaryMap[key]; exists2 {
				// Key exists in both maps, merge the values
				mergedValue := MergeValuesForKey(key, node1, node2)
				addKeyValue(key, mergedValue)
			} else {
				// Key exists only in the primary map, use its value
				addKeyValue(key, node1)
			}
		} else if node2, exists := secondaryMap[key]; exists {
			// Key exists only in the primary map, use its value
			addKeyValue(key, node2)
		}
		// If the key is not in either map, it's skipped in this loop.
		// It might be added later if it's in the remaining keys.
	}

	// Add any additional keys not listed in preferredKeyOrder
	for key, node := range primaryMap {
		if !seenKeys[key] {
			addKeyValue(key, node)
		}
	}
	for key, node := range secondaryMap {
		if !seenKeys[key] {
			addKeyValue(key, node)
		}
	}
	return mergedNode
}

// MergeValuesForKey merges values based on their type, especially handling arrays.
func MergeValuesForKey(key string, value1, value2 *yaml.Node) *yaml.Node {
	if value1.Kind == yaml.SequenceNode && value2.Kind == yaml.SequenceNode {
		return MergeArraysUniquely(value1, value2)
	}
	// By default, override with value2
	return value2
}

// MergeArraysUniquely merges two YAML arrays avoiding duplicate entries (by serialized representation).
func MergeArraysUniquely(array1, array2 *yaml.Node) *yaml.Node {
	mergedArray := &yaml.Node{Kind: yaml.SequenceNode}
	seenSerializedValues := map[string]bool{}

	// This code defines an anonymous function. Its primary purpose is to add
	// a YAML node (item) to a merged array (mergedArray) only if an equivalent
	// node hasn't already been added. This is done by serializing the YAML node
	// to a string and checking if that string representation has been seen before.
	appendIfNotSeen := func(item *yaml.Node) {
		var buffer bytes.Buffer
		encoder := yaml.NewEncoder(&buffer)
		if encErr := encoder.Encode(item); encErr != nil { // Use local encErr
			common.Logger("error", "Failed to encode YAML node: %v", encErr)
			return
		}
		// String representation of the YAML node
		serialized := buffer.String()
		if !seenSerializedValues[serialized] {
			seenSerializedValues[serialized] = true
			mergedArray.Content = append(mergedArray.Content, item)
		}
	}

	// Merge two YAML arrays (array1 and array2) while ensuring that there are
	// no duplicate entries in the resulting merged array.
	// Iterate through the first array and add each item to the merged array if it's not already present.
	for _, item := range array1.Content {
		appendIfNotSeen(item)
	}
	// Iterate through the second array and add each item to the merged array if it's not already present.
	for _, item := range array2.Content {
		appendIfNotSeen(item)
	}

	return mergedArray
}

// CopyAndMergeYAMLDir copies files from an embedded source to a target directory.
// If a YAML file exists at the destination, it's merged with the embedded version.
// embeddedSourceDirRelToInternalEmbeds is path like "templates/common".
func CopyAndMergeYAMLDir(embeddedSourceDirRelToInternalEmbeds string, targetDir string) error {
	fullEmbedSourcePath := path.Join("internalembeds", embeddedSourceDirRelToInternalEmbeds)

	return fs.WalkDir(internalFS, fullEmbedSourcePath, func(embedPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("[ERROR] Failed to access embedded path %s: %w", embedPath, walkErr)
		}

		relPath, errRel := filepath.Rel(fullEmbedSourcePath, embedPath)
		if errRel != nil {
			return fmt.Errorf("[ERROR] Failed to compute relative path for %s from %s: %w", embedPath, fullEmbedSourcePath, errRel)
		}

		if relPath == "." { // Skip the root source directory itself
			return nil
		}
		destPath := filepath.Join(targetDir, relPath)

		// If it's a directory, create it if needed
		if d.IsDir() {
			return os.MkdirAll(destPath, config.PermissionDir)
		}

		// Ensure parent directory for the file exists before writing/merging
		if errMkdir := os.MkdirAll(filepath.Dir(destPath), config.PermissionDir); errMkdir != nil {
			return fmt.Errorf("[ERROR] Error creating directory for %s: %w", destPath, errMkdir)
		}

		// Handle only files from here
		if IsYAMLFile(embedPath) && FileExists(destPath) {
			common.Logger("debug", "YAML file exists at destination %s, attempting merge with embedded %s.", destPath, embedPath)

			embeddedFileData, errRead := internalFS.ReadFile(embedPath)
			if errRead != nil {
				return fmt.Errorf("[ERROR] Failed to read embedded YAML file %s for merging: %w", embedPath, errRead)
			}

			tmpEmbedFile, errTmp := os.CreateTemp("", "embed-*.yaml")
			if errTmp != nil {
				return fmt.Errorf("[ERROR] Failed to create temporary file for embedded YAML %s: %w", embedPath, errTmp)
			}
			// Defer cleanup of the temporary file
			defer func() {
				tmpEmbedFile.Close()           // Ensure it's closed
				os.Remove(tmpEmbedFile.Name()) // Then remove
			}()

			if _, errWriteTmp := tmpEmbedFile.Write(embeddedFileData); errWriteTmp != nil {
				return fmt.Errorf("[ERROR] Failed to write embedded YAML %s to temporary file %s: %w", embedPath, tmpEmbedFile.Name(), errWriteTmp)
			}

			// IMPORTANT: Close the file before MergeYAMLFiles attempts to read it.
			if errClose := tmpEmbedFile.Close(); errClose != nil {
				// If close fails, still attempt removal via defer, but log error.
				common.Logger("warning", "Failed to close temporary file %s before merge: %v", tmpEmbedFile.Name(), errClose)
				// Depending on OS, MergeYAMLFiles might fail if file not properly closed.
			}

			merged, errMerge := MergeYAMLFiles(destPath, tmpEmbedFile.Name()) // destPath is existing, tmpEmbedFile.Name() is new from embed
			if errMerge != nil {
				return fmt.Errorf("[ERROR] Failed to merge %s and embedded %s (from temp %s): %w", destPath, embedPath, tmpEmbedFile.Name(), errMerge)
			}
			errWrite := os.WriteFile(destPath, []byte(merged), config.PermissionFile)
			if errWrite != nil {
				return fmt.Errorf("[ERROR] Failed to write merged YAML to %s: %w", destPath, errWrite)
			}
			common.Logger("debug", "Merged YAML file: %s with embedded %s. Final content written to %s.", destPath, embedPath, destPath)
			return nil
		}

		// Standard copy for non-YAML files or if destination YAML doesn't exist
		fileData, errRead := internalFS.ReadFile(embedPath)
		if errRead != nil {
			return fmt.Errorf("[ERROR] Error reading embedded file %s: %w", embedPath, errRead)
		}
		if errWrite := os.WriteFile(destPath, fileData, config.PermissionFile); errWrite != nil {
			return fmt.Errorf("[ERROR] Error writing file to %s: %w", destPath, errWrite)
		}
		common.Logger("debug", "Copied embedded file %s to %s", embedPath, destPath)
		return nil
	})
}

// IsYAMLFile checks if the filename has a YAML extension (.yaml or .yml), excluding patch files.
func IsYAMLFile(filename string) bool {
	// Conditional used to avoid merge *.patch.yaml file
	// Skip non-YAML files and *.patch.yaml and *.patch.yml files
	if HasAnySuffix(filename, ".patch.yaml", ".patch.yml") {
		common.Logger("debug", "Skipping *.patch.yaml or *.patch.yml file: %s", filename)
		return false
	}
	if HasAnySuffix(filename, ".yaml", ".yml") {
		return true
	}
	common.Logger("debug", "Skipping non-YAML file: %s", filename)
	return false
}

// FileExists checks if a file exists and is not a directory.
func FileExists(path string) bool {
	info, errStat := os.Stat(path)
	return errStat == nil && !info.IsDir()
}

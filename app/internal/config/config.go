// Package config set global variables and constants
package config

import (
	"os"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

// PropertiesStruct is a struct defined in the global context of the CLI
// to group all the properties that can be used/changed in different contexts
// and that can have custom values ​​according to the arguments of each subcommand
type PropertiesStruct struct {
	// For more details about tag validation: https://github.com/go-playground/validator
	// File/Directory Paths (Basic check: required)
	// Attention!!! The validator do not support ˜, $HOME or file globbing in values.
	DefaultConfigFile  string `mapstructure:"cli_config_file" validate:"omitempty"`
	DefaultEnvironment string `mapstructure:"cli_environment" validate:"required,lowercase,oneof=dev staging production"`
	// GCP Settings (Basic checks)
	// 'alphanum' allows only letters and numbers. Might need a custom validator
	// for hyphens if project IDs can contain them (e.g., register a custom 'alphanumhyphen').
	DefaultGCPProject         string `mapstructure:"cli_gcp_project" validate:"required,lowercase"`
	DefaultGCPRegion          string `mapstructure:"cli_gcp_region" validate:"required,lowercase"`
	DefaultDatabaseType       string `mapstructure:"cli_database_type" validate:"required,lowercase,oneof=postgresql mongodb none"`
	DefaultVPNAddressTarget   string `mapstructure:"cli_vpn_host_target" validate:"required,lowercase,noUnderscore,http_url"`
	DefaultGSABaseAccountName string `mapstructure:"cli_gsa_base_account" validate:"required,lowercase,noUnderscore,max=30"`
	DefaultGSAAccountName     string `mapstructure:"cli_gsa_account" validate:"required,lowercase"`
}

// Global variables
var (
	// Version is set during build time
	// Given a version number MAJOR.MINOR.PATCH, increment the:
	// MAJOR version when you make incompatible changes, like: API, arguments or big code refactory
	// MINOR version when you add functionality in a backward compatible manner
	// PATCH version when you make backward compatible bug fixes
	// Reference: https://semver.org/
	CLIVersion = "0.2.0"
	CLIName    = "pires-cli"

	CommandsToCheck = []string{"git", "kubectl", "gcloud"}

	// Properties is a global variable of PropertiesStruct type
	Properties PropertiesStruct

	// Log configurations
	Debug *bool

	//----------------------------
	// Kubernetes configurations
	//----------------------------
	// Define the preferred order for Kubernetes manifest keys
	K8sYamlManifestsPreferredKeyOrder = []string{
		"apiVersion", "kind", "metadata", "namespace", "spec", "resources", "images", "patches",
	}

	//----------------------------
	// Linux/Unix configurations
	//----------------------------
	// 0755 => 0 -> selects attributes for the set user ID
	//         7 -> (U)ser/owner can read, can write and can execute.
	//         5 -> (G)roup can read, can't write and can execute.
	//         5 -> (O)thers can read, can't write and can
	// References:
	// https://chmodcommand.com/chmod-0755/
	// https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permissions
	PermissionDir    os.FileMode = 0755
	PermissionBinary os.FileMode = 0755

	// 0644 => 0 -> selects attributes for the set user ID
	//         6 -> (U)ser/owner can read, can write and can't execute.
	//         4 -> (G)roup can read, can't write and can't execute.
	//         4 -> (O)thers can read, can't write and can't execute.
	// References:
	// https://chmodcommand.com/chmod-0644/
	// https://stackoverflow.com/questions/14249467/os-mkdir-and-os-mkdirall-permissions
	PermissionFile os.FileMode = 0644

	//----------------------------
	// GCP/gcloud configurations
	//----------------------------
	// Role required by perform the actions on GCP
	GCPRequiredRole string = "roles/owner"
	// Default output type for firewall rules export
	GCPFirewallRulesOutputType string = "csv"
	GCPFirewallRulesPrefix     string = "gcp-firewall-rules"

	//----------------------------
	// VPN configurations
	//----------------------------
	VPNCheckConnection bool
	VPNTimeout         time.Duration = 15
)

// Config set default values to Properties variable
func Config() {
	Properties.DefaultConfigFile = ".env"
	// Attention!!! The validator do not support ˜, $HOME or file globbing in values.
	Properties.DefaultEnvironment = "dev"
	Properties.DefaultGCPProject = "change-here"
	Properties.DefaultGCPRegion = "change-here"
	Properties.DefaultDatabaseType = "none"
	Properties.DefaultVPNAddressTarget = "http://change-here.com"
	Properties.DefaultGSABaseAccountName = "change-here-gsa"
	Properties.DefaultGSAAccountName = Properties.DefaultGSABaseAccountName + "@" + Properties.DefaultGCPProject + ".iam.gserviceaccount.com"
}

// NoUnderscores is a custom validator to reject string with underscore '_'
func NoUnderscores(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString(`_`, fl.Field().String())
	return !matched
}

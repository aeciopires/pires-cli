package cmd

import (
	"errors" // Required for errors.As
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/internal/getinfo"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Local variables
var (
	longVersion  *bool
	shortVersion *bool

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "pires-cli",
		Short: "My CLI, developed in Golang, to perform Ops activities",
		Long:  `My CLI, developed in Golang, to perform Ops activities... See: https://github.com/aeciopires/pires-cli`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		Run: func(cmd *cobra.Command, args []string) {
			// If the user ran the command without providing any arguments and without setting any flags.
			// If both of those conditions are met, it assumes the user needs help and displays the command's help text.
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				cmd.Help()
				return
			}
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	// Show longVersion. *longVersion contains the pointer address. If the content is true print longVersion, system and arch
	if *longVersion {
		getinfo.PrintLongVersion()
		getinfo.ShowOperatingSystem()
		getinfo.ShowSystemArch()
	}

	// Show shortVersion. *shortVersion contains the pointer address. If the content is true print shortVersion, system and arch
	if *shortVersion {
		getinfo.PrintShortVersion()
	}

	// Debug message is displayed if -D option was passed
	common.Logger("debug", "====> Values loaded in cmd/root.go")
	auxValue := reflect.ValueOf(config.Properties)
	auxType := reflect.TypeOf(config.Properties)

	// Interate over the fields of the struct
	for i := 0; i < auxValue.NumField(); i++ {
		fieldName := auxType.Field(i).Name
		fieldValue := auxValue.Field(i).Interface()
		common.Logger("debug", "Field: %s, Value: %v", fieldName, fieldValue)
	}
}

func init() {
	config.Config()
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&config.Properties.DefaultConfigFile, "config-file", "C", config.Properties.DefaultConfigFile, "config file path")
	rootCmd.PersistentFlags().StringVarP(&config.Properties.DefaultEnvironment, "environment", "E", config.Properties.DefaultEnvironment, "Name of environment. Supported values: dev, staging or production")
	rootCmd.PersistentFlags().StringVarP(&config.Properties.DefaultGCPProject, "gcp-project", "P", config.Properties.DefaultGCPProject, "GCP name project.")
	rootCmd.PersistentFlags().StringVarP(&config.Properties.DefaultGCPRegion, "gcp-region", "R", config.Properties.DefaultGCPRegion, "GCP region.")
	rootCmd.PersistentFlags().StringVarP(&config.Properties.DefaultDatabaseType, "database-type", "T", config.Properties.DefaultDatabaseType, "Database type. Supported values: postgresql or mongodb or none")
	rootCmd.PersistentFlags().StringVarP(&config.Properties.DefaultVPNAddressTarget, "vpn-address-target", "I", config.Properties.DefaultVPNAddressTarget, "Address for VPN connectivity check. Required if --vpn-check-connection is true. Must be a valid URL (http or https).")
	rootCmd.PersistentFlags().BoolVarP(&config.VPNCheckConnection, "vpn-check-connection", "J", false, "VPN check or not connection. If true, it will check the VPN connection using the --vpn-address-target flag.")

	config.Debug = rootCmd.PersistentFlags().BoolP("debug", "D", false, "Enable debug mode.")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	longVersion = rootCmd.Flags().BoolP("long-version", "V", false, "Show long version")
	shortVersion = rootCmd.Flags().BoolP("version", "v", false, "Show short version")

	// Flags are required
	//rootCmd.MarkPersistentFlagRequired(
	//	"app-name",
	//	"vpn-address-target",
	//)

	// Flags must be provided together
	rootCmd.MarkFlagsRequiredTogether(
		//"config-file",
		"environment",
		"gcp-project",
		"gcp-region",
	)

}

// initConfig reads in config file and ENV variables if set.
// This function is performaded in cmd/root.go and cmd/subcommand.go
func initConfig() {
	// Environment variables expect with prefix CLI_ . This helps avoid conflicts.
	viper.SetEnvPrefix("cli")
	// Type file
	viper.SetConfigType("env")
	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --gcp-region to CLI_GCP_REGION
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	// Attempt to read the SPECIFIC config file (passed by default value or -c option)
	common.Logger("debug", "Attempting to read specific config file: %s", config.Properties.DefaultConfigFile)
	// Tell Viper the exact file path
	viper.SetConfigFile(config.Properties.DefaultConfigFile)
	// Attempt to read the specific file
	err := viper.ReadInConfig()
	// Handle outcome of reading the specific file
	if err == nil {
		// SUCCESS reading specific file
		common.Logger("debug", "Using config file: %v", viper.ConfigFileUsed())
	} else {
		// FAILURE reading specific file - Log details and attempt fallback
		common.Logger("error", "Could not read specific config file '%s': %v\n", viper.ConfigFileUsed(), err)
		// Check if the error was specifically "file not found"
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			common.Logger("info", "Specific config file not found. Falling back to search for '.env' file.")
		} else {
			// A different error occurred (permissions, format, etc.)
			common.Logger("warning", "Error occurred while reading specific config file '%s'.: %v\n", viper.ConfigFileUsed(), err)
			common.Logger("warning", "Check %v file permissions and format.", viper.ConfigFileUsed())
		}

		// Configure and attempt fallback search for ".env"
		common.Logger("debug", "Setting up fallback search for '.env' in paths: '.', '/app'")
		viper.SetConfigName(".env") // Target filename for fallback
		viper.SetConfigType("env")  // Expected format for fallback
		viper.AddConfigPath(".")    // Search current directory
		viper.AddConfigPath("/app") // Search /app directory

		// Attempt to read AGAIN, performing the search defined above
		if fallbackErr := viper.ReadInConfig(); fallbackErr == nil {
			// SUCCESS reading fallback .env file
			common.Logger("debug", "Using fallback config file: %v", viper.ConfigFileUsed())
		} else {
			// FAILURE reading fallback .env file
			if errors.As(fallbackErr, &configFileNotFoundError) {
				// This is expected if no .env file exists in the search paths
				common.Logger("info", "No '.env' config file found in search paths either. Using defaults and environment variables.")
			} else {
				// An error occurred reading the fallback .env file (permissions, format?)
				common.Logger("warning", "Error reading fallback '.env' file: %v\n", fallbackErr)
				common.Logger("warning", "Check %v file permissions and format.", viper.ConfigFileUsed())
			}
		}
	}

	// Read in environment variables that match Viper keys or have the CLI_ prefix
	// Read environment variables *now*. They might be overridden by config file.
	viper.AutomaticEnv()

	// Unmarshal the final configuration
	// Viper now contains the merged view: Defaults overridden by Env Vars overridden by (potentially) a loaded Config File.
	common.Logger("debug", "Unmarshaling final configuration into struct.")
	if err := viper.Unmarshal(&config.Properties); err != nil {
		common.Logger("fatal", "Error unmarshaling config: %s", err)
	}

	// Redefining variables
	config.Properties.DefaultGSABaseAccountName = "todo-gsa"
	config.Properties.DefaultGSAAccountName = config.Properties.DefaultGSABaseAccountName + "@" + config.Properties.DefaultGCPProject + ".iam.gserviceaccount.com"

	// Validate the populated struct
	common.Logger("debug", "Validating final configuration...")
	// Create a new validator instance
	validate := validator.New(validator.WithRequiredStructEnabled())
	// Register custom validators
	validate.RegisterValidation("noUnderscore", config.NoUnderscores)

	// Validate the Properties struct (pass by reference)
	if err := validate.Struct(&config.Properties); err != nil {
		// Check if the error is specifically validation errors
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Build a user-friendly error message
			errorMsg := "Configuration validation failed:\n"
			for _, fieldErr := range validationErrors {
				errorMsg += fmt.Sprintf("  - Field '%s': Failed on validation rule '%s'. Value: '%v'\n",
					fieldErr.StructNamespace(), // e.g., PropertiesStruct.DefaultEnvironment
					fieldErr.Tag(),             // e.g., "required", "oneof"
					fieldErr.Value(),           // The actual invalid value
				)
			}
			// Log as fatal error and exit
			common.Logger("fatal", "%s", errorMsg)
		} else {
			// Handle other potential errors during validation itself (less common)
			common.Logger("fatal", "An unexpected error occurred during configuration validation: %s", err)
		}
	}

	// Optional: Log the final loaded configuration for verification
	finalConfigBytes, _ := yaml.Marshal(config.Properties) // Or use json.MarshalIndent
	common.Logger("debug", "Final Configuration Loaded:\n%s\n", string(finalConfigBytes))

}

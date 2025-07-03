// Package common has common functions reusable
package common

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/go-playground/validator/v10"

	pkgerrors "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	zerolog_pkgerrors "github.com/rs/zerolog/pkgerrors"
)

// FindExecutable checks if a file exists at the given path and is executable.
func FindExecutable(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil // File doesn't exist, not an error for the search logic itself
		}
		// Other error (e.g., permission denied to stat)
		return false, fmt.Errorf("[ERROR] Have permissions to execute? Check path '%s': %w", path, err)
	}

	// Check if it's a regular file and has execute permissions (for user, group, or others)
	// Mode()&0111 checks the execute bits.
	if !info.IsDir() && info.Mode()&0111 != 0 {
		return true, nil // Found and executable
	}

	// Exists but is a directory or not executable
	return false, nil
}

// CreateValidationErrorMessage based on validation tags.
// If the tag is not recognized, it will return a generic message.
func CreateValidationErrorMessage(err error, data interface{}) string {
	var message string

	for _, err := range err.(validator.ValidationErrors) {
		if err.Tag() == "required" {
			message += fmt.Sprintf("%s is required. ", err.Field())
		} else if err.Tag() == "len" {
			message += fmt.Sprintf("%s must have %s characters. ", err.Field(), err.Param())
		} else if err.Tag() == "oneof" {
			message += fmt.Sprintf("%s must be one of [%s]. ", err.Field(), strings.Join(strings.Split(err.Param(), " "), ", "))
		} else if err.Tag() == "email" {
			message += fmt.Sprintf("%s must be a valid email. ", err.Field())
		} else if err.Tag() == "min" {
			message += fmt.Sprintf("%s must be greater than %s. ", err.Field(), err.Param())
		} else if err.Tag() == "max" {
			message += fmt.Sprintf("%s must be less than %s. ", err.Field(), err.Param())
		} else if err.Tag() == "uuid" {
			message += fmt.Sprintf("%s must be a valid UUID. ", err.Field())
		} else if err.Tag() == "number" {
			message += fmt.Sprintf("%s must be a number. ", err.Field())
		} else if err.Tag() == "boolean" {
			message += fmt.Sprintf("%s must be a boolean. ", err.Field())
		} else if err.Tag() == "string" {
			message += fmt.Sprintf("%s must be a string. ", err.Field())
		} else if err.Tag() == "nefield" {
			param := GetParamName(data, err.Param())
			message += fmt.Sprintf("%s must be different from %s. ", err.Field(), param)
		} else {
			message += fmt.Sprintf("%s is invalid. ", err.Field())
		}
	}

	return strings.TrimSpace(message)
}

// GetParamName returns the name of the parameter based on the struct tag.
// If the tag is not found, it returns the parameter name.
func GetParamName(data interface{}, param string) string {
	topStruct := reflect.TypeOf(data)

	if topStruct.Kind() == reflect.Ptr {
		topStruct = topStruct.Elem()
	}

	if field, ok := topStruct.FieldByName(param); ok {
		tag := field.Tag.Get("json")
		split := strings.SplitN(tag, ",", 2)
		if split[0] != "" {
			return split[0]
		}
	}
	return param
}

// Logger print message log accoding level, timestamp and trace.
// Support the message with same behavour of fmt.Sprintf.
//
// References:
//
// https://www.geeksforgeeks.org/fmt-sprintf-function-in-golang-with-examples/
// https://pkg.go.dev/fmt
// https://pkg.go.dev/fmt#Sprintf
// https://github.com/rs/zerolog
//
// Examples:
//
// common.Logger("fatal","[ERROR] Example message")
// common.Logger("warning","[WARNING] config.Debug %v", *config.Debug)
// common.Logger("DEBUG","[DEBUG] config.Debug %v", *config.Debug)
// common.Logger("info","[INFO] Hello world!")
//
// Output:
//
// 2025-04-22T19:29:04-03:00 ERROR <nil> error="[DEBUG] config.Debug %!s(bool=true) stack=[{"func":"Logger","line":"146","source":"common.go"},{"func":"Execute","line":"64","source":"root.go"},{"func":"main","line":"8","source":"main.go"},{"func":"main","line":"283","source":"proc.go"},{"func":"goexit","line":"1223","source":"asm_arm64.s"}]
// 2025-04-22T19:29:04-03:00 WARNING [WARNING] config.Debug true
// 2025-04-22T19:29:04-03:00 DEBUG [DEBUG] config.Debug true
// 2025-04-22T19:29:04-03:00 INFO [INFO] Hello world!
func Logger(level string, message string, args ...interface{}) {
	level = strings.ToLower(level)

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprint(i))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprint(i)
		},
		FormatTimestamp: func(i interface{}) string {
			if ts, ok := i.(string); ok {
				return ts
			}
			if t, ok := i.(time.Time); ok {
				return t.Format("2006-01-02 15:04:05")
			}
			return fmt.Sprint(i)
		},
	})

	// Set time some configurations of zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = zerolog_pkgerrors.MarshalStack

	// Default level is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if config.Debug != nil && *config.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Get the message and arguments from Sprintf
	formatted := fmt.Sprintf(message, args...)

	// Get stack trace with line and file where the error occurred
	if level == "error" || level == "fatal" || level == "panic" {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			errWithStack := pkgerrors.WithStack(fmt.Errorf("%s (%s:%d)", formatted, file, line))
			switch level {
			case "error":
				// This log type does not interrupt the program
				log.Error().Stack().Err(errWithStack).Msg(formatted)
			case "fatal":
				// This log type interrupt the program with error code 1
				log.Fatal().Stack().Err(errWithStack).Msg(formatted)
			case "panic":
				// This log type interrupt the program with error code 1
				log.Panic().Stack().Err(errWithStack).Msg(formatted)
			}
			return
		}
	}

	// Levels below error (with stack trace)
	switch level {
	case "debug":
		log.Debug().Msg(formatted)
	case "warn", "warning":
		log.Warn().Msg(formatted)
	default:
		log.Info().Msg(formatted)
	}
}

// StringToEnvVar transform strings to uppercase and substitue '-' by '_' if exists
func StringToEnvVar(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

// CheckVPNConnection attempts a basic check for VPN connectivity.
// This is a placeholder and might not be reliable for all VPN setups.
// It tries to HTTP request GET a host that should only be accessible via VPN.
func CheckVPNConnection(vpnCheckURL string) {
	// Parse the URL to validate the format
	parsedURL, err := url.Parse(vpnCheckURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		Logger("fatal", "Invalid URL format to check VPN connection. Expected format protocol://host:port")
		return
	}

	// Log the attempt to connect
	Logger("debug", "Attempting VPN connection check by requesting %s...", vpnCheckURL)

	// Create an HTTP client with a timeout
	client := http.Client{
		Timeout: config.VPNTimeout * time.Second,
	}

	// Send a GET request to the URL
	resp, err := client.Get(vpnCheckURL)
	if err != nil {
		Logger("fatal", "VPN connection check failed: Could not connect to %s. Ensure VPN is active.", vpnCheckURL)
	}
	// Ensure the response body is closed
	// client.Get() can be return resp == nil along with an err != nil:
	// The code calls defer resp.Body.Close() after ensuring that resp is not nil.
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	// Check if the status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		Logger("fatal", "VPN connection check failed: Received HTTP status %d from %s", resp.StatusCode, vpnCheckURL)
		return
	}

	Logger("debug", "VPN connection check successful to %s.", vpnCheckURL)
}

// CheckCommandsAvailable verifies if all specified command-line tools are installed
// and accessible in the system's PATH.
// It returns a list of missing commands and an error if any are not found.
// If all commands are found, it returns nil, nil.
func CheckCommandsAvailable(commands []string) {
	missingCommands := []string{}

	if len(commands) == 0 {
		Logger("debug", "No commands specified for availability check.")
	}

	Logger("debug", "Checking availability of required commands: %v", commands)

	for _, cmdName := range commands {
		if strings.TrimSpace(cmdName) == "" {
			Logger("warning", "Empty command name provided in the list, skipping.")
			continue
		}
		// exec.LookPath searches for an executable named file in the directories
		// named by the PATH environment variable.
		// If file contains a slash, it is tried directly and the PATH is not consulted.
		// The result may be an absolute path or a path relative to the current directory.
		_, findErr := exec.LookPath(cmdName)
		if findErr != nil {
			// Error typically means the command was not found in PATH.
			// It could also be a permission issue for directories in PATH, but "not found" is most common.
			Logger("warning", "Command '%s' not found in system PATH: %v", cmdName, findErr)
			missingCommands = append(missingCommands, cmdName)
		} else {
			Logger("debug", "Command '%s' found in system PATH.", cmdName)
		}
	}

	if len(missingCommands) > 0 {
		Logger("fatal", "the following required command(s) were not found in your system PATH: %s. Please install them and ensure they are accessible.", strings.Join(missingCommands, ", "))
	}

	Logger("debug", "All specified commands (%v) are available in system PATH.", commands)
}

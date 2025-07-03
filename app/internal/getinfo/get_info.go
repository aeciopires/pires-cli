// Package getinfo provides getinfo and version messages
package getinfo

import (
	"fmt"
	"os"
	"runtime"

	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
)

// PrintLongVersion prints the application version
func PrintLongVersion() {
	fmt.Printf("Version: %s\n", config.CLIVersion)
}

// PrintShortVersion prints only number of the application version
func PrintShortVersion() {
	fmt.Printf("%s\n", config.CLIVersion)
}

// ShowOperatingSystem prints the operating system
func ShowOperatingSystem() {
	osName := runtime.GOOS
	switch osName {
	case "darwin", "linux":
		fmt.Println("Operating system:", osName)
	default:
		fmt.Printf("%s is not supported.", osName)
		os.Exit(1)
	}
}

// CheckOperatingSystem check if operating system is supported
func CheckOperatingSystem() {
	osName := runtime.GOOS
	switch osName {
	case "darwin", "linux":
		common.Logger("debug", "Operating system: %s", osName)
	default:
		common.Logger("fatal", "%s is not supported.", osName)
	}
}

// ShowSystemArch prints the system arch
func ShowSystemArch() {
	arch := runtime.GOARCH
	fmt.Println("System Arch:", arch)
}

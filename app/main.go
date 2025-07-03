package main

import (
	"github.com/aeciopires/pires-cli/cmd"
	"github.com/aeciopires/pires-cli/internal/config"
	"github.com/aeciopires/pires-cli/internal/getinfo"
	"github.com/aeciopires/pires-cli/pkg/pireslib/common"
	"github.com/aeciopires/pires-cli/pkg/pireslib/fileeditor"
)

func main() {
	getinfo.CheckOperatingSystem()
	fileeditor.GetYqPath()
	common.CheckCommandsAvailable(config.CommandsToCheck)
	// ToDO: Here we have a bug, because the flags values is not loaded yet.
	// This code block should be moved to the root command Run function and replicated to subcommands.
	if config.VPNCheckConnection != nil && *config.VPNCheckConnection {
		common.CheckVPNConnection(config.Properties.DefaultVPNAddressTarget)
	}
	cmd.Execute()
}

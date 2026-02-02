// tftest is a CLI tool to help with testing Terraform modules and
// configurations.
//
// Usage:
//
//	tftest [command] [flags]
//
// Available Commands:
//
//	tftest completion  Generate shell completion scripts
//	tftest help        Help about any command
//	tftest test        Run tests for Terraform modules
//
// Global Flags:
//
//	--terraform-version string   Specify the Terraform version to use (default to system terraform)
package main

import (
	"os"

	"github.com/kapetndev/tftest/internal/cmd"
	cliruntime "github.com/tomasbasham/cli-runtime"
)

func main() {
	command := cmd.NewRootCommand()
	if code := cliruntime.Run(command); code != 0 {
		os.Exit(code)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/intelchain-itc/intelchain/internal/cli"
	"github.com/spf13/cobra"
)

const (
	versionFormat = "Intelchain (C) 2023. %v, version %v-%v (%v %v)"
)

// Version string variables
var (
	version string
	builtBy string
	builtAt string
	commit  string
)

var versionFlag = cli.BoolFlag{
	Name:      "version",
	Shorthand: "V",
	Usage:     "display version info",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version of the intelchain binary",
	Long:  "print version of the intelchain binary",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
		os.Exit(0)
	},
}

func getIntelchainVersion() string {
	return fmt.Sprintf(versionFormat, "intelchain", version, commit, builtBy, builtAt)
}

func printVersion() {
	fmt.Println(getIntelchainVersion())
}

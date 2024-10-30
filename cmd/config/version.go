package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zennittians/intelchain/internal/cli"
)

const (
	versionFormat = "Intelchain (C) 2023. %v, version %v-%v (%v %v)"
)

var VersionMetaData []interface{}

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
		PrintVersion()
		os.Exit(0)
	},
}

func VersionFlag() cli.BoolFlag {
	return versionFlag
}

func GetIntelchainVersion() string {
	return fmt.Sprintf(versionFormat, VersionMetaData[:5]...) // "Intelchain", version, commit, builtBy, builtAt
}

func PrintVersion() {
	fmt.Println(GetIntelchainVersion())
}

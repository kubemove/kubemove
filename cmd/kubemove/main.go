package main

import (
	"os"

	"github.com/spf13/pflag"

	"github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"
)

func main() {
	flags := pflag.NewFlagSet("kubemove", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := plugin.NewCmdKubeMove()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"os"

	"github.com/spf13/pflag"

	plugin "github.com/kubemove/kubemove/pkg/plugin/v1alpha1"
)

func main() {
	flags := pflag.NewFlagSet("kubemove", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := plugin.NewCmdKubeMove()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

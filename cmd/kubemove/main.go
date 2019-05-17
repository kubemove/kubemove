package main

import (
	"os"

	"github.com/spf13/pflag"

	"github.com/kubemove/kubemove/pkg/plugin/v1alpha1"
)

func main() {
	flags := pflag.NewFlagSet("kubemove", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := v1alpha1.NewCmdKubeMove()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

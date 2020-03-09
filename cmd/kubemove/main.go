package main

import (
	"github.com/spf13/pflag"
)

func main() {
	flags := pflag.NewFlagSet("kubemove", pflag.ExitOnError)
	pflag.CommandLine = flags

	/*
	// TODO
	root := plugin.NewCmdKubeMove()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
	 */

}

package v1alpha1

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// NewCmdVersion returns kubemove version command
func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints version and other details relevant to kubemove plugin",
		Long: `Prints version and other details relevant to kubemove plugin

Usage:
kubemove version
	`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", "v0.0.1-unreleased")
			fmt.Printf("GO Version: %s\n", runtime.Version())
			fmt.Printf("GO ARCH: %s\n", runtime.GOARCH)
			fmt.Printf("GO OS: %s\n", runtime.GOOS)
		},
	}

	return cmd
}

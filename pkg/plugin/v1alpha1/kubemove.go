package v1alpha1

import "github.com/spf13/cobra"

// NewCmdKubeMove returns kubemove root command
func NewCmdKubeMove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubemove",
		Short: "kubemove is used for interacting with kubemove operators",
		Long:  "kubemove is used for interacting with kubemove operators",
	}
	cmd.AddCommand(NewCmdVersion())
	return cmd
}

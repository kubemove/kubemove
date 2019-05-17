package v1alpha1

import "github.com/spf13/cobra"

// NewCmdKubeMove returns kubemove root command
func NewCmdKubeMove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl kubemove",
		Short: "kubemove cli-plugin is used for interacting with kmov operators",
		Long:  "kubemove cli-plugin is used for interacting with kmov operators",
	}
	cmd.AddCommand(NewCmdVersion())
	return cmd
}

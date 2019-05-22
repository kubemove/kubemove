package v1alpha1

import (
	"fmt"

	"github.com/spf13/cobra"
)

const dummyData = `
Name        Type
----        ---- 
CR1         KubeMovePair
CR2         KubeMoveSync	
`

// NewCmdList returns kubemove list command to list the kubemove CRs available in the cluster
func NewCmdList() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Prints List of available kubemove CRDs",
		Long: `Prints List of available kubemove CRDs
		Usage:
		kubemove list`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s", dummyData)
		},
	}
	return listCmd
}

package controller

import (
	"github.com/kubemove/kubemove/pkg/controller/moveswitch"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, moveswitch.Add)
}

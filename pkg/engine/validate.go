package engine

import (
	"strings"

	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	"github.com/pkg/errors"
)

func ValidateEngine(mov *v1alpha1.MoveEngine) error {
	var ret []string

	if len(mov.Spec.MovePair) == 0 {
		ret = append(ret, "MovePair not given")
	}

	if len(ret) == 0 {
		return nil
	}

	return errors.Errorf(strings.Join(ret, ""))
}

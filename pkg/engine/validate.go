package engine

import (
	"fmt"
	"strings"

	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
)

func ValidateEngine(mov *v1alpha1.MoveEngine) error {
	var validationError []string

	if len(mov.Spec.MovePair) == 0 {
		validationError = append(validationError, "MovePair not given")
	}

	if len(mov.Spec.SyncPeriod) == 0 {
		return nil
	}

	_, err := cron.ParseStandard(mov.Spec.SyncPeriod)
	if err != nil {
		validationError = append(validationError, fmt.Sprintf("SyncPeriod is invalid.. %v", err))
	}

	if len(validationError) == 0 {
		return nil
	}
	return errors.Errorf(strings.Join(validationError, ""))
}

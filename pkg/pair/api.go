package mpair

import (
	"context"

	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Get(key client.ObjectKey, c client.Client) (*v1alpha1.MovePair, error) {
	m := &v1alpha1.MovePair{}
	err := c.Get(context.TODO(), key, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

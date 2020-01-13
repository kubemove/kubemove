package main

import (
	"github.com/go-logr/logr"
	client "github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type dummyDDM struct {
	log logr.Logger
	p   client.Plugin
}

var _ client.Plugin = (*dummyDDM)(nil)

func main() {
	var log = logf.Log.WithName("Plugin")
	//	err := client.Register("kubemove.io/dummy",
	err := client.Register("dummy",
		&dummyDDM{
			log: log,
		})
	if err != nil {
		log.Error(err, "Failed to create a plugin")
		return
	}

}

func (d *dummyDDM) Init(param map[string]string) error {
	d.log.Info("Initializing plugin with config %v", param)
	return nil
}

func (d *dummyDDM) Sync(engine string, isOwner bool, param map[string]string, vol []*client.Volume) (string, error) {
	d.log.Info("Syncing engine %v", engine)
	d.log.Info("Synced vol %v", vol)
	return "dummy_id", nil
}

func (d *dummyDDM) Status(id string) (int32, error) {
	d.log.Info("got request for id %v", id)
	return 0, nil
}

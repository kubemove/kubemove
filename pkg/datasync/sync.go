package datasync

import (
	"github.com/pkg/errors"
)

func (d *DataSync) Sync() error {
	plugin := d.ds.Spec.PluginProvider

	_, err := d.ddm.SyncData(plugin, *d.ds)
	if err != nil {
		return errors.Wrapf(err, "Failed to sync data")
	}
	return nil
}

func (d *DataSync) SyncStatus() (int32, error) {
	plugin := d.ds.Spec.PluginProvider
	return d.ddm.SyncStatus(plugin, *d.ds)
}

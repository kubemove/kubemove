package datasync

import (
	"github.com/pkg/errors"
)

func (d *DataSync) Sync() error {
	plugin := d.ds.Spec.PluginProvider
	if d.ds.Status.Status == "Completed" {
		return nil
	}

	// Sync
	// TODO id ignored
	_, err := d.ddm.SyncData(plugin, *d.ds)
	if err != nil {
		return errors.Wrapf(err, "Failed to sync data")
	}

	// TODO Update CR

	return nil
}

package main

type SyncPhase string
type StatusType string

const (
	SyncPhaseBackup  SyncPhase = "backup"
	SyncPhaseRestore SyncPhase = "restore"

	BackupStatus  StatusType = "backup-status"
	RestoreStatus StatusType = "restore-status"
)

// triggerSnapshot hits the Snapshot API of Elasticsearch to trigger a snapshot
func (d *eckDDM)triggerSnapshot(params map[string]string) (string, error) {
	// TODO

	return "<snapshot name?>", nil
}

// triggerRestore hits the Recovery API of Elasticsearch to trigger a restore process
func (d *eckDDM)triggerRestore(params map[string]string) (string, error) {
	// TODO

	return "what should we return?", nil
}

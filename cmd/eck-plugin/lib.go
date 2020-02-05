package main

type SyncPhase string
type StatusType string

const (
	SyncPhaseBackup  SyncPhase = "backup"
	SyncPhaseRestore SyncPhase = "restore"

	BackupStatus  StatusType = "backup-status"
	RestoreStatus StatusType = "restore-status"
)

// registerSnapshotRepository hits Snapshot API of Elasticsearch to register a repository
func registerSnapshotRepository(params map[string]string) error {
	// TODO
	return nil
}

// triggerSnapshot hits the Snapshot API of Elasticsearch to trigger a snapshot
func triggerSnapshot(params map[string]string) (string, error) {
	// TODO

	return "<snapshot name?>", nil
}

// triggerRestore hits the Recovery API of Elasticsearch to trigger a restore process
func triggerRestore(params map[string]string) (string, error) {
	// TODO

	return "what should we return?", nil
}

// retrieveSnapshotState hits the Snapshot API of Elasticsearch to retrieve the backup status of a Snapshot
func retrieveSnapshotState(snapshotName string) error {
	// TODO

	return nil
}

// retrieveRestoreState hits the Recovery API of Elasticsearch to retrieve the restore status of a ES.
func retrieveRestoreState(snapshotName string) error {
	// TODO

	return nil
}

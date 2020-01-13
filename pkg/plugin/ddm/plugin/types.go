package plugin

type Volume struct {
	VolumeName        string
	VolumeClaim       string
	RemoteVolumeClaim string
	LocalNS           string
	RemoteNS          string
}

type Plugin interface {
	Status(string) (int32, error)
	Sync(string, bool, map[string]string, []*Volume) (string, error)
	Init(map[string]string) error
}

const (
	Completed = iota
	InProgress
	Invalid
	Canceled
	Errored
)

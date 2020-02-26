package plugin

type Volume struct {
	VolumeName        string
	VolumeClaim       string
	RemoteVolumeClaim string
	LocalNS           string
	RemoteNS          string
}

type Plugin interface {
	Init(map[string]string) error
	Sync(map[string]string, []*Volume) (string, error)
	Status(map[string]string) (int32, error)
}

const (
	Completed = iota
	InProgress
	Invalid
	Canceled
	Errored
	Failed
	Unknown
)

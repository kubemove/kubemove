package plugin

type Volume struct {
	VolumeName        string
	VolumeClaim       string
	RemoteVolumeClaim string
	LocalNS           string
	RemoteNS          string
}

type Plugin interface {
	Init(map[string]interface{}) error
	Sync(map[string]interface{}, []*Volume) (string, error)
	Status(map[string]interface{}) (int32, error)
}

const (
	Completed = iota
	InProgress
	Invalid
	Canceled
	Errored
)
